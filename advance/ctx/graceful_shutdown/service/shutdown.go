package service

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Option func(*App)

type ShutdownCallback func(ctx context.Context)

func WithShutdownCallbacks(cbs ...ShutdownCallback) Option {
	return func(app *App) {
		app.cbs = cbs
	}
}

type App struct {
	servers []*Server

	shutdownTimeout time.Duration

	waitTime  time.Duration
	cbTimeout time.Duration

	cbs []ShutdownCallback
}

func NewApp(servers []*Servers, opts ...Option) *App {
	res := &App{
		servers:         servers,
		waitTime:        10 * time.Second,
		cbTimeout:       3 * time.Second,
		shutdownTimeout: 30 * time.Second,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

type Server struct {
	srv  *http.Server
	name string
	mux  *serverMux
}

type serverMux struct {
	reject bool
	*http.ServeMux
}

func (s *serverMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 只是在考虑到 CPU 高速缓存的时候，会存在短时间的不一致性
	if s.reject {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("服务已关闭"))
		return
	}
	s.ServeMux.ServeHTTP(w, r)
}

func (app *App) StartAndServe() {
	for _, s := range app.servers {
		srv := s
		go func() {
			if err := srv.Start(); err != nil {
				if err == http.ErrServerClosed {
					log.Printf("服务器%s已关闭", srv.name)
				} else {
					log.Printf("服务器%s异常退出", srv.name)
				}
			}
		}()
	}

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, signals...)
	go func() {
		select {
		case <-ch:
			log.Printf("强制退出")
			os.Exit(1)
		case <-time.After(app.shutdownTimeout):
			log.Printf("超时强制退出")
			os.Exit(1)
		}
	}()
	app.shutdown()
}

func (app *App) shutdown() {
	log.Println("开始关闭应用，停止接收新请求")
	for _, s := range app.servers {
		s.rejectReq()
	}
	log.Println("等待正在执行请求完毕")
	time.Sleep(app.waitTime)
	log.Println("开始关闭服务器")

	var wg sync.WaitGroup
	wg.Add(len(app.servers))
	for _, srv := range app.servers {
		srvCp := srv
		go func() {
			if err := srvCp.stop(); err != nil {
				log.Println("关闭服务失败", srvCp.name)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	log.Println("开始执行自定义回调")
	wg.Add(len(app.cbs))
	for _, cb := range app.cbs {
		c := cb
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), app.cbTimeout)
			c(ctx)
			cancel()
			wg.Done()
		}()
	}
	wg.Wait()
	log.Println("开始释放资源")
	app.close()
}

func (app *App) close() {
	// 在这里释放掉一些可能的资源
	time.Sleep(time.Second)
	log.Println("应用关闭")
}

func NewServer(name string, addr string) *Server {
	mux := &serverMux{ServeMux: http.NewServeMux()}
	return &Server{
		name: name,
		mux:  mux,
		srv: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}
}

func (s *Server) Handle(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
}

func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

func (s *Server) rejectReq() {
	s.mux.reject = true
}

func (s *Server) stop() error {
	log.Printf("服务器%s关闭中", s.name)
	return s.srv.Shutdown(context.Background())
}

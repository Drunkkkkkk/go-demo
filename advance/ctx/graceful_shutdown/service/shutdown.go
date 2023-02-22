package service

import (
	"context"
	"log"
	"net/http"
	"time"
)

type Option func(*App)

type ShutdownCallback func(ctx context.Context)

func WithShutdownCallbacks(cbs ...ShutdownCallback) Option {
	panic("implement me")
}

type App struct {
	servers []*http.Server

	shutdownTimeout time.Duration

	waitTime  time.Duration
	cbTimeout time.Duration

	cbs []ShutdownCallback
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
			if err := srv.Star
		}()
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

func (s *Server) stop(ctx context.Context) error {
	log.Printf("服务器%s关闭中", s.name)
	return s.srv.Shutdown(ctx)
}


package main

import (
	"context"
	"go-demo/advance/ctx/graceful_shutdown/service"
	"log"
	"net/http"
	"time"
)

func main() {
	s1 := service.NewServer("business", "127.0.0.1:8080")
	s1.Handle("/", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte("hello"))
	}))
	s2 := service.NewServer("admin", "localhost:8081")
	app := service.NewApp([]*service.Server{s1, s2}, service.WithShutdownCallbacks(StoreCacheToDBCallBack))
	app.StartAndServe()
}

func StoreCacheToDBCallBack(ctx context.Context) {
	done := make(chan struct{}, 1)
	go func() {
		// 你的业务逻辑，比如说这里我们模拟的是将本地缓存刷新到数据库里面
		// 这里我们简单的睡一段时间来模拟
		log.Printf("刷新缓存中……")
		time.Sleep(1 * time.Second)
	}()
	select {
	case <-ctx.Done():
		log.Println("刷新缓存超时")
	case <-done:
		log.Println("缓存被刷新到了 DB")
	}
}

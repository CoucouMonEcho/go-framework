package main

import (
	"code-practise/shutdown/service"
	"context"
	"log"
	"net/http"
)

func main() {
	s1 := service.NewServer("business", "localhost:8080")
	s1.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte("hello world"))
	}))
	s2 := service.NewServer("business", "localhost:8081")
	app := service.NewApp([]*service.Server{s1, s2}, service.WithShutdownCallbacks(StoreCacheToDBCallback))
	app.Run()
}

func StoreCacheToDBCallback(ctx context.Context) {
	done := make(chan struct{}, 1)
	go func() {
		log.Println("store cache to db")
	}()
	select {
	case <-ctx.Done():
		log.Println("store cache timeout")
	case <-done:
		log.Println("store cache to db done")
	}
}

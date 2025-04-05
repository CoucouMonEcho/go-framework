package web

import (
	"fmt"
	"net/http"
	"testing"
)

func TestHTTPServer(t *testing.T) {
	h := NewHTTPServer()
	h.middlewares = []Middleware{
		func(next HandlerFunc) HandlerFunc {
			return func(ctx *Context) {
				fmt.Println("hello 1 before")
				next(ctx)
				fmt.Println("hello 1 after")
			}
		},
		func(next HandlerFunc) HandlerFunc {
			return func(ctx *Context) {
				fmt.Println("hello 2 before")
				next(ctx)
				fmt.Println("hello 2 after")
			}
		},
		func(next HandlerFunc) HandlerFunc {
			return func(ctx *Context) {
				fmt.Println("break 3")
			}
		},
		func(next HandlerFunc) HandlerFunc {
			return func(ctx *Context) {
				fmt.Println("can not reach 4")
			}
		},
	}
	h.ServeHTTP(nil, &http.Request{})
	//h.Get("/user/login", func(ctx *Context) {
	//	ctx.Resp.Write([]byte("<h1>Hello World</h1>"))
	//})
	//h.Start(":8081")
}

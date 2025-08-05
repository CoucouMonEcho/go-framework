package accesslog

import (
	"fmt"
	"github.com/CoucouMonEcho/go-framework/web"
	"log"
	"net/http/httptest"
	"testing"
)

func TestMiddleBuilder_Build(t *testing.T) {
	builder := NewMiddlewareBuilder()
	builder.LogFunc(func(log string) {
		fmt.Println(log)
	})
	h := web.NewHTTPServer(web.ServerWithMiddlewares(builder.Build()), web.ServerWithLogger(func(msg string, args ...any) {
		log.Println(msg, args)
	}))
	h.Get("/user/*", func(ctx *web.Context) {
		fmt.Println("A")
	})
	request := httptest.NewRequest("GET", "/user/login", nil)
	h.ServeHTTP(nil, request)
	//h.Get("/user/login", func(ctx *web.Context) {
	//	ctx.Resp.Write([]byte("<h1>Hello World</h1>"))
	//})
	//h.Start(":8081")
}

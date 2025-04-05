package accesslog

import (
	"code-practise/web"
	"fmt"
	"net/http/httptest"
	"testing"
)

func TestMiddleBuilder_Build(t *testing.T) {
	builder := MiddlewareBuilder{}
	builder.LogFunc(func(log string) {
		fmt.Println(log)
	})
	h := web.NewHTTPServer(web.ServerWithMiddlewares(builder.Build()))
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

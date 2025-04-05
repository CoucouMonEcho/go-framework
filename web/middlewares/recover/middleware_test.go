package recover

import (
	"code-practise/web"
	"fmt"
	"net/http"
	"testing"
)

func TestMiddlewareBuilder_Build(t *testing.T) {
	builder := MiddlewareBuilder{
		Code: http.StatusInternalServerError,
		Data: []byte("panic"),
		Log: func(ctx *web.Context) {
			fmt.Printf("panic path: %s", ctx.Req.URL.String())
		},
	}

	server := web.NewHTTPServer(web.ServerWithMiddlewares(builder.Build()))
	server.Get("/user", func(ctx *web.Context) {
		panic("panic")
	})
	server.Start(":8081")
}

package recover

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"go-framework/web"
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
	err := server.Start(":8081")
	require.NoError(t, err)
}

//go:build e2e

package prometheus

import (
	"github.com/CoucouMonEcho/go-framework/web"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

func TestMiddlewareBuilder_Build(t *testing.T) {
	builder := MiddlewareBuilder{
		Namespace: "test",
		Subsystem: "web",
		Name:      "prometheus",
	}
	server := web.NewHTTPServer(web.ServerWithMiddlewares(builder.Build()))
	type User struct {
		Name string
	}
	server.Get("/metrics", func(ctx *web.Context) {
		val := rand.Intn(1000) + 1
		time.Sleep(time.Duration(val) * time.Millisecond)
		ctx.RespJSON(http.StatusOK, User{Name: "Tom"})
	})

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":9090", nil)
	}()

	server.Start(":8081")
}

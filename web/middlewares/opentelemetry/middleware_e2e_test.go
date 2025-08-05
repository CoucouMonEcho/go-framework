//go:build e2e

package opentelemetry

import (
	"github.com/CoucouMonEcho/go-framework/web"
	"go.opentelemetry.io/otel"
	"testing"
	"time"
)

func TestMiddlewareBuilder_Build(t *testing.T) {
	tracer := otel.Tracer(instrumentationName)
	builder := MiddlewareBuilder{
		Tracer: tracer,
	}
	server := web.NewHTTPServer(web.ServerWithMiddlewares(builder.Build()))
	server.Get("/user", func(ctx *web.Context) {
		_, span := tracer.Start(ctx.Req.Context(), "first_layer")
		defer span.End()

		_, second := tracer.Start(ctx.Req.Context(), "second_layer")
		time.Sleep(time.Second)
		_, third1 := tracer.Start(ctx.Req.Context(), "third_layer_1")
		time.Sleep(100 * time.Millisecond)
		third1.End()
		_, third2 := tracer.Start(ctx.Req.Context(), "third_layer_2")
		time.Sleep(300 * time.Millisecond)
		third2.End()
		second.End()
	})
	server.Start(":8081")
}

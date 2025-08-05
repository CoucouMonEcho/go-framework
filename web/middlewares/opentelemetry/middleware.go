package opentelemetry

import (
	"go-framework/web"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "go-framework/web/middlewares/opentelemetry"

type MiddlewareBuilder struct {
	Tracer trace.Tracer
}

func (m MiddlewareBuilder) Build() web.Middleware {
	if m.Tracer == nil {
		m.Tracer = otel.GetTracerProvider().Tracer(instrumentationName)
	}

	return func(next web.Handler) web.Handler {
		return func(ctx *web.Context) {
			reqCtx := ctx.Req.Context()
			// get client trace
			reqCtx = otel.GetTextMapPropagator().Extract(reqCtx, propagation.HeaderCarrier(ctx.Req.Header))
			// this req ctx need reset
			reqCtx, span := m.Tracer.Start(reqCtx, "unknown")
			defer func() {
				span.SetName(ctx.MatchedRoute)
				span.SetAttributes(
					attribute.Int("http.status", ctx.RespCode),
				)
				span.End()
			}()
			span.SetAttributes(
				attribute.String("http.method", ctx.Req.Method),
				attribute.String("http.url", ctx.Req.URL.String()),
				attribute.String("http.scheme", ctx.Req.URL.Scheme),
				attribute.String("http.hostname", ctx.Req.Host),
				attribute.String("http.proto", ctx.Req.Proto),
			)
			ctx.Req = ctx.Req.WithContext(reqCtx)
			next(ctx)
		}
	}
}

package opentelemetry

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net"
)

const instrumentationName = "go-framework/micro/observability/opentelemetry"

type ServerOpenTelemetryBuilder struct {
	Tracer trace.Tracer
	Port   int
}

func (s *ServerOpenTelemetryBuilder) Build() grpc.UnaryServerInterceptor {
	if s.Tracer == nil {
		s.Tracer = otel.GetTracerProvider().Tracer(instrumentationName)
	}
	address := getOutboundIP()
	if s.Port != 0 {
		address = fmt.Sprintf("%s:%d", address, s.Port)
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		reqCtx := s.extract(ctx)
		reqCtx, span := s.Tracer.Start(reqCtx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))

		defer func() {
			if err != nil {
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
			}
			span.End()
		}()
		span.SetAttributes(
			attribute.String("address", address),
		)

		resp, err = handler(ctx, req)
		return
	}
}

func (s *ServerOpenTelemetryBuilder) extract(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	return otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(md))
}

func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer func() { _ = conn.Close() }()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

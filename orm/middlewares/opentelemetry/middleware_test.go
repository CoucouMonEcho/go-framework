package opentelemetry

import (
	"go.opentelemetry.io/otel/trace"
	"testing"
)

func TestNewMiddlewareBuilder(t *testing.T) {
	var t1 trace.Tracer
	NewMiddlewareBuilder(t1)
}

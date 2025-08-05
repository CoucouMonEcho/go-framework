package opentelemetry

import (
	"context"
	"fmt"
	"go-framework/orm"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "go-framework/orm/middlewares/opentelemetry"

type MiddlewareBuilder struct {
	Tracer trace.Tracer
}

func NewMiddlewareBuilder(tracer trace.Tracer) *MiddlewareBuilder {
	return &MiddlewareBuilder{
		Tracer: tracer,
	}
}

func (m MiddlewareBuilder) Build() orm.Middleware {
	if m.Tracer == nil {
		m.Tracer = otel.GetTracerProvider().Tracer(instrumentationName)
	}
	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {

			// span name: SELECT-test_model INSERT-test_model
			tableName := qc.Model.TableName
			spanCtx, span := m.Tracer.Start(ctx, fmt.Sprintf("%s-%s", qc.Type, tableName))
			defer span.End()
			query, _ := qc.Builder.Build()
			if query != nil {
				span.SetAttributes(
					attribute.String("sql", query.SQL),
				)
			}
			span.SetAttributes(
				attribute.String("component", "orm"),
				attribute.String("table", tableName),
			)

			res := next(spanCtx, qc)

			if res.Err != nil {
				span.RecordError(res.Err)
			}
			return res
		}
	}
}

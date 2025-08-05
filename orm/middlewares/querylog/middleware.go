package querylog

import (
	"context"
	"github.com/CoucouMonEcho/go-framework/orm"
	"log"
)

type MiddlewareBuilder struct {
	logFunc func(sql string, args []any)
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{
		logFunc: func(sql string, args []any) {
			log.Printf("sql: %s, args: %v \n", sql, args)
		},
	}
}

func (m *MiddlewareBuilder) LogFunc(logFunc func(sql string, args []any)) *MiddlewareBuilder {
	m.logFunc = logFunc
	return m
}

func (m *MiddlewareBuilder) Build() orm.Middleware {
	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			query, err := qc.Builder.Build()
			if err != nil {
				return &orm.QueryResult{
					Err: err,
				}
			}
			m.logFunc(query.SQL, query.Args)

			res := next(ctx, qc)
			return res
		}
	}
}

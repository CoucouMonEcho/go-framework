package querylog

import (
	"context"
	"github.com/CoucouMonEcho/go-framework/orm"
	"log"
	"time"
)

type MiddlewareBuilder struct {
	threshold time.Duration
	logFunc   func(sql string, args []any)
}

func NewMiddlewareBuilder(threshold time.Duration) *MiddlewareBuilder {
	return &MiddlewareBuilder{
		// 100ms
		threshold: threshold,
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
			start := time.Now()
			defer func() {
				if time.Since(start) < m.threshold {
					return
				}
				query, err := qc.Builder.Build()
				if err == nil {
					m.logFunc(query.SQL, query.Args)
				}
			}()
			res := next(ctx, qc)
			return res
		}
	}
}

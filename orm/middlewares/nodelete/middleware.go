package nodelete

import (
	"code-practise/orm"
	"context"
	"errors"
)

type MiddlewareBuilder struct {
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{}
}

func (m MiddlewareBuilder) Build() orm.Middleware {
	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			if qc.Type == "DELETE" {
				return &orm.QueryResult{Err: errors.New("DELETE statement is disabled")}
			}
			res := next(ctx, qc)
			return res
		}
	}
}

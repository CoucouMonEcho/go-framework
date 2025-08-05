package nodelete

import (
	"context"
	"errors"
	"go-framework/orm"
	"strings"
)

type MiddlewareBuilder struct {
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{}
}

func (m MiddlewareBuilder) Build() orm.Middleware {
	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			// UPDATE DELETE [SELECT] statement WHERE must be used
			if qc.Type == "UPDATE" || qc.Type == "DELETE" {
				query, err := qc.Builder.Build()
				if err != nil {
					return &orm.QueryResult{Err: err}
				}
				if !strings.Contains(query.SQL, "WHERE") {
					return &orm.QueryResult{Err: errors.New("unsafe dml statement")}
				}
			}
			res := next(ctx, qc)
			return res
		}
	}
}

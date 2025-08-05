package orm

import (
	"context"
	"github.com/CoucouMonEcho/go-framework/orm/model"
)

type QueryContext struct {
	Type    string
	Builder QueryBuilder

	Model *model.Model
}

type QueryResult struct {
	// Result SELECT: *T []*T others: sql.Result
	Result any
	Err    error
}

type Handler func(ctx context.Context, qc *QueryContext) *QueryResult

type Middleware func(next Handler) Handler

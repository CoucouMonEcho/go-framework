package orm

import "context"

type QueryContext struct {
	Type    string
	Builder QueryBuilder
}

type QueryResult struct {
	// Result SELECT: *T []*T others: Result
	Result any
	Err    error
}

type Handler func(ctx context.Context, qc *QueryContext) *QueryResult

type Middleware func(next Handler) Handler

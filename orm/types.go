package orm

import (
	"context"
)

var (
	_ Querier[any] = &Selector[any]{}
	_ Querier[any] = &RawQuerier[any]{}
)

// Querier used for SELECT
type Querier[T any] interface {
	// Get Using a pointer result may cause memory escape issues,
	// and without using pointers may cause reflection issues.
	Get(ctx context.Context) (*T, error)
	GetMulti(ctx context.Context) ([]*T, error)
}

var (
	_ Executor = &Inserter[any]{}
	_ Executor = &Deleter[any]{}
	_ Executor = &RawQuerier[any]{}
)

// Executor used for INSERT, DELETE and UPDATE
type Executor interface {
	Exec(ctx context.Context) Result
}

var (
	_ QueryBuilder = &Selector[any]{}
	_ QueryBuilder = &Inserter[any]{}
	_ QueryBuilder = &Deleter[any]{}
	_ QueryBuilder = &RawQuerier[any]{}
)

type QueryBuilder interface {
	// Build Result Query using pointers is easy for AOP.
	Build() (*Query, error)
}

type Query struct {
	SQL  string
	Args []any
}

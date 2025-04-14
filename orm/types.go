package orm

import (
	"context"
	"database/sql"
)

// Querier used for SELECT
type Querier[T any] interface {
	// Get Using a pointer result may cause memory escape issues,
	// and without using pointers may cause reflection issues.
	Get(ctx context.Context) (*T, error)
	GetMulti(ctx context.Context) ([]*T, error)
}

// Executor used for INSERT, DELETE and UPDATE
type Executor interface {
	Exec(ctx context.Context) (sql.Result, error)
}

type QueryBuilder interface {
	// Build Result Query using pointers is easy for AOP.
	Build() (*Query, error)
}

type Query struct {
	SQL  string
	Args []any
}

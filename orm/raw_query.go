package orm

import (
	"context"
	"database/sql"
)

var _ Querier[any] = &RawQuerier[any]{}

var _ QueryBuilder = &RawQuerier[any]{}

type RawQuerier[T any] struct {
	core
	sess Session
	sql  string
	args []any
}

func RawQuery[T any](sess Session, query string, args ...any) *RawQuerier[T] {
	c := sess.getCore()
	return &RawQuerier[T]{
		core: c,
		sess: sess,
		sql:  query,
		args: args,
	}
}

func (r *RawQuerier[T]) Build() (*Query, error) {
	return &Query{SQL: r.sql, Args: r.args}, nil
}

func (r *RawQuerier[T]) Get(ctx context.Context) (*T, error) {
	// initialize model
	var err error
	if r.model, err = r.r.Get(new(T)); err != nil {
		return nil, err
	}
	res := get[T](ctx, r.sess, r.core, &QueryContext{
		Type:    "RAW",
		Builder: r,
		Model:   r.model,
	})
	if res.Result != nil {
		return res.Result.(*T), res.Err
	}
	return nil, res.Err
}

func (r *RawQuerier[T]) GetMulti(ctx context.Context) ([]*T, error) {
	// initialize model
	var err error
	if r.model, err = r.r.Get(new(T)); err != nil {
		return nil, err
	}
	res := getMulti[T](ctx, r.sess, r.core, &QueryContext{
		Type:    "RAW",
		Builder: r,
		Model:   r.model,
	})
	if res.Result != nil {
		return res.Result.([]*T), res.Err
	}
	return nil, res.Err
}

func (r *RawQuerier[T]) Exec(ctx context.Context) Result {
	// initialize model
	var err error
	if r.model, err = r.r.Get(new(T)); err != nil {
		return Result{err: err}
	}
	res := exec(ctx, r.sess, r.core, &QueryContext{
		Type:    "RAW",
		Builder: r,
		Model:   r.model,
	})
	if res.Result == nil {
		return Result{err: res.Err}
	}
	return Result{
		res: res.Result.(sql.Result),
		err: res.Err,
	}
}

package orm

import (
	"context"
	"github.com/CoucouMonEcho/go-framework/orm/internal/accessor"
	"github.com/CoucouMonEcho/go-framework/orm/model"
)

type core struct {
	model   *model.Model
	dialect Dialect
	creator accessor.Creator
	r       model.Registry

	middlewares []Middleware
}

func get[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	var root Handler = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return getHandler[T](ctx, sess, c, qc)
	}
	// handler not executed -> build method not executed -> model need initialize
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		root = c.middlewares[i](root)
	}
	return root(ctx, qc)
}

func getHandler[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	query, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{Err: err}
	}

	rows, err := sess.queryContext(ctx, query.SQL, query.Args...)
	if err != nil {
		return &QueryResult{Err: err}
	}
	defer func() {
		if rows != nil {
			err = rows.Close()
		}
	}()
	if !rows.Next() {
		return &QueryResult{Err: ErrNoRows}
	}

	tp := new(T)
	acc := c.creator(qc.Model, tp)
	err = acc.SetColumns(rows)

	return &QueryResult{
		Result: tp,
		Err:    err,
	}
}

func getMulti[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	var root Handler = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return getMultiHandler[T](ctx, sess, c, qc)
	}
	// handler not executed -> build method not executed -> model need initialize
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		root = c.middlewares[i](root)
	}
	return root(ctx, qc)
}

func getMultiHandler[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	query, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{Err: err}
	}

	rows, err := sess.queryContext(ctx, query.SQL, query.Args...)
	if err != nil {
		return &QueryResult{Err: err}
	}
	defer func() {
		if rows != nil {
			err = rows.Close()
		}
	}()

	tps := make([]*T, 16)
	for rows.Next() {
		tp := new(T)
		acc := c.creator(qc.Model, tp)
		err = acc.SetColumns(rows)

		tps = append(tps, tp)
	}

	return &QueryResult{
		Result: tps,
		Err:    err,
	}
}

func exec(ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	var root Handler = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return execHandler(ctx, sess, qc)
	}
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		root = c.middlewares[i](root)
	}
	return root(ctx, qc)
}

func execHandler(ctx context.Context, sess Session, qc *QueryContext) *QueryResult {
	query, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{
			Result: Result{
				err: err,
			},
		}
	}
	res, err := sess.execContext(ctx, query.SQL, query.Args...)
	return &QueryResult{
		Result: Result{
			res: res,
			err: err,
		},
	}
}

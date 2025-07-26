package orm

import (
	"context"
	"database/sql"
)

type Deleter[T any] struct {
	builder
	table string

	sess Session

	where []Predicate
}

func (d *Deleter[T]) Exec(ctx context.Context) Result {
	// initialize model
	var err error
	if d.model, err = d.r.Get(new(T)); err != nil {
		return Result{err: err}
	}
	res := exec(ctx, d.sess, d.core, &QueryContext{
		Type:    "DELETE",
		Builder: d,
		Model:   d.model,
	})
	if res.Result == nil {
		return Result{err: res.Err}
	}
	return Result{
		res: res.Result.(sql.Result),
		err: res.Err,
	}
}

func NewDeleter[T any](sess Session) *Deleter[T] {
	c := sess.getCore()
	return &Deleter[T]{
		builder: builder{
			core:   c,
			quoter: c.dialect.quoter(),
		},
		sess: sess,
	}
}

func (d *Deleter[T]) Build() (*Query, error) {
	var err error
	if d.model == nil {
		if d.model, err = d.r.Get(new(T)); err != nil {
			return nil, err
		}
	}

	d.sb.WriteString("DELETE FROM ")

	// table name
	if d.table == "" {
		d.quote(d.model.TableName)
	} else {
		d.sb.WriteString(d.table)
	}

	// where
	if len(d.where) > 0 {
		d.sb.WriteString(" WHERE ")
		err = d.buildPredicates(d.where)
		if err != nil {
			return nil, err
		}
	}

	d.sb.WriteByte(';')
	return &Query{
		SQL:  d.sb.String(),
		Args: d.args,
	}, nil
}

// Where
// id := []int{1, 2, 3}
// wrong -> s.Where("id in (?, ?, ?)", ids)
// right -> s.Where("id in (?, ?, ?)", ids...)
func (d *Deleter[T]) Where(ps ...Predicate) *Deleter[T] {
	d.where = ps
	return d
}

func (d *Deleter[T]) From(table string) *Deleter[T] {
	d.table = table
	return d
}

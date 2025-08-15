package orm

import (
	"context"
	"database/sql"
)

type Updater[T any] struct {
	builder
	table   string
	columns []Column

	sess Session

	where []Predicate
}

func NewUpdater[T any](sess Session) *Updater[T] {
	c := sess.getCore()
	return &Updater[T]{
		builder: builder{
			core:   c,
			quoter: c.dialect.quoter(),
		},
		sess: sess,
	}
}

func (u *Updater[T]) Build() (*Query, error) {
	var err error
	if u.model == nil {
		if u.model, err = u.r.Get(new(T)); err != nil {
			return nil, err
		}
	}

	u.sb.WriteString("UPDATE ")

	// table name
	if u.table == "" {
		u.quote(u.model.TableName)
	} else {
		u.sb.WriteString(u.table)
	}

	u.sb.WriteString(" SET ")

	// columns
	if len(u.columns) > 0 {
		for i, col := range u.columns {
			if i > 0 {
				u.sb.WriteString(", ")
			}
			if err = u.buildColumn(col); err != nil {
				return nil, err
			}
			u.sb.WriteString(" = ?")
			i++
		}
	}

	// where
	if len(u.where) > 0 {
		u.sb.WriteString(" WHERE ")
		err = u.buildPredicates(u.where)
		if err != nil {
			return nil, err
		}
	}

	u.sb.WriteByte(';')
	return &Query{
		SQL:  u.sb.String(),
		Args: u.args,
	}, err
}

func (u *Updater[T]) Update(table string) *Updater[T] {
	u.table = table
	return u
}

func (u *Updater[T]) Set(col Column, val any) *Updater[T] {
	u.columns = append(u.columns, col)
	u.args = append(u.args, val)
	return u
}

func (u *Updater[T]) Where(ps ...Predicate) *Updater[T] {
	u.where = ps
	return u
}

func (u *Updater[T]) Exec(ctx context.Context) Result {
	// initialize model
	var err error
	if u.model, err = u.r.Get(new(T)); err != nil {
		return Result{err: err}
	}
	u.sb.Reset()
	res := exec(ctx, u.sess, u.core, &QueryContext{
		Type:    "UPDATE",
		Builder: u,
		Model:   u.model,
	})
	if res.Result == nil {
		return Result{err: res.Err}
	}
	return Result{
		res: res.Result.(sql.Result),
		err: res.Err,
	}
}

package orm

import (
	"context"
	"database/sql"
	"go-framework/orm/internal/errs"
	"go-framework/orm/model"
)

type Assignable interface {
	assign()
}

type Upsert struct {
	assigns         []Assignable
	conflictColumns []string
}

type UpsertBuilder[T any] struct {
	i               *Inserter[T]
	conflictColumns []string
}

func (o *UpsertBuilder[T]) ConflictColumns(cols ...string) *UpsertBuilder[T] {
	o.conflictColumns = cols
	return o
}

func (o *UpsertBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.upsert = &Upsert{
		assigns:         assigns,
		conflictColumns: o.conflictColumns,
	}
	return o.i
}

type Inserter[T any] struct {
	builder
	table   string
	columns []string

	sess Session

	values []*T
	upsert *Upsert
	//onDuplicate []Assignable
}

func NewInserter[T any](sess Session) *Inserter[T] {
	c := sess.getCore()
	return &Inserter[T]{
		builder: builder{
			core:   c,
			quoter: c.dialect.quoter(),
		},
		sess: sess,
	}
}

func (i *Inserter[T]) Build() (*Query, error) {
	if len(i.values) == 0 {
		return nil, errs.ErrInsertZeroRow
	}
	var err error
	if i.model == nil {
		if i.model, err = i.r.Get(new(T)); err != nil {
			return nil, err
		}
	}

	i.sb.WriteString("INSERT INTO ")

	// table name
	if i.table == "" {
		i.quote(i.model.TableName)
	} else {
		i.sb.WriteString(i.table)
	}

	// the order in which the map is traversed is random
	fields := i.model.Fields
	if len(i.columns) > 0 {
		fields = make([]*model.Field, 0, len(i.columns))
		for _, col := range i.columns {
			fd, ok := i.model.FieldMap[col]
			if !ok {
				return nil, errs.NewErrUnknownField(col)
			}
			fields = append(fields, fd)
		}
	}

	// columns
	i.sb.WriteByte('(')
	for i1, field := range fields {
		if i1 > 0 {
			i.sb.WriteString(", ")
		}
		i.quote(field.ColName)
		i1++
	}
	i.sb.WriteByte(')')

	i.sb.WriteString(" VALUES ")

	// values
	i.args = make([]any, 0, len(i.values)*len(fields))
	for i1, val := range i.values {
		if i1 > 0 {
			i.sb.WriteByte(',')
		}
		i.sb.WriteByte('(')
		acc := i.creator(i.model, val)
		for i2, field := range fields {
			if i2 > 0 {
				i.sb.WriteString(", ")
			}
			i.sb.WriteByte('?')
			arg, err := acc.Field(field.GoName)
			if err != nil {
				return nil, err
			}
			i.addArgs(arg)
		}
		i.sb.WriteByte(')')
	}

	// upsert
	if i.upsert != nil {
		if err = i.dialect.buildUpsert(&i.builder, i.upsert); err != nil {
			return nil, err
		}
	}

	i.sb.WriteByte(';')
	return &Query{
		SQL:  i.sb.String(),
		Args: i.args,
	}, err
}

func (i *Inserter[T]) Into(table string) *Inserter[T] {
	i.table = table
	return i
}

func (i *Inserter[T]) Columns(cols ...string) *Inserter[T] {
	i.columns = cols
	return i
}

func (i *Inserter[T]) Values(vals ...*T) *Inserter[T] {
	i.values = vals
	return i
}

//func (i *Inserter[T]) onDuplicate(assigns ...Assignable) *Inserter[T] {
//	i.onDuplicate = assigns
//	return i
//}

func (i *Inserter[T]) OnDuplicateKey() *UpsertBuilder[T] {
	return &UpsertBuilder[T]{
		i: i,
	}
}

func (i *Inserter[T]) Exec(ctx context.Context) Result {
	// initialize model
	var err error
	if i.model, err = i.r.Get(new(T)); err != nil {
		return Result{err: err}
	}
	res := exec(ctx, i.sess, i.core, &QueryContext{
		Type:    "INSERT",
		Builder: i,
		Model:   i.model,
	})
	if res.Result == nil {
		return Result{err: res.Err}
	}
	return Result{
		res: res.Result.(sql.Result),
		err: res.Err,
	}
}

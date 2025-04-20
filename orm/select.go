package orm

import (
	"code-practise/orm/internal/errs"
	"context"
	"database/sql"
	"reflect"
)

type Selector[T any] struct {
	builder
	table string

	db *DB

	where []Predicate
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		db: db,
	}
}

func (s *Selector[T]) Build() (*Query, error) {
	var err error
	s.model, err = s.db.r.Register(new(T))
	if err != nil {
		return nil, err
	}

	s.sb.WriteString("SELECT * FROM ")

	// table name
	if s.table == "" {
		s.sb.WriteByte('`')
		s.sb.WriteString(s.model.tableName)
		s.sb.WriteByte('`')
	} else {
		// s.sb.WriteByte('`')
		s.sb.WriteString(s.table)
		// s.sb.WriteByte('`')
	}

	// where condition
	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		err = s.buildPredicates(s.where)
		if err != nil {
			return nil, err
		}
	}

	s.sb.WriteByte(';')
	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) addArg(val any) *Selector[T] {
	if s.args == nil {
		s.args = make([]any, 0, 4)
	}
	s.args = append(s.args, val)
	return s
}

// id := []int{1, 2, 3}
// wrong -> s.Where("id in (?, ?, ?)", ids)
// right -> s.Where("id in (?, ?, ?)", ids...)
func (s *Selector[T]) Where(ps ...Predicate) *Selector[T] {
	s.where = ps
	return s
}

func (s *Selector[T]) From(table string) *Selector[T] {
	s.table = table
	return s
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	query, err := s.Build()
	if err != nil {
		return nil, err
	}

	db := s.db.db
	rows, err := db.QueryContext(ctx, query.SQL, query.Args...)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, ErrNoRows
	}

	tp, err := s.doGet(rows)
	if err != nil {
		return nil, err
	}

	return tp, nil
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	query, err := s.Build()
	if err != nil {
		return nil, err
	}

	db := s.db.db
	rows, err := db.QueryContext(ctx, query.SQL, query.Args...)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	tps := make([]*T, 16)

	for rows.Next() {
		tp, err := s.doGet(rows)
		if err != nil {
			return nil, err
		}
		tps = append(tps, tp)
	}
	return tps, nil
}

func (s *Selector[T]) doGet(rows *sql.Rows) (*T, error) {
	tp := new(T)
	meta, err := s.db.r.Get(tp)
	if err != nil {
		return nil, err
	}
	// select column
	cs, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	if len(cs) > len(meta.fieldMap) {
		return nil, errs.ErrTooManyColumns
	}
	// scan
	vals := make([]any, 0, len(cs))
	valElems := make([]reflect.Value, 0, len(cs))
	for _, c := range cs {
		fd, ok := s.model.columnMap[c]
		if !ok {
			return nil, errs.NewErrUnknownColumn(c)
		}
		// return pointer
		val := reflect.New(fd.typ)

		vals = append(vals, val.Interface())
		valElems = append(valElems, val.Elem())
	}
	err = rows.Scan(vals...)
	if err != nil {
		return nil, err
	}
	// struct
	tpValue := reflect.ValueOf(tp)
	for i, c := range cs {
		fd, ok := s.model.columnMap[c]
		if !ok {
			return nil, errs.NewErrUnknownColumn(c)
		}
		tpValue.Elem().FieldByName(fd.goName).Set(valElems[i])
	}
	return tp, nil
}

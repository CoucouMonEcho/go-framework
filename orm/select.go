package orm

import (
	"context"
)

// Selectable tag interface,
// SELECT {selectable}... for prevent SQL injection problems
type Selectable interface {
	selectable()
}

type Selector[T any] struct {
	builder
	table   string
	columns []Selectable

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

	s.sb.WriteString("SELECT")

	if len(s.columns) > 0 {
		for i, col := range s.columns {
			if i > 0 {
				s.sb.WriteByte(',')
			}
			s.sb.WriteByte(' ')
			err = s.buildColumns(col.(Column))
			if err != nil {
				return nil, err
			}
		}
	} else {
		s.sb.WriteString(" *")
	}

	s.sb.WriteString(" FROM ")

	// table name
	if s.table == "" {
		s.sb.WriteByte('`')
		s.sb.WriteString(s.model.TableName)
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

//func (s *Selector[T]) Select(cols ...string) *Selector[T] {
//	s.columns = cols
//	return s
//}

func (s *Selector[T]) Select(cols ...Selectable) *Selector[T] {
	s.columns = cols
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
	if err != nil {
		return nil, err
	}
	defer func() {
		if rows != nil {
			err = rows.Close()
		}
	}()
	if !rows.Next() {
		return nil, ErrNoRows
	}

	tp := new(T)
	acc := s.db.creator(s.model, tp)
	err = acc.SetColumns(rows)

	return tp, err
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	query, err := s.Build()
	if err != nil {
		return nil, err
	}

	db := s.db.db
	rows, err := db.QueryContext(ctx, query.SQL, query.Args...)
	defer func() {
		if rows != nil {
			err = rows.Close()
		}
	}()
	if err != nil {
		return nil, err
	}
	tps := make([]*T, 16)

	for rows.Next() {
		tp := new(T)
		acc := s.db.creator(s.model, tp)
		err = acc.SetColumns(rows)

		tps = append(tps, tp)
	}
	return tps, nil
}

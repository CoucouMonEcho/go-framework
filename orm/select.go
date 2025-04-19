package orm

import (
	"context"
)

type Selector[T any] struct {
	builder
	table string

	where []Predicate
}

func (s *Selector[T]) Build() (*Query, error) {
	var err error
	s.model, err = parseModel(new(T))
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
	//TODO implement me
	panic("implement me")
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	//TODO implement me
	panic("implement me")
}

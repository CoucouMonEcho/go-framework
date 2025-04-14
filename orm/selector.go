package orm

import (
	"context"
	"reflect"
	"strings"
)

type Selector[T any] struct {
	table string
}

func (s *Selector[T]) Build() (*Query, error) {
	var sb strings.Builder
	sb.WriteString("SELECT * FROM ")

	// table name
	if s.table == "" {
		var t T
		sb.WriteByte('`')
		sb.WriteString(reflect.TypeOf(t).Name())
		sb.WriteByte('`')
	} else {
		// sb.WriteByte('`')
		sb.WriteString(s.table)
		// sb.WriteByte('`')
	}

	sb.WriteByte(';')
	return &Query{
		SQL: sb.String(),
	}, nil
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

package orm

import (
	"context"
)

// Selectable tag interface,
// SELECT {selectable}... for prevent SQL injection problems
type Selectable interface {
	selectable()
}

var _ Querier[any] = &Selector[any]{}

type Selector[T any] struct {
	builder
	table   string
	columns []Selectable

	db *DB

	groupBy []Column
	having  []Predicate
	where   []Predicate
	orderBy []OrderBy
	limit   int
	offset  int
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		builder: builder{
			dialect: db.dialect,
			quoter:  db.dialect.quoter(),
		},
		db: db,
	}
}

func (s *Selector[T]) Build() (*Query, error) {
	var err error
	s.model, err = s.db.r.Get(new(T))
	if err != nil {
		return nil, err
	}

	s.sb.WriteString("SELECT")

	// columns
	if err = s.buildColumns(); err != nil {
		return nil, err
	}
	s.sb.WriteString(" FROM ")

	// table name
	if s.table == "" {
		s.quote(s.model.TableName)
	} else {
		s.sb.WriteString(s.table)
	}

	// where
	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		err = s.buildPredicates(s.where)
		if err != nil {
			return nil, err
		}
	}

	// order by
	if len(s.groupBy) > 0 {
		s.sb.WriteString(" GROUP BY ")
		for i, col := range s.groupBy {
			if i > 0 {
				s.sb.WriteString(", ")
			}
			if err = s.buildColumn(col.name); err != nil {
				return nil, err
			}
		}
	}

	// having
	if len(s.having) > 0 {
		s.sb.WriteString(" HAVING ")
		err = s.buildPredicates(s.having)
		if err != nil {
			return nil, err
		}
	}

	// order by
	if len(s.orderBy) > 0 {
		s.sb.WriteString(" ORDER BY ")
		for i, ob := range s.orderBy {
			if i > 0 {
				s.sb.WriteString(", ")
			}
			if err = s.buildColumn(ob.column.name); err != nil {
				return nil, err
			}
			s.sb.WriteByte(' ')
			s.sb.WriteString(ob.order)
		}
	}

	// limit
	if s.limit > 0 {
		s.sb.WriteString(" LIMIT ?")
		s.addArgs(s.limit)
	}

	// offset
	if s.offset > 0 {
		s.sb.WriteString(" OFFSET ?")
		s.addArgs(s.offset)
	}

	s.sb.WriteByte(';')
	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) buildColumns() error {
	if len(s.columns) == 0 {
		s.sb.WriteString(" *")
		return nil
	}
	for i, col := range s.columns {
		if i > 0 {
			s.sb.WriteByte(',')
		}
		s.sb.WriteByte(' ')

		switch c := col.(type) {
		case Column:
			if err := s.buildColumn(c.name); err != nil {
				return err
			}
			if c.alias != "" {
				s.sb.WriteString(" AS ")
				s.quote(c.alias)
			}
		case Aggregate:
			if err := s.buildAggregate(c); err != nil {
				return err
			}
			if c.alias != "" {
				s.sb.WriteString(" AS ")
				s.quote(c.alias)
			}
		case RawExpr:
			s.sb.WriteString(c.raw)
			s.addArgs(c.args...)
		}
	}
	return nil
}

//func (s *Selector[T]) Select(cols ...string) *Selector[T] {
//	s.columns = cols
//	return s
//}

func (s *Selector[T]) Select(cols ...Selectable) *Selector[T] {
	s.columns = cols
	return s
}

func (s *Selector[T]) From(table string) *Selector[T] {
	s.table = table
	return s
}

func (s *Selector[T]) Where(ps ...Predicate) *Selector[T] {
	s.where = ps
	return s
}

func (s *Selector[T]) GroupBy(cols ...Column) *Selector[T] {
	s.groupBy = cols
	return s
}

func (s *Selector[T]) Having(ps ...Predicate) *Selector[T] {
	s.having = ps
	return s
}

type OrderBy struct {
	column Column
	order  string
}

func Asc(field string) OrderBy {
	return OrderBy{
		column: C(field),
		order:  "ASC",
	}
}

func Desc(field string) OrderBy {
	return OrderBy{
		column: C(field),
		order:  "DESC",
	}
}

func (s *Selector[T]) OrderBy(obs ...OrderBy) *Selector[T] {
	s.orderBy = obs
	return s
}

func (s *Selector[T]) Offset(offset int) *Selector[T] {
	s.offset = offset
	return s
}

func (s *Selector[T]) Limit(limit int) *Selector[T] {
	s.limit = limit
	return s
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	query, err := s.Limit(1).Build()
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

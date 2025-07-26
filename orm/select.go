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
	table   TableReference
	columns []Selectable

	sess Session

	groupBy []Column
	having  []Predicate
	where   []Predicate
	orderBy []OrderBy
	limit   int
	offset  int
}

func NewSelector[T any](sess Session) *Selector[T] {
	c := sess.getCore()
	return &Selector[T]{
		builder: builder{
			core:   c,
			quoter: c.dialect.quoter(),
		},
		sess: sess,
	}
}

func (s *Selector[T]) Build() (*Query, error) {
	var err error
	if s.model == nil {
		if s.model, err = s.r.Get(new(T)); err != nil {
			return nil, err
		}
	}

	s.sb.WriteString("SELECT")

	// columns
	if err = s.buildColumns(); err != nil {
		return nil, err
	}
	s.sb.WriteString(" FROM ")

	// table name
	if err = s.buildTable(s.table); err != nil {
		return nil, err
	}

	// where
	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		err = s.buildPredicates(s.where)
		if err != nil {
			return nil, err
		}
	}

	// group by
	if len(s.groupBy) > 0 {
		s.sb.WriteString(" GROUP BY ")
		for i, col := range s.groupBy {
			if i > 0 {
				s.sb.WriteString(", ")
			}
			if err = s.buildColumn(col); err != nil {
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
			if err = s.buildColumn(ob.column); err != nil {
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
			if err := s.buildColumn(c); err != nil {
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

func (s *Selector[T]) From(table TableReference) *Selector[T] {
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

func (s *Selector[T]) AsSubquery(alias string) Subquery {
	t := s.table
	if t == nil {
		t = TableOf(new(T))
	}
	return Subquery{
		t:       t,
		builder: s,
		columns: s.columns,
		alias:   alias,
	}
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	// initialize model
	var err error
	if s.model, err = s.r.Get(new(T)); err != nil {
		return nil, err
	}
	res := get[T](ctx, s.sess, s.core, &QueryContext{
		Type:    "SELECT",
		Builder: s.Limit(1),
		Model:   s.model,
	})
	if res.Result != nil {
		return res.Result.(*T), res.Err
	}
	return nil, res.Err
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	// initialize model
	var err error
	if s.model, err = s.r.Get(new(T)); err != nil {
		return nil, err
	}
	res := getMulti[T](ctx, s.sess, s.core, &QueryContext{
		Type:    "SELECT",
		Builder: s,
		Model:   s.model,
	})
	if res.Result != nil {
		return res.Result.([]*T), res.Err
	}
	return nil, res.Err
}

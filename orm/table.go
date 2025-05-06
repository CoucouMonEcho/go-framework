package orm

var (
	_ TableReference = &Table{}
	_ TableReference = &Join{}
	_ TableReference = &Subquery{}
)

type TableReference interface {
	table()
}

type Table struct {
	entity any
	alias  string
}

func TableOf(entity any) Table {
	return Table{entity: entity}
}

func (Table) table() {}

func (t Table) As(alias string) Table {
	return Table{
		entity: t.entity,
		alias:  alias,
	}
}

func (t Table) C(name string) Column {
	return Column{
		name:  name,
		table: t,
	}
}

func (t Table) Join(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: right,
		typ:   "JOIN",
	}
}

func (t Table) LeftJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: right,
		typ:   "LEFT JOIN",
	}
}

func (t Table) RightJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: right,
		typ:   "RIGHT JOIN",
	}
}

type Join struct {
	left  TableReference
	right TableReference
	typ   string
	on    []Predicate
	using []string
}

func (Join) table() {}

func (j Join) Join(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: right,
		typ:   "JOIN",
	}
}

func (j Join) LeftJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: right,
		typ:   "LEFT JOIN",
	}
}

func (j Join) RightJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: right,
		typ:   "RIGHT JOIN",
	}
}

type JoinBuilder struct {
	left  TableReference
	right TableReference
	typ   string
}

func (j *JoinBuilder) On(ps ...Predicate) Join {
	return Join{
		left:  j.left,
		right: j.right,
		typ:   j.typ,
		on:    ps,
	}
}

func (j *JoinBuilder) Using(cols ...string) Join {
	return Join{
		left:  j.left,
		right: j.right,
		typ:   j.typ,
		using: cols,
	}
}

type Subquery struct {
	t       TableReference
	builder QueryBuilder
	columns []Selectable
	alias   string
}

func (Subquery) table() {}

func (Subquery) expr() {}

func (s Subquery) tableAlias() string {
	return s.alias
}

func (s Subquery) Join(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  s,
		right: right,
		typ:   "JOIN",
	}
}

func (s Subquery) LeftJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  s,
		right: right,
		typ:   "LEFT JOIN",
	}
}

func (s Subquery) RightJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  s,
		right: right,
		typ:   "RIGHT JOIN",
	}
}

func Any(sub Subquery) SubqueryExpr {
	return SubqueryExpr{
		s:    sub,
		pred: "ANY",
	}
}

func All(sub Subquery) SubqueryExpr {
	return SubqueryExpr{
		s:    sub,
		pred: "ALL",
	}
}

func Some(sub Subquery) SubqueryExpr {
	return SubqueryExpr{
		s:    sub,
		pred: "SOME",
	}
}

func Exists(sub Subquery) Predicate {
	return Predicate{
		op:    opExists,
		right: sub,
	}
}

type SubqueryExpr struct {
	s    Subquery
	pred string
}

func (SubqueryExpr) expr() {}

package orm

var (
	_ Expression = &Aggregate{}
	_ Expression = &Column{}
	_ Expression = &Predicate{}
	_ Expression = &RawExpr{}
	_ Expression = &value{}
	_ Expression = &Subquery{}
	_ Expression = &SubqueryExpr{}
)

// Expression tag interface,
// which represents an expression
type Expression interface {
	expr()
}

type RawExpr struct {
	raw  string
	args []any
}

func (RawExpr) expr() {}

func (RawExpr) selectable() {}

func Raw(raw string, args ...any) RawExpr {
	return RawExpr{
		raw:  raw,
		args: args,
	}
}

func (r RawExpr) AsPredicate() Predicate {
	return Predicate{
		left: r,
	}
}

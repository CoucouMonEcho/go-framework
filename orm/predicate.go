package orm

// nickname
// type op = string

// derivation type
type op string

const (
	opEQ  op = "="
	opLT  op = "<"
	opGT  op = ">"
	opNOT op = "NOT"
	opAND op = "AND"
	opOR  op = "OR"
)

func (o op) String() string {
	return string(o)
}

// Expression tag interface,
// which represents an expression
type Expression interface {
	expr()
}

func (Predicate) expr() {}

type Predicate struct {
	left  Expression
	op    op
	right Expression
}

// Eq("id", 123)
// subquery Eq(sub, "id", 123)
// subquery Eq(sub.id, 123)
//func Eq(column string, right any) Predicate {
//	return Predicate{
//		Column: column,
//		Op:     "=",
//		Arg:    right,
//	}
//}

func (value) expr() {}

type value struct {
	val any
}

// Eq C("id").Eq(123)
// sub query sub.C("id").Eq(123)
func (c Column) Eq(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opEQ,
		right: value{val: arg},
	}
}

func (c Column) Lt(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opLT,
		right: value{val: arg},
	}
}

func (c Column) Gt(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opGT,
		right: value{val: arg},
	}
}

// Not Not(C("name").Eq("user1"))
func Not(p Predicate) Predicate {
	return Predicate{
		op:    opNOT,
		right: p,
	}
}

// And C("id").Eq(123).And(C("name").Eq("user1"))
func (left Predicate) And(right Predicate) Predicate {
	return Predicate{
		left:  left,
		op:    opAND,
		right: right,
	}
}

// Or C("id").Eq(123).Or(C("name").Eq("user1"))
func (left Predicate) Or(right Predicate) Predicate {
	return Predicate{
		left:  left,
		op:    opOR,
		right: right,
	}
}

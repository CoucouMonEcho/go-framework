package orm

// nickname
// type fn = string

// derivation type
type fn string

const (
	fnCOUNT fn = "COUNT"
	fnAVG   fn = "AVG"
	fnSUM   fn = "SUM"
	fnMAX   fn = "MAX"
	fnMIN   fn = "MIN"
)

// Aggregate
// AVG("age"), SUM("score"), COUNT("id"), MAX("create_time"), MIN("update_time")
type Aggregate struct {
	fn    fn
	arg   string
	alias string
}

func (Aggregate) expr() {}

func (Aggregate) selectable() {}

func Count(col string) Aggregate {
	return Aggregate{
		fn:  fnCOUNT,
		arg: col,
	}
}

func Avg(col string) Aggregate {
	return Aggregate{
		fn:  fnAVG,
		arg: col,
	}
}

func Sum(col string) Aggregate {
	return Aggregate{
		fn:  fnSUM,
		arg: col,
	}
}

func Max(col string) Aggregate {
	return Aggregate{
		fn:  fnMAX,
		arg: col,
	}
}

func Min(col string) Aggregate {
	return Aggregate{
		fn:  fnMIN,
		arg: col,
	}
}

func (a Aggregate) As(alias string) Aggregate {
	return Aggregate{
		fn:    a.fn,
		arg:   a.arg,
		alias: alias,
	}
}

func (a Aggregate) Eq(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opEQ,
		right: value{arg},
	}
}

func (a Aggregate) Lt(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opLT,
		right: value{arg},
	}
}

func (a Aggregate) Gt(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opGT,
		right: value{arg},
	}
}

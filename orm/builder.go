package orm

import (
	"code-practise/orm/internal/errs"
	"strings"
)

type builder struct {
	core
	sb     strings.Builder
	args   []any
	quoter byte
}

func (b *builder) quote(name string) {
	b.sb.WriteByte(b.quoter)
	b.sb.WriteString(name)
	b.sb.WriteByte(b.quoter)
}

func (b *builder) buildPredicates(ps []Predicate) error {
	p := ps[0]
	for i := 1; i < len(ps); i++ {
		p = p.And(ps[i])
	}
	return b.buildExpression(p)
}

func (b *builder) buildExpression(expr Expression) error {
	switch exprTrans := expr.(type) {
	case nil:
	case Predicate:
		// left
		_, ok := exprTrans.left.(Predicate)
		if ok {
			b.sb.WriteByte('(')
		}
		if err := b.buildExpression(exprTrans.left); err != nil {
			return err
		}
		if ok {
			b.sb.WriteByte(')')
		}
		// op
		if exprTrans.op != "" {
			if exprTrans.left != nil {
				b.sb.WriteByte(' ')
			}
			b.sb.WriteString(exprTrans.op.String())
			b.sb.WriteByte(' ')
		}
		// right
		_, ok = exprTrans.right.(Predicate)
		if ok {
			b.sb.WriteByte('(')
		}
		if err := b.buildExpression(exprTrans.right); err != nil {
			return err
		}
		if ok {
			b.sb.WriteByte(')')
		}
	case Column:
		// implicitly forbidden to use alias in where statements
		//exprTrans.alias = ""
		return b.buildColumn(exprTrans)
	case Aggregate:
		return b.buildAggregate(exprTrans)
	case value:
		b.sb.WriteByte('?')
		b.addArgs(exprTrans.val)
	case RawExpr:
		b.sb.WriteString(exprTrans.raw)
		b.addArgs(exprTrans.args...)
	default:
		return errs.NewErrUnsupportedExpression(expr)
	}
	return nil
}

func (b *builder) buildColumn(c Column) error {
	switch table := c.table.(type) {
	case nil:
		fd, ok := b.model.FieldMap[c.name]
		if !ok {
			return errs.NewErrUnknownField(c.name)
		}
		b.quote(fd.ColName)
	case Table:
		m, err := b.r.Get(table.entity)
		if err != nil {
			return err
		}
		fd, ok := m.FieldMap[c.name]
		if !ok {
			return errs.NewErrUnknownField(c.name)
		}
		if table.alias != "" {
			b.quote(table.alias)
		} else {
			b.quote(m.TableName)
		}
		b.sb.WriteByte('.')
		b.quote(fd.ColName)
	default:
		return errs.NewErrUnsupportedTableReference(table)
	}
	return nil
}

func (b *builder) buildAggregate(a Aggregate) error {
	b.sb.WriteString(string(a.fn))
	b.sb.WriteByte('(')
	if err := b.buildColumn(C(a.arg)); err != nil {
		return err
	}
	b.sb.WriteByte(')')
	return nil
}

func (b *builder) addArgs(vals ...any) {
	if len(vals) == 0 {
		return
	}
	if b.args == nil {
		b.args = make([]any, 0, 8)
	}
	b.args = append(b.args, vals...)
	return
}

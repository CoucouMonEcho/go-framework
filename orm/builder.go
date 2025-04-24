package orm

import (
	"code-practise/orm/internal/errs"
	model2 "code-practise/orm/model"
	"strings"
)

type builder struct {
	sb    strings.Builder
	args  []any
	model *model2.Model
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
		if exprTrans.left != nil {
			b.sb.WriteByte(' ')
		}
		b.sb.WriteString(exprTrans.op.String())
		b.sb.WriteByte(' ')
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
		return b.buildColumns(exprTrans)
	case value:
		b.args = append(b.args, exprTrans.val)
		b.sb.WriteByte('?')
	default:
		return errs.NewErrUnsupportedExpression(expr)
	}
	return nil
}

func (b *builder) buildColumns(c Column) error {
	fd, ok := b.model.FieldMap[c.name]
	if !ok {
		return errs.NewErrUnknownField(c.name)
	}
	b.sb.WriteByte('`')
	b.sb.WriteString(fd.ColName)
	b.sb.WriteByte('`')
	return nil
}

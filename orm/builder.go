package orm

import (
	"code-practise/orm/internal/errs"
	"code-practise/orm/model"
	"strings"
)

type builder struct {
	sb    strings.Builder
	args  []any
	model *model.Model

	dialect Dialect
	quoter  byte
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
		exprTrans.alias = ""
		return b.buildColumn(exprTrans.name)
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

func (b *builder) buildColumn(name string) error {
	fd, ok := b.model.FieldMap[name]
	if !ok {
		return errs.NewErrUnknownField(name)
	}
	b.quote(fd.ColName)
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

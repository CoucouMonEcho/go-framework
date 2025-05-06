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
	case Subquery:
		if err := b.buildSubquery(exprTrans); err != nil {
			return err
		}
	case SubqueryExpr:
		b.sb.WriteString(exprTrans.pred)
		b.sb.WriteByte(' ')
		if err := b.buildSubquery(exprTrans.s); err != nil {
			return err
		}
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
	case Join:
		if b.buildColumn(Column{table: table.left, name: c.name}) != nil {
			return b.buildColumn(Column{table: table.right, name: c.name})
		}
	case Subquery:
		if len(table.columns) > 0 {
			//FIXME
			for _, col := range table.columns {
				if colTrans, ok := col.(Column); ok && colTrans.name == c.name {
					return b.buildColumn(colTrans)
				}
			}
			return errs.NewErrUnknownField(c.name)
		}
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

func (b *builder) buildSubquery(s Subquery) error {
	query, err := s.builder.Build()
	if err != nil {
		return err
	}
	b.sb.WriteByte('(')
	b.sb.WriteString(query.SQL[:len(query.SQL)-1])
	if len(query.Args) > 0 {
		b.addArgs(query.Args)
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

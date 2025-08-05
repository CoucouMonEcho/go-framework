package orm

import (
	"github.com/CoucouMonEcho/go-framework/orm/internal/errs"
)

var (
	DialectMySQL      Dialect = mysqlDialect{}
	DialectSQLite     Dialect = sqliteDialect{}
	DialectPostgreSQL Dialect = postgresqlDialect{}
)

type Dialect interface {
	// quoter, MYSQL(`),POSTGRESQL('), ORACLE(")
	quoter() byte
	buildUpsert(b *builder, upsert *Upsert) error
}

var _ Dialect = standardSQL{}

type standardSQL struct {
}

func (s standardSQL) quoter() byte {
	//TODO implement me
	panic("implement me")
}

func (s standardSQL) buildUpsert(b *builder, upsert *Upsert) error {
	b.sb.WriteString(" ON CONFLICT (")
	for i, col := range upsert.conflictColumns {
		if i > 0 {
			b.sb.WriteString(", ")
		}
		if err := b.buildColumn(C(col)); err != nil {
			return err
		}
	}
	b.sb.WriteString(") DO UPDATE SET ")
	for i, assign := range upsert.assigns {
		if i > 0 {
			b.sb.WriteString(", ")
		}
		switch a := assign.(type) {
		case Assignment:
			fd, ok := b.model.FieldMap[a.col]
			if !ok {
				return errs.NewErrUnknownField(a.col)
			}
			b.quote(fd.ColName)
			b.sb.WriteString(" = ?")
			b.addArgs(a.val)
		case Column:
			fd, ok := b.model.FieldMap[a.name]
			if !ok {
				return errs.NewErrUnknownField(a.name)
			}
			b.quote(fd.ColName)
			b.sb.WriteString(" = excluded.")
			b.quote(fd.ColName)
		default:
			return errs.NewErrUnsupportedAssignable(a)
		}
	}
	return nil
}

type mysqlDialect struct {
	standardSQL
	quote byte
}

func (m mysqlDialect) quoter() byte {
	return '`'
}

func (m mysqlDialect) buildUpsert(b *builder, upsert *Upsert) error {
	b.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
	for i, assign := range upsert.assigns {
		if i > 0 {
			b.sb.WriteString(", ")
		}
		switch a := assign.(type) {
		case Assignment:
			fd, ok := b.model.FieldMap[a.col]
			if !ok {
				return errs.NewErrUnknownField(a.col)
			}
			b.quote(fd.ColName)
			b.sb.WriteString(" = ?")
			b.addArgs(a.val)
		case Column:
			fd, ok := b.model.FieldMap[a.name]
			if !ok {
				return errs.NewErrUnknownField(a.name)
			}
			b.quote(fd.ColName)
			b.sb.WriteString(" = VALUES(")
			b.quote(fd.ColName)
			b.sb.WriteString(")")
		default:
			return errs.NewErrUnsupportedAssignable(a)
		}
	}
	return nil
}

type sqliteDialect struct {
	standardSQL
}

func (s sqliteDialect) quoter() byte {
	return '`'
}

type postgresqlDialect struct {
	standardSQL
}

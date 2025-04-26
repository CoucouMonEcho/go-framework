package orm

import (
	"code-practise/orm/internal/errs"
)

var (
	DialectMySQL      Dialect = mysqlDialect{}
	DialectSQLite     Dialect = sqliteDialect{}
	DialectPostgreSQL Dialect = postgresqlDialect{}
)

type Dialect interface {
	// quoter, MYSQL(`),POSTGRESQL('), ORACLE(")
	quoter() byte
	buildOnDuplicateKey(b *builder, odk *OnDuplicateKey) error
}

var _ Dialect = standardSQL{}

type standardSQL struct {
}

func (s standardSQL) quoter() byte {
	//TODO implement me
	panic("implement me")
}

func (s standardSQL) buildOnDuplicateKey(b *builder, odk *OnDuplicateKey) error {
	//TODO implement me
	panic("implement me")
}

type mysqlDialect struct {
	standardSQL
	quote byte
}

func (m mysqlDialect) quoter() byte {
	return '`'
}

func (m mysqlDialect) buildOnDuplicateKey(b *builder, odk *OnDuplicateKey) error {
	b.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
	for i1, assign := range odk.assigns {
		if i1 > 0 {
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

type postgresqlDialect struct {
	standardSQL
}

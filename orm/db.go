package orm

import (
	"code-practise/orm/internal/accessor"
	"code-practise/orm/model"
	"database/sql"
)

// DB is decorator of sql.DB
type DB struct {
	r       model.Registry
	db      *sql.DB
	creator accessor.Creator
}

type DBOption func(db *DB)

func Open(driver string, dataSourceName string, opts ...DBOption) (*DB, error) {
	db, err := sql.Open(driver, dataSourceName)
	if err != nil {
		return nil, err
	}
	return OpenDB(db, opts...)
}

func OpenDB(db *sql.DB, opts ...DBOption) (*DB, error) {
	res := &DB{
		r:       model.NewRegistry(),
		db:      db,
		creator: accessor.NewUnsafeAccess,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func DBUseReflect() DBOption {
	return func(db *DB) {
		db.creator = accessor.NewReflectAccess
	}
}

func MustOpen(driver string, dataSourceName string, opts ...DBOption) *DB {
	res, err := Open(driver, dataSourceName, opts...)
	if err != nil {
		panic(err)
	}
	return res
}

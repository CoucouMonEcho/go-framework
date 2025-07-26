package orm

import (
	"code-practise/orm/internal/accessor"
	"code-practise/orm/internal/errs"
	"code-practise/orm/model"
	"context"
	"database/sql"
)

// DB is decorator of sql.DB
type DB struct {
	core
	db *sql.DB
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
		db: db,
		core: core{
			r:       model.NewRegistry(),
			creator: accessor.NewUnsafeAccess,
			dialect: DialectMySQL,
		},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx}, nil
}

func (db *DB) getCore() core {
	return db.core
}

func (db *DB) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.db.QueryContext(ctx, query, args...)
}

func (db *DB) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.db.ExecContext(ctx, query, args...)
}

func (db *DB) DoTx(ctx context.Context,
	fn func(ctx context.Context, tx *Tx) error,
	opts *sql.TxOptions) (err error) {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	panicked := true
	defer func() {
		if panicked || err != nil {
			err = errs.NewErrFailedToRollbackTx(err, tx.Rollback(), panicked)
		} else {
			err = tx.Commit()
		}
	}()
	err = fn(ctx, tx)
	panicked = false
	return err
}

func DBWithMiddlewares(middlewares ...Middleware) DBOption {
	return func(db *DB) {
		db.middlewares = middlewares
	}
}

func DBWithDialect(dialect Dialect) DBOption {
	return func(db *DB) {
		db.dialect = dialect
	}
}

func DBUseReflect() DBOption {
	return func(db *DB) {
		db.creator = accessor.NewReflectAccess
	}
}

// DBWithRegistry can be used to share the same registry with different dbs
func DBWithRegistry(r model.Registry) DBOption {
	return func(db *DB) {
		db.r = r
	}
}

func MustOpen(driver string, dataSourceName string, opts ...DBOption) *DB {
	res, err := Open(driver, dataSourceName, opts...)
	if err != nil {
		panic(err)
	}
	return res
}

package orm

import (
	"context"
	"database/sql"
)

var (
	_ Session = &DB{}
	_ Session = &Tx{}
)

type Session interface {
	getCore() core
	queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	execContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type Tx struct {
	tx *sql.Tx
	db *DB
}

func (t *Tx) getCore() core {
	return t.db.getCore()
}

func (t *Tx) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

func (t *Tx) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

func (t *Tx) Commit() error {
	return t.tx.Commit()
}

func (t *Tx) Rollback() error {
	return t.tx.Rollback()
}

func (t *Tx) RollbackIfNotCommit() error {
	err := t.tx.Rollback()
	if err != sql.ErrTxDone {
		return err
	}
	return nil
}

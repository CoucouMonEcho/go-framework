package orm

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDB_Tx(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()
	db, err := OpenDB(mockDB)
	require.NoError(t, err)

	mock.ExpectBegin()
	mock.ExpectCommit()

	err = db.DoTx(context.Background(), func(ctx context.Context, tx *Tx) error {
		// do something
		return nil
	}, &sql.TxOptions{})
	require.NoError(t, err)
}

package sql

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
	"time"
)

func TestSQLMock(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	mockRows := sqlmock.NewRows([]string{"id", "first_name"}).AddRow(1, "Tom")
	mock.ExpectQuery("SELECT `id`, `first_name` FROM `user`.*").WillReturnRows(mockRows)
	mock.ExpectQuery("SELECT `id` FROM `user`.*").WillReturnError(errors.New("error"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	require.NoError(t, err)
	defer cancel()

	rows, err := db.QueryContext(ctx, "SELECT `id`, `first_name` FROM `user` WHERE `id` = ?", 1)
	require.NoError(t, err)
	for rows.Next() {
		tm := TestModel{}
		err = rows.Scan(&tm.Id, &tm.FirstName)
		require.NoError(t, err)
		log.Println(tm)
	}

	_, err = db.QueryContext(ctx, "SELECT `id` FROM `user` WHERE `id` = ?", 1)
	require.Error(t, err)
}

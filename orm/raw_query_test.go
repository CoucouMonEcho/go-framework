package orm

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRawQuery_Get(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		err = mockDB.Close()
	}()
	db, err := OpenDB(mockDB)
	require.NoError(t, err)

	// query error
	mock.ExpectQuery("SELECT.*").WillReturnError(errors.New("query error"))

	// no rows
	rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"})
	mock.ExpectQuery("SELECT.*").WillReturnRows(rows)

	// data
	rows = sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"}).
		AddRow("1", "Tom", "18", "Jerry")
	mock.ExpectQuery("SELECT.*").WillReturnRows(rows)

	//// scan error
	//rows = sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"}).
	//	AddRow("string", "Tom", "18", "Jerry")
	//mock.ExpectQuery("SELECT.*").WillReturnRows(rows)

	testCases := []struct {
		name string
		r    *RawQuerier[TestModel]

		wantErr error
		wantRes *TestModel
	}{
		{
			name:    "query error",
			r:       RawQuery[TestModel](db, "SELECT * FROM `test_model`"),
			wantErr: errors.New("query error"),
		},
		{
			name:    "no rows",
			r:       RawQuery[TestModel](db, "SELECT * FROM `test_model` WHERE `id` = ?", -1),
			wantErr: ErrNoRows,
		},
		{
			name: "data",
			r:    RawQuery[TestModel](db, "SELECT * FROM `test_model` WHERE `id` = ?", 1),
			wantRes: &TestModel{
				Id:        1,
				Age:       18,
				FirstName: "Tom",
				LastName:  &sql.NullString{Valid: true, String: "Jerry"},
			},
		},
		//{
		//	name:    "scan error",
		//	r:       RawQuery[TestModel](db,"SELECT * FROM `test_model`",),
		//	wantErr: ErrNoRows,
		//},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.r.Get(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

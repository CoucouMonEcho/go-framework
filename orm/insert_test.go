package orm

import (
	"code-practise/orm/internal/errs"
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInserter_Build(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()
	db, err := OpenDB(mockDB)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		inserter QueryBuilder
		want     *Query
		wantErr  error
	}{
		{
			name: "multiple row",
			inserter: NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       17,
				LastName:  &sql.NullString{String: "Jane", Valid: true},
			}, &TestModel{
				Id:        13,
				FirstName: "Tom",
				Age:       19,
				LastName:  &sql.NullString{String: "Jane", Valid: true},
			}),
			want: &Query{
				"INSERT INTO `test_model`(`id`, `age`, `first_name`, `last_name`) VALUES (?, ?, ?, ?),(?, ?, ?, ?);",
				[]any{
					int64(12), int8(17), "Tom", &sql.NullString{String: "Jane", Valid: true},
					int64(13), int8(19), "Tom", &sql.NullString{String: "Jane", Valid: true},
				},
			},
		},
		{
			name: "single rows",
			inserter: NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       17,
				LastName:  &sql.NullString{String: "Jane", Valid: true},
			}),
			want: &Query{
				"INSERT INTO `test_model`(`id`, `age`, `first_name`, `last_name`) VALUES (?, ?, ?, ?);",
				[]any{int64(12), int8(17), "Tom", &sql.NullString{String: "Jane", Valid: true}},
			},
		},
		{
			name:     "empty",
			inserter: NewInserter[TestModel](db).Values(),
			wantErr:  errs.ErrInsertZeroRow,
		},
		{
			name: "partial rows",
			inserter: NewInserter[TestModel](db).Columns("Id", "FirstName", "Age").Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       17,
				LastName:  &sql.NullString{String: "Jane", Valid: true},
			}),
			want: &Query{
				"INSERT INTO `test_model`(`id`, `first_name`, `age`) VALUES (?, ?, ?);",
				[]any{int64(12), "Tom", int8(17)},
			},
		},
		{
			name: "upsert-update value",
			inserter: NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       17,
				LastName:  &sql.NullString{String: "Jane", Valid: true},
			}).OnDuplicateKey().Update(
				Assign("FirstName", "haha"),
				Assign("Age", 5)),
			want: &Query{
				"INSERT INTO `test_model`(`id`, `age`, `first_name`, `last_name`) VALUES (?, ?, ?, ?)" +
					" ON DUPLICATE KEY UPDATE `first_name` = ?, `age` = ?;",
				[]any{int64(12), int8(17), "Tom", &sql.NullString{String: "Jane", Valid: true}, "haha", 5},
			},
		},
		{
			name: "upsert-update column",
			inserter: NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       17,
				LastName:  &sql.NullString{String: "Jane", Valid: true},
			}, &TestModel{
				Id:        13,
				FirstName: "Tom",
				Age:       19,
				LastName:  &sql.NullString{String: "Jane", Valid: true},
			}).OnDuplicateKey().Update(C("FirstName"), C("Age")),
			want: &Query{
				"INSERT INTO `test_model`(`id`, `age`, `first_name`, `last_name`) VALUES (?, ?, ?, ?),(?, ?, ?, ?)" +
					" ON DUPLICATE KEY UPDATE `first_name` = VALUES(`first_name`), `age` = VALUES(`age`);",
				[]any{
					int64(12), int8(17), "Tom", &sql.NullString{String: "Jane", Valid: true},
					int64(13), int8(19), "Tom", &sql.NullString{String: "Jane", Valid: true},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.inserter.Build()
			require.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			require.Equal(t, tc.want, q)
		})
	}
}

func TestInserter_SQLite_upsert(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()
	db, err := OpenDB(mockDB, DBWithDialect(DialectSQLite))
	require.NoError(t, err)

	testCases := []struct {
		name     string
		inserter QueryBuilder
		want     *Query
		wantErr  error
	}{
		{
			name: "upsert-update value",
			inserter: NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       17,
				LastName:  &sql.NullString{String: "Jane", Valid: true},
			}).OnDuplicateKey().ConflictColumns("Id").Update(
				Assign("FirstName", "haha"),
				Assign("Age", 5)),
			want: &Query{
				"INSERT INTO `test_model`(`id`, `age`, `first_name`, `last_name`) VALUES (?, ?, ?, ?)" +
					" ON CONFLICT (`id`) DO UPDATE SET `first_name` = ?, `age` = ?;",
				[]any{int64(12), int8(17), "Tom", &sql.NullString{String: "Jane", Valid: true}, "haha", 5},
			},
		},
		{
			name: "upsert-update column",
			inserter: NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       17,
				LastName:  &sql.NullString{String: "Jane", Valid: true},
			}, &TestModel{
				Id:        13,
				FirstName: "Tom",
				Age:       19,
				LastName:  &sql.NullString{String: "Jane", Valid: true},
			}).OnDuplicateKey().ConflictColumns("FirstName", "Age").Update(C("FirstName"), C("Age")),
			want: &Query{
				"INSERT INTO `test_model`(`id`, `age`, `first_name`, `last_name`) VALUES (?, ?, ?, ?),(?, ?, ?, ?)" +
					" ON CONFLICT (`first_name`, `age`) DO UPDATE SET `first_name` = excluded.`first_name`, `age` = excluded.`age`;",
				[]any{
					int64(12), int8(17), "Tom", &sql.NullString{String: "Jane", Valid: true},
					int64(13), int8(19), "Tom", &sql.NullString{String: "Jane", Valid: true},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.inserter.Build()
			require.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			require.Equal(t, tc.want, q)
		})
	}
}

func TestInserter_Exec(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if mockDB != nil {
			err = mockDB.Close()
		}
	}()
	db, err := OpenDB(mockDB, DBWithDialect(DialectSQLite))
	require.NoError(t, err)

	testCases := []struct {
		name     string
		i        *Inserter[TestModel]
		wantErr  error
		affected int64
	}{
		{
			name: "db error",
			i: func() *Inserter[TestModel] {
				mock.ExpectExec(`INSERT INTO .*`).
					WillReturnError(errors.New("db error"))
				return NewInserter[TestModel](db).Values(&TestModel{})
			}(),
			wantErr: errors.New("db error"),
		},
		{
			name: "query error",
			i: func() *Inserter[TestModel] {
				return NewInserter[TestModel](db).Values(&TestModel{}).Columns("INVALID")
			}(),
			wantErr: errs.NewErrUnknownField("INVALID"),
		},
		{
			name: "exec",
			i: func() *Inserter[TestModel] {
				res := driver.RowsAffected(1)
				mock.ExpectExec(`INSERT INTO .*`).
					WillReturnResult(res)
				return NewInserter[TestModel](db).Values(&TestModel{})
			}(),
			affected: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.i.Exec(context.Background())
			affected, err := res.RowsAffected()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.affected, affected)
		})
	}
}

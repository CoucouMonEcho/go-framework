package orm

import (
	"code-practise/orm/internal/errs"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
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

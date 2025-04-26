package orm

import (
	"code-practise/orm/internal/errs"
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSelector_Build(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()
	db, err := OpenDB(mockDB)
	require.NoError(t, err)

	testCases := []struct {
		name    string
		builder QueryBuilder

		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "no from",
			builder: NewSelector[TestModel](db),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
		},
		{
			name:    "from",
			builder: (NewSelector[TestModel](db)).From("test_model"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model;",
				Args: nil,
			},
		},
		{
			name:    "empty from",
			builder: (NewSelector[TestModel](db)).From(""),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
		},
		{
			name:    "db from",
			builder: (NewSelector[TestModel](db)).From("`test_db`.`test_model`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_db`.`test_model`;",
				Args: nil,
			},
		},
		{
			name:    "where",
			builder: (NewSelector[TestModel](db)).Where(C("Age").Eq(18)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `age` = ?;",
				Args: []any{18},
			},
		},
		{
			name:    "not",
			builder: (NewSelector[TestModel](db)).Where(Not(C("Age").Eq(18))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE NOT (`age` = ?);",
				Args: []any{18},
			},
		},
		{
			name:    "and",
			builder: (NewSelector[TestModel](db)).Where(C("Age").Eq(18).And(C("FirstName").Eq("user1"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` = ?) AND (`first_name` = ?);",
				Args: []any{18, "user1"},
			},
		},
		{
			name:    "or",
			builder: (NewSelector[TestModel](db)).Where(C("Age").Eq(18).Or(C("FirstName").Eq("user1"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` = ?) OR (`first_name` = ?);",
				Args: []any{18, "user1"},
			},
		},
		{
			name:    "empty where",
			builder: (NewSelector[TestModel](db)).Where(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
		},
		{
			name: "complex select",
			builder: (NewSelector[TestModel](db)).
				Where(C("Age").Eq(18).Or(C("FirstName").Eq("user1"))).
				GroupBy(C("FirstName"), C("Age")).
				Having(Avg("FirstName").Eq("user1")).
				OrderBy(Asc("Id"), Desc("Age")).
				Limit(2).
				Offset(10),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` WHERE (`age` = ?) OR (`first_name` = ?) " +
					"GROUP BY `first_name`, `age` HAVING AVG(`first_name`) = ? ORDER BY `id` ASC, `age` DESC LIMIT ? OFFSET ?;",
				Args: []any{18, "user1", "user1", 2, 10},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.builder.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}

func TestSelector_Select(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()
	db, err := OpenDB(mockDB)
	require.NoError(t, err)

	testCases := []struct {
		name    string
		builder QueryBuilder

		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "where",
			builder: (NewSelector[TestModel](db)).Where(C("Age").Eq(18).And(C("FirstName").Eq("user1"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` = ?) AND (`first_name` = ?);",
				Args: []any{18, "user1"},
			},
		},
		{
			name:    "single columns",
			builder: (NewSelector[TestModel](db)).Select(C("LastName")).Where(C("Age").Eq(18).And(C("FirstName").Eq("user1"))),
			wantQuery: &Query{
				SQL:  "SELECT `last_name` FROM `test_model` WHERE (`age` = ?) AND (`first_name` = ?);",
				Args: []any{18, "user1"},
			},
		},
		{
			name:    "multiple columns",
			builder: (NewSelector[TestModel](db)).Select(C("FirstName"), C("LastName")).Where(C("Age").Eq(18).And(C("FirstName").Eq("user1"))),
			wantQuery: &Query{
				SQL:  "SELECT `first_name`, `last_name` FROM `test_model` WHERE (`age` = ?) AND (`first_name` = ?);",
				Args: []any{18, "user1"},
			},
		},
		{
			name:    "invalid field",
			builder: (NewSelector[TestModel](db)).Select(C("Field"), C("LastName")).Where(C("Age").Eq(18).And(C("FirstName").Eq("user1"))),
			wantErr: errs.NewErrUnknownField("Field"),
		},
		{
			name:    "single aggregate",
			builder: (NewSelector[TestModel](db)).Select(Avg("LastName")).Where(C("Age").Eq(18).And(C("FirstName").Eq("user1"))),
			wantQuery: &Query{
				SQL:  "SELECT AVG(`last_name`) FROM `test_model` WHERE (`age` = ?) AND (`first_name` = ?);",
				Args: []any{18, "user1"},
			},
		},
		{
			name:    "multiple aggregate",
			builder: (NewSelector[TestModel](db)).Select(Avg("FirstName"), Avg("LastName")).Where(C("Age").Eq(18).And(C("FirstName").Eq("user1"))),
			wantQuery: &Query{
				SQL:  "SELECT AVG(`first_name`), AVG(`last_name`) FROM `test_model` WHERE (`age` = ?) AND (`first_name` = ?);",
				Args: []any{18, "user1"},
			},
		},
		{
			name:    "raw",
			builder: (NewSelector[TestModel](db)).Select(Raw("COUNT(DISTINCT `first_name`)")).Where(C("Age").Eq(18).And(C("FirstName").Eq("user1"))),
			wantQuery: &Query{
				SQL:  "SELECT COUNT(DISTINCT `first_name`) FROM `test_model` WHERE (`age` = ?) AND (`first_name` = ?);",
				Args: []any{18, "user1"},
			},
		},
		{
			name:    "raw as predicate",
			builder: (NewSelector[TestModel](db)).Select(C("FirstName"), C("LastName")).Where(Raw("(`age` = ?) AND (`first_name` = ?)", 18, "user1").AsPredicate()),
			wantQuery: &Query{
				SQL:  "SELECT `first_name`, `last_name` FROM `test_model` WHERE (`age` = ?) AND (`first_name` = ?);",
				Args: []any{18, "user1"},
			},
		},
		{
			name:    "raw used in predicate",
			builder: (NewSelector[TestModel](db)).Select(C("FirstName"), C("LastName")).Where(C("Id").Eq(Raw("`age` + ?", 1).AsPredicate())),
			wantQuery: &Query{
				SQL:  "SELECT `first_name`, `last_name` FROM `test_model` WHERE `id` = (`age` + ?);",
				Args: []any{1},
			},
		},
		{
			name:    "column as",
			builder: (NewSelector[TestModel](db)).Select(C("FirstName").As("my_name"), Max("Age").As("max")).Where(C("Age").As("aaa").Eq(18).And(C("FirstName").Eq("user1"))),
			wantQuery: &Query{
				SQL:  "SELECT `first_name` AS `my_name`, MAX(`age`) AS `max` FROM `test_model` WHERE (`age` = ?) AND (`first_name` = ?);",
				Args: []any{18, "user1"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.builder.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}

func TestSelector_Get(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()
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
		s    *Selector[TestModel]

		wantErr error
		wantRes *TestModel
	}{
		{
			name:    "invalid query",
			s:       NewSelector[TestModel](db).Where(C("BB").Eq(18)),
			wantErr: errs.NewErrUnknownField("BB"),
		},
		{
			name:    "query error",
			s:       NewSelector[TestModel](db).Where(C("Id").Eq(1)),
			wantErr: errors.New("query error"),
		},
		{
			name:    "no rows",
			s:       NewSelector[TestModel](db).Where(C("Id").Eq(1)),
			wantErr: ErrNoRows,
		},
		{
			name: "data",
			s:    NewSelector[TestModel](db).Where(C("Id").Eq(1)),
			wantRes: &TestModel{
				Id:        1,
				Age:       18,
				FirstName: "Tom",
				LastName:  &sql.NullString{Valid: true, String: "Jerry"},
			},
		},
		//{
		//	name:    "scan error",
		//	s:       NewSelector[TestModel](db).Where(C("Id").Eq(1)),
		//	wantErr: ErrNoRows,
		//},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.s.Get(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

type TestModel struct {
	Id        int64           `orm:"id()"`
	Age       int8            `orm:"age(18)"`
	FirstName string          `orm:"name(first_name)"`
	LastName  *sql.NullString `orm:"name(last_name)"`
}

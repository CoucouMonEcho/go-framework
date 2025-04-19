package orm

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSelector_Build(t *testing.T) {
	testCases := []struct {
		name    string
		builder QueryBuilder

		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "no from",
			builder: NewSelector[TestModel](MustNewDB()),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
		},
		{
			name:    "from",
			builder: (NewSelector[TestModel](MustNewDB())).From("test_model"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model;",
				Args: nil,
			},
		},
		{
			name:    "empty from",
			builder: (NewSelector[TestModel](MustNewDB())).From(""),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
		},
		{
			name:    "db from",
			builder: (NewSelector[TestModel](MustNewDB())).From("`test_db`.`test_model`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_db`.`test_model`;",
				Args: nil,
			},
		},
		{
			name:    "where",
			builder: (NewSelector[TestModel](MustNewDB())).Where(C("Age").Eq(18)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `age` = ?;",
				Args: []any{18},
			},
		},
		{
			name:    "not",
			builder: (NewSelector[TestModel](MustNewDB())).Where(Not(C("Age").Eq(18))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE NOT (`age` = ?);",
				Args: []any{18},
			},
		},
		{
			name:    "and",
			builder: (NewSelector[TestModel](MustNewDB())).Where(C("Age").Eq(18).And(C("FirstName").Eq("user1"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` = ?) AND (`first_name` = ?);",
				Args: []any{18, "user1"},
			},
		},
		{
			name:    "or",
			builder: (NewSelector[TestModel](MustNewDB())).Where(C("Age").Eq(18).Or(C("FirstName").Eq("user1"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` = ?) OR (`first_name` = ?);",
				Args: []any{18, "user1"},
			},
		},
		{
			name:    "empty where",
			builder: (NewSelector[TestModel](MustNewDB())).Where(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model`;",
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

type TestModel struct {
	Id        int64           `orm:"-"`
	Age       int8            `orm:"age(18)"`
	FirstName string          `orm:"name(first_name)"`
	LastName  *sql.NullString `orm:"name(last_name)"`
}

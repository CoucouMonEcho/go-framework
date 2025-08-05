package querylog

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-framework/orm"
	"testing"
)

func TestNewMiddlewareBuilder(t *testing.T) {
	var querySQL string
	var queryArgs []any

	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		_ = mockDB.Close()
	}()
	db, err := orm.OpenDB(mockDB, orm.DBWithMiddlewares(NewMiddlewareBuilder().LogFunc(func(sql string, args []any) {
		querySQL = sql
		queryArgs = args
	}).Build()))
	require.NoError(t, err)

	_, _ = orm.NewSelector[TestModel](db).Where(orm.C("Id").Eq(123)).Get(context.Background())
	assert.Equal(t, "SELECT * FROM `test_model` WHERE `id` = ?;", querySQL)
	assert.Equal(t, []any{123}, queryArgs)

	orm.NewInserter[TestModel](db).Values(&TestModel{Id: 6}).Exec(context.Background())
	assert.Equal(t, "INSERT INTO `test_model`(`id`, `age`, `first_name`, `last_name`) VALUES (?, ?, ?, ?);", querySQL)
	assert.Equal(t, []any{int64(6), int8(0), "", (*sql.NullString)(nil)}, queryArgs)

}

type TestModel struct {
	Id        int64           `orm:"id()"`
	Age       int8            `orm:"age(18)"`
	FirstName string          `orm:"name(first_name)"`
	LastName  *sql.NullString `orm:"name(last_name)"`
}

package sql

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
	"time"
)

//FIXME go env -w CGO_ENABLED=0
// install TDM-GCC

func Test_Curd(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	require.NoError(t, err)
	err = db.Ping()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer db.Close()
	defer cancel()

	_, err = db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS test_model (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name TEXT,
    age INTEGER,
    last_name TEXT NOT NULL
)
`)
	require.NoError(t, err)
	res, err := db.ExecContext(ctx, "INSERT INTO test_model(`id`, `first_name`, `age`, `last_name`) VALUES (?, ?, ?, ?)",
		1, "Tom", 18, "Jerry")
	require.NoError(t, err)
	affected, err := res.RowsAffected()
	require.NoError(t, err)
	log.Println("affected:", affected)
	id, err := res.LastInsertId()
	require.NoError(t, err)
	log.Println("last insert id", id)

	//sql/convert.go/convertAssignRows

	row := db.QueryRowContext(ctx, "SELECT `id`, `first_name`, `age`, `last_name` FROM `test_model` WHERE `id` = ?",
		1)
	require.NoError(t, row.Err())
	tm := TestModel{}
	err = row.Scan(&tm.Id, &tm.FirstName, &tm.Age, &tm.LastName)
	require.NoError(t, err)

	//row := db.QueryRowContext(ctx, "SELECT `id`, `first_name`, `age`, `last_name` FROM `test_model` WHERE `id` = ?",
	//	2)
	//require.NoError(t, row.Err())
	//tm := TestModel{}
	//err = row.Scan(&tm.Id, &tm.FirstName, &tm.Age, &tm.LastName)
	//require.Error(t, sql.ErrNoRows, err)

	//rows, err := db.QueryContext(ctx, "SELECT `id`, `first_name`, `age`, `last_name` FROM `test_model` WHERE `id` = ?",
	//	1)
	//require.NoError(t, err)
	//for rows.Next() {
	//	tm := TestModel{}
	//	err = rows.Scan(&tm.Id, &tm.FirstName, &tm.Age, &tm.LastName)
	//	require.NoError(t, err)
	//}

}

func Test_Transaction(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	require.NoError(t, err)
	err = db.Ping()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer db.Close()
	defer cancel()

	_, err = db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS test_model (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name TEXT,
    age INTEGER,
    last_name TEXT NOT NULL
)
`)
	require.NoError(t, err)

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	require.NoError(t, err)

	res, err := tx.ExecContext(ctx, "INSERT INTO test_model(`id`, `first_name`, `age`, `last_name`) VALUES (?, ?, ?, ?)",
		1, "Tom", 18, "Jerry")
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	// other tx...

	err = tx.Commit()

	affected, err := res.RowsAffected()
	require.NoError(t, err)
	log.Println("affected:", affected)
	id, err := res.LastInsertId()
	require.NoError(t, err)
	log.Println("last insert id", id)

}

func TestPrepareStatement(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	require.NoError(t, err)
	err = db.Ping()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer db.Close()
	defer cancel()

	_, err = db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS test_model (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name TEXT,
    age INTEGER,
    last_name TEXT NOT NULL
)
`)
	require.NoError(t, err)
	res, err := db.ExecContext(ctx, "INSERT INTO test_model(`id`, `first_name`, `age`, `last_name`) VALUES (?, ?, ?, ?)",
		1, "Tom", 18, "Jerry")
	require.NoError(t, err)
	affected, err := res.RowsAffected()
	require.NoError(t, err)
	log.Println("affected:", affected)
	id, err := res.LastInsertId()
	require.NoError(t, err)
	log.Println("last insert id", id)

	stmt, err := db.PrepareContext(ctx, "SELECT `id`, `first_name`, `age`, `last_name` FROM `test_model` WHERE `id` = ?")
	//In order to guarantee no more than 10,000, it needs to be closed
	//In a concurrent scenario, the same STMT may be used concurrently
	//There is also no guarantee of DB isolation
	//
	//1. Reference Counting: When a threshold is exceeded, the least used STMT is closed
	//2. In "IN" query, STMT is not used or a one-time STMT is used
	defer stmt.Close()
	require.NoError(t, err)
	rows, err := stmt.Query(1)
	require.NoError(t, err)
	for rows.Next() {
		tm := TestModel{}
		err = rows.Scan(&tm.Id, &tm.FirstName, &tm.Age, &tm.LastName)
		require.NoError(t, err)
	}

	// stmt more than 10,000 panic(every in query)

	//stmt, err = db.PrepareContext(ctx,
	//	"SELECT `id`, `first_name`, `age`, `last_name` FROM `test_model` WHERE `id` IN (?, ?)")
	//stmt, err = db.PrepareContext(ctx,
	//	"SELECT `id`, `first_name`, `age`, `last_name` FROM `test_model` WHERE `id` IN (?, ?, ?)")

}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}

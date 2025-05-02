package testdata

import (
	sql "database/sql"
)

type User struct {
	Name     string
	Age      *int
	NickName *sql.NullString
	Picture  []byte
}

type UserDetail struct {
	Address string
}

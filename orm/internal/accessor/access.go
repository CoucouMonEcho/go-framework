package accessor

import (
	"code-practise/orm/model"
	"database/sql"
)

type Access interface {
	SetColumns(rows *sql.Rows) error
}

type Creator func(model *model.Model, val any) Access

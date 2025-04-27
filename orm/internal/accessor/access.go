package accessor

import (
	"code-practise/orm/model"
	"database/sql"
)

var (
	_ Access = &unsafeAccess{}
	_ Access = &reflectAccess{}
)

type Access interface {
	Field(name string) (any, error)
	SetColumns(rows *sql.Rows) error
}

var (
	_ Creator = NewUnsafeAccess
	_ Creator = NewReflectAccess
)

type Creator func(model *model.Model, val any) Access

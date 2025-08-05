package accessor

import (
	"database/sql"
	"github.com/CoucouMonEcho/go-framework/orm/model"
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

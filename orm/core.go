package orm

import (
	"code-practise/orm/internal/accessor"
	"code-practise/orm/model"
)

type core struct {
	model   *model.Model
	dialect Dialect
	creator accessor.Creator
	r       model.Registry
}

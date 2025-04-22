package accessor

import (
	"code-practise/orm/internal/errs"
	"code-practise/orm/model"
	"database/sql"
	"reflect"
)

var _ Access = &reflectAccess{}

type reflectAccess struct {
	model *model.Model
	// val is pointer of T
	val any
}

var _ Creator = NewReflectAccess

func NewReflectAccess(model *model.Model, val any) Access {
	return reflectAccess{
		model: model,
		val:   val,
	}
}

func (r reflectAccess) SetColumns(rows *sql.Rows) error {
	// select column
	cs, err := rows.Columns()
	if err != nil {
		return err
	}
	// scan
	vals := make([]any, 0, len(cs))
	valElems := make([]reflect.Value, 0, len(cs))
	for _, c := range cs {
		fd, ok := r.model.ColumnMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		// return pointer
		val := reflect.New(fd.Type)

		vals = append(vals, val.Interface())
		valElems = append(valElems, val.Elem())
	}
	err = rows.Scan(vals...)
	if err != nil {
		return err
	}
	// struct
	tpValue := reflect.ValueOf(r.val)
	for i, c := range cs {
		fd, ok := r.model.ColumnMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		tpValue.Elem().FieldByName(fd.GoName).Set(valElems[i])
	}
	return err
}

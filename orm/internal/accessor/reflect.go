package accessor

import (
	"database/sql"
	"go-framework/orm/internal/errs"
	"go-framework/orm/model"
	"reflect"
)

type reflectAccess struct {
	model *model.Model
	// val is pointer of T
	//val any
	val reflect.Value
}

func NewReflectAccess(model *model.Model, val any) Access {
	return reflectAccess{
		model: model,
		val:   reflect.ValueOf(val).Elem(),
	}
}

func (r reflectAccess) Field(name string) (any, error) {
	//_, ok := r.val.Type().FieldByName(name)
	//if !ok {
	//	return nil, errs.NewErrUnknownField(name)
	//}
	res := r.val.FieldByName(name)
	if res == (reflect.Value{}) {
		return nil, errs.NewErrUnknownField(name)
	}
	return res.Interface(), nil
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
	for i, c := range cs {
		fd, ok := r.model.ColumnMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		r.val.FieldByName(fd.GoName).Set(valElems[i])
	}
	return err
}

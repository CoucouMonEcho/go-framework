package accessor

import (
	"database/sql"
	"github.com/CoucouMonEcho/go-framework/orm/internal/errs"
	"github.com/CoucouMonEcho/go-framework/orm/model"
	"reflect"
	"unsafe"
)

type unsafeAccess struct {
	model *model.Model
	// reference address
	address unsafe.Pointer
}

func NewUnsafeAccess(model *model.Model, val any) Access {
	address := reflect.ValueOf(val).UnsafePointer()
	return unsafeAccess{
		model:   model,
		address: address,
	}
}

func (u unsafeAccess) Field(name string) (any, error) {
	fd, ok := u.model.FieldMap[name]
	if !ok {
		return nil, errs.NewErrUnknownField(name)
	}
	// return pointer
	// address + offset
	return reflect.NewAt(fd.Type, unsafe.Pointer(uintptr(u.address)+fd.Offset)).Elem().Interface(), nil
}

func (u unsafeAccess) SetColumns(rows *sql.Rows) error {
	// select column
	cs, err := rows.Columns()
	if err != nil {
		return err
	}
	// scan
	var vals []any
	for _, c := range cs {
		fd, ok := u.model.ColumnMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		// return pointer
		// address + offset
		val := reflect.NewAt(fd.Type, unsafe.Pointer(uintptr(u.address)+fd.Offset))
		//val := reflect.New(fd.Type)

		vals = append(vals, val.Interface())
	}
	err = rows.Scan(vals...)
	return err
}

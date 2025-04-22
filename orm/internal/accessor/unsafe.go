package accessor

import (
	"code-practise/orm/internal/errs"
	"code-practise/orm/model"
	"database/sql"
	"reflect"
	"unsafe"
)

var _ Access = &unsafeAccess{}

type unsafeAccess struct {
	model *model.Model
	// val is pointer of T
	val any
}

var _ Creator = NewUnsafeAccess

func NewUnsafeAccess(model *model.Model, val any) Access {
	return unsafeAccess{
		model: model,
		val:   val,
	}
}

func (u unsafeAccess) SetColumns(rows *sql.Rows) error {
	// select column
	cs, err := rows.Columns()
	if err != nil {
		return err
	}
	// scan
	var vals []any
	address := reflect.ValueOf(u.val).UnsafePointer()
	for _, c := range cs {
		fd, ok := u.model.ColumnMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		// return pointer
		// address + offset
		val := reflect.NewAt(fd.Type, unsafe.Pointer(uintptr(address)+fd.Offset))
		//val := reflect.New(fd.Type)

		vals = append(vals, val.Interface())
	}
	err = rows.Scan(vals...)
	return err
}

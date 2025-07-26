package unsafe

import (
	"errors"
	"reflect"
	"unsafe"
)

type Accessor struct {
	// address unsafe.Pointer go will help maintain,
	// and gc still points to the changed address before and after
	address unsafe.Pointer

	fields map[string]FieldMeta
}

func NewUnsafeAccessor(entity any) (*Accessor, error) {
	if entity == nil {
		return nil, errors.New("invalid entity")
	}
	typ := reflect.TypeOf(entity)
	if typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct {
		return nil, errors.New("invalid entity")
	}
	fields := make(map[string]FieldMeta, typ.Elem().NumField())
	elemTyp := typ.Elem()
	for i := 0; i < elemTyp.NumField(); i++ {
		fd := elemTyp.Field(i)
		fields[fd.Name] = FieldMeta{
			offset: fd.Offset,
			typ:    fd.Type,
		}
	}
	// UnsafeAddr will change after gc, UnsafePointer will not
	return &Accessor{
		address: reflect.ValueOf(entity).UnsafePointer(),
		fields:  fields,
	}, nil
}

func (a *Accessor) Field(field string) (any, error) {
	fd, ok := a.fields[field]
	if !ok {
		return nil, errors.New("invalid field")
	}
	fdAddr := unsafe.Pointer(uintptr(a.address) + fd.offset)

	// if field type is known
	//return *(*int)(fdAddr), nil

	// type is unknown
	return reflect.NewAt(fd.typ, fdAddr).Elem().Interface(), nil
}

func (a *Accessor) SetField(field string, val any) error {
	fd, ok := a.fields[field]
	if !ok {
		return errors.New("invalid field")
	}
	fdAddr := unsafe.Pointer(uintptr(a.address) + fd.offset)

	// if field type is known
	//*(*int)(fdAddr) = val.(int)

	// type is unknown
	reflect.NewAt(fd.typ, fdAddr).Elem().Set(reflect.ValueOf(val))

	return nil
}

type FieldMeta struct {
	// in fact, offset is only a number
	// uintptr is usually only used when performing address operations
	offset uintptr

	typ reflect.Type
}

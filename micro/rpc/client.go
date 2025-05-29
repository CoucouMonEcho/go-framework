package rpc

import (
	"context"
	"errors"
	"reflect"
)

// InitClientProxy assign values to function type fields
func InitClientProxy(service Service) error {
	return setFuncField(service, nil)
}

func setFuncField(service Service, p Proxy) error {
	if service == nil {
		return errors.New("rpc: service is nil")
	}
	val := reflect.ValueOf(service)
	typ := val.Type()
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return errors.New("rpc: service must be a pointer to struct")
	}

	val = val.Elem()
	typ = typ.Elem()

	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		fieldTyp := typ.Field(i)
		fieldVal := val.Field(i)

		if fieldVal.CanSet() {
			fn := reflect.MakeFunc(fieldTyp.Type, func(args []reflect.Value) []reflect.Value {
				ctx := args[0].Interface().(context.Context)
				req := &Request{
					ServiceName: service.Name(),
					MethodName:  fieldTyp.Name,
					Arg:         args[1].Interface(),
				}
				retVal := reflect.New(fieldTyp.Type.Out(0)).Elem()
				_, err := p.Invoke(ctx, req)
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}
				return []reflect.Value{retVal, reflect.Zero(reflect.TypeOf(new(error)).Elem())}
			})
			fieldVal.Set(fn)
		}
	}
	return nil
}

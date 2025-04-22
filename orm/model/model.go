package model

import (
	"code-practise/orm/internal/errs"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

const (
	tagKeyOrm    = "orm"
	tagKeyColumn = "column"
)

type TableName interface {
	TableName() string
}

type Registry interface {
	Get(val any) (*Model, error)
	Register(val any, opts ...Option) (*Model, error)
}

// Model is public but not recommended to use,
// it is in order to solve circular dependencies
type Model struct {
	TableName string
	FieldMap  map[string]*Field
	ColumnMap map[string]*Field
}

type Option func(*Model) error

func WithTableName(tableName string) Option {
	return func(m *Model) error {
		if tableName == "" {
			return errs.ErrIllegalTableName
		}
		m.TableName = tableName
		return nil
	}
}

func WithColumnName(field string, colName string) Option {
	return func(m *Model) error {
		fd, ok := m.FieldMap[field]
		if !ok {
			return errs.NewErrUnknownField(field)
		}
		if colName == "" {
			return errs.ErrIllegalColumnName
		}
		delete(m.ColumnMap, fd.ColName)
		fd.ColName = colName
		m.ColumnMap[colName] = fd
		return nil
	}
}

type Field struct {
	GoName  string
	ColName string
	Type    reflect.Type
	Offset  uintptr
}

type registry struct {
	//lock   sync.RWMutex
	//models map[reflect.Type]*Model

	models sync.Map
}

func NewRegistry() Registry {
	return &registry{}
}

func (r *registry) Get(entity any) (*Model, error) {
	pointerType := reflect.TypeOf(entity)
	m, ok := r.models.Load(pointerType)
	if ok {
		return m.(*Model), nil
	}
	m, err := r.Register(entity)
	if err != nil {
		return nil, err
	}

	return m.(*Model), nil
}

//func (r *registry) Get(entity any) (*Model, error) {
//	entityType := reflect.TypeOf(entity)
//	r.lock.RLock()
//	m, ok := r.models[entityType]
//	r.lock.RUnlock()
//	if ok {
//		return m, nil
//	}
//
//	r.lock.Lock()
//	defer r.lock.Unlock()
//
//	r.lock.RLock()
//	m, ok = r.models[entityType]
//	r.lock.RUnlock()
//	if ok {
//		return m, nil
//	}
//
//	var err error
//	m, err = r.Register(entity)
//	if err != nil {
//		return nil, err
//	}
//	r.models[entityType] = m
//
//	return m, nil
//}

func (r *registry) Register(entity any, opts ...Option) (*Model, error) {
	pointerType := reflect.TypeOf(entity)
	if pointerType.Kind() != reflect.Ptr || pointerType.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrModelNotPointer
	}
	entityType := pointerType.Elem()
	numField := entityType.NumField()
	fieldMap := make(map[string]*Field, numField)
	columnMap := make(map[string]*Field, numField)
	for i := 0; i < numField; i++ {
		fd := entityType.Field(i)
		pairs, err := r.parseTag(fd.Tag)
		if err != nil {
			return nil, err
		}
		colName := pairs[tagKeyColumn]
		if colName == "" {
			colName = toUnderscore(fd.Name)
		}
		fdMeta := &Field{
			GoName:  fd.Name,
			ColName: colName,
			Type:    fd.Type,
			Offset:  fd.Offset,
		}
		fieldMap[fd.Name] = fdMeta
		columnMap[colName] = fdMeta
	}

	var tableName string
	if tbl, ok := entity.(TableName); ok {
		tableName = tbl.TableName()
	}
	if tableName == "" {
		tableName = toUnderscore(entityType.Name())
	}

	res := &Model{
		TableName: tableName,
		FieldMap:  fieldMap,
		ColumnMap: columnMap,
	}

	for _, opt := range opts {
		err := opt(res)
		if err != nil {
			return nil, err
		}
	}
	r.models.Store(pointerType, res)

	return res, nil
}

func (r *registry) parseTag(tag reflect.StructTag) (map[string]string, error) {
	ormTag, ok := tag.Lookup(tagKeyOrm)
	if !ok {
		return map[string]string{}, nil
	}
	pairs := strings.Split(ormTag, ",")
	res := map[string]string{}
	for _, pair := range pairs {
		segs := strings.Split(pair, "(")
		if len(segs) != 2 {
			return nil, errs.NewErrInvalidTagContent(pair)
		}
		key := segs[0]
		val := strings.Split(segs[1], ")")[0]
		res[key] = val
	}
	return res, nil
}

func toUnderscore(name string) string {
	return strings.ToLower(regexp.MustCompile("([a-z0-9])([A-Z])").ReplaceAllString(name, "${1}_${2}"))
}

package orm

import (
	"code-practise/orm/internal/errs"
	"reflect"
	"regexp"
	"strings"
)

type model struct {
	tableName string
	fields    map[string]*field
}

type field struct {
	colName string
}

type registry struct {
	models map[reflect.Type]*model
}

func newRegistry() *registry {
	return &registry{
		models: make(map[reflect.Type]*model, 64),
	}
}

func (r *registry) get(entity any) (*model, error) {
	entityType := reflect.TypeOf(entity)
	m, ok := r.models[entityType]
	if !ok {
		var err error
		m, err = r.parseModel(entity)
		if err != nil {
			return nil, err
		}
		r.models[entityType] = m
	}
	return m, nil
}

func (r *registry) parseModel(entity any) (*model, error) {
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() != reflect.Ptr || entityType.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrModelNotPointer
	}
	entityType = entityType.Elem()
	numField := entityType.NumField()
	fields := make(map[string]*field, numField)
	for i := range numField {
		structField := entityType.Field(i)
		fields[structField.Name] = &field{
			colName: toUnderscore(structField.Name),
		}
	}
	return &model{
		tableName: toUnderscore(entityType.Name()),
		fields:    fields,
	}, nil
}

func toUnderscore(name string) string {
	return strings.ToLower(regexp.MustCompile("([a-z0-9])([A-Z])").ReplaceAllString(name, "${1}_${2}"))
}

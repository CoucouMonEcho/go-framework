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

func parseModel(entity any) (*model, error) {
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

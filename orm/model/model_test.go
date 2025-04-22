package model

import (
	"code-practise/orm/internal/errs"
	"database/sql"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_parseModel(t *testing.T) {
	fields := []*Field{
		{
			GoName:  "Id",
			ColName: "id",
			Type:    reflect.TypeOf(int64(0)),
		},
		{
			GoName:  "Age",
			ColName: "age",
			Type:    reflect.TypeOf(int8(0)),
			Offset:  8,
		},
		{
			GoName:  "FirstName",
			ColName: "first_name",
			Type:    reflect.TypeOf(""),
			Offset:  16,
		},
		{
			GoName:  "LastName",
			ColName: "last_name",
			Type:    reflect.TypeOf(&sql.NullString{}),
			Offset:  32,
		},
	}
	tests := []struct {
		name   string
		entity any

		fields []*Field

		want    *Model
		wantErr error
	}{
		{
			name:   "test Model",
			fields: fields,
			entity: &TestModel{},
			want: &Model{
				TableName: "test_model",
			},
		},
		{
			name:    "struct",
			entity:  TestModel{},
			wantErr: errs.ErrModelNotPointer,
		},
	}
	r := &registry{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.Register(tt.entity)
			assert.Equal(t, tt.wantErr, err)
			if err != nil {
				return
			}

			fieldMap := make(map[string]*Field)
			columnMap := make(map[string]*Field)
			for _, field := range tt.fields {
				fieldMap[field.GoName] = field
				columnMap[field.ColName] = field
			}
			tt.want.FieldMap = fieldMap
			tt.want.ColumnMap = columnMap

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_registryGet(t *testing.T) {
	tests := []struct {
		name   string
		entity any
		opts   []Option

		fields []*Field

		want    *Model
		wantErr error

		cacheSize int
	}{
		{
			name:   "test Model",
			entity: &TestModel{},
			fields: []*Field{
				{
					GoName:  "Id",
					ColName: "new_column",
					Type:    reflect.TypeOf(int64(0)),
				},
				{
					GoName:  "Age",
					ColName: "age",
					Type:    reflect.TypeOf(int8(0)),
					Offset:  8,
				},
				{
					GoName:  "FirstName",
					ColName: "first_name",
					Type:    reflect.TypeOf(string("")),
					Offset:  16,
				},
				{
					GoName:  "LastName",
					ColName: "last_name",
					Type:    reflect.TypeOf(&sql.NullString{}),
					Offset:  32,
				},
			},
			opts: []Option{
				WithColumnName("Id", "new_column"),
				WithTableName("new_table"),
			},
			want: &Model{
				TableName: "new_table",
			},
			cacheSize: 1,
		},
		{
			name:    "struct",
			entity:  TestModel{},
			wantErr: errs.ErrModelNotPointer,
		},
		{
			name: "tag",
			entity: func() any {
				type TagTable struct {
					Name string `orm:"column(name_column)"`
				}
				return &TagTable{}
			}(),
			fields: []*Field{
				{
					GoName:  "Name",
					ColName: "name_column",
					Type:    reflect.TypeOf(""),
				},
			},
			want: &Model{
				TableName: "tag_table",
			},
		},
		{
			name: "empty tag",
			entity: func() any {
				type TagTable struct {
					Name string `orm:"column()"`
				}
				return &TagTable{}
			}(),
			fields: []*Field{
				{
					GoName:  "Name",
					ColName: "name",
					Type:    reflect.TypeOf(""),
				},
			},
			want: &Model{
				TableName: "tag_table",
				FieldMap: map[string]*Field{
					"Name": {
						ColName: "name",
					},
				},
				ColumnMap: map[string]*Field{
					"name": {
						ColName: "name",
					},
				},
			},
		},
		{
			name: "err",
			entity: func() any {
				type TagTable struct {
					Name string `orm:"column"`
				}
				return &TagTable{}
			}(),
			wantErr: errs.NewErrInvalidTagContent("column"),
		},
		{
			name:   "table name",
			entity: &CustomTableName{},
			fields: []*Field{
				{
					GoName:  "FirstName",
					ColName: "first_name_c",
					Type:    reflect.TypeOf(""),
				},
			},
			want: &Model{
				TableName: "custom_table_name_t",
			},
		},
		{
			name:   "table name",
			entity: &CustomTableNamePtr{},
			fields: []*Field{
				{
					GoName:  "FirstName",
					ColName: "first_name_c",
					Type:    reflect.TypeOf(""),
				},
			},
			want: &Model{
				TableName: "custom_table_name_p",
				FieldMap: map[string]*Field{
					"FirstName": {
						ColName: "first_name_c",
					},
				},
			},
		},
	}
	r := NewRegistry()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//got, err := r.Get(tt.entity)
			got, err := r.Register(tt.entity, tt.opts...)
			assert.Equal(t, tt.wantErr, err)
			if err != nil {
				return
			}

			fieldMap := make(map[string]*Field)
			columnMap := make(map[string]*Field)
			for _, field := range tt.fields {
				fieldMap[field.GoName] = field
				columnMap[field.ColName] = field
			}
			tt.want.FieldMap = fieldMap
			tt.want.ColumnMap = columnMap

			assert.Equal(t, tt.want, got)
			typ := reflect.TypeOf(tt.entity)
			cache, ok := r.(*registry).models.Load(typ)
			assert.True(t, ok)
			assert.Equal(t, tt.want, cache)
		})
	}
}

type CustomTableName struct {
	FirstName string `orm:"column(first_name_c)"`
}

func (c CustomTableName) TableName() string {
	return "custom_table_name_t"
}

type CustomTableNamePtr struct {
	FirstName string `orm:"column(first_name_c)"`
}

func (c *CustomTableNamePtr) TableName() string {
	return "custom_table_name_p"
}

type TestModel struct {
	Id        int64           `orm:"id()"`
	Age       int8            `orm:"age(18)"`
	FirstName string          `orm:"name(first_name)"`
	LastName  *sql.NullString `orm:"name(last_name)"`
}

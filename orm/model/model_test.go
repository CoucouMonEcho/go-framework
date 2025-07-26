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
	testCases := []struct {
		name   string
		entity any

		want    *Model
		wantErr error
	}{
		{
			name:   "test Model",
			entity: &TestModel{},
			want: &Model{
				TableName: "test_model",
				Fields:    fields,
			},
		},
		{
			name:    "struct",
			entity:  TestModel{},
			wantErr: errs.ErrModelNotPointer,
		},
	}
	r := &registry{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := r.Register(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}

			fieldMap := make(map[string]*Field)
			columnMap := make(map[string]*Field)
			for _, field := range tc.want.Fields {
				fieldMap[field.GoName] = field
				columnMap[field.ColName] = field
			}
			tc.want.FieldMap = fieldMap
			tc.want.ColumnMap = columnMap

			assert.Equal(t, tc.want, got)
		})
	}
}

func Test_registryGet(t *testing.T) {
	testCases := []struct {
		name   string
		entity any
		opts   []Option

		want    *Model
		wantErr error

		cacheSize int
	}{
		{
			name:   "test Model",
			entity: &TestModel{},
			opts: []Option{
				WithColumnName("Id", "new_column"),
				WithTableName("new_table"),
			},
			want: &Model{
				TableName: "new_table",
				Fields: []*Field{
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
						Type:    reflect.TypeOf(""),
						Offset:  16,
					},
					{
						GoName:  "LastName",
						ColName: "last_name",
						Type:    reflect.TypeOf(&sql.NullString{}),
						Offset:  32,
					},
				},
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
			want: &Model{
				TableName: "tag_table",
				Fields: []*Field{
					{
						GoName:  "Name",
						ColName: "name_column",
						Type:    reflect.TypeOf(""),
					},
				},
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
				Fields: []*Field{
					{
						GoName:  "Name",
						ColName: "name",
						Type:    reflect.TypeOf(""),
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
			want: &Model{
				TableName: "custom_table_name_t",
				Fields: []*Field{
					{
						GoName:  "FirstName",
						ColName: "first_name_c",
						Type:    reflect.TypeOf(""),
					},
				},
			},
		},
		{
			name:   "table name",
			entity: &CustomTableNamePtr{},
			want: &Model{
				TableName: "custom_table_name_p",
				FieldMap: map[string]*Field{
					"FirstName": {
						ColName: "first_name_c",
					},
				},
				Fields: []*Field{
					{
						GoName:  "FirstName",
						ColName: "first_name_c",
						Type:    reflect.TypeOf(""),
					},
				},
			},
		},
	}
	r := NewRegistry()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			//got, err := r.Get(tc.entity)
			got, err := r.Register(tc.entity, tc.opts...)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}

			fieldMap := make(map[string]*Field)
			columnMap := make(map[string]*Field)
			for _, field := range tc.want.Fields {
				fieldMap[field.GoName] = field
				columnMap[field.ColName] = field
			}
			tc.want.FieldMap = fieldMap
			tc.want.ColumnMap = columnMap

			assert.Equal(t, tc.want, got)
			typ := reflect.TypeOf(tc.entity)
			cache, ok := r.(*registry).models.Load(typ)
			assert.True(t, ok)
			assert.Equal(t, tc.want, cache)
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

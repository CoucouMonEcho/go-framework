package orm

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
			goName:  "Id",
			colName: "id",
			typ:     reflect.TypeOf(int64(0)),
		},
		{
			goName:  "Age",
			colName: "age",
			typ:     reflect.TypeOf(int8(0)),
		},
		{
			goName:  "FirstName",
			colName: "first_name",
			typ:     reflect.TypeOf(""),
		},
		{
			goName:  "LastName",
			colName: "last_name",
			typ:     reflect.TypeOf(&sql.NullString{}),
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
				tableName: "test_model",
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
				fieldMap[field.goName] = field
				columnMap[field.colName] = field
			}
			tt.want.fieldMap = fieldMap
			tt.want.columnMap = columnMap

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_registryGet(t *testing.T) {
	tests := []struct {
		name   string
		entity any
		opts   []ModelOption

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
					goName:  "Id",
					colName: "new_column",
					typ:     reflect.TypeOf(int64(0)),
				},
				{
					goName:  "Age",
					colName: "age",
					typ:     reflect.TypeOf(int8(0)),
				},
				{
					goName:  "FirstName",
					colName: "first_name",
					typ:     reflect.TypeOf(string("")),
				},
				{
					goName:  "LastName",
					colName: "last_name",
					typ:     reflect.TypeOf(&sql.NullString{}),
				},
			},
			opts: []ModelOption{
				ModelWithColumnName("Id", "new_column"),
				ModelWithTableName("new_table"),
			},
			want: &Model{
				tableName: "new_table",
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
					goName:  "Name",
					colName: "name_column",
					typ:     reflect.TypeOf(""),
				},
			},
			want: &Model{
				tableName: "tag_table",
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
					goName:  "Name",
					colName: "name",
					typ:     reflect.TypeOf(""),
				},
			},
			want: &Model{
				tableName: "tag_table",
				fieldMap: map[string]*Field{
					"Name": {
						colName: "name",
					},
				},
				columnMap: map[string]*Field{
					"name": {
						colName: "name",
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
					goName:  "FirstName",
					colName: "first_name_c",
					typ:     reflect.TypeOf(""),
				},
			},
			want: &Model{
				tableName: "custom_table_name_t",
			},
		},
		{
			name:   "table name",
			entity: &CustomTableNamePtr{},
			fields: []*Field{
				{
					goName:  "FirstName",
					colName: "first_name_c",
					typ:     reflect.TypeOf(""),
				},
			},
			want: &Model{
				tableName: "custom_table_name_p",
				fieldMap: map[string]*Field{
					"FirstName": {
						colName: "first_name_c",
					},
				},
			},
		},
	}
	r := newRegistry()
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
				fieldMap[field.goName] = field
				columnMap[field.colName] = field
			}
			tt.want.fieldMap = fieldMap
			tt.want.columnMap = columnMap

			assert.Equal(t, tt.want, got)
			typ := reflect.TypeOf(tt.entity)
			cache, ok := r.models.Load(typ)
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

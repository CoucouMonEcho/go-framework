package orm

import (
	"code-practise/orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_parseModel(t *testing.T) {
	tests := []struct {
		name    string
		entity  any
		want    *model
		wantErr error
	}{
		{
			name:   "test model",
			entity: &TestModel{},
			want: &model{
				tableName: "test_model",
				fields: map[string]*field{
					"Id": {
						colName: "id",
					},
					"Age": {
						colName: "age",
					},
					"FirstName": {
						colName: "first_name",
					},
					"LastName": {
						colName: "last_name",
					},
				},
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
			got, err := r.parseModel(tt.entity)
			assert.Equal(t, tt.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_registryGet(t *testing.T) {
	tests := []struct {
		name    string
		entity  any
		want    *model
		wantErr error

		cacheSize int
	}{
		{
			name:   "test model",
			entity: &TestModel{},
			want: &model{
				tableName: "test_model",
				fields: map[string]*field{
					"Id": {
						colName: "id",
					},
					"Age": {
						colName: "age",
					},
					"FirstName": {
						colName: "first_name",
					},
					"LastName": {
						colName: "last_name",
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
	}
	r := newRegistry()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.get(tt.entity)
			assert.Equal(t, tt.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.cacheSize, len(r.models))
			typ := reflect.TypeOf(tt.entity)
			m, ok := r.models[typ]
			assert.True(t, ok)
			assert.Equal(t, tt.want, m)
		})
	}
}

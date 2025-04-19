package orm

import (
	"code-practise/orm/internal/errs"
	"github.com/stretchr/testify/assert"
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseModel(tt.entity)
			assert.Equal(t, tt.wantErr, err)
			if err != nil {
				return
			}
			assert.Equalf(t, tt.want, got, "parseModel(%v)", tt.entity)
		})
	}
}

package accessor

import (
	"database/sql"
	"database/sql/driver"
	"github.com/CoucouMonEcho/go-framework/orm/model"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewReflectAccess(t *testing.T) {

}

func TestNewUnsafeAccess(t *testing.T) {

}

func BenchmarkAccessSetColumns(b *testing.B) {

	fn := func(b *testing.B, creator Creator) {

		mockDB, mock, err := sqlmock.New()
		require.NoError(b, err)
		defer func() {
			_ = mockDB.Close()
		}()

		mockRows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
		row := []driver.Value{"1", "Tom", "18", "Jerry"}
		for i := 0; i < b.N; i++ {
			mockRows.AddRow(row...)
		}
		mock.ExpectQuery("SELECT XX").WillReturnRows(mockRows)
		rows, err := mockDB.Query("SELECT XX")
		require.NoError(b, err)

		r := model.NewRegistry()
		m, err := r.Get(&TestModel{})
		require.NoError(b, err)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			rows.Next()
			acc := creator(m, &TestModel{})
			_ = acc.SetColumns(rows)
		}

	}

	//TODO go test -bench=BenchmarkAccessSetColumns -benchtime=10000x -benchmem

	b.Run("reflect", func(b *testing.B) {
		fn(b, NewReflectAccess)
	})

	b.Run("unsafe", func(b *testing.B) {
		fn(b, NewUnsafeAccess)
	})

}

func Test_reflectAccess_SetColumns(t *testing.T) {
	testSetColumns(t, NewReflectAccess)
}

func Test_unsafeAccess_SetColumns(t *testing.T) {
	testSetColumns(t, NewUnsafeAccess)
}

func testSetColumns(t *testing.T, creator Creator) {
	testCases := []struct {
		name   string
		entity any
		rows   func() *sqlmock.Rows

		wantErr    error
		wantEntity any
	}{
		{
			name:   "success",
			entity: &TestModel{},
			rows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
				rows.AddRow("1", "Tom", "18", "Jerry")
				return rows
			},
			wantEntity: &TestModel{
				Id:        1,
				Age:       18,
				FirstName: "Tom",
				LastName: &sql.NullString{
					String: "Jerry",
					Valid:  true,
				},
			},
		},
		{
			name:   "partial column",
			entity: &TestModel{},
			rows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "first_name"})
				rows.AddRow("1", "Tom")
				return rows
			},
			wantEntity: &TestModel{
				Id: 1,
				//Age:       18,
				FirstName: "Tom",
				//LastName: &sql.NullString{
				//	String: "Jerry",
				//	Valid:  true,
				//},
			},
		},
	}

	r := model.NewRegistry()

	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		_ = mockDB.Close()
	}()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			mock.ExpectQuery("SELECT XX").WillReturnRows(tc.rows())
			rows, err := mockDB.Query("SELECT XX")
			require.NoError(t, err)

			rows.Next()

			m, err := r.Get(tc.entity)
			require.NoError(t, err)
			access := creator(m, tc.entity)
			err = access.SetColumns(rows)
			assert.Equal(t, err, tc.wantErr)
			if err != nil {
				return
			}
			assert.Equal(t, tc.entity, tc.wantEntity)

		})
	}
}

type TestModel struct {
	Id        int64           `orm:"id()"`
	Age       int8            `orm:"age(18)"`
	FirstName string          `orm:"name(first_name)"`
	LastName  *sql.NullString `orm:"name(last_name)"`
}

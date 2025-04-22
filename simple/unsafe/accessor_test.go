package unsafe

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUnsafeAccessor_Field(t *testing.T) {
	type user struct {
		Name string
		Age  int
	}

	u := &user{Name: "user", Age: 18}
	accessor, err := NewUnsafeAccessor(u)
	require.NoError(t, err)
	val, err := accessor.Field("Age")
	require.NoError(t, err)
	assert.Equal(t, val, 18)

	err = accessor.SetField("Age", 20)
	require.NoError(t, err)
	assert.Equal(t, u.Age, 20)

}

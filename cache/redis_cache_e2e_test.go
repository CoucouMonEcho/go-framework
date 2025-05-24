//go:build e2e

package cache

import (
	"testing"
)

func TestRedisCache_e2e_Get(t *testing.T) {
	//rdb := redis.NewClient(&redis.Options{
	//	Addr: "localhost:6379",
	//})
	//c := NewRedisCache(rdb)
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	//defer cancel()
	//err := c.Set(ctx, "key1", "value1", time.Minute)
	//require.NoError(t, err)
	//
	//val, err := c.Get(ctx, "key1")
	//require.NoError(t, err)
	//assert.Equal(t, "value1", val)
}

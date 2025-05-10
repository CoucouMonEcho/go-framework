package cache

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestBuildInMapCache_Get(t *testing.T) {
	testCases := []struct {
		name string

		k     string
		v     any
		cache func() *BuildInMapCache

		wantErr error
	}{
		{
			name: "ok",
			k:    "k1",
			cache: func() *BuildInMapCache {
				res := NewBuildInMapCache(10 * time.Second)
				err := res.Set(context.Background(), "k1", 123, time.Minute)
				require.NoError(t, err)
				return res
			},
			v: 123,
		},
		{
			name: "key not found",
			k:    "not exist",
			cache: func() *BuildInMapCache {
				return NewBuildInMapCache(10 * time.Second)
			},
			wantErr: errKeyNotFound,
		},
		{
			name: "expired",
			k:    "k1",
			cache: func() *BuildInMapCache {
				res := NewBuildInMapCache(10 * time.Second)
				err := res.Set(context.Background(), "k1", 123, time.Second)
				time.Sleep(time.Second)
				require.NoError(t, err)
				return res
			},
			wantErr: errKeyNotFound,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := tc.cache()
			v, err := cache.Get(context.Background(), tc.k)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.v, v)
		})
	}
}

func TestBuildInMapCache_Loop(t *testing.T) {
	cnt := 0
	c := NewBuildInMapCache(time.Second, BuildInMapCacheWithEvictedCallback(func(k string, v any) {
		cnt++
	}))
	require.NoError(t, c.Set(context.Background(), "k1", 123, time.Second))
	time.Sleep(3 * time.Second)
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	_, ok := c.data["k1"]
	require.False(t, ok)
	assert.Equal(t, 1, cnt)
}

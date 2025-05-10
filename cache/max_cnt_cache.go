package cache

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

var (
	errOverCapacity = errors.New("cache: ver capacity")
)

type MaxCntCache struct {
	*BuildInMapCache
	cnt int32
	max int32
}

func NewMaxCntCache(max int32, c *BuildInMapCache) *MaxCntCache {
	res := &MaxCntCache{
		BuildInMapCache: c,
		max:             max,
	}
	origin := c.onEvicted
	c.onEvicted = func(k string, v any) {
		atomic.AddInt32(&res.cnt, -1)
		if origin != nil {
			origin(k, v)
		}
	}
	return res
}

func (m *MaxCntCache) Set(ctx context.Context, k string, v any, expire time.Duration) error {
	// The duplicate key issue cannot be resolved
	//cnt := atomic.AddInt32(&m.cnt, 1)
	//if cnt > m.max {
	//	atomic.AddInt32(&m.cnt, -1)
	//	return errOverCapacity
	//}
	//return m.c.Set(ctx, k, v, expire)

	// Unlocking with SET non-atomicity, CNT may increase repeatedly
	//m.mutex.Lock()
	//_, ok := m.data[k]
	//if !ok {
	//	m.cnt++
	//}
	//if m.cnt >= m.max {
	//	return errOverCapacity
	//}
	//m.mutex.Unlock()
	//return m.BuildInMapCache.Set(ctx, k, v, expire)

	m.mutex.Lock()
	defer m.mutex.Unlock()
	_, ok := m.data[k]
	if !ok {
		if m.cnt >= m.max {
			//TODO LRU or LFU
			return errOverCapacity
		}
		m.cnt++
	}
	return m.set(ctx, k, v, expire)
}

func (m *MaxCntCache) Get(ctx context.Context, k string) (any, error) {
	return m.Get(ctx, k)
}

func (m *MaxCntCache) Del(ctx context.Context, k string) error {
	return m.Del(ctx, k)
}

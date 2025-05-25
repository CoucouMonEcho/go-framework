package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	errKeyNotFound   = errors.New("cache: key not found")
	errAlreadyClosed = errors.New("cache: already closed")
)

type BuildInMapCache struct {
	data      map[string]item
	mutex     sync.RWMutex
	close     chan struct{}
	onEvicted func(k string, v any)
}

type BuildInMapCacheOption func(b *BuildInMapCache)

func BuildInMapCacheWithEvictedCallback(onEvicted func(k string, v any)) BuildInMapCacheOption {
	return func(b *BuildInMapCache) {
		b.onEvicted = onEvicted
	}
}

func NewBuildInMapCache(interval time.Duration, opts ...BuildInMapCacheOption) *BuildInMapCache {
	res := &BuildInMapCache{
		data:      make(map[string]item, 128),
		close:     make(chan struct{}),
		onEvicted: func(k string, v any) {},
	}
	for i := len(opts) - 1; i >= 0; i-- {
		opts[i](res)
	}
	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case t := <-ticker.C:
				res.mutex.Lock()
				i := 0
				for k, v := range res.data {
					if i > 10000 {
						break
					}
					if !v.expired(t) {
						res.delete(k)
					}
					i++
				}
				res.mutex.Unlock()
			case <-res.close:
				return
			}
		}
	}()
	return res
}

func (b *BuildInMapCache) Set(ctx context.Context, k string, v any, expire time.Duration) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.set(ctx, k, v, expire)
}

func (b *BuildInMapCache) set(ctx context.Context, k string, v any, expire time.Duration) error {
	var dl time.Time
	if expire > 0 {
		dl = time.Now().Add(expire)
	}
	b.data[k] = item{v, dl}
	return nil
}

func (b *BuildInMapCache) Get(ctx context.Context, k string) (any, error) {
	b.mutex.RLock()
	itm, ok := b.data[k]
	b.mutex.RUnlock()
	if !ok {
		return nil, errKeyNotFound
	}
	now := time.Now()
	if itm.expired(now) {
		b.mutex.Lock()
		defer b.mutex.Unlock()
		itm, ok = b.data[k]
		if !ok {
			return nil, errKeyNotFound
		}
		if itm.expired(now) {
			b.delete(k)
			return nil, errKeyNotFound
		}
	}
	return itm.v, nil
}

func (b *BuildInMapCache) Del(ctx context.Context, k string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.delete(k)
	return nil
}

func (b *BuildInMapCache) LoadAndDelete(ctx context.Context, k string) (any, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	itm, ok := b.data[k]
	if !ok {
		return nil, errKeyNotFound
	}
	delete(b.data, k)
	b.onEvicted(k, itm.v)
	return itm.v, nil
}

func (b *BuildInMapCache) delete(k string) {
	itm, ok := b.data[k]
	if !ok {
		return
	}
	delete(b.data, k)
	b.onEvicted(k, itm.v)
}

func (b *BuildInMapCache) Close() error {
	select {
	case b.close <- struct{}{}:
	default:
		return errAlreadyClosed
	}
	return nil
}

func (b *BuildInMapCache) OnEvicted(f func(k string, v any)) {
	//TODO implement me
	panic("implement me")
}

type item struct {
	v  any
	dl time.Time
}

func (i *item) expired(t time.Time) bool {
	return !i.dl.IsZero() && t.After(i.dl)
}

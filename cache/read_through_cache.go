package cache

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/singleflight"
	"time"
)

var (
	ErrFailedToRefreshCache = errors.New("cache: failed to refresh cache")
)

type ReadThroughCache struct {
	Cache
	LoadFunc   func(ctx context.Context, k string) (any, error)
	Expiration time.Duration
	g          singleflight.Group
}

// Get single flight
func (r *ReadThroughCache) Get(ctx context.Context, k string) (any, error) {
	v, err := r.Cache.Get(ctx, k)
	if err == errKeyNotFound {
		v, err, _ = r.g.Do(k, func() (any, error) {
			val, er := r.LoadFunc(ctx, k)
			if er != nil {
				return v, er
			}
			if er = r.Cache.Set(ctx, k, val, r.Expiration); err != nil {
				return nil, fmt.Errorf("%w: %s", ErrFailedToRefreshCache, err)
			}
			return val, err
		})
	}
	return v, err
}

// Get semi asynchronous
//func (r *ReadThroughCache) Get(ctx context.Context, k string) (any, error) {
//	v, err := r.Cache.Get(ctx, k)
//	if err == errKeyNotFound {
//		v, err = r.LoadFunc(ctx, k)
//		if err != nil {
//			return v, err
//		}
//		go func() {
//			_ = r.Cache.Set(ctx, k, v, r.Expiration)
//		}()
//	}
//	return v, err
//}

// Get asynchronous
//func (r *ReadThroughCache) Get(ctx context.Context, k string) (any, error) {
//	v, err := r.Cache.Get(ctx, k)
//	if err == errKeyNotFound {
//		go func() {
//			v, err = r.LoadFunc(ctx, k)
//			if err != nil {
//				return
//			}
//			_ = r.Cache.Set(ctx, k, v, r.Expiration)
//		}()
//	}
//	return v, err
//}

// Get synchronous
//func (r *ReadThroughCache) Get(ctx context.Context, k string) (any, error) {
//	v, err := r.Cache.Get(ctx, k)
//	if err == errKeyNotFound {
//		v, err = r.LoadFunc(ctx, k)
//		if err != nil {
//			return v, err
//		}
//		if err = r.Cache.Set(ctx, k, v, r.Expiration); err != nil {
//			return nil, fmt.Errorf("%w: %s", ErrFailedToRefreshCache, err)
//		}
//	}
//	return v, err
//}

package cache

import (
	"context"
	"golang.org/x/sync/singleflight"
	"time"
)

type SingleflightCache struct {
	ReadThroughCache
}

func NewSingleflightCache(cache Cache,
	loadFunc func(ctx context.Context, k string) (any, error),
	expiration time.Duration) Cache {
	g := &singleflight.Group{}
	return &SingleflightCache{
		ReadThroughCache: ReadThroughCache{
			Cache: cache,
			LoadFunc: func(ctx context.Context, k string) (any, error) {
				v, err, _ := g.Do(k, func() (any, error) {
					return loadFunc(ctx, k)
				})
				return v, err
			},
			Expiration: expiration,
		},
	}
}

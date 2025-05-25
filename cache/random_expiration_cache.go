package cache

import (
	"context"
	"math/rand"
	"time"
)

type RandomExpirationCache struct {
	Cache
}

func (r *RandomExpirationCache) Set(ctx context.Context, k string, v any, expire time.Duration) error {
	if expire > 0 {
		offset := time.Duration(rand.Intn(300)) * time.Second
		return r.Cache.Set(ctx, k, v, expire+offset)
	}
	return r.Cache.Set(ctx, k, v, expire)
}

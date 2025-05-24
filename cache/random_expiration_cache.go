package cache

import (
	"context"
	"math/rand"
	"time"
)

type RandomExpirationCache struct {
	Cache
}

func (r *RandomExpirationCache) Set(ctx context.Context, k string, v any, expiration time.Duration) error {
	if expiration > 0 {
		offset := time.Duration(rand.Intn(300)) * time.Second
		return r.Cache.Set(ctx, k, v, expiration+offset)
	}
	return r.Cache.Set(ctx, k, v, expiration)
}

package cache

import (
	"context"
)

type BloomFilterCache struct {
	ReadThroughCache
}

func NewBloomFilterCache(cache Cache, bf BloomFilter,
	loadFunc func(ctx context.Context, k string) (any, error)) *BloomFilterCache {
	return &BloomFilterCache{
		ReadThroughCache{
			Cache: cache,
			LoadFunc: func(ctx context.Context, k string) (any, error) {
				if !bf.HasKey(ctx, k) {
					return nil, nil
				}
				return loadFunc(ctx, k)
			},
		},
	}
}

type BloomFilter interface {
	HasKey(ctx context.Context, k string) bool
}

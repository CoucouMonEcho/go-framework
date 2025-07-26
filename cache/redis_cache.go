package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

// go install github.com/golang/mock/mockgen@v1.6.0
// mockgen -destination=cache/mocks/mock_redis_cmdable.gen.go -package=mocks github.com/redis/go-redis/v9 Cmdable

var (
	errFailedToSetCache = errors.New("cache: failed to set cache")
)

type RedisCache struct {
	client redis.Cmdable
}

func NewRedisCache(client redis.Cmdable) *RedisCache {
	return &RedisCache{
		client: client,
	}
}

func (r *RedisCache) Set(ctx context.Context, k string, v any, expire time.Duration) error {
	res, err := r.client.Set(ctx, k, v, expire).Result()
	if err != nil {
		return err
	}
	if res != "OK" {
		return fmt.Errorf("%w, %s", errFailedToSetCache, res)
	}
	return nil
}

func (r *RedisCache) Get(ctx context.Context, k string) (any, error) {
	return r.client.Get(ctx, k).Result()
}

func (r *RedisCache) Del(ctx context.Context, k string) error {
	_, err := r.client.Del(ctx, k).Result()
	return err
}

func (r *RedisCache) LoadAndDelete(ctx context.Context, k string) (any, error) {
	return r.client.GetDel(ctx, k).Result()
}

func (r *RedisCache) OnEvicted(_ func(k string, v any)) {
	// do nothing
}

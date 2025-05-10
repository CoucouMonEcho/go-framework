package cache

import (
	"context"
	"time"
)

type Cache interface {
	Set(ctx context.Context, k string, v any, expire time.Duration) error
	Get(ctx context.Context, k string) (any, error)
	Del(ctx context.Context, k string) error
}

package cache

import (
	"context"
	"time"
)

type WriteThroughCache struct {
	Cache
	StoreFunc  func(ctx context.Context, k string, v any) error
	Expiration time.Duration
}

// Set semi asynchronous
func (w *WriteThroughCache) Set(ctx context.Context, k string, v any) error {
	err := w.StoreFunc(ctx, k, v)
	go func() {
		_ = w.Cache.Set(ctx, k, v, w.Expiration)
	}()
	return err
}

// Set asynchronous
//func (w *WriteThroughCache) Set(ctx context.Context, k string, v any) error {
//	go func() {
//		_ = w.StoreFunc(ctx, k, v)
//		_ = w.Cache.Set(ctx, k, v, w.Expiration)
//	}()
//	return nil
//}

// Set synchronous
//func (w *WriteThroughCache) Set(ctx context.Context, k string, v any) error {
//	err := w.StoreFunc(ctx, k, v)
//	if err != nil {
//		return err
//	}
//	go func() {
//		_ = w.Cache.Set(ctx, k, v, w.Expiration)
//	}()
//	return err
//}

package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"time"
)

var (
	ErrFailedToPreemptLock  = errors.New("redis-lock: failed to preempt lock")
	ErrLockNotHold          = errors.New("redis-lock: lock not hold")
	ErrRetryLimitIsExceeded = errors.New("redis-lock: retry limit is exceeded")

	//go:embed lua/unlock.lua
	luaUnlock string
	//go:embed lua/refresh.lua
	luaRefresh string
	//go:embed lua/lock.lua
	luaLock string
)

type Client struct {
	client redis.Cmdable
	g      singleflight.Group
}

func NewClient(client redis.Cmdable) (*Client, error) {
	return &Client{client: client}, nil
}

func (c *Client) SingleflightLock(ctx context.Context,
	key string,
	expire time.Duration,
	timeout time.Duration,
	retry RetryStrategy) (*Lock, error) {
	for {
		flag := false
		resChan := c.g.DoChan(key, func() (any, error) {
			flag = true
			return c.Lock(ctx, key, expire, timeout, retry)
		})
		select {
		case res := <-resChan:
			if flag {
				c.g.Forget(key)
				if res.Err != nil {
					return nil, res.Err
				}
				return res.Val.(*Lock), nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (c *Client) Lock(ctx context.Context,
	key string,
	expire time.Duration,
	timeout time.Duration,
	retry RetryStrategy) (*Lock, error) {
	var timer *time.Timer
	val := uuid.New().String()
	for {
		lctx, cancel := context.WithTimeout(ctx, timeout)
		res, err := c.client.Eval(lctx, luaLock, []string{key}, val, expire.Seconds()).Result()
		cancel()
		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		if res == "OK" {
			return &Lock{
				client:     c.client,
				key:        key,
				value:      val,
				expire:     expire,
				unlockChan: make(chan struct{}, 1),
			}, nil
		}
		interval, ok := retry.Next()
		if !ok {
			return nil, ErrRetryLimitIsExceeded
		}
		if timer == nil {
			timer = time.NewTimer(interval)
		} else {
			timer.Reset(interval)
		}
		select {
		case <-timer.C:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (c *Client) TryLock(ctx context.Context,
	key string,
	expire time.Duration) (*Lock, error) {
	val := uuid.New().String()
	ok, err := c.client.SetNX(ctx, key, "lock", expire).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrFailedToPreemptLock
	}
	return &Lock{
		client:     c.client,
		key:        key,
		value:      val,
		expire:     expire,
		unlockChan: make(chan struct{}, 1),
	}, nil
}

type Lock struct {
	client redis.Cmdable
	key    string
	value  string
	expire time.Duration

	unlockChan chan struct{}
}

func (l *Lock) AutoRefresh(interval time.Duration, timeout time.Duration) error {
	var lock Lock
	timeoutChan := make(chan struct{}, 1)
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			err := lock.Refresh(ctx)
			cancel()
			if errors.Is(err, context.DeadlineExceeded) {
				timeoutChan <- struct{}{}
				continue
			}
			if err != nil {
				return err
			}
		case <-timeoutChan:
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			err := lock.Refresh(ctx)
			cancel()
			if errors.Is(err, context.DeadlineExceeded) {
				timeoutChan <- struct{}{}
				continue
			}
			if err != nil {
				return err
			}
		case <-l.unlockChan:
			return nil
		}
	}
}

func (l *Lock) Unlock(ctx context.Context) error {
	res, err := l.client.Eval(ctx, luaUnlock, []string{l.key}, l.value).Int64()
	defer func() {
		select {
		case l.unlockChan <- struct{}{}:
		default:
		}
	}()
	if errors.Is(err, redis.Nil) {
		return fmt.Errorf("%w: %s", ErrLockNotHold, l.key)
	}
	if err != nil {
		return err
	}
	if res != 1 {
		return fmt.Errorf("%w: %s", ErrLockNotHold, l.key)
	}
	return nil
}

func (l *Lock) Refresh(ctx context.Context) error {
	res, err := l.client.Eval(ctx, luaRefresh, []string{l.key}, l.value, l.expire.Seconds()).Int64()
	//defer func() {
	//	l.unlock <- struct{}{}
	//}()
	if errors.Is(err, redis.Nil) {
		return fmt.Errorf("%w: %s", ErrLockNotHold, l.key)
	}
	if err != nil {
		return err
	}
	if res != 1 {
		return fmt.Errorf("%w: %s", ErrLockNotHold, l.key)
	}
	return nil
}

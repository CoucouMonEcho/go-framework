package rate_limit

import (
	"context"
	_ "embed"
	"errors"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"time"
)

//go:embed lua/fix_window.lua
var luaFixWindow string

type RedisFixWindowLimiter struct {
	client   redis.Cmdable
	service  string
	interval time.Duration
	rate     int
}

func NewRedisFixWindowLimiter(client redis.Cmdable, service string, interval time.Duration, rate int) *RedisFixWindowLimiter {
	return &RedisFixWindowLimiter{
		client:   client,
		service:  service,
		interval: interval,
		rate:     rate,
	}
}

func (r *RedisFixWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		// use method to control the throttling dimension
		// method, instance, service
		limit, err := r.limit(ctx)
		if err != nil {
			return
		}
		if limit {
			err = errors.New("micro: redis fix window limit exceeded")
			return
		}
		resp, err = handler(ctx, req)
		return
	}
}

func (r *RedisFixWindowLimiter) BuildClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		//TODO
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (r *RedisFixWindowLimiter) limit(ctx context.Context) (bool, error) {
	return r.client.Eval(ctx, luaFixWindow, []string{r.service},
		r.rate, r.interval.Milliseconds()).Bool()
}

func (r *RedisFixWindowLimiter) Close() error {
	//TODO
	return nil
}

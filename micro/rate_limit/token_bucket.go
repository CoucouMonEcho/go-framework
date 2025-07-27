package rate_limit

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"time"
)

type TokenBucketLimiter struct {
	tokens chan struct{}
	close  chan struct{}
}

func NewTokenBucketLimiter(capacity int, interval time.Duration) *TokenBucketLimiter {
	tokens := make(chan struct{}, capacity)
	closeCh := make(chan struct{})
	producer := time.NewTicker(interval)
	go func() {
		defer producer.Stop()
		for {
			select {
			case <-producer.C:
				select {
				case tokens <- struct{}{}:
				default:
					// no one get token
				}
			case <-closeCh:
				return
			}
		}
	}()
	return &TokenBucketLimiter{
		tokens: tokens,
		close:  closeCh,
	}
}

func (t *TokenBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		select {
		case <-ctx.Done():
			err = ctx.Err()
		case <-t.close:
			resp, err = handler(ctx, req)
		case <-t.tokens:
			resp, err = handler(ctx, req)
		default:
			err = errors.New("micro: token limit exceeded")
		}
		return
	}
}

func (t *TokenBucketLimiter) BuildClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		//TODO
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (t *TokenBucketLimiter) Close() error {
	close(t.close)
	return nil
}

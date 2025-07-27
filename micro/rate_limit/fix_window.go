package rate_limit

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"sync/atomic"
	"time"
)

type FixWindowLimiter struct {
	timeStamp int64
	interval  int64

	rate int32
	cnt  int32

	//mutex sync.Mutex
}

func NewFixWindowLimiter(interval time.Duration, rate int32) *FixWindowLimiter {
	return &FixWindowLimiter{
		timeStamp: time.Now().UnixNano(),
		interval:  interval.Nanoseconds(),
		rate:      rate,
	}
}

func (f *FixWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if f.rate < 0 {
			resp, err = handler(ctx, req)
			return
		}
		//f.mutex.Lock()
		now := time.Now().UnixNano()
		timeStamp := atomic.LoadInt64(&f.timeStamp)
		cnt := atomic.LoadInt32(&f.cnt)
		if timeStamp+f.interval < now {
			if atomic.CompareAndSwapInt64(&f.timeStamp, timeStamp, now) {
				//atomic.StoreInt64(&f.cnt, 0)
				atomic.CompareAndSwapInt32(&f.cnt, cnt, 0)
			}
		}
		cnt = atomic.AddInt32(&f.cnt, 1)
		if cnt > f.rate {
			err = errors.New("micro: fix window limit exceeded")
			//atomic.AddInt32(&f.cnt, -1)
			//f.mutex.Unlock()
			return
		}
		//f.mutex.Unlock()
		resp, err = handler(ctx, req)
		return
	}
}

func (f *FixWindowLimiter) BuildClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		//TODO
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (f *FixWindowLimiter) Close() error {
	f.rate = -1
	return nil
}

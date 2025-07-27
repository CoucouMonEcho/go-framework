package rate_limit

import (
	"container/list"
	"context"
	"errors"
	"google.golang.org/grpc"
	"math"
	"sync"
	"time"
)

type SlideWindowLimiter struct {
	queue    *list.List
	interval int64

	rate  int
	mutex sync.Mutex
}

func NewSlideWindowLimiter(interval time.Duration, rate int) *SlideWindowLimiter {
	return &SlideWindowLimiter{
		queue:    list.New(),
		interval: interval.Nanoseconds(),
		rate:     rate,
	}
}

func (s *SlideWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if s.limit() {
			err = errors.New("micro: slide window limit exceeded")
			return
		}
		resp, err = handler(ctx, req)
		return
	}
}

func (s *SlideWindowLimiter) BuildClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		//TODO
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (s *SlideWindowLimiter) limit() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// fast path
	now := time.Now().UnixNano()
	if s.queue.Len() < s.rate {
		s.queue.PushBack(now)
		return false
	}
	// slow path
	boundary := now - s.interval
	timeStamp := s.queue.Front()
	for timeStamp != nil && timeStamp.Value.(int64) < boundary {
		s.queue.Remove(timeStamp)
		timeStamp = s.queue.Front()
	}
	if s.queue.Len() < s.rate {
		s.queue.PushBack(now)
		return false
	}
	return true
}

func (s *SlideWindowLimiter) Close() error {
	s.rate = math.MaxInt32
	return nil
}

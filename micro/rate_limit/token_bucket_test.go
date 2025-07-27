package rate_limit

import (
	"code-practise/micro/rpc/proto/gen"
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestTokenBucketLimiter_BuildServerInterceptor(t *testing.T) {
	testCases := []struct {
		name    string
		b       func() *TokenBucketLimiter
		ctx     context.Context
		handler func(ctx context.Context, req any) (any, error)

		wantErr  error
		wantResp any
	}{
		{
			name: "success",
			b: func() *TokenBucketLimiter {
				res := NewTokenBucketLimiter(1, time.Second)
				time.Sleep(time.Second)
				return res
			},
			handler: func(ctx context.Context, req any) (any, error) {
				return gen.GetByIdResp{}, nil
			},
			ctx:      context.Background(),
			wantResp: gen.GetByIdResp{},
		},
		{
			name: "closed",
			b: func() *TokenBucketLimiter {
				closeCh := make(chan struct{})
				close(closeCh)
				return &TokenBucketLimiter{
					tokens: make(chan struct{}, 1),
					close:  closeCh,
				}
			},
			handler: func(ctx context.Context, req any) (any, error) {
				return nil, nil
			},
			ctx: context.Background(),
			//wantErr: context.Canceled,
		},
		{
			name: "canceled",
			b: func() *TokenBucketLimiter {
				return NewTokenBucketLimiter(1, time.Millisecond*10)
			},
			handler: func(ctx context.Context, req any) (any, error) {
				return nil, nil
			},
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			interceptor := tc.b().BuildServerInterceptor()
			resp, err := interceptor(tc.ctx, &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, tc.handler)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}

func TestTokenBucketLimiter_Tokens(t *testing.T) {
	limiter := NewTokenBucketLimiter(10, time.Millisecond*900)
	interceptor := limiter.BuildServerInterceptor()
	cnt := 0
	handler := func(ctx context.Context, req any) (any, error) {
		cnt++
		return gen.GetByIdResp{}, nil
	}

	time.Sleep(time.Second)

	resp, err := interceptor(context.Background(), &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
	require.NoError(t, err)
	assert.Equal(t, gen.GetByIdResp{}, resp)

	resp, err = interceptor(context.Background(), &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
	assert.Equal(t, errors.New("micro: token limit exceeded"), err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	time.Sleep(time.Second)
	defer cancel()
	resp, err = interceptor(ctx, &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
	assert.Equal(t, context.DeadlineExceeded, err)

}

package rate_limit

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go-framework/micro/rpc/proto/gen"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestLeakyBucketLimiter_BuildServerInterceptor(t *testing.T) {
	testCases := []struct {
		name    string
		b       func() *LeakyBucketLimiter
		ctx     context.Context
		handler func(ctx context.Context, req any) (any, error)

		wantErr  error
		wantResp any
	}{
		{
			name: "success",
			b: func() *LeakyBucketLimiter {
				res := NewLeakyBucketLimiter(time.Second)
				time.Sleep(time.Second)
				return res
			},
			handler: func(ctx context.Context, req any) (any, error) {
				return gen.GetByIdResp{}, nil
			},
			ctx:      context.Background(),
			wantResp: gen.GetByIdResp{},
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

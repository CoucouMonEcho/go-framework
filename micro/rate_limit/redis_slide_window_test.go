package rate_limit

import (
	"code-practise/micro/rate_limit/mocks"
	"code-practise/micro/rpc/proto/gen"
	"context"
	"github.com/golang/mock/gomock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestRedisSlideWindowLimiter_BuildServerInterceptor(t *testing.T) {
	//ctrl := gomock.NewController(t)
	//defer ctrl.Finish()
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) redis.Cmdable

		wantErr error
	}{
		{
			name: "success",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal("false")
				cmd.EXPECT().
					Eval(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(res)
				return cmd
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			limiter := NewRedisSlideWindowLimiter(tc.mock(ctrl), "test-service", time.Second, 1)
			interceptor := limiter.BuildServerInterceptor()
			cnt := 0
			handler := func(ctx context.Context, req any) (any, error) {
				cnt++
				return gen.GetByIdResp{}, nil
			}

			resp, err := interceptor(context.Background(), &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
			require.NoError(t, err)
			assert.Equal(t, gen.GetByIdResp{}, resp)

		})
	}
}

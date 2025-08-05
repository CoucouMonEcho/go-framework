package rate_limit

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-framework/micro/rpc/proto/gen"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestFixWindowLimiter_BuildServerInterceptor(t *testing.T) {
	limiter := NewFixWindowLimiter(time.Second*3, 1)
	interceptor := limiter.BuildServerInterceptor()
	cnt := 0
	handler := func(ctx context.Context, req any) (any, error) {
		cnt++
		return gen.GetByIdResp{}, nil
	}

	resp, err := interceptor(context.Background(), &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
	require.NoError(t, err)
	assert.Equal(t, gen.GetByIdResp{}, resp)

	resp, err = interceptor(context.Background(), &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
	assert.Equal(t, errors.New("micro: fix window limit exceeded"), err)

	time.Sleep(time.Second * 3)

	resp, err = interceptor(context.Background(), &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
	require.NoError(t, err)
	assert.Equal(t, gen.GetByIdResp{}, resp)

}

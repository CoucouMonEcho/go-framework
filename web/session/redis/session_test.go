package redis

import (
	"code-practise/cache/mocks"
	"context"
	"github.com/golang/mock/gomock"
	"github.com/redis/go-redis/v9"
	"testing"
	"time"
)

func TestNewStore(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cmd := mocks.NewMockCmdable(ctrl)
	status := redis.NewStatusCmd(context.Background())
	status.SetVal("OK")
	cmd.EXPECT().
		Set(context.Background(), "key1", "value1", time.Second).
		Return(status)
	NewStore(cmd, StoreWithPrefix("session"))
}

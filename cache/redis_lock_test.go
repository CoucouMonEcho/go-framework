package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/CoucouMonEcho/go-framework/cache/mocks"
	"github.com/golang/mock/gomock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
	"time"
)

func TestClient_TryLock(t *testing.T) {
	testCases := []struct {
		name string
		key  string

		mock func(ctrl *gomock.Controller) redis.Cmdable

		wantErr  error
		wantLock *Lock
	}{
		{
			name: "success",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(true, nil)
				cmd.EXPECT().
					SetNX(context.Background(), "key1", gomock.Any(), time.Minute).
					Return(res)
				return cmd
			},
			key: "key1",
			wantLock: &Lock{
				key:    "key1",
				expire: time.Minute,
			},
		},
		{
			name: "set nx error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(false, context.DeadlineExceeded)
				cmd.EXPECT().
					SetNX(context.Background(), "key1", gomock.Any(), time.Minute).
					Return(res)
				return cmd
			},
			key:     "key1",
			wantErr: context.DeadlineExceeded,
		},
		{
			name: "fail to preempt lock",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(false, nil)
				cmd.EXPECT().
					SetNX(context.Background(), "key1", gomock.Any(), time.Minute).
					Return(res)
				return cmd
			},
			key:     "key1",
			wantErr: ErrFailedToPreemptLock,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			client, err := NewClient(tc.mock(ctrl))
			require.NoError(t, err)
			lock, err := client.TryLock(context.Background(), tc.key, time.Minute)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantLock.key, lock.key)
			assert.Equal(t, tc.wantLock.expire, lock.expire)
			assert.NotEmpty(t, lock.value)
		})
	}
}

func TestClient_UnLock(t *testing.T) {
	testCases := []struct {
		name  string
		key   string
		value string

		mock func(ctrl *gomock.Controller) redis.Cmdable

		wantErr  error
		wantLock *Lock
	}{
		{
			name:  "success",
			key:   "key1",
			value: "value1",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(1))
				cmd.EXPECT().
					Eval(context.Background(), luaUnlock, []string{"key1"}, "value1").
					Return(res)
				return cmd
			},
			wantErr: nil,
		},
		{
			name:  "eval error",
			key:   "key1",
			value: "value1",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(context.DeadlineExceeded)
				cmd.EXPECT().
					Eval(context.Background(), luaUnlock, []string{"key1"}, "value1").
					Return(res)
				return cmd
			},
			wantErr: context.DeadlineExceeded,
		},
		{
			name:  "lock not hold",
			key:   "key1",
			value: "value1",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(0))
				cmd.EXPECT().
					Eval(context.Background(), luaUnlock, []string{"key1"}, "value1").
					Return(res)
				return cmd
			},
			wantErr: fmt.Errorf("%w: %s", ErrLockNotHold, "key1"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			lock := &Lock{
				key:    tc.key,
				value:  tc.value,
				client: tc.mock(ctrl),
			}
			err := lock.Unlock(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
		})
	}
}

func TestClient_Refresh(t *testing.T) {
	testCases := []struct {
		name   string
		key    string
		value  string
		expire time.Duration

		mock func(ctrl *gomock.Controller) redis.Cmdable

		wantErr  error
		wantLock *Lock
	}{
		{
			name:  "success",
			key:   "key1",
			value: "value1",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(1))
				cmd.EXPECT().
					Eval(context.Background(), luaRefresh, []string{"key1"}, "value1", float64(60)).
					Return(res)
				return cmd
			},
			expire:  time.Minute,
			wantErr: nil,
		},
		{
			name:  "eval error",
			key:   "key1",
			value: "value1",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(context.DeadlineExceeded)
				cmd.EXPECT().
					Eval(context.Background(), luaRefresh, []string{"key1"}, "value1", float64(60)).
					Return(res)
				return cmd
			},
			expire:  time.Minute,
			wantErr: context.DeadlineExceeded,
		},
		{
			name:  "lock not hold",
			key:   "key1",
			value: "value1",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(0))
				cmd.EXPECT().
					Eval(context.Background(), luaRefresh, []string{"key1"}, "value1", float64(60)).
					Return(res)
				return cmd
			},
			expire:  time.Minute,
			wantErr: fmt.Errorf("%w: %s", ErrLockNotHold, "key1"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			lock := &Lock{
				key:    tc.key,
				value:  tc.value,
				client: tc.mock(ctrl),
				expire: tc.expire,
			}
			err := lock.Refresh(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
		})
	}
}

// manual renewal template
func ExampleLock_Refresh() {
	var lock *Lock
	stopChan := make(chan struct{})
	errChan := make(chan error)
	timeoutChan := make(chan struct{}, 1)
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		timeoutRetry := 0
		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				err := lock.Refresh(ctx)
				cancel()
				if errors.Is(err, context.DeadlineExceeded) {
					timeoutChan <- struct{}{}
					continue
				}
				if err != nil {
					errChan <- err
					return
				}
			case <-timeoutChan:
				timeoutRetry++
				if timeoutRetry == 3 {
					errChan <- context.DeadlineExceeded
					return
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				err := lock.Refresh(ctx)
				cancel()
				if errors.Is(err, context.DeadlineExceeded) {
					timeoutChan <- struct{}{}
					continue
				}
				if err != nil {
					errChan <- err
					return
				}
			case <-stopChan:
				return
			}
		}
	}()

	// step 1
	select {
	case err := <-errChan:
		log.Fatalln(err)
		return
	default:
		// do something
		time.Sleep(1 * time.Second)
	}
	// step 2
	select {
	case err := <-errChan:
		log.Fatalln(err)
		return
	default:
		// do something
		time.Sleep(1 * time.Second)
	}
	// step ...

	// done
	err := lock.Unlock(context.Background())
	if err != nil {
		errChan <- err
	}
	stopChan <- struct{}{}

}

// automatic renewal template
func ExampleLock_AutoRefresh() {
	var lock *Lock
	var errChan chan error
	go func() {
		err := lock.AutoRefresh(time.Second*8, time.Second*10)
		if err != nil {
			errChan <- err
		}
	}()

	// step 1
	select {
	case err := <-errChan:
		log.Fatalln(err)
		return
	default:
		// do something
		time.Sleep(1 * time.Second)
	}
	// step 2
	select {
	case err := <-errChan:
		log.Fatalln(err)
		return
	default:
		// do something
		time.Sleep(1 * time.Second)
	}
	// step ...

	// done
	err := lock.Unlock(context.Background())
	if err != nil {
		errChan <- err
	}
}

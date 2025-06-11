package rpc

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
	"time"
)

func TestInitClientProxy(t *testing.T) {
	// server
	server := NewServer()
	serviceServer := &TestServiceServer{}
	server.RegisterService(serviceServer)
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second)

	// client
	serviceClient := &TestService{}
	err := InitClientProxy(":8081", serviceClient)
	require.NoError(t, err)

	testCases := []struct {
		name string

		mock func()

		wangErr  error
		wantResp *GetByIdResp
	}{
		{
			name: "success",
			mock: func() {
				serviceServer.Err = errors.New("mock error")
				serviceServer.Msg = "hello, world"
			},
			wantResp: &GetByIdResp{
				Msg: "hello, world",
			},
			wangErr: errors.New("mock error"),
		},
		{
			name: "no error",
			mock: func() {
				serviceServer.Err = nil
				serviceServer.Msg = "hello, world"
			},
			wantResp: &GetByIdResp{
				Msg: "hello, world",
			},
		},
		{
			name: "error",
			mock: func() {
				serviceServer.Err = errors.New("mock error")
			},
			wangErr:  errors.New("mock error"),
			wantResp: &GetByIdResp{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			resp, er := serviceClient.GetById(context.Background(), &GetByIdReq{Id: 123})
			assert.Equal(t, tc.wangErr, er)
			if er != nil {
				return
			}
			assert.Equal(t, resp, tc.wantResp)
		})
	}
}

// TestServiceServer rpc server
type TestServiceServer struct {
	Err error
	Msg string
}

func (t *TestServiceServer) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {

	// do something
	log.Println(req)

	return &GetByIdResp{
		Msg: t.Msg,
	}, t.Err
}

func (t *TestServiceServer) Name() string {
	return "test-service"
}

// TestService rpc client
type TestService struct {
	GetById func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error)
}

func (t TestService) Name() string {
	return "test-service"
}

type GetByIdReq struct {
	Id int64
}

type GetByIdResp struct {
	Msg string
}

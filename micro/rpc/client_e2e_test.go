package rpc

import (
	"code-practise/micro/rpc/proto/gen"
	"code-practise/micro/rpc/serialize/json"
	"code-practise/micro/rpc/serialize/proto"
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"sync"
	"testing"
	"time"
)

func TestInitClientProxyProto(t *testing.T) {
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
	client, err := NewClient(":8081", ClientWithSerializer(&proto.Serializer{}))
	require.NoError(t, err)
	err = client.InitService(serviceClient)
	require.NoError(t, err)

	testCases := []struct {
		name string

		mock func()

		wangErr  error
		wantResp *gen.GetByIdResp
	}{
		{
			name: "success",
			mock: func() {
				serviceServer.Err = errors.New("mock error")
				serviceServer.Msg = "hello, world"
			},
			wantResp: &gen.GetByIdResp{
				User: &gen.User{
					Id:  123,
					Msg: "hello, world",
				},
			},
			wangErr: errors.New("mock error"),
		},
		{
			name: "no error",
			mock: func() {
				serviceServer.Err = nil
				serviceServer.Msg = "hello, world"
			},
			wantResp: &gen.GetByIdResp{
				User: &gen.User{
					Id:  123,
					Msg: "hello, world",
				},
			},
		},
		{
			name: "error",
			mock: func() {
				serviceServer.Err = errors.New("mock error")
			},
			wangErr:  errors.New("mock error"),
			wantResp: &gen.GetByIdResp{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			resp, er := serviceClient.GetByIdProto(context.Background(), &gen.GetByIdReq{Id: 123})
			assert.Equal(t, tc.wangErr, er)
			if er != nil && resp != nil && resp.User != nil {
				return
			}
			//assert.Equal(t, resp, tc.wantResp)
			assert.Equal(t, resp.User, tc.wantResp.User)
		})
	}
}

func TestInitClientProxyJson(t *testing.T) {
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
	client, err := NewClient(":8081", ClientWithSerializer(&json.Serializer{}))
	require.NoError(t, err)
	err = client.InitService(serviceClient)
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

			var wg sync.WaitGroup
			wg.Add(1)

			var respAsync *GetByIdResp
			go func() {
				respAsync, err = serviceClient.GetById(context.Background(), &GetByIdReq{Id: 123})
				wg.Done()
			}()

			// do something

			wg.Wait()
			assert.Equal(t, tc.wangErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, respAsync, tc.wantResp)
		})
	}
}

func TestOneway(t *testing.T) {
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
	client, err := NewClient(":8081", ClientWithSerializer(&proto.Serializer{}))
	require.NoError(t, err)
	err = client.InitService(serviceClient)
	require.NoError(t, err)

	testCases := []struct {
		name string

		mock func()

		wangErr  error
		wantResp *gen.GetByIdResp
	}{
		{
			name: "success",
			mock: func() {
				serviceServer.Err = errors.New("mock error")
				serviceServer.Msg = "hello, world"
			},
			wantResp: &gen.GetByIdResp{},
			wangErr:  errors.New("rpc: oneway should not handle result"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			ctx := context.Background()
			// oneway
			ctx = CtxWithOneway(ctx)

			resp, er := serviceClient.GetByIdProto(ctx, &gen.GetByIdReq{Id: 123})
			assert.Equal(t, tc.wangErr, er)
			if er != nil && resp != nil && resp.User != nil {
				return
			}
			//assert.Equal(t, resp, tc.wantResp)
			assert.Equal(t, resp.User, tc.wantResp.User)
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

func (t *TestServiceServer) GetByIdProto(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {

	// do something
	log.Println(req)

	return &gen.GetByIdResp{
		User: &gen.User{
			Id:  req.Id,
			Msg: t.Msg,
		},
	}, t.Err
}

func (t *TestServiceServer) Name() string {
	return "test-service"
}

// TestService rpc client
type TestService struct {
	GetById      func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error)
	GetByIdProto func(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error)
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

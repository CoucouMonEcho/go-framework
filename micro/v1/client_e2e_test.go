package v1

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
	"time"
)

func TestInitClientProxy(t *testing.T) {
	// server
	server := NewServer()
	server.RegisterService(&TestServiceServer{})
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second)

	// client
	service := &TestService{}
	err := InitClientProxy(":8081", service)
	require.NoError(t, err)
	resp, err := service.GetById(context.Background(), &GetByIdReq{Id: 123})
	require.NoError(t, err)
	assert.Equal(t, resp.Msg, "hello, world")
}

// TestServiceServer rpc server
type TestServiceServer struct {
}

func (t *TestServiceServer) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {

	// do something
	log.Println(req)

	return &GetByIdResp{
		Msg: "hello, world",
	}, nil
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

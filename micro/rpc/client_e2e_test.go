package rpc

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

type TestServiceServer struct {
}

func (t *TestServiceServer) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	log.Println(req)
	return &GetByIdResp{
		Msg: "hello, world",
	}, nil
}

func (t *TestServiceServer) Name() string {
	return "test-service"
}

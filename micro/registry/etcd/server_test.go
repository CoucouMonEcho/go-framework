package etcd

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"go-framework/micro"
	"go-framework/micro/rpc/proto/gen"
	clientv3 "go.etcd.io/etcd/client/v3"
	"testing"
)

func TestServer(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)

	r, err := NewRegistry(etcdClient)
	require.NoError(t, err)

	ts := &TestServiceServer{}
	server, err := micro.NewServer("test-service", micro.ServerWithRegistry(r))
	require.NoError(t, err)

	gen.RegisterTestServiceServer(server, ts)
	err = server.Start(":8081")
	t.Log(err)
}

type TestServiceServer struct {
	gen.UnimplementedTestServiceServer
}

func (s TestServiceServer) GetById(_ context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	fmt.Println(req)
	return &gen.GetByIdResp{
		User: &gen.User{
			Id:  1,
			Msg: "test",
		},
	}, nil
}

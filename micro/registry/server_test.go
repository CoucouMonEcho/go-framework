package registry

import (
	"code-practise/micro"
	"code-practise/micro/registry/etcd"
	"code-practise/micro/rpc/proto/gen"
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"testing"
)

func TestServer(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)

	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)

	us := &TestServiceServer{}
	server, err := micro.NewServer("test-service", micro.ServerWithRegistry(r))
	require.NoError(t, err)

	gen.RegisterTestServiceServer(server, us)
	err = server.Start(":8081")
	t.Log(err)
}

type TestServiceServer struct {
	gen.UnimplementedTestServiceServer
}

func (s TestServiceServer) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	fmt.Println(req)
	return &gen.GetByIdResp{
		User: &gen.User{
			Id:  1,
			Msg: "test",
		},
	}, nil
}

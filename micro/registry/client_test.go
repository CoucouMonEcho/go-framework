package registry

import (
	"code-practise/micro"
	"code-practise/micro/registry/etcd"
	"code-practise/micro/rpc/proto/gen"
	"context"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)

	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)

	c, err := micro.NewClient(micro.ClientWithInsecure(),
		micro.ClientWithRegistry(r, time.Second*3))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	cc, err := c.Dial(ctx, "test-service")
	require.NoError(t, err)

	client := gen.NewTestServiceClient(cc)
	resp, err := client.GetById(ctx, &gen.GetByIdReq{Id: 1})
	require.NoError(t, err)
	t.Log(resp)
}

package route

import (
	"code-practise/micro"
	"code-practise/micro/registry/etcd"
	"code-practise/micro/route/loadbalance/round_robin"
	"code-practise/micro/rpc/proto/gen"
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"testing"
	"time"
)

func TestGroupRoute(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)

	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)

	var eg errgroup.Group
	for i := range 3 {
		group := "A"
		if i == 0 {
			group = "B"
		}
		server, err := micro.NewServer("test-service", micro.ServerWithRegistry(r), micro.ServerWithGroup(group))
		require.NoError(t, err)

		us := &TestServiceServer{}
		gen.RegisterTestServiceServer(server, us)

		eg.Go(func() error {
			return server.Start(fmt.Sprintf(":808%d", i))
		})
	}

	t.Log(err)

	time.Sleep(1 * time.Second)
	balancer.Register(base.NewBalancerBuilder("DEMO_ROUND_ROBIN", &round_robin.BalancerBuilder{
		Filter: GroupFilterBuilder{}.Build()}, base.Config{HealthCheck: true}))
	// client
	cc, err := grpc.Dial("localhost:8081", grpc.WithInsecure(),
		//grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`))
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"DEMO_ROUND_ROBIN"}`))
	require.NoError(t, err)

	client := gen.NewTestServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	ctx = context.WithValue(ctx, "group", "A")
	defer cancel()

	for range 10 {
		resp, err := client.GetById(ctx, &gen.GetByIdReq{Id: 1})
		require.NoError(t, err)
		t.Log(resp)
	}

}

type TestServiceServer struct {
	group string
	gen.UnimplementedTestServiceServer
}

func (s TestServiceServer) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	fmt.Println(req)
	fmt.Println(s.group)
	return &gen.GetByIdResp{
		User: &gen.User{
			Id:  1,
			Msg: "test",
		},
	}, nil
}

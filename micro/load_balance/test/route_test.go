package test

import (
	"code-practise/micro"
	"code-practise/micro/load_balance/round_robin"
	"code-practise/micro/registry/memery"
	"code-practise/micro/route"
	"code-practise/micro/rpc/proto/gen"
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"testing"
	"time"
)

func TestGroupRoute(t *testing.T) {
	r := memery.NewRegistry()

	var eg errgroup.Group
	for i := range 3 {
		group := "A"
		if i == 0 {
			group = "B"
		}
		server, err := micro.NewServer("test-service", micro.ServerWithRegistry(r), micro.ServerWithGroup(group), micro.ServerWithWeight(100))
		require.NoError(t, err)

		ts := &TestServiceServer{group: group}
		gen.RegisterTestServiceServer(server, ts)

		eg.Go(func() error {
			return server.Start(fmt.Sprintf(":808%d", i))
		})
	}

	time.Sleep(1 * time.Second)

	c, err := micro.NewClient(micro.ClientWithInsecure(),
		micro.ClientWithRegistry(r, time.Second*3),
		micro.ClientWithPickerBuilder(&round_robin.BalancerBuilder{
			Filter: route.GroupFilterBuilder{}.Build()}))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	cc, err := c.Dial(ctx, "test-service")
	require.NoError(t, err)

	client := gen.NewTestServiceClient(cc)
	ctx = context.WithValue(ctx, "group", "B")
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

func (s TestServiceServer) GetById(_ context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	fmt.Println(req)
	fmt.Println(fmt.Sprintf("group is %s", s.group))
	return &gen.GetByIdResp{
		User: &gen.User{
			Id:  1,
			Msg: "test",
		},
	}, nil
}

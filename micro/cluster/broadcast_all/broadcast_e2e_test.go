package broadcast

import (
	"code-practise/micro"
	"code-practise/micro/load_balance/round_robin"
	"code-practise/micro/registry/memery"
	"code-practise/micro/rpc/proto/gen"
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
	"time"
)

func TestUsedBroadCast(t *testing.T) {
	r := memery.NewRegistry()

	var eg errgroup.Group
	var servers []*micro.Server
	var serviceServers []*TestServiceServer
	for i := range 3 {
		server, err := micro.NewServer("test-service", micro.ServerWithRegistry(r))
		servers = append(servers, server)
		require.NoError(t, err)

		ts := &TestServiceServer{idx: i}
		serviceServers = append(serviceServers, ts)
		gen.RegisterTestServiceServer(server, ts)

		eg.Go(func() error {
			return server.Start(fmt.Sprintf(":808%d", i))
		})
	}
	defer func() {
		for _, server := range servers {
			_ = server.Close()
		}
	}()

	time.Sleep(1 * time.Second)

	c, err := micro.NewClient(micro.ClientWithInsecure(),
		micro.ClientWithRegistry(r, time.Second*3),
		micro.ClientWithPickerBuilder(&round_robin.BalancerBuilder{}))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	ctx, ch := UsedBroadCast(ctx)
	bd := NewClusterBuilder(r, "test-service", grpc.WithTransportCredentials(insecure.NewCredentials()))
	cc, err := c.Dial(ctx, "test-service",
		grpc.WithUnaryInterceptor(bd.BuildUnaryClientInterceptor()))
	require.NoError(t, err)

	client := gen.NewTestServiceClient(cc)
	defer cancel()

	go func() {
		// <- ch
		for res := range ch {
			fmt.Println(res)
		}
	}()
	// ch <-
	resp, err := client.GetById(ctx, &gen.GetByIdReq{Id: 1})
	require.NoError(t, err)
	t.Log(resp)

	for _, server := range serviceServers {
		require.Equal(t, 1, server.cnt)
	}

}

type TestServiceServer struct {
	idx int
	cnt int
	gen.UnimplementedTestServiceServer
}

func (s *TestServiceServer) GetById(_ context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	fmt.Println(req)
	s.cnt++
	return &gen.GetByIdResp{
		User: &gen.User{
			Id:  1,
			Msg: fmt.Sprintf("test-%d", s.idx),
		},
	}, nil
}

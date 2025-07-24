package round_robin

import (
	"code-practise/micro/rpc/proto/gen"
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"net"
	"testing"
	"time"
)

func TestBalancer_e2e_Pick(t *testing.T) {
	go func() {
		// server
		us := &TestServiceServer{}
		server := grpc.NewServer()

		gen.RegisterTestServiceServer(server, us)
		l, err := net.Listen("tcp", ":8081")
		require.NoError(t, err)
		err = server.Serve(l)
		t.Log(err)
	}()

	time.Sleep(1 * time.Second)
	balancer.Register(base.NewBalancerBuilder("DEMO_ROUND_ROBIN", &BalancerBuilder{}, base.Config{HealthCheck: true}))
	// client
	cc, err := grpc.Dial("localhost:8081", grpc.WithInsecure(),
		//grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`))
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"DEMO_ROUND_ROBIN"}`))
	require.NoError(t, err)

	client := gen.NewTestServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	resp, err := client.GetById(ctx, &gen.GetByIdReq{Id: 1})
	require.NoError(t, err)
	t.Log(resp)
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

package memery

import (
	"code-practise/micro"
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

func TestRegistry_e2e_Register(t *testing.T) {
	r := NewRegistry()

	var eg errgroup.Group
	for i := range 3 {
		group := "A"
		if i == 0 {
			group = "B"
		}
		server, err := micro.NewServer("test-service", micro.ServerWithRegistry(r), micro.ServerWithGroup(group))
		require.NoError(t, err)

		ts := &TestServiceServer{}
		gen.RegisterTestServiceServer(server, ts)

		eg.Go(func() error {
			return server.Start(fmt.Sprintf(":808%d", i))
		})
	}

	time.Sleep(1 * time.Second)

	// client
	cc, err := grpc.NewClient("localhost:8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
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

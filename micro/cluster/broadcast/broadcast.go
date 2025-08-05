package broadcast

import (
	"context"
	"github.com/CoucouMonEcho/go-framework/micro/registry"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type ClusterBuilder struct {
	registry registry.Registry
	service  string
	opts     []grpc.DialOption
}

func NewClusterBuilder(registry registry.Registry, service string, opts ...grpc.DialOption) *ClusterBuilder {
	return &ClusterBuilder{
		registry: registry,
		service:  service,
		opts:     opts,
	}
}

func (b ClusterBuilder) BuildUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if !isBroadCast(ctx) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		instances, err := b.registry.ListServices(ctx, b.service)
		if err != nil {
			return err
		}
		var eg errgroup.Group
		for _, instance := range instances {
			address := instance.Address
			eg.Go(func() error {
				// 1. too many NewClient() cuz slow
				// 2. need reuse instanceCC
				instanceCC, er := grpc.NewClient(address, b.opts...)
				if er != nil {
					return er
				}
				return invoker(ctx, method, req, reply, instanceCC, opts...)
			})
		}
		return eg.Wait()
	}
}

func UsedBroadCast(ctx context.Context) context.Context {
	return context.WithValue(ctx, broadcastKey{}, true)
}

type broadcastKey struct{}

func isBroadCast(ctx context.Context) bool {
	val, ok := ctx.Value(broadcastKey{}).(bool)
	return ok && val
}

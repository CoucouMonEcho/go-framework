package broadcast

import (
	"context"
	"fmt"
	"github.com/CoucouMonEcho/go-framework/micro/registry"
	"google.golang.org/grpc"
	"reflect"
	"sync"
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
		ok, ch := isBroadCast(ctx)
		if !ok {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		defer close(ch)
		instances, err := b.registry.ListServices(ctx, b.service)
		if err != nil {
			return err
		}
		var wg sync.WaitGroup
		wg.Add(len(instances))
		typ := reflect.TypeOf(reply).Elem()
		for _, instance := range instances {
			address := instance.Address
			go func() {
				// 1. too many NewClient() cuz slow
				// 2. need reuse instanceCC
				defer wg.Done()
				instanceCC, er := grpc.NewClient(address, b.opts...)
				if er != nil {
					ch <- Resp{Err: er}
					return
				}
				newReply := reflect.New(typ).Interface()
				er = invoker(ctx, method, req, newReply, instanceCC, opts...)
				select {
				case ch <- Resp{Reply: newReply}:
				case <-ctx.Done():
					err = fmt.Errorf("micro: broadcast all call timeout: %w", ctx.Err())
				}
			}()

		}
		wg.Wait()
		return nil
	}
}

func UsedBroadCast(ctx context.Context) (context.Context, <-chan Resp) {
	ch := make(chan Resp)
	return context.WithValue(ctx, broadcastKey{}, ch), ch
}

type broadcastKey struct{}

func isBroadCast(ctx context.Context) (bool, chan<- Resp) {
	val, ok := ctx.Value(broadcastKey{}).(chan Resp)
	return ok, val
}

type Resp struct {
	Err   error
	Reply any
}

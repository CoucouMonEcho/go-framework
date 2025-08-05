package micro

import (
	"context"
	"go-framework/micro/registry"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"time"
)

type grpcResolverBuilder struct {
	r       registry.Registry
	timeout time.Duration
}

//func NewResolverBuilder(r registry.Registry, timeout time.Duration) (*grpcResolverBuilder, error) {
//	return &grpcResolverBuilder{
//		r:       r,
//		timeout: timeout,
//	}, nil
//}

func (b *grpcResolverBuilder) Build(target resolver.Target,
	cc resolver.ClientConn,
	_ resolver.BuildOptions) (resolver.Resolver, error) {
	r := &grpcResolver{
		cc:      cc,
		r:       b.r,
		target:  target,
		timeout: b.timeout,
	}
	r.resolve()
	go r.watch()
	return r, nil
}

func (b *grpcResolverBuilder) Scheme() string {
	return "registry"
}

type grpcResolver struct {
	target  resolver.Target
	r       registry.Registry
	cc      resolver.ClientConn
	timeout time.Duration
	close   chan struct{}
}

func (g *grpcResolver) ResolveNow(_ resolver.ResolveNowOptions) {
	g.resolve()
}

func (g *grpcResolver) watch() {
	events, err := g.r.Subscribe(g.target.Endpoint())
	if err != nil {
		g.cc.ReportError(err)
		return
	}
	for {
		select {
		case <-events:
			g.resolve()
		//case event := <-events:
		//	switch event.Type {
		//	case "DELETE":
		//	case "ADD":
		//	}
		case <-g.close:
			return
		}
	}
}

func (g *grpcResolver) resolve() {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	instances, err := g.r.ListServices(ctx, g.target.Endpoint())
	if err != nil {
		g.cc.ReportError(err)
		return
	}
	addresses := make([]resolver.Address, 0, len(instances))
	for _, instance := range instances {
		addresses = append(addresses, resolver.Address{
			Addr: instance.Address,
			Attributes: attributes.New("weight", instance.Weight).
				WithValue("group", instance.Group),
		})
	}
	err = g.cc.UpdateState(resolver.State{
		Addresses: addresses,
	})
	if err != nil {
		g.cc.ReportError(err)
		return
	}
}

func (g *grpcResolver) Close() {
	close(g.close)
}

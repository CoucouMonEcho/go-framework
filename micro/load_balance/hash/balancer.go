package round_robin

import (
	"go-framework/micro/route"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

type Balancer struct {
	connes []balancer.SubConn
	length int
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if b.length == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	// can't get IP or other req info
	idx := info.Ctx.Value("hash_code").(int)
	conn := b.connes[idx]
	return balancer.PickResult{
		SubConn: conn,
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

var _ route.BalancerBuilder = &BalancerBuilder{}

type BalancerBuilder struct {
}

func (b *BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]balancer.SubConn, 0, len(info.ReadySCs))
	for conn := range info.ReadySCs {
		connections = append(connections, conn)
	}
	return &Balancer{
		connes: connections,
		length: len(connections),
	}
}

func (b *BalancerBuilder) Name() string {
	return "HASH"
}

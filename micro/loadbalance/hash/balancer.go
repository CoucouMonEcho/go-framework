package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

type Balancer struct {
	conns  []balancer.SubConn
	length int
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if b.length == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	// can't get IP or other req info
	idx := info.Ctx.Value("hash_code").(int)
	c := b.conns[idx]
	return balancer.PickResult{
		SubConn: c,
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

type Builder struct {
}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]balancer.SubConn, 0, len(info.ReadySCs))
	for conn := range info.ReadySCs {
		connections = append(connections, conn)
	}
	return &Balancer{
		conns:  connections,
		length: len(connections),
	}
}

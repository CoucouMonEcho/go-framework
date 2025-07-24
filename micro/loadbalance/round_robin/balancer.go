package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"sync/atomic"
)

type Balancer struct {
	index  int32
	conns  []balancer.SubConn
	length int32
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if b.length == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	idx := atomic.AddInt32(&b.index, 1)
	conn := b.conns[idx%b.length]
	return balancer.PickResult{
		SubConn: conn,
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

type BalancerBuilder struct {
}

func (b *BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]balancer.SubConn, 0, len(info.ReadySCs))
	for conn := range info.ReadySCs {
		conns = append(conns, conn)
	}
	return &Balancer{
		conns:  conns,
		index:  -1,
		length: int32(len(conns)),
	}
}

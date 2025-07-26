package round_robin

import (
	"code-practise/micro/route"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

type ConsistentHashBalancer struct {
	connes []balancer.SubConn
	length int
}

func (c *ConsistentHashBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if c.length == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	//FIXME can't get IP or other req info
	idx := info.Ctx.Value("hash_code").(int)
	conn := c.connes[idx]
	return balancer.PickResult{
		SubConn: conn,
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

var _ route.BalancerBuilder = &ConsistentHashBalancerBuilder{}

type ConsistentHashBalancerBuilder struct {
}

func (c *ConsistentHashBalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	//FIXME unable to establish consistent hash mapping information
	connections := make([]balancer.SubConn, 0, len(info.ReadySCs))
	for conn := range info.ReadySCs {
		connections = append(connections, conn)
	}
	return &ConsistentHashBalancer{
		connes: connections,
		length: len(connections),
	}
}

func (c *ConsistentHashBalancerBuilder) Name() string {
	return "CONSISTENT_HASH"
}

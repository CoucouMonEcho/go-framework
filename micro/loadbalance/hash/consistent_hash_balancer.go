package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

type ConsistentHashBalancer struct {
	conns  []balancer.SubConn
	length int
}

func (c *ConsistentHashBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if c.length == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	//FIXME can't get IP or other req info
	idx := info.Ctx.Value("hash_code").(int)
	conn := c.conns[idx]
	return balancer.PickResult{
		SubConn: conn,
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

type ConsistentHashBalancerBuilder struct {
}

func (c *ConsistentHashBalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	//FIXME unable to establish consistent hash mapping information
	connections := make([]balancer.SubConn, 0, len(info.ReadySCs))
	for conn := range info.ReadySCs {
		connections = append(connections, conn)
	}
	return &ConsistentHashBalancer{
		conns:  connections,
		length: len(connections),
	}
}

package random

import (
	"code-practise/micro/route"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"math/rand"
)

type Balancer struct {
	connes []*subConn
	filter route.Filter
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	candidates := make([]*subConn, 0, len(b.connes))
	for _, conn := range b.connes {
		if b.filter != nil && !b.filter(info, conn.addr) {
			continue
		}
		candidates = append(candidates, conn)
	}
	if len(candidates) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	idx := rand.Intn(len(candidates))
	conn := b.connes[idx]
	return balancer.PickResult{
		SubConn: conn.conn,
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

var _ route.BalancerBuilder = &BalancerBuilder{}

type BalancerBuilder struct {
	Filter route.Filter
}

func (b *BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]*subConn, 0, len(info.ReadySCs))
	for conn, connInfo := range info.ReadySCs {
		connections = append(connections, &subConn{
			conn: conn,
			addr: connInfo.Address,
		})
	}
	return &Balancer{
		connes: connections,
		filter: b.Filter,
	}
}

func (b *BalancerBuilder) Name() string {
	return "RANDOM"
}

type subConn struct {
	conn balancer.SubConn
	addr resolver.Address
}

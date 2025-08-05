package round_robin

import (
	"github.com/CoucouMonEcho/go-framework/micro/route"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"sync/atomic"
)

type Balancer struct {
	index  int32
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
		// can also return to the default node
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	idx := atomic.AddInt32(&b.index, 1)
	conn := candidates[idx%int32(len(candidates))]
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
	connes := make([]*subConn, 0, len(info.ReadySCs))
	for conn, connInfo := range info.ReadySCs {
		connes = append(connes, &subConn{
			conn: conn,
			addr: connInfo.Address,
		})
	}
	return &Balancer{
		connes: connes,
		index:  -1,
		filter: b.Filter,
	}
}

func (b *BalancerBuilder) Name() string {
	return "ROUND_ROBIN"
}

type subConn struct {
	conn balancer.SubConn
	addr resolver.Address
}

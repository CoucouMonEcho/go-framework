package round_robin

import (
	"code-practise/micro/route"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"sync/atomic"
)

type Balancer struct {
	index  int32
	conns  []*subConn
	length int32
	filter route.Filter
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	candidates := make([]*subConn, 0, len(b.conns))
	for _, conn := range b.conns {
		if b.filter != nil && b.filter(info, conn.addr) {
			candidates = append(candidates, conn)
		}
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

type BalancerBuilder struct {
	Filter route.Filter
}

func (b *BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*subConn, 0, len(info.ReadySCs))
	for conn, connInfo := range info.ReadySCs {
		conns = append(conns, &subConn{
			conn: conn,
			addr: connInfo.Address,
		})
	}
	return &Balancer{
		conns:  conns,
		index:  -1,
		length: int32(len(conns)),
		filter: b.Filter,
	}
}

type subConn struct {
	conn balancer.SubConn
	addr resolver.Address
}

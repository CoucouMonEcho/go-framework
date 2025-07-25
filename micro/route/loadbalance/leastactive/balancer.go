package leastactive

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math"
	"sync/atomic"
)

type Balancer struct {
	conns []*activeConn
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(b.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	res := &activeConn{
		cnt: math.MaxUint32,
	}
	for _, conn := range b.conns {
		if atomic.LoadUint32(&conn.cnt) < res.cnt {
			res = conn
		}
	}
	if res == nil {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	atomic.AddUint32(&res.cnt, 1)
	return balancer.PickResult{
		SubConn: res.conn,
		Done: func(info balancer.DoneInfo) {
			atomic.AddUint32(&res.cnt, -1)
		},
	}, nil
}

type BalancerBuilder struct {
}

func (b *BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*activeConn, 0, len(info.ReadySCs))
	for conn := range info.ReadySCs {
		conns = append(conns, &activeConn{
			conn: conn,
		})
	}
	return &Balancer{
		conns: conns,
	}
}

type activeConn struct {
	cnt  uint32
	conn balancer.SubConn
}

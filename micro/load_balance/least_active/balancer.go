package least_active

import (
	"go-framework/micro/route"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math"
	"sync/atomic"
)

type Balancer struct {
	connes []*activeConn
	//TODO filter
}

func (b *Balancer) Pick(_ balancer.PickInfo) (balancer.PickResult, error) {
	if len(b.connes) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	res := &activeConn{
		cnt: math.MaxUint32,
	}
	for _, conn := range b.connes {
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

var _ route.BalancerBuilder = &BalancerBuilder{}

type BalancerBuilder struct {
}

func (b *BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	connes := make([]*activeConn, 0, len(info.ReadySCs))
	for conn := range info.ReadySCs {
		connes = append(connes, &activeConn{
			conn: conn,
		})
	}
	return &Balancer{
		connes: connes,
	}
}

func (b *BalancerBuilder) Name() string {
	return "LEAST_ACTIVE"
}

type activeConn struct {
	cnt  uint32
	conn balancer.SubConn
}

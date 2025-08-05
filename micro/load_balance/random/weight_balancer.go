package random

import (
	"go-framework/micro/route"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"math/rand"
)

type WeightBalancer struct {
	connes []*weightConn
	filter route.Filter
}

func (w *WeightBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	candidates := make([]*weightConn, 0, len(w.connes))
	var totalWeight uint32
	for _, conn := range w.connes {
		if w.filter != nil && !w.filter(info, conn.addr) {
			continue
		}
		candidates = append(candidates, conn)
		totalWeight += conn.weight
	}
	if len(candidates) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	tgt := rand.Intn(int(totalWeight) + 1)
	var idx int
	for i, conn := range candidates {
		if w.filter != nil && !w.filter(info, conn.addr) {
			continue
		}
		tgt -= int(conn.weight)
		if tgt <= 0 {
			idx = i
			break
		}
	}
	res := candidates[idx]
	return balancer.PickResult{
		SubConn: res.conn,
		Done: func(info balancer.DoneInfo) {
			// weight modified is unnecessary with random
		},
	}, nil
}

var _ route.BalancerBuilder = &WeightBalancerBuilder{}

type WeightBalancerBuilder struct {
	Filter route.Filter
}

func (w *WeightBalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	connes := make([]*weightConn, 0, len(info.ReadySCs))
	for conn, connInfo := range info.ReadySCs {
		//weight, err := strconv.ParseInt(weightStr, 10, 64)
		//if err != nil {
		//	panic(err)
		//}
		weight := connInfo.Address.Attributes.Value("weight").(uint32)
		connes = append(connes, &weightConn{
			conn:   conn,
			weight: weight,
			addr:   connInfo.Address,
		})
	}
	return &WeightBalancer{
		connes: connes,
		filter: w.Filter,
	}
}

func (w *WeightBalancerBuilder) Name() string {
	return "WEIGHT_RANDOM"
}

type weightConn struct {
	conn   balancer.SubConn
	weight uint32
	addr   resolver.Address
}

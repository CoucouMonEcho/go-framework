package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math/rand"
)

type WeightBalancer struct {
	conns       []*weightConn
	totalWeight uint32
}

func (w *WeightBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(w.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	tgt := rand.Intn(int(w.totalWeight) + 1)
	var idx int
	for i, conn := range w.conns {
		tgt -= int(conn.weight)
		if tgt <= 0 {
			idx = i
			break
		}
	}
	res := w.conns[idx]
	return balancer.PickResult{
		SubConn: res.conn,
		Done: func(info balancer.DoneInfo) {
			// weight modified is unnecessary with random
		},
	}, nil
}

type WeightBalancerBuilder struct {
}

func (w *WeightBalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*weightConn, 0, len(info.ReadySCs))
	var totalWeight uint32
	for conn, connInfo := range info.ReadySCs {
		//weight, err := strconv.ParseInt(weightStr, 10, 64)
		//if err != nil {
		//	panic(err)
		//}
		weight := connInfo.Address.Attributes.Value("weight").(uint32)
		totalWeight += weight
		conns = append(conns, &weightConn{
			conn:   conn,
			weight: weight,
		})
	}
	return &WeightBalancer{
		conns:       conns,
		totalWeight: totalWeight,
	}
}

type weightConn struct {
	conn   balancer.SubConn
	weight uint32
}

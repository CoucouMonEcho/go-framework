package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math"
	"sync"
	"sync/atomic"
)

type WeightBalancer struct {
	conns []*weightConn
}

func (w *WeightBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(w.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	var totalWeight uint32
	var res *weightConn

	for _, conn := range w.conns {
		conn.mutex.Lock()
		totalWeight += conn.currentWeight
		conn.currentWeight += conn.efficientWeight
		conn.mutex.Unlock()
		if res == nil || res.currentWeight < conn.currentWeight {
			res = conn
		}
	}
	if res == nil {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	res.mutex.Lock()
	res.currentWeight -= totalWeight
	res.mutex.Unlock()
	return balancer.PickResult{
		SubConn: res.conn,
		Done: func(info balancer.DoneInfo) {
			res.mutex.Lock()
			weight := atomic.LoadUint32(&res.efficientWeight)
			if info.Err != nil && weight == 0 {
				return
			}
			if info.Err == nil && weight == math.MaxUint32 {
				return
			}
			if info.Err == nil {
				res.efficientWeight++
			} else {
				res.efficientWeight--
			}
			res.mutex.Unlock()

			//for {
			//	weight := atomic.LoadUint32(&res.efficientWeight)
			//	if info.Err != nil && weight == 0 {
			//		return
			//	}
			//	if info.Err == nil && weight == math.MaxUint32 {
			//		return
			//	}
			//	newWeight := weight
			//	if info.Err == nil {
			//		newWeight++
			//	} else {
			//		newWeight--
			//	}
			//	if atomic.CompareAndSwapUint32(&(res.efficientWeight), weight, newWeight) {
			//		return
			//	}
			//}
		},
	}, nil
}

func (w *WeightBalancer) done(conn *weightConn) func(info balancer.DoneInfo) {
	return func(info balancer.DoneInfo) {

	}
}

type WeightBalancerBuilder struct {
}

func (w *WeightBalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]*weightConn, 0, len(info.ReadySCs))
	for conn, connInfo := range info.ReadySCs {
		//weight, err := strconv.ParseInt(weightStr, 10, 64)
		//if err != nil {
		//	panic(err)
		//}
		weight := connInfo.Address.Attributes.Value("weight").(uint32)
		connections = append(connections, &weightConn{
			conn:            conn,
			weight:          weight,
			currentWeight:   weight,
			efficientWeight: weight,
		})
	}
	return &WeightBalancer{
		conns: connections,
	}
}

type weightConn struct {
	conn            balancer.SubConn
	weight          uint32
	currentWeight   uint32
	efficientWeight uint32
	mutex           sync.Mutex
}

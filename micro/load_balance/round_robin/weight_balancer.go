package round_robin

import (
	"github.com/CoucouMonEcho/go-framework/micro/route"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"math"
	"sync"
	"sync/atomic"
)

type WeightBalancer struct {
	connes []*weightConn
	filter route.Filter
}

func (w *WeightBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(w.connes) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	var totalWeight uint32
	var res *weightConn

	for _, conn := range w.connes {
		if w.filter != nil && !w.filter(info, conn.addr) {
			continue
		}
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

func (w *WeightBalancer) done(_ *weightConn) func(info balancer.DoneInfo) {
	return func(info balancer.DoneInfo) {

	}
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
			conn:            conn,
			weight:          weight,
			currentWeight:   weight,
			efficientWeight: weight,
			addr:            connInfo.Address,
		})
	}
	return &WeightBalancer{
		connes: connes,
	}
}

func (w *WeightBalancerBuilder) Name() string {
	return "WEIGHT_ROUND_ROBIN"
}

type weightConn struct {
	conn            balancer.SubConn
	weight          uint32
	currentWeight   uint32
	efficientWeight uint32
	mutex           sync.Mutex
	addr            resolver.Address
}

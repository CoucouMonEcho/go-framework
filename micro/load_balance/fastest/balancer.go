package fastest

import (
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type Balancer struct {
	mutex    sync.RWMutex
	connes   []*subConn
	lastSync time.Time
	endpoint string
}

func (b *Balancer) Pick(_ balancer.PickInfo) (balancer.PickResult, error) {
	b.mutex.RLock()
	if len(b.connes) == 0 {
		b.mutex.RUnlock()
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	var res *subConn
	for _, conn := range b.connes {
		if res == nil || res.respTime > conn.respTime {
			res = conn
		}
	}
	b.mutex.RUnlock()

	return balancer.PickResult{
		SubConn: res,
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

func (b BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	connes := make([]*subConn, 0, len(info.ReadySCs))
	for conn, connInfo := range info.ReadySCs {
		connes = append(connes, &subConn{
			SubConn:  conn,
			addr:     connInfo.Address,
			respTime: time.Millisecond * 100,
		})
	}
	res := &Balancer{
		connes: connes,
		//filter: flt
	}

	// non close(), use runtime.SetFinalizer() to end go routine
	ch := make(chan struct{}, 1)
	runtime.SetFinalizer(res, func() {
		ch <- struct{}{}
	})
	go func() {
		ticker := time.NewTicker(b.Interval)
		for {
			select {
			case <-ticker.C:
				res.updateRespTime(b.Endpoint, b.Query)
			case <-ch:
				return
			}
		}
	}()

	return res
}

func (b *Balancer) updateRespTime(endpoint, query string) {
	httpResp, err := http.Get(fmt.Sprintf("%s/api/v1/query?query=%s", endpoint, query))
	if err != nil {
		log.Fatalln("micro: load balance update resp time connect fail:", err)
		return
	}
	//body, err := ioutil.ReadAll(httpResp.Body)
	//if err != nil {
	//	return
	//}
	//log.Println(string(body))
	decoder := json.NewDecoder(httpResp.Body)

	var resp response
	err = decoder.Decode(&resp)
	if err != nil {
		log.Fatalln("micro: load balance update resp time decode fail:", err)
		return
	}
	if resp.Status != "success" {
		log.Fatalln("micro: load balance update resp time bad response")
		return
	}
	for _, promRes := range resp.Data.Result {
		address, ok := promRes.Metric["address"]
		if !ok {
			continue
		}
		for _, conn := range b.connes {
			if conn.addr.Addr == address {
				ms, er := strconv.ParseInt(promRes.Value[1].(string), 10, 64)
				if er != nil {
					continue
				}
				conn.respTime = time.Duration(ms) * time.Millisecond
			}
		}
	}

}

type BalancerBuilder struct {
	// prometheus address
	Endpoint string
	Query    string
	// update duration
	Interval time.Duration
}

type subConn struct {
	balancer.SubConn
	addr     resolver.Address
	respTime time.Duration
}

type response struct {
	Status string `json:"status"`
	Data   data   `json:"data"`
}

type data struct {
	ResultType string   `json:"resultType"`
	Result     []Result `json:"result"`
}

type Result struct {
	Metric map[string]string `json:"metric"`
	Value  []any             `json:"value"`
}

package metrics

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"net"
	"time"
)

type ServerMetricsBuilder struct {
	Namespace string
	Subsystem string
	Port      int
}

func (s *ServerMetricsBuilder) Build() grpc.UnaryServerInterceptor {
	address := getOutboundIP()
	if s.Port != 0 {
		address = fmt.Sprintf("%s:%d", address, s.Port)
	}

	reqGuage := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: s.Namespace,
		Subsystem: s.Subsystem,
		Name:      "active_request_cnt",
		Help:      "Number of requests currently active",
		ConstLabels: map[string]string{
			"component": "server",
			"address":   address,
			//...
		},
	}, []string{"service"})
	prometheus.MustRegister(reqGuage)

	errCnt := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: s.Namespace,
		Subsystem: s.Subsystem,
		Name:      "err_request_cnt",
		Help:      "Number of requests that have failed",
		ConstLabels: map[string]string{
			"component": "server",
			"address":   address,
			//...
		},
	}, []string{"service"})
	prometheus.MustRegister(errCnt)

	respSummary := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: s.Namespace,
		Subsystem: s.Subsystem,
		Name:      "active_response_time",
		Help:      "Response time of requests currently active",
		ConstLabels: map[string]string{
			"component": "server",
			"address":   address,
			//...
		},
	}, []string{"service"})
	prometheus.MustRegister(respSummary)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		startTime := time.Now()
		reqGuage.WithLabelValues(info.FullMethod).Add(1)
		defer func() {
			reqGuage.WithLabelValues(info.FullMethod).Add(-1)
			if err != nil {
				errCnt.WithLabelValues(info.FullMethod).Add(-1)
			}
			respSummary.WithLabelValues(info.FullMethod).Observe(float64(time.Since(startTime).Milliseconds()))
		}()
		resp, err = handler(ctx, req)
		return
	}
}

func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer func() { _ = conn.Close() }()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

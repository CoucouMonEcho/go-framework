package prometheus

import (
	"code-practise/web"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
)

type MiddlewareBuilder struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
}

func (m MiddlewareBuilder) Build() web.Middleware {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: m.Namespace,
		Subsystem: m.Subsystem,
		Name:      m.Name,
		Help:      m.Help,
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.90:  0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, []string{"pattern", "method", "status"})
	prometheus.MustRegister(vector)
	return func(next web.HandlerFunc) web.HandlerFunc {
		return func(ctx *web.Context) {
			start := time.Now()
			defer func() {
				go report(start, ctx, vector)
			}()
			next(ctx)
		}
	}
}

func report(start time.Time, ctx *web.Context, vector *prometheus.SummaryVec) {
	pattern := ctx.MatchedRoute
	if pattern == "" {
		pattern = "unknown"
	}
	vector.WithLabelValues(
		pattern,
		ctx.Req.Method,
		strconv.Itoa(ctx.RespCode),
	).Observe(float64(time.Since(start).Milliseconds()))
}

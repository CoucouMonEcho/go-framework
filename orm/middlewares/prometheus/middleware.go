package prometheus

import (
	"context"
	"github.com/CoucouMonEcho/go-framework/orm"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type MiddlewareBuilder struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
}

func NewMiddlewareBuilder(namespace, subsystem, name, help string) *MiddlewareBuilder {
	return &MiddlewareBuilder{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	}
}

func (m MiddlewareBuilder) Build() orm.Middleware {
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
	}, []string{"type", "table"})
	prometheus.MustRegister(vector)

	// err counter, histogram, active

	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			start := time.Now()
			defer func() {
				go report(start, qc, vector)
			}()
			res := next(ctx, qc)
			return res
		}
	}
}

func report(start time.Time, qc *orm.QueryContext, vector *prometheus.SummaryVec) {
	vector.WithLabelValues(
		qc.Type,
		qc.Model.TableName,
	).Observe(float64(time.Since(start).Milliseconds()))
}

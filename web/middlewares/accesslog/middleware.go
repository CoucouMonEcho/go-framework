package accesslog

import (
	"code-practise/web"
	"encoding/json"
	"fmt"
)

type MiddlewareBuilder struct {
	logFunc func(log string)
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{
		logFunc: func(log string) {
			fmt.Println(log)
		},
	}
}

func (m *MiddlewareBuilder) LogFunc(logFunc func(log string)) *MiddlewareBuilder {
	m.logFunc = logFunc
	return m
}

func (m *MiddlewareBuilder) Build() web.Middleware {
	return func(next web.Handler) web.Handler {
		return func(ctx *web.Context) {
			// finish route handler to get matched route
			defer func() {
				data, _ := json.Marshal(AccessLog{
					Host:       ctx.Req.Host,
					Route:      ctx.MatchedRoute,
					HTTPMethod: ctx.Req.Method,
					Path:       ctx.Req.URL.Path,
				})
				m.logFunc(string(data))
			}()
			next(ctx)
		}
	}
}

type AccessLog struct {
	Host       string `json:"host,omitempty"`
	Route      string `json:"route,omitempty"`
	HTTPMethod string `json:"http_method,omitempty"`
	Path       string `json:"path,omitempty"`
}

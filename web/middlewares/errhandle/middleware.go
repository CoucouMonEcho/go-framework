package errhandle

import "code-practise/web"

type MiddlewareBuilder struct {
	// static page
	resp map[int][]byte
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{
		resp: make(map[int][]byte, 64),
	}
}

func (m *MiddlewareBuilder) RegisterError(status int, data []byte) *MiddlewareBuilder {
	m.resp[status] = data
	return m
}

func (m MiddlewareBuilder) Build() web.Middleware {
	return func(next web.Handler) web.Handler {
		return func(ctx *web.Context) {
			next(ctx)
			resp, ok := m.resp[ctx.RespCode]
			if ok {
				// static page
				ctx.RespData = resp
			}
		}
	}
}

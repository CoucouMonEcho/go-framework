package recover

import (
	"code-practise/web"
)

type MiddlewareBuilder struct {
	Code int
	Data []byte
	// log func(err any)
	Log func(ctx *web.Context)
	// log func(stack string)
}

func (m MiddlewareBuilder) Build() web.Middleware {
	return func(next web.Handler) web.Handler {
		return func(ctx *web.Context) {
			defer func() {
				if err := recover(); err != nil {
					ctx.RespData = m.Data
					ctx.RespCode = m.Code
					m.Log(ctx)
				}
			}()
			next(ctx)
		}
	}
}

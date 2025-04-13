// Package session
// in fact, session should not belong to web framework
// under the web package now because it is only used for practice
package test

import (
	"code-practise/web"
	"code-practise/web/session"
	"code-practise/web/session/cookie"
	"code-practise/web/session/memory"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestHttpServer(t *testing.T) {
	var m *session.Manager = &session.Manager{
		Propagator: cookie.NewPropagator(),
		Store:      memory.NewStore(time.Minute * 15),
		CtxSessKey: "sessionKey",
	}
	// login session
	server := web.NewHTTPServer(web.ServerWithMiddlewares(func(next web.HandlerFunc) web.HandlerFunc {
		return func(ctx *web.Context) {
			if ctx.Req.URL.Path != "/login" {
				_, err := m.GetSession(ctx)
				if err != nil {
					ctx.RespCode = http.StatusUnauthorized
					ctx.RespData = []byte(err.Error())
					return
				}
				// refresh
				err = m.RefreshSession(ctx)
				if err != nil {
					log.Println("refresh session error", err)
				}
			}
			next(ctx)
		}
	}))
	server.Get("/login", func(ctx *web.Context) {

		// check auth

		// session
		sess, err := m.InitSession(ctx)
		if err != nil {
			ctx.RespCode = http.StatusInternalServerError
			ctx.RespData = []byte(err.Error())
			return
		}
		err = sess.Set(ctx.Req.Context(), "name", "john")
		if err != nil {
			ctx.RespCode = http.StatusInternalServerError
			ctx.RespData = []byte(err.Error())
			return
		}
		ctx.RespCode = http.StatusOK
		ctx.RespData = []byte("ok")
		return
	})
	server.Get("/logout", func(ctx *web.Context) {

		// delete data

		// session
		err := m.RemoveSession(ctx)
		if err != nil {
			ctx.RespCode = http.StatusInternalServerError
			ctx.RespData = []byte(err.Error())
			return
		}
		// redirect
		ctx.RespCode = http.StatusOK
		ctx.RespData = []byte("ok")
		return
	})
	server.Get("/user", func(ctx *web.Context) {
		sess, err := m.GetSession(ctx)
		if err != nil {
			ctx.RespCode = http.StatusInternalServerError
			ctx.RespData = []byte(err.Error())
			return
		}
		val, err := sess.Get(ctx.Req.Context(), "name")
		if err != nil {
			ctx.RespCode = http.StatusInternalServerError
			ctx.RespData = []byte(err.Error())
			return
		}
		ctx.RespCode = http.StatusOK
		ctx.RespData = []byte(val.(string))
		return
	})
	server.Start(":8081")
}

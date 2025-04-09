// Package session
// in fact, session should not belong to web framework
// under the web package now because it is only used for practice
package session

import (
	"code-practise/web"
	"log"
	"net/http"
	"testing"
)

func TestHttpServer(t *testing.T) {
	var m Manager
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
		sess, _ := m.GetSession(ctx)
		sess.Get(ctx.Req.Context(), "name")
	})
	server.Start(":8081")
}

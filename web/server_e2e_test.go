//go:build e2e

package web

import (
	"testing"
)

func TestServer(t *testing.T) {
	h := NewHTTPServer()
	type User struct {
		Name string `json:"name"`
	}
	h.Get("/user", func(ctx *Context) {
		ctx.RespJSON(200, User{
			Name: "dsahjkfdsahdls",
		})
	})
	h.Start(":8080")
}

package web

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

var r router
var mockHandler Handler

func init() {
	testRoutes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/"},
		{http.MethodGet, "/order/detail"},
		{http.MethodGet, "/order/*"},
		{http.MethodGet, "/order/*/:id"},
		{http.MethodGet, "/order/detail/:orderId([0-9]+)"},
		{http.MethodGet, "/order/detail/:orderId2([a-z]+)"},
	}

	mockHandler = func(ctx *Context) {

	}
	r = newRouter()
	for _, testRoute := range testRoutes {
		r.addRoute(testRoute.method, testRoute.path, mockHandler)
	}
}

func TestRouter_AddRoute(t *testing.T) {

	wantRouter := router{
		trees: map[string]*node{
			http.MethodGet: {
				path:     "/",
				nodeType: nodeTypeStatic,
				handler:  mockHandler,
				children: []*node{
					{
						path:     "order",
						nodeType: nodeTypeStatic,
						children: []*node{
							{
								path:     "detail",
								nodeType: nodeTypeStatic,
								handler:  mockHandler,
								children: []*node{
									{
										path:     ":orderId",
										nodeType: nodeTypeRegular,
										handler:  mockHandler,
									},
									{
										path:     ":orderId2",
										nodeType: nodeTypeRegular,
										handler:  mockHandler,
									},
								},
							},
							{
								path:     "*",
								nodeType: nodeTypeWildcard,
								handler:  mockHandler,
								children: []*node{
									{
										path:     ":id",
										nodeType: nodeTypePathParam,
										handler:  mockHandler,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	msg, ok := wantRouter.equal(&r)
	assert.True(t, ok, msg)

	r2 := newRouter()
	r2.addRoute(http.MethodGet, "/*", mockHandler)
	assert.Panics(t, func() {
		r2.addRoute(http.MethodGet, "/:id", mockHandler)
	})
}

func TestRouter_getRoute(t *testing.T) {
	testCases := []struct {
		name      string
		method    string
		path      string
		wantFound bool
		info      *matchInfo
	}{
		{"method not exist", http.MethodOptions, "/order/detail", false, nil},
		{"static", http.MethodGet, "/order/detail", true, &matchInfo{node: &node{path: "detail", nodeType: nodeTypeStatic, handler: mockHandler, children: []*node{{path: ":orderId", nodeType: nodeTypeRegular, handler: mockHandler}}}}},
		{"wild card", http.MethodGet, "/order/aaa", true, &matchInfo{node: &node{path: "*", nodeType: nodeTypeWildcard, handler: mockHandler, children: []*node{{path: ":id", nodeType: nodeTypePathParam, handler: mockHandler}}}}},
		{"path param", http.MethodGet, "/order/detail/123o", true, &matchInfo{node: &node{path: ":orderId", nodeType: nodeTypeRegular, handler: mockHandler}, pathParams: map[string]string{"orderId": "123o"}}},
		{"path param", http.MethodGet, "/order/detail/123o", true, &matchInfo{node: &node{path: ":orderId2", nodeType: nodeTypeRegular, handler: mockHandler}, pathParams: map[string]string{"orderId2": "123o"}}},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			info, found := r.route(testCase.method, testCase.path)
			assert.Equal(t, testCase.wantFound, found)
			if !found {
				return
			}
			msg, ok := testCase.info.node.equal(info.node)
			assert.True(t, ok, msg)
			assert.Equal(t, testCase.info.pathParams, info.pathParams)
		})
	}
}

func (r *router) equal(r2 *router) (string, bool) {
	if r == r2 {
		return "", true
	}
	if len(r.trees) != len(r2.trees) {
		return "can not found method", false
	}
	for k, v := range r.trees {
		dst, ok := r2.trees[k]
		if !ok {
			return "can not found method", false
		}
		msg, ok := v.equal(dst)
		if !ok {
			return msg, false
		}
	}
	return "", true
}

func (n *node) equal(n2 *node) (string, bool) {
	if n == n2 {
		return "", true
	}
	if n.path != n2.path {
		return "path not match", false
	}
	if len(n.children) != len(n2.children) {
		return "children len not match", false
	}
	if n.nodeType != n2.nodeType {
		return "node type not match", false
	}
	if reflect.ValueOf(n.handler) != reflect.ValueOf(n2.handler) {
		return "handler not match", false
	}
	for _, n1Child := range n.children {
		found := false
		var msg string
		var ok bool
		for _, n2Child := range n2.children {
			msg, ok = n1Child.equal(n2Child)
			if ok {
				found = true
			}
		}
		if !found {
			return msg, false
		}
	}
	return "", true
}

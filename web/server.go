package web

import (
	"log"
	"net"
	"net/http"
)

type HandlerFunc func(ctx *Context)

type Server interface {
	http.Handler
	Start(addr string) error

	addRoute(method string, path string, handlerFunc HandlerFunc)
}

type HTTPServerOption func(server *HTTPServer)

var _ Server = &HTTPServer{}

type HTTPServer struct {
	router
	middlewares    []Middleware
	logger         func(msg string, args ...any)
	templateEngine TemplateEngine
}

func NewHTTPServer(opts ...HTTPServerOption) *HTTPServer {
	res := &HTTPServer{
		router: newRouter(),
		logger: func(msg string, args ...any) {
			log.Fatalf(msg, args)
		},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func ServerWithMiddlewares(middlewares ...Middleware) HTTPServerOption {
	return func(server *HTTPServer) {
		server.middlewares = middlewares
	}
}

func ServerWithLogger(logger func(msg string, args ...any)) HTTPServerOption {
	return func(server *HTTPServer) {
		server.logger = logger
	}
}

func ServerWithTemplateEngine(templateEngine TemplateEngine) HTTPServerOption {
	return func(server *HTTPServer) {
		server.templateEngine = templateEngine
	}
}

func (h *HTTPServer) Get(path string, handlerFunc HandlerFunc) {
	h.router.addRoute(http.MethodGet, path, handlerFunc)
}

func (h *HTTPServer) Post(path string, handlerFunc HandlerFunc) {
	h.router.addRoute(http.MethodPost, path, handlerFunc)
}

// ServeHTTP deal request
func (h *HTTPServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	ctx := &Context{
		Req:            req,
		Resp:           resp,
		templateEngine: h.templateEngine,
	}
	root := h.serve
	for i := len(h.middlewares) - 1; i >= 0; i-- {
		root = h.middlewares[i](root)
	}
	m := func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) {
			next(ctx)
			h.flushResp(ctx)
		}
	}
	root = m(root)
	root(ctx)
}

func (h *HTTPServer) flushResp(ctx *Context) {
	if ctx.RespCode != 0 {
		ctx.Resp.WriteHeader(ctx.RespCode)
	}
	n, err := ctx.Resp.Write(ctx.RespData)
	if err != nil || n != len(ctx.RespData) {
		h.logger("web: failed to write response body %v", err)
	}
}

func (h *HTTPServer) serve(ctx *Context) {
	info, ok := h.router.route(ctx.Req.Method, ctx.Req.URL.Path)
	if !ok {
		ctx.RespCode = http.StatusNotFound
		ctx.RespData = []byte("404 page not found")
		return
	}
	ctx.pathParams = info.pathParams
	ctx.MatchedRoute = info.n.route
	info.n.handler(ctx)
	return
}

func (h *HTTPServer) Start(addr string) error {
	listenr, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return http.Serve(listenr, h)
}

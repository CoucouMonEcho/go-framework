package web

type Middleware func(next HandlerFunc) HandlerFunc

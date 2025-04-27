package web

type Middleware func(next Handler) Handler

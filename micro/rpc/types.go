package rpc

import (
	"context"
	"github.com/CoucouMonEcho/go-framework/micro/rpc/message"
)

type Service interface {
	Name() string
}

// mockgen -destination=micro/rpc/mocks/rpc_proxy.gen_test.go -package=mocks -package=rpc -source=micro/rpc/types.go Proxy

type Proxy interface {
	Invoke(ctx context.Context, req *message.Request) (*message.Response, error)
}

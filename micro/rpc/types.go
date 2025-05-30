package rpc

import "context"

type Service interface {
	Name() string
}

//  mockgen -destination=micro/rpc/mocks/rpc_proxy.gen_test.go -package=mocks -package=rpc -source=micro/rpc/types.go Proxy

type Proxy interface {
	Invoke(ctx context.Context, req *Request) (*Response, error)
}

type Request struct {
	ServiceName string
	MethodName  string
	//Arg         any
	// Arg  use any can not confirm type
	Arg []byte
}

type Response struct {
	Data []byte
}

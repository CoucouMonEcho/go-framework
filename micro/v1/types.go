package v1

import "context"

type Service interface {
	Name() string
}

//  mockgen -destination=micro/v1/mocks/rpc_proxy.gen_test.go -package=mocks -package=v1 -source=micro/v1/types.go Proxy

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

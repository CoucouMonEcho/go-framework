package registry

import (
	"context"
	"io"
)

type Registry interface {
	Register(ctx context.Context, si ServiceInstance) error
	UnRegister(ctx context.Context, si ServiceInstance) error
	LisServices(ctx context.Context, name string) ([]ServiceInstance, error)
	Subscribe(name string) (<-chan Event, error)
	io.Closer
}

type ServiceInstance struct {
	Name    string
	Address string
}

type Event struct {
	Type string
}

package micro

import (
	"code-practise/micro/registry"
	"context"
	"google.golang.org/grpc"
	"net"
	"time"
)

type ServerOption func(*Server)

type Server struct {
	name            string
	registry        registry.Registry
	registerTimeout time.Duration
	*grpc.Server
	listener net.Listener
	si       registry.ServiceInstance
	weight   uint32
}

func NewServer(name string, opts ...ServerOption) (*Server, error) {
	res := &Server{
		name:            name,
		Server:          grpc.NewServer(),
		registerTimeout: 10 * time.Second,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func (s *Server) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = listener
	if s.registry != nil {
		// register server
		s.si = registry.ServiceInstance{
			Name: s.name,
			//FIXME the container returns 127.0.0.1
			Address: listener.Addr().String(),
			Weight:  s.weight,
		}
		ctx, cancel := context.WithTimeout(context.Background(), s.registerTimeout)
		defer cancel()
		err = s.registry.Register(ctx, s.si)
		if err != nil {
			return err
		}
		//defer func() {
		//	_ = s.registry.Close()
		//	_ = s.registry.UnRegister(ctx, s.si)
		//}()
	}
	err = s.Serve(listener)
	return err
}

func (s *Server) Close() error {
	if s.registry != nil {
		err := s.registry.Close()
		if err != nil {
			return err
		}
	}
	s.GracefulStop()
	return nil
}

func ServerWithRegistry(registry registry.Registry) ServerOption {
	return func(server *Server) {
		server.registry = registry
	}
}

func ServerWithWeight(weight uint32) ServerOption {
	return func(server *Server) {
		server.weight = weight
	}
}

package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"reflect"
)

type Server struct {
	services map[string]*reflectionStub
}

func NewServer() *Server {
	return &Server{
		services: make(map[string]*reflectionStub, 4),
	}
}

func (s *Server) RegisterService(service Service) {
	s.services[service.Name()] = &reflectionStub{
		s:     service,
		value: reflect.ValueOf(service),
	}
}

func (s *Server) Start(network, addr string) error {
	listener, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			if er := s.handleConn(conn); er != nil {
				_ = conn.Close()
			}
		}()
	}
}

func (s *Server) handleConn(conn net.Conn) error {
	for {
		// read
		reqBs, err := ReadMsg(conn)
		if err != nil {
			return err
		}

		req := &Request{}
		err = json.Unmarshal(reqBs, req)
		if err != nil {
			return err
		}
		resp, err := s.Invoke(context.Background(), req)
		if err != nil {
			// business err
		}

		// write message body
		_, err = conn.Write(EncodeMsg(resp.Data))
		if err != nil {
			return err
		}
	}
}

func (s *Server) Invoke(ctx context.Context, req *Request) (*Response, error) {
	stub, ok := s.services[req.ServiceName]
	if !ok {
		return nil, errors.New("rpc: service not found")
	}
	// service
	data, err := stub.invoke(ctx, req.MethodName, req.Arg)
	if err != nil {
		return nil, err
	}
	return &Response{
		Data: data,
	}, nil
}

// reflectionStub consider using unsafe in the future
type reflectionStub struct {
	s     Service
	value reflect.Value
}

func (s *reflectionStub) invoke(ctx context.Context, methodName string, data []byte) ([]byte, error) {
	// method
	method := s.value.MethodByName(methodName)
	// arg
	in := reflect.New(method.Type().In(1).Elem())
	err := json.Unmarshal(data, in.Interface())
	if err != nil {
		return nil, err
	}
	// call
	out := method.Call([]reflect.Value{reflect.ValueOf(ctx), in})
	if out[1].Interface() != nil {
		return nil, out[1].Interface().(error)
	}
	return json.Marshal(out[0].Interface())
}

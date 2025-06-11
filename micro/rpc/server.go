package rpc

import (
	"code-practise/micro/rpc/message"
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

		req := message.DecodeReq(reqBs)
		if err != nil {
			return err
		}
		resp, err := s.Invoke(context.Background(), req)
		if err != nil {
			// business err
			resp.Error = []byte(err.Error())
		}
		resp.CalculateHeaderLength()
		resp.CalculateBodyLength()

		// write message body
		_, err = conn.Write(message.EncodeResp(resp))
		if err != nil {
			return err
		}
	}
}

func (s *Server) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	stub, ok := s.services[req.ServiceName]
	resp := &message.Response{
		MessageId:  req.MessageId,
		Version:    req.Version,
		Compress:   req.Compress,
		Serializer: req.Serializer,
	}
	if !ok {
		return nil, errors.New("rpc: service not found")
	}
	// service
	data, err := stub.invoke(ctx, req.MethodName, req.Data)
	resp.Data = data
	if err != nil {
		return resp, err
	}
	return resp, nil
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
		err = out[1].Interface().(error)
	}
	if out[0].IsNil() {
		return nil, err
	}
	res, er := json.Marshal(out[0].Interface())
	if er != nil {
		return nil, er
	}
	return res, err
}

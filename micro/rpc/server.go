package rpc

import (
	"code-practise/micro/rpc/message"
	"code-practise/micro/rpc/serialize"
	"code-practise/micro/rpc/serialize/json"
	"code-practise/micro/rpc/serialize/proto"
	"context"
	"errors"
	"net"
	"reflect"
	"strconv"
	"time"
)

var (
	errOnewayServer = errors.New("rpc: server receives oneway request")
)

type Server struct {
	services    map[string]*reflectionStub
	serializers map[uint8]serialize.Serializer
}

func NewServer() *Server {
	res := &Server{
		services:    make(map[string]*reflectionStub, 16),
		serializers: make(map[uint8]serialize.Serializer, 4),
	}
	res.RegisterSerializer(&proto.Serializer{})
	res.RegisterSerializer(&json.Serializer{})
	return res
}

func (s *Server) RegisterSerializer(sl serialize.Serializer) {
	s.serializers[sl.Code()] = sl
}

func (s *Server) RegisterService(service Service) {
	s.services[service.Name()] = &reflectionStub{
		s:           service,
		value:       reflect.ValueOf(service),
		serializers: s.serializers,
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

		// meta
		ctx := context.Background()
		cancel := func() {}
		if deadlineStr, ok := req.Meta["deadline"]; ok {
			if deadline, er := strconv.ParseInt(deadlineStr, 10, 64); er == nil {
				ctx, cancel = context.WithDeadline(ctx, time.UnixMilli(deadline))
			}
		}
		oneway, ok := req.Meta["one-way"]
		if ok && oneway == "true" {
			ctx = CtxWithOneway(ctx)
		}

		resp, err := s.Invoke(ctx, req)
		cancel()
		if errors.Is(err, errOnewayServer) {
			return nil
		}
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
		return resp, errors.New("rpc: service not found")
	}

	if isOneway(ctx) {
		go func() {
			_, _ = stub.invoke(ctx, req)
		}()
		return resp, errOnewayServer
	}

	// service
	data, err := stub.invoke(ctx, req)
	resp.Data = data
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// reflectionStub consider using unsafe in the future
type reflectionStub struct {
	s           Service
	value       reflect.Value
	serializers map[uint8]serialize.Serializer
}

func (s *reflectionStub) invoke(ctx context.Context, req *message.Request) ([]byte, error) {
	// method
	method := s.value.MethodByName(req.MethodName)
	// arg
	in := reflect.New(method.Type().In(1).Elem())
	serializer, ok := s.serializers[req.Serializer]
	if !ok {
		return nil, errors.New("rpc: serializer not found")
	}
	err := serializer.Decode(req.Data, in.Interface())
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
	res, er := serializer.Encode(out[0].Interface())
	if er != nil {
		return nil, er
	}
	return res, err
}

package rpc

import (
	"context"
	"errors"
	"go-framework/micro/pool"
	"go-framework/micro/rpc/compress"
	"go-framework/micro/rpc/compress/donothing"
	"go-framework/micro/rpc/message"
	"go-framework/micro/rpc/serialize"
	"go-framework/micro/rpc/serialize/proto"
	"net"
	"reflect"
	"strconv"
	"time"
)

var (
	errOnewayClient = errors.New("rpc: oneway should not handle result")
)

// InitService assign values to function type fields
func (c *Client) InitService(service Service) error {
	return setFuncField(service, c, c.serializer, c.compressor)
}

func setFuncField(service Service, p Proxy,
	s serialize.Serializer,
	c compress.Compressor) error {
	if service == nil {
		return errors.New("rpc: service is nil")
	}
	val := reflect.ValueOf(service)
	typ := val.Type()
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return errors.New("rpc: service must be a pointer to struct")
	}

	val = val.Elem()
	typ = typ.Elem()

	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		fieldTyp := typ.Field(i)
		fieldVal := val.Field(i)

		if fieldVal.CanSet() {
			fn := reflect.MakeFunc(fieldTyp.Type, func(args []reflect.Value) []reflect.Value {
				retVal := reflect.New(fieldTyp.Type.Out(0).Elem())
				ctx := args[0].Interface().(context.Context)
				data, err := s.Encode(args[1].Interface())
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}
				reqData, err := c.Compress(data)
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				// meta
				meta := make(map[string]string, 2)
				if deadline, ok := ctx.Deadline(); ok {
					meta["deadline"] = strconv.FormatInt(deadline.UnixMilli(), 10)
				}
				if isOneway(ctx) {
					meta["one-way"] = "true"
				}

				// req param
				req := &message.Request{
					ServiceName: service.Name(),
					MethodName:  fieldTyp.Name,
					Data:        reqData,
					Serializer:  s.Code(),
					Compressor:  c.Code(),
					Meta:        meta,
				}
				req.CalculateHeaderLength()
				req.CalculateBodyLength()

				resp, err := p.Invoke(ctx, req)
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}
				var retErr error
				if len(resp.Error) > 0 {
					// business err
					retErr = errors.New(string(resp.Error))
				}
				if len(resp.Data) > 0 {
					respData, er := c.Uncompress(resp.Data)
					if er != nil {
						return []reflect.Value{retVal, reflect.ValueOf(er)}
					}
					err = s.Decode(respData, retVal.Interface())
					if err != nil {
						return []reflect.Value{retVal, reflect.ValueOf(err)}
					}
				}
				if retErr == nil {
					return []reflect.Value{retVal, reflect.Zero(reflect.TypeOf(new(error)).Elem())}
				}
				return []reflect.Value{retVal, reflect.ValueOf(retErr)}
			})
			fieldVal.Set(fn)
		}
	}
	return nil
}

type Client struct {
	pool       *pool.Pool
	serializer serialize.Serializer
	compressor compress.Compressor
}

type ClientOption func(client *Client)

func NewClient(addr string, opts ...ClientOption) (*Client, error) {
	p, err := pool.NewPool(&pool.Config{
		InitialCap:  1,
		MaxCap:      30,
		MaxIdle:     10,
		IdleTimeout: time.Second * 5,
		Factory: func() (any, error) {
			return net.DialTimeout("tcp", addr, time.Second*3)
		},
		Close: func(conn any) error {
			return conn.(net.Conn).Close()
		},
	})
	if err != nil {
		return nil, err
	}
	res := &Client{
		pool:       p,
		serializer: &proto.Serializer{},
		compressor: &donothing.Compressor{},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func ClientWithSerializer(sl serialize.Serializer) ClientOption {
	return func(client *Client) {
		client.serializer = sl
	}
}

func ClientWithCompressor(compressor compress.Compressor) ClientOption {
	return func(c *Client) {
		c.compressor = compressor
	}
}

func (c *Client) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	var (
		resp *message.Response
		err  error
	)
	ch := make(chan struct{})
	go func() {
		resp, err = c.doInvoke(ctx, req)
		ch <- struct{}{}
		close(ch)
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-ch:
		return resp, err
	}
}

func (c *Client) doInvoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	data := message.EncodeReq(req)
	resp, err := c.send(ctx, data)
	if err != nil {
		return nil, err
	}
	return message.DecodeResp(resp), nil
}

func (c *Client) send(ctx context.Context, data []byte) ([]byte, error) {
	val, err := c.pool.Get(ctx)
	if err != nil {
		return nil, err
	}
	conn := val.(net.Conn)
	defer func() {
		_ = conn.Close()
	}()

	// write message body
	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}

	// oneway
	if isOneway(ctx) {
		return nil, errOnewayClient
	}

	// read
	respBs, err := ReadMsg(conn)
	if err != nil {
		return nil, err
	}

	return respBs, nil

}

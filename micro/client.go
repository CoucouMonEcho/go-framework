package micro

import (
	"code-practise/micro/registry"
	"code-practise/micro/route"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

type ClientOption func(*Client)

type Client struct {
	insecure bool
	r        registry.Registry
	timeout  time.Duration

	bb route.BalancerBuilder
}

func NewClient(opts ...ClientOption) (*Client, error) {
	res := &Client{}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func ClientWithInsecure() ClientOption {
	return func(c *Client) {
		c.insecure = true
	}
}

func ClientWithRegistry(r registry.Registry, timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.r = r
		c.timeout = timeout
	}
}

func ClientWithPickerBuilder(bb route.BalancerBuilder) ClientOption {
	return func(c *Client) {
		c.bb = bb
	}
}

func (c *Client) Dial(_ context.Context, service string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if c.r != nil {
		rb := &grpcResolverBuilder{
			r:       c.r,
			timeout: c.timeout,
		}
		opts = append(opts, grpc.WithResolvers(rb))
	}
	if c.insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	if c.bb != nil {
		balancer.Register(base.NewBalancerBuilder(c.bb.Name(), c.bb, base.Config{HealthCheck: true}))
		opts = append(opts, grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy":"%s"}`, c.bb.Name())))
	}
	cc, err := grpc.NewClient(fmt.Sprintf("registry:///%s", service), opts...)
	return cc, err
}

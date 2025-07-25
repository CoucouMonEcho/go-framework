package micro

import (
	"code-practise/micro/registry"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"time"
)

type ClientOption func(*Client)

type Client struct {
	insecure bool
	r        registry.Registry
	timeout  time.Duration

	picker string
	//pickerBuilder base.PickerBuilder
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

func ClientWithPickerBuilder(name string, pickerBuilder base.PickerBuilder) ClientOption {
	return func(c *Client) {
		c.picker = name
		balancer.Register(base.NewBalancerBuilder("DEMO_ROUND_ROBIN", pickerBuilder, base.Config{HealthCheck: true}))
	}
}

func (c *Client) Dial(ctx context.Context, service string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	if c.r != nil {
		rb, err := NewResolverBuilder(c.r, c.timeout)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithResolvers(rb))
	}
	if c.insecure {
		opts = append(opts, grpc.WithInsecure())
	}
	if c.picker != "" {
		//FIXME ???
		opts = append(opts, grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy":"%s"}`, c.picker)))
	}
	cc, err := grpc.DialContext(ctx, fmt.Sprintf("registry:///%s", service), opts...)
	return cc, err
}

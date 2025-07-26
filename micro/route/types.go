package route

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

// Filter returns TRUE when selected
type Filter func(info balancer.PickInfo, addr resolver.Address) bool

type GroupFilterBuilder struct{}

func (g GroupFilterBuilder) Build() Filter {
	return func(info balancer.PickInfo, addr resolver.Address) bool {
		tgt := addr.Attributes.Value("group")
		input := info.Ctx.Value("group")
		return tgt == input
	}
}

type BalancerBuilder interface {
	Build(info base.PickerBuildInfo) balancer.Picker
	Name() string
}

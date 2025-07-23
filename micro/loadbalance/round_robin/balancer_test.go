package round_robin

import (
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/balancer"
	"testing"
)

func TestBalancer_Pick(t *testing.T) {
	testCases := []struct {
		name string
		b    *Balancer

		wantErr           error
		wantSubConn       Subconn
		wantBalancerIndex int32
	}{
		{
			name: "start",
			b: &Balancer{
				index: -1,
				conns: []balancer.SubConn{
					Subconn{name: "127.0.0.1:8080"},
					Subconn{name: "127.0.0.1:8081"},
				},
				length: 2,
			},
			wantSubConn:       Subconn{name: "127.0.0.1:8080"},
			wantBalancerIndex: 0,
		},
		{
			name: "empty",
			b: &Balancer{
				index: 1,
				conns: []balancer.SubConn{},
			},
			wantErr: balancer.ErrNoSubConnAvailable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.b.Pick(balancer.PickInfo{})
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantSubConn.name, res.SubConn.(Subconn).name)
			assert.NotNil(t, res.Done)
			assert.Equal(t, tc.wantBalancerIndex, tc.b.index)
		})
	}
}

type Subconn struct {
	name string
	balancer.SubConn
}

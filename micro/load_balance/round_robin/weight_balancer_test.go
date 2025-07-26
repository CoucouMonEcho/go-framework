package round_robin

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/balancer"
	"testing"
)

func TestWeightBalancer_Pick(t *testing.T) {
	b := &WeightBalancer{
		connes: []*weightConn{
			{
				conn: Subconn{
					name: "weight-5",
				},
				weight:          5,
				efficientWeight: 5,
				currentWeight:   5,
			},
			{
				conn: Subconn{
					name: "weight-4",
				},
				weight:          4,
				efficientWeight: 4,
				currentWeight:   4,
			},
			{
				conn: Subconn{
					name: "weight-3",
				},
				weight:          3,
				efficientWeight: 3,
				currentWeight:   3,
			},
		},
	}
	pickRes, err := b.Pick(balancer.PickInfo{})
	require.NoError(t, err)

	pickRes.Done(balancer.DoneInfo{})
	assert.Equal(t, uint32(6), b.connes[0].efficientWeight)

	pickRes, err = b.Pick(balancer.PickInfo{})
	require.NoError(t, err)
	assert.Equal(t, "weight-4", pickRes.SubConn.(Subconn).name)

	pickRes, err = b.Pick(balancer.PickInfo{})
	require.NoError(t, err)
	assert.Equal(t, "weight-3", pickRes.SubConn.(Subconn).name)

	pickRes, err = b.Pick(balancer.PickInfo{})
	require.NoError(t, err)
	assert.Equal(t, "weight-5", pickRes.SubConn.(Subconn).name)

	pickRes, err = b.Pick(balancer.PickInfo{})
	require.NoError(t, err)
	assert.Equal(t, "weight-4", pickRes.SubConn.(Subconn).name)

}

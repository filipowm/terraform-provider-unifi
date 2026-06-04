package device

import (
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToPortOverrideAggregateTranslation(t *testing.T) {
	t.Run("CountBecomesContiguousMemberRange", func(t *testing.T) {
		po, err := toPortOverride(map[string]interface{}{
			"number":              5,
			"name":                "lag uplink",
			"port_profile_id":     "",
			"op_mode":             "aggregate",
			"poe_mode":            "",
			"aggregate_num_ports": 3,
		})
		require.NoError(t, err)
		assert.Equal(t, []int{5, 6, 7}, po.AggregateMembers)
	})

	t.Run("UnsetLeavesMembersNil", func(t *testing.T) {
		po, err := toPortOverride(map[string]interface{}{
			"number":              1,
			"name":                "plain port",
			"port_profile_id":     "abc123",
			"op_mode":             "switch",
			"poe_mode":            "auto",
			"aggregate_num_ports": 0,
		})
		require.NoError(t, err)
		assert.Nil(t, po.AggregateMembers)
	})
}

func TestFromPortOverrideAggregateTranslation(t *testing.T) {
	t.Run("MemberListLengthBecomesCount", func(t *testing.T) {
		m, err := fromPortOverride(unifi.DevicePortOverrides{
			PortIDX:          5,
			OpMode:           "aggregate",
			AggregateMembers: []int{5, 6, 7},
		})
		require.NoError(t, err)
		assert.Equal(t, 3, m["aggregate_num_ports"])
	})

	t.Run("NilMembersReadBackAsZero", func(t *testing.T) {
		m, err := fromPortOverride(unifi.DevicePortOverrides{
			PortIDX: 1,
			OpMode:  "switch",
		})
		require.NoError(t, err)
		assert.Equal(t, 0, m["aggregate_num_ports"])
	})
}

func TestPortOverrideAggregateRoundTrip(t *testing.T) {
	in := map[string]interface{}{
		"number":              2,
		"name":                "lag",
		"port_profile_id":     "",
		"op_mode":             "aggregate",
		"poe_mode":            "",
		"aggregate_num_ports": 4,
	}
	po, err := toPortOverride(in)
	require.NoError(t, err)
	out, err := fromPortOverride(po)
	require.NoError(t, err)
	assert.Equal(t, in["aggregate_num_ports"], out["aggregate_num_ports"])
}

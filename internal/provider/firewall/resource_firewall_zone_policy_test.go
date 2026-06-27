package firewall

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFirewallPolicyTargetModelPortParsing(t *testing.T) {
	t.Run("NumericPortParses", func(t *testing.T) {
		m := NewFirewallPolicyTargetModel("", nil, false, false, "8443", "", "zone1")
		assert.False(t, m.Port.IsNull())
		assert.Equal(t, int32(8443), m.Port.ValueInt32())
	})

	t.Run("EmptyPortIsNull", func(t *testing.T) {
		m := NewFirewallPolicyTargetModel("", nil, false, false, "", "", "zone1")
		assert.True(t, m.Port.IsNull())
	})

	t.Run("RangeValueIsNull", func(t *testing.T) {
		// The controller accepts ranges/lists ("8000-8010,9443") but the
		// schema attribute is an Int32 — unrepresentable values read as null.
		m := NewFirewallPolicyTargetModel("", nil, false, false, "8000-8010", "", "zone1")
		assert.True(t, m.Port.IsNull())
	})
}

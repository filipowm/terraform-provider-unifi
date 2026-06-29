package utils

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
)

// TestRawConfigSet covers the "is the attribute explicitly configured"
// decision that gates the conditional write in some resources. Only a
// known, non-empty raw-config value must be sent to the controller; null and
// empty-string must be treated as "not set" so omitempty drops the field and a
// managed resource is not clobbered. Unknown (interpolated)
// counts as set — it is resolved to a real value by apply time, when GetResourceData runs.
func TestRawConfigSet(t *testing.T) {
	tests := []struct {
		name string
		raw  cty.Value
		want bool
	}{
		{
			name: "null config (attribute omitted)",
			raw:  cty.ObjectVal(map[string]cty.Value{"firewall_zone_id": cty.NullVal(cty.String)}),
			want: false,
		},
		{
			name: "explicit empty string",
			raw:  cty.ObjectVal(map[string]cty.Value{"firewall_zone_id": cty.StringVal("")}),
			want: false,
		},
		{
			name: "known non-empty value",
			raw:  cty.ObjectVal(map[string]cty.Value{"firewall_zone_id": cty.StringVal("zoneABC")}),
			want: true,
		},
		{
			name: "unknown (interpolated) value",
			raw:  cty.ObjectVal(map[string]cty.Value{"firewall_zone_id": cty.UnknownVal(cty.String)}),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRawConfigSet(tt.raw, "firewall_zone_id"); got != tt.want {
				t.Errorf("utils.IsRawConfigSet(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

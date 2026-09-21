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

// TestIsRawConfigSet_bool guards a panic: every type that is not String or Number
// fell through to LengthInt(), which is only defined for collections, so a bool
// attribute crashed the provider during plan rather than returning an answer.
//
// `false` counts as set. Unlike "" or 0 it is not ambiguous with absence -- the
// raw config spells absence as null -- so a caller asking "did the user write
// this attribute?" gets the accurate answer.
func TestIsRawConfigSet_bool(t *testing.T) {
	tests := []struct {
		name string
		raw  cty.Value
		want bool
	}{
		{
			name: "explicit true",
			raw:  cty.ObjectVal(map[string]cty.Value{"uid_vpn_sync_public_ip": cty.True}),
			want: true,
		},
		{
			name: "explicit false",
			raw:  cty.ObjectVal(map[string]cty.Value{"uid_vpn_sync_public_ip": cty.False}),
			want: true,
		},
		{
			name: "null (absent)",
			raw:  cty.ObjectVal(map[string]cty.Value{"uid_vpn_sync_public_ip": cty.NullVal(cty.Bool)}),
			want: false,
		},
		{
			name: "unknown (interpolated)",
			raw:  cty.ObjectVal(map[string]cty.Value{"uid_vpn_sync_public_ip": cty.UnknownVal(cty.Bool)}),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRawConfigSet(tt.raw, "uid_vpn_sync_public_ip"); got != tt.want {
				t.Errorf("utils.IsRawConfigSet(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

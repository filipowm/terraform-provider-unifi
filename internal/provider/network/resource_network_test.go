package network

import (
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TestResourceNetworkSetResourceData_readsFirewallZoneID is a regression guard for
// the read-back gap that caused issue #94: the resource never surfaced
// unifi.Network.FirewallZoneID, hiding zone drift entirely. SetResourceData must now
// always populate firewall_zone_id so Read/import round-trip it. nil meta is safe
// here — GetResourceData is not called and SetResourceData references no client.
func TestResourceNetworkSetResourceData_readsFirewallZoneID(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceNetwork().Schema, map[string]interface{}{})

	resp := &unifi.Network{FirewallZoneID: "zoneABC"}
	if diags := resourceNetworkSetResourceData(resp, d, "default"); diags.HasError() {
		t.Fatalf("resourceNetworkSetResourceData returned diagnostics: %v", diags)
	}

	if got := d.Get("firewall_zone_id").(string); got != "zoneABC" {
		t.Errorf("firewall_zone_id = %q, want %q", got, "zoneABC")
	}
}

// TestRawConfigSet_firewallZoneID covers the "is the attribute explicitly configured"
// decision that gates the conditional write in resourceNetworkGetResourceData. Only a
// known, non-empty raw-config value must be sent to the controller; null and
// empty-string must be treated as "not set" so omitempty drops the field and a zone
// managed via unifi_firewall_zone.networks is not clobbered. Unknown (interpolated)
// counts as set — it is resolved to a real value by apply time, when GetResourceData runs.
func TestRawConfigSet_firewallZoneID(t *testing.T) {
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
			if got := rawConfigSet(tt.raw, "firewall_zone_id"); got != tt.want {
				t.Errorf("rawConfigSet(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

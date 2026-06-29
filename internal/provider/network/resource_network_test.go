package network

import (
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TestResourceNetworkGetResourceData_dhcpGuarding is a regression guard for issue
// #123: the Create/Update mapper must carry dhcp_guarding into the go-unifi request
// struct. Before the fix the field was never assigned, so every PUT serialized
// dhcpguard_enabled:false (the field has no omitempty), silently clearing a value
// the user enabled in the controller UI.
func TestResourceNetworkGetResourceData_dhcpGuarding(t *testing.T) {
	raw := map[string]interface{}{
		"name":          "tfacc-dhcp-guarding",
		"purpose":       "corporate",
		"dhcp_guarding": true,
	}

	d := schema.TestResourceDataRaw(t, ResourceNetwork().Schema, raw)

	// meta is not dereferenced by resourceNetworkGetResourceData, so nil is safe.
	req, err := resourceNetworkGetResourceData(d, nil)
	if err != nil {
		t.Fatalf("resourceNetworkGetResourceData returned error: %s", err)
	}
	if !req.DHCPguardEnabled {
		t.Fatalf("expected DHCPguardEnabled to be true, got false")
	}
}

// TestResourceNetworkGetResourceData_dhcpGuardingExplicitFalse ensures an explicit
// false in config is honored (the disable path), not dropped.
func TestResourceNetworkGetResourceData_dhcpGuardingExplicitFalse(t *testing.T) {
	raw := map[string]interface{}{
		"name":          "tfacc-dhcp-guarding",
		"purpose":       "corporate",
		"dhcp_guarding": false,
	}

	d := schema.TestResourceDataRaw(t, ResourceNetwork().Schema, raw)

	req, err := resourceNetworkGetResourceData(d, nil)
	if err != nil {
		t.Fatalf("resourceNetworkGetResourceData returned error: %s", err)
	}
	if req.DHCPguardEnabled {
		t.Fatalf("expected DHCPguardEnabled to be false, got true")
	}
}

// TestResourceNetworkSetResourceData_dhcpGuarding is the symmetric read-path guard:
// the controller value must be flattened back into the dhcp_guarding attribute so an
// omitted config inherits the controller's real value (Optional+Computed).
func TestResourceNetworkSetResourceData_dhcpGuarding(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceNetwork().Schema, map[string]interface{}{})

	resp := &unifi.Network{DHCPguardEnabled: true}
	if diags := resourceNetworkSetResourceData(resp, d, "default"); diags.HasError() {
		t.Fatalf("resourceNetworkSetResourceData returned diagnostics: %v", diags)
	}
	if got := d.Get("dhcp_guarding").(bool); !got {
		t.Fatalf("expected dhcp_guarding to be true in state, got false")
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

package network

import (
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/go-cty/cty"
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
	}
}

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

// TestValidateDHCPGuardingRawConfig is the plan-time guard for issue #123. The
// trusted-server requirement must be driven off the raw config, not d.Get: an Update
// that omits dhcp_guarding (inheriting a previously-enabled value plus its trusted
// servers) must NOT trip the gate, because in a ResourceDiff a Computed list reads
// back empty while the scalar still surfaces the inherited true. The gate fires only
// when the user explicitly enables guarding in *this* config without a trusted server.
func TestValidateDHCPGuardingRawConfig(t *testing.T) {
	servers := cty.ListVal([]cty.Value{cty.StringVal("10.0.0.1")})
	nullServers := cty.NullVal(cty.List(cty.String))
	emptyServers := cty.ListValEmpty(cty.String)

	rawConfig := func(guarding, trustedServers cty.Value) cty.Value {
		return cty.ObjectVal(map[string]cty.Value{
			"dhcp_guarding":                 guarding,
			"dhcp_guarding_trusted_servers": trustedServers,
		})
	}

	tests := []struct {
		name    string
		raw     cty.Value
		wantErr bool
	}{
		{
			// The decisive #123 case: dhcp_guarding omitted on an Update.
			name: "guarding omitted (inherited)",
			raw:  rawConfig(cty.NullVal(cty.Bool), nullServers),
		},
		{
			name: "guarding enabled with trusted server",
			raw:  rawConfig(cty.True, servers),
		},
		{
			name:    "guarding enabled, trusted servers omitted",
			raw:     rawConfig(cty.True, nullServers),
			wantErr: true,
		},
		{
			name:    "guarding enabled, trusted servers explicitly empty",
			raw:     rawConfig(cty.True, emptyServers),
			wantErr: true,
		},
		{
			name: "guarding explicitly disabled",
			raw:  rawConfig(cty.False, nullServers),
		},
		{
			// Interpolated (var.x) — unresolved at plan, can't validate, must not error.
			name: "guarding unknown",
			raw:  rawConfig(cty.UnknownVal(cty.Bool), nullServers),
		},
		{
			// No config at all (e.g. destroy).
			name: "null raw config",
			raw:  cty.NullVal(cty.Object(map[string]cty.Type{"dhcp_guarding": cty.Bool, "dhcp_guarding_trusted_servers": cty.List(cty.String)})),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDHCPGuardingRawConfig(tt.raw)
			if tt.wantErr && err == nil {
				t.Fatalf("expected an error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got: %s", err)
			}
		})
	}
}

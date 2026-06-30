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

// TestValidateDefaultGatewayRawConfig covers the cross-field rule for the DHCP
// default-gateway override (issue #120). Like the DHCP Guarding gate it is driven off
// the raw config so an Update that omits the attributes (inheriting the controller's
// values via Optional+Computed) does not false-positive; each rule keys on the
// explicit config of its counterpart.
func TestValidateDefaultGatewayRawConfig(t *testing.T) {
	gw := cty.StringVal("10.0.0.1")
	nullGw := cty.NullVal(cty.String)
	emptyGw := cty.StringVal("")

	rawConfig := func(enabled, gateway cty.Value) cty.Value {
		return cty.ObjectVal(map[string]cty.Value{
			"dhcpd_gateway_enabled": enabled,
			"dhcpd_gateway":         gateway,
		})
	}

	tests := []struct {
		name    string
		raw     cty.Value
		wantErr bool
	}{
		{
			// Decisive #120/#123-style case: both omitted on an Update (inherited).
			name: "both omitted (inherited)",
			raw:  rawConfig(cty.NullVal(cty.Bool), nullGw),
		},
		{
			name: "override enabled with gateway",
			raw:  rawConfig(cty.True, gw),
		},
		{
			name:    "override enabled, gateway omitted",
			raw:     rawConfig(cty.True, nullGw),
			wantErr: true,
		},
		{
			name:    "override enabled, gateway explicitly empty",
			raw:     rawConfig(cty.True, emptyGw),
			wantErr: true,
		},
		{
			name:    "gateway set, override explicitly disabled",
			raw:     rawConfig(cty.False, gw),
			wantErr: true,
		},
		{
			name: "override explicitly disabled, no gateway",
			raw:  rawConfig(cty.False, nullGw),
		},
		{
			// Gateway set while the toggle is omitted: it may inherit true, so the
			// conservative gate does not fire (matches the inherit-from-controller path).
			name: "gateway set, override omitted (inherited)",
			raw:  rawConfig(cty.NullVal(cty.Bool), gw),
		},
		{
			// Interpolated toggle (var.x) — unresolved at plan, can't validate.
			name: "override unknown, gateway omitted",
			raw:  rawConfig(cty.UnknownVal(cty.Bool), nullGw),
		},
		{
			name: "override unknown, gateway set",
			raw:  rawConfig(cty.UnknownVal(cty.Bool), gw),
		},
		{
			// No config at all (e.g. destroy).
			name: "null raw config",
			raw:  cty.NullVal(cty.Object(map[string]cty.Type{"dhcpd_gateway_enabled": cty.Bool, "dhcpd_gateway": cty.String})),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDefaultGatewayRawConfig(tt.raw)
			if tt.wantErr && err == nil {
				t.Fatalf("expected an error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got: %s", err)
			}
		})
	}
}

// TestResourceNetworkGetResourceData_defaultGateway guards the write path: the mapper
// must carry the override toggle and gateway IP into the go-unifi request struct
// (neither field has omitempty, so an unmapped value would silently serialize as
// auto/empty on every PUT — the issue #120 gap).
func TestResourceNetworkGetResourceData_defaultGateway(t *testing.T) {
	raw := map[string]interface{}{
		"name":                  "tfacc-default-gateway",
		"purpose":               "corporate",
		"dhcpd_gateway_enabled": true,
		"dhcpd_gateway":         "10.0.0.1",
	}

	d := schema.TestResourceDataRaw(t, ResourceNetwork().Schema, raw)

	req, err := resourceNetworkGetResourceData(d, nil)
	if err != nil {
		t.Fatalf("resourceNetworkGetResourceData returned error: %s", err)
	}
	if !req.DHCPDGatewayEnabled {
		t.Fatalf("expected DHCPDGatewayEnabled to be true, got false")
	}
	if req.DHCPDGateway != "10.0.0.1" {
		t.Fatalf("expected DHCPDGateway to be %q, got %q", "10.0.0.1", req.DHCPDGateway)
	}
}

// TestResourceNetworkSetResourceData_defaultGateway is the symmetric read-path guard:
// the controller values must be flattened back so an omitted config inherits them
// (Optional+Computed) and imports round-trip.
func TestResourceNetworkSetResourceData_defaultGateway(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceNetwork().Schema, map[string]interface{}{})

	resp := &unifi.Network{DHCPDGatewayEnabled: true, DHCPDGateway: "10.0.0.1"}
	if diags := resourceNetworkSetResourceData(resp, d, "default"); diags.HasError() {
		t.Fatalf("resourceNetworkSetResourceData returned diagnostics: %v", diags)
	}
	if got := d.Get("dhcpd_gateway_enabled").(bool); !got {
		t.Fatalf("expected dhcpd_gateway_enabled to be true in state, got false")
	}
	if got := d.Get("dhcpd_gateway").(string); got != "10.0.0.1" {
		t.Fatalf("expected dhcpd_gateway to be %q in state, got %q", "10.0.0.1", got)
	}
}

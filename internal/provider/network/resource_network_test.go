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
	}
}

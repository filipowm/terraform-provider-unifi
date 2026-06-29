package radius

import (
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TestResourceAccountGetResourceData_mapsVLAN verifies that the `vlan` attribute is
// mapped onto the go-unifi Account struct on the write path. Regression lock for #95.
func TestResourceAccountGetResourceData_mapsVLAN(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceAccount().Schema, map[string]interface{}{
		"name":     "tfacc",
		"password": "secure",
		"vlan":     90,
	})

	req, err := resourceAccountGetResourceData(d)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if req.VLAN != 90 {
		t.Fatalf("expected Account.VLAN == 90, got %d", req.VLAN)
	}
}

// TestResourceAccountSetResourceData_readsVLAN verifies that the controller-returned
// VLAN is written back into state on the read path. Regression lock for #95.
func TestResourceAccountSetResourceData_readsVLAN(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceAccount().Schema, map[string]interface{}{})

	if diags := resourceAccountSetResourceData(&unifi.Account{Name: "x", VLAN: 90}, d, "default"); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if got := d.Get("vlan").(int); got != 90 {
		t.Fatalf("expected vlan == 90 in state, got %d", got)
	}
}

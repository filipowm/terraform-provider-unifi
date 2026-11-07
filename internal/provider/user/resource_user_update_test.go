package user

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// importedUserState is what `terraform import` leaves behind: the controller's
// fields, but none of the provider-only flags, which have no controller value.
func importedUserState() *terraform.InstanceState {
	return &terraform.InstanceState{
		ID: "3e697a9e-9855-4480-bf9e-da769db430e3",
		Attributes: map[string]string{
			"id":              "3e697a9e-9855-4480-bf9e-da769db430e3",
			"site":            "default",
			"mac":             "48:e1:5c:78:b2:8c",
			"name":            "Apple TV 4K",
			"fixed_ip":        "192.168.44.4",
			"blocked":         "false",
			"dev_id_override": "0",
		},
	}
}

// applyUser plans cfg against the imported state and applies it with a nil
// client, so any attempt to reach the controller fails with an error.
func applyUser(t *testing.T, cfg map[string]interface{}) (*terraform.InstanceState, diag.Diagnostics) {
	t.Helper()
	r := ResourceUser()
	ctx := context.Background()
	s := importedUserState()
	d, err := r.Diff(ctx, s, terraform.NewResourceConfigRaw(cfg), nil)
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	if d == nil || d.Empty() {
		t.Fatal("expected a diff")
	}
	return r.Apply(ctx, s, d, nil)
}

func TestUserUpdateWithOnlyProviderFlagsSkipsController(t *testing.T) {
	st, diags := applyUser(t, map[string]interface{}{
		"mac":                    "48:e1:5c:78:b2:8c",
		"name":                   "Apple TV 4K",
		"fixed_ip":               "192.168.44.4",
		"skip_forget_on_destroy": true,
	})
	if diags.HasError() {
		t.Fatalf("update reached the controller for a flag-only change: %v", diags)
	}
	if got := st.Attributes["allow_existing"]; got != "true" {
		t.Errorf("allow_existing = %q, want true", got)
	}
	if got := st.Attributes["skip_forget_on_destroy"]; got != "true" {
		t.Errorf("skip_forget_on_destroy = %q, want true", got)
	}
	if got := st.Attributes["name"]; got != "Apple TV 4K" {
		t.Errorf("name = %q, controller fields must be kept", got)
	}
}

func TestUserUpdateWithControllerFieldStillCallsController(t *testing.T) {
	_, diags := applyUser(t, map[string]interface{}{
		"mac":                    "48:e1:5c:78:b2:8c",
		"name":                   "Living Room Apple TV",
		"fixed_ip":               "192.168.44.4",
		"skip_forget_on_destroy": true,
	})
	if !diags.HasError() {
		t.Fatal("a name change must be sent to the controller, but the update returned without calling it")
	}
}

package validators_test

import (
	"context"
	"testing"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestMacRegex pins the set of MAC address strings MacRegex accepts and rejects.
//
// MacRegex accepts both ':' and '-' separators (and, as a property of the
// pattern, mixed separators such as "00-11:22:33-44:55"). Mixed separators are
// intentionally accepted here: resources that hold MAC sets use the MACType
// custom type (see types.MACType), which treats values that differ only in case
// or separator as semantically equal, so any accepted variant round-trips to the
// controller's canonical form without a diff.
func TestMacRegex(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    string
		expectOK bool
	}{
		// Accepted — colon separated
		"colon lowercase":       {"00:11:22:33:44:55", true},
		"colon uppercase":       {"AA:BB:CC:DD:EE:FF", true},
		"colon mixed case":      {"aA:bB:cC:dD:eE:fF", true},
		"colon all zeros":       {"00:00:00:00:00:00", true},
		"colon all f":           {"ff:ff:ff:ff:ff:ff", true},
		"colon unicast example": {"00:15:6d:00:00:01", true},

		// Accepted — dash separated (the form this PR widened the regex to allow)
		"dash lowercase": {"00-11-22-33-44-55", true},
		"dash uppercase": {"AA-BB-CC-DD-EE-FF", true},

		// Accepted — mixed separators (documented decision: accepted, normalized downstream)
		"mixed separators": {"00-11:22:33-44:55", true},

		// Rejected — wrong format / notation
		"empty":           {"", false},
		"garbage":         {"invalid-mac-address", false},
		"dot notation":    {"0011.2233.4455", false},
		"no separators":   {"001122334455", false},
		"space separated": {"00 11 22 33 44 55", false},
		"trailing space":  {"00:11:22:33:44:55 ", false},
		"leading space":   {" 00:11:22:33:44:55", false},

		// Rejected — wrong length
		"too short (5 octets)": {"00:11:22:33:44", false},
		"too long (7 octets)":  {"00:11:22:33:44:55:66", false},
		"single hex digit":     {"0:11:22:33:44:55", false},

		// Rejected — non-hex characters
		"non-hex g": {"0g:11:22:33:44:55", false},
		"non-hex z": {"zz:11:22:33:44:55", false},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := validators.MacRegex.MatchString(test.input); got != test.expectOK {
				t.Fatalf("MacRegex.MatchString(%q) = %v, want %v", test.input, got, test.expectOK)
			}
		})
	}
}

// TestMacValidator exercises the framework validator wired on MacRegex to ensure
// the validator (not just the raw regex) accepts/rejects as expected, and that
// unknown/null values are skipped rather than erroring.
func TestMacValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         types.String
		expectError bool
	}
	tests := map[string]testCase{
		"unknown":         {val: types.StringUnknown()},
		"null":            {val: types.StringNull()},
		"valid colon":     {val: types.StringValue("00:11:22:33:44:55")},
		"valid dash":      {val: types.StringValue("00-11-22-33-44-55")},
		"valid uppercase": {val: types.StringValue("AA:BB:CC:DD:EE:FF")},
		"invalid":         {val: types.StringValue("invalid-mac-address"), expectError: true},
		"empty":           {val: types.StringValue(""), expectError: true},
		"too short":       {val: types.StringValue("00:11:22:33:44"), expectError: true},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			req := validator.StringRequest{ConfigValue: test.val}
			resp := validator.StringResponse{}
			validators.Mac.ValidateString(context.Background(), req, &resp)

			if !test.expectError && resp.Diagnostics.HasError() {
				t.Fatalf("got unexpected error: %s", resp.Diagnostics.Errors()[0].Detail())
			}
			if test.expectError && !resp.Diagnostics.HasError() {
				t.Fatalf("expected error but got none")
			}
		})
	}
}

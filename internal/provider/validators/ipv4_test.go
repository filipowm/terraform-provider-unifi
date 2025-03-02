package validators_test

import (
	"context"
	"testing"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestIPv4Validator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         types.String
		expectError bool
	}
	tests := map[string]testCase{
		"unknown": {
			val: types.StringUnknown(),
		},
		"null": {
			val: types.StringNull(),
		},
		"empty": {
			val: types.StringValue(""),
		},
		"valid ipv4": {
			val: types.StringValue("192.168.1.1"),
		},
		"valid ipv4 with leading zeros": {
			val:         types.StringValue("192.168.001.001"),
			expectError: true,
		},
		"invalid ipv4 - out of range": {
			val:         types.StringValue("192.168.1.256"),
			expectError: true,
		},
		"invalid ipv4 - incomplete": {
			val:         types.StringValue("192.168.1"),
			expectError: true,
		},
		"invalid ipv4 - ipv6": {
			val:         types.StringValue("::1"),
			expectError: true,
		},
		"invalid ipv4 - characters": {
			val:         types.StringValue("not-an-ip"),
			expectError: true,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			req := validator.StringRequest{
				ConfigValue: test.val,
			}
			resp := validator.StringResponse{}
			validators.IPv4().ValidateString(context.Background(), req, &resp)

			if !test.expectError && resp.Diagnostics.HasError() {
				t.Fatalf("got unexpected error: %s", resp.Diagnostics.Errors()[0].Detail())
			}

			if test.expectError && !resp.Diagnostics.HasError() {
				t.Fatalf("expected error but got none")
			}
		})
	}
}

package validators_test

import (
	"context"
	"testing"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestIPv6Validator(t *testing.T) {
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
		"valid ipv6 full": {
			val: types.StringValue("2001:0db8:85a3:0000:0000:8a2e:0370:7334"),
		},
		"valid ipv6 compressed": {
			val: types.StringValue("2001:db8:85a3::8a2e:370:7334"),
		},
		"valid ipv6 loopback": {
			val: types.StringValue("::1"),
		},
		"valid ipv6 unspecified": {
			val: types.StringValue("::"),
		},
		"valid ipv6 with zone": {
			val: types.StringValue("fe80::1ff:fe23:4567:890a%eth0"),
		},
		"valid ipv6 ipv4-mapped": {
			val: types.StringValue("::ffff:192.0.2.128"),
		},
		"invalid ipv6 - too many segments": {
			val:         types.StringValue("2001:0db8:85a3:0000:0000:8a2e:0370:7334:1111"),
			expectError: true,
		},
		"invalid ipv6 - out of range": {
			val:         types.StringValue("2001:0db8:85a3:0000:0000:8a2e:0370:GGGG"),
			expectError: true,
		},
		"invalid ipv6 - incomplete": {
			val:         types.StringValue("2001:0db8:85a3"),
			expectError: true,
		},
		"invalid ipv6 - ipv4": {
			val:         types.StringValue("192.168.1.1"),
			expectError: true,
		},
		"invalid ipv6 - characters": {
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
			validators.IPv6().ValidateString(context.Background(), req, &resp)

			if !test.expectError && resp.Diagnostics.HasError() {
				t.Fatalf("got unexpected error: %s", resp.Diagnostics.Errors()[0].Detail())
			}

			if test.expectError && !resp.Diagnostics.HasError() {
				t.Fatalf("expected error but got none")
			}
		})
	}
}

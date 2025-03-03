package validators_test

import (
	"context"
	"testing"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestCIDR(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         types.String
		expectError bool
	}
	tests := map[string]testCase{
		"unknown": {
			val:         types.StringUnknown(),
			expectError: false,
		},
		"null": {
			val:         types.StringNull(),
			expectError: false,
		},
		"empty": {
			val:         types.StringValue(""),
			expectError: true,
		},
		"valid-ipv4": {
			val:         types.StringValue("192.168.1.0/24"),
			expectError: false,
		},
		"invalid-ipv4": {
			val:         types.StringValue("192.168.1.0"),
			expectError: true,
		},
		"valid-ipv6": {
			val:         types.StringValue("2001:db8::/32"),
			expectError: false,
		},
		"invalid-ipv6": {
			val:         types.StringValue("2001:db8::"),
			expectError: true,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			request := validator.StringRequest{
				ConfigValue: test.val,
			}
			response := validator.StringResponse{}
			validators.CIDR().ValidateString(context.Background(), request, &response)

			if test.expectError {
				require.NotEmpty(t, response.Diagnostics)
				return
			}
			require.Empty(t, response.Diagnostics)
		})
	}
}

func TestCIDROrEmpty(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         types.String
		expectError bool
	}
	tests := map[string]testCase{
		"unknown": {
			val:         types.StringUnknown(),
			expectError: false,
		},
		"null": {
			val:         types.StringNull(),
			expectError: false,
		},
		"empty": {
			val:         types.StringValue(""),
			expectError: false,
		},
		"valid-ipv4": {
			val:         types.StringValue("192.168.1.0/24"),
			expectError: false,
		},
		"invalid-ipv4": {
			val:         types.StringValue("192.168.1.0"),
			expectError: true,
		},
		"valid-ipv6": {
			val:         types.StringValue("2001:db8::/32"),
			expectError: false,
		},
		"invalid-ipv6": {
			val:         types.StringValue("2001:db8::"),
			expectError: true,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			request := validator.StringRequest{
				ConfigValue: test.val,
			}
			response := validator.StringResponse{}
			validators.CIDROrEmpty().ValidateString(context.Background(), request, &response)

			if test.expectError {
				require.NotEmpty(t, response.Diagnostics)
				return
			}
			require.Empty(t, response.Diagnostics)
		})
	}
}

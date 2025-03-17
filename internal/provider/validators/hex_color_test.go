package validators_test

import (
	"context"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestHexColorValidator(t *testing.T) {
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
		"valid-6-digits": {
			val: types.StringValue("#123456"),
		},
		"valid-3-digits": {
			val: types.StringValue("#123"),
		},
		"valid-uppercase": {
			val: types.StringValue("#ABCDEF"),
		},
		"valid-mixed-case": {
			val: types.StringValue("#aBcDeF"),
		},
		"invalid-missing-hash": {
			val:         types.StringValue("123456"),
			expectError: true,
		},
		"invalid-too-short": {
			val:         types.StringValue("#12"),
			expectError: true,
		},
		"invalid-too-long": {
			val:         types.StringValue("#1234567"),
			expectError: true,
		},
		"invalid-wrong-chars": {
			val:         types.StringValue("#12345G"),
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
			validators.HexColor().ValidateString(context.Background(), request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}

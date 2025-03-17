package validators_test

import (
	"context"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEmailValidator(t *testing.T) {
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
		"valid-simple": {
			val: types.StringValue("test@example.com"),
		},
		"valid-with-dots": {
			val: types.StringValue("john.doe@example.com"),
		},
		"valid-with-plus": {
			val: types.StringValue("john+test@example.com"),
		},
		"valid-with-subdomain": {
			val: types.StringValue("john@sub.example.com"),
		},
		"valid-with-numbers": {
			val: types.StringValue("user123@example.com"),
		},
		"invalid-no-at": {
			val:         types.StringValue("testexample.com"),
			expectError: true,
		},
		"invalid-no-domain": {
			val:         types.StringValue("test@"),
			expectError: true,
		},
		"invalid-no-tld": {
			val:         types.StringValue("test@example"),
			expectError: true,
		},
		"invalid-space": {
			val:         types.StringValue("test user@example.com"),
			expectError: true,
		},
		"invalid-special-chars": {
			val:         types.StringValue("test*user@example.com"),
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
			validators.Email().ValidateString(context.Background(), request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}

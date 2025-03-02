package validators_test

import (
	"context"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestTimezoneValidator(t *testing.T) {
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
		"valid-america": {
			val: types.StringValue("America/Los_Angeles"),
		},
		"valid-europe": {
			val: types.StringValue("Europe/London"),
		},
		"valid-asia": {
			val: types.StringValue("Asia/Tokyo"),
		},
		"valid-australia": {
			val: types.StringValue("Australia/Sydney"),
		},
		"valid-utc": {
			val: types.StringValue("UTC"),
		},
		"invalid-with-space": {
			val:         types.StringValue("America/New York"),
			expectError: true,
		},
		"invalid-nonexistent": {
			val:         types.StringValue("NonExistent/Timezone"),
			expectError: true,
		},
		"invalid-empty-string": {
			val:         types.StringValue(""),
			expectError: true,
		},
		"invalid-just-region": {
			val:         types.StringValue("America"),
			expectError: true,
		},
		"invalid-lowercase": {
			val:         types.StringValue("america/los_angeles"),
			expectError: true,
		},
		"invalid-utc-offset": {
			val:         types.StringValue("UTC+01:00"),
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
			validators.Timezone().ValidateString(context.Background(), request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}

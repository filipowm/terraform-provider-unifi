package validators_test

import (
	"context"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestHostnameValidator(t *testing.T) {
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
			val: types.StringValue("example.com"),
		},
		"valid-subdomain": {
			val: types.StringValue("sub.example.com"),
		},
		"valid-multiple-subdomains": {
			val: types.StringValue("a.b.c.example.com"),
		},
		"valid-with-hyphen": {
			val: types.StringValue("my-hostname.example.com"),
		},
		"valid-with-numbers": {
			val: types.StringValue("example123.com"),
		},
		"valid-tld-with-numbers": {
			val: types.StringValue("example.co2"),
		},
		"invalid-with-scheme": {
			val:         types.StringValue("http://example.com"),
			expectError: true,
		},
		"invalid-with-https-scheme": {
			val:         types.StringValue("https://example.com"),
			expectError: true,
		},
		"invalid-with-path": {
			val:         types.StringValue("example.com/path"),
			expectError: true,
		},
		"invalid-with-port": {
			val:         types.StringValue("example.com:8080"),
			expectError: true,
		},
		"invalid-with-underscore": {
			val:         types.StringValue("invalid_hostname.com"),
			expectError: true,
		},
		"invalid-single-label": {
			val:         types.StringValue("localhost"),
			expectError: true,
		},
		"invalid-ends-with-hyphen": {
			val:         types.StringValue("hostname-.com"),
			expectError: true,
		},
		"invalid-begins-with-hyphen": {
			val:         types.StringValue("-hostname.com"),
			expectError: true,
		},
		"invalid-empty-string": {
			val:         types.StringValue(""),
			expectError: true,
		},
		"invalid-special-chars": {
			val:         types.StringValue("hostname!.com"),
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
			validators.Hostname().ValidateString(context.Background(), request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}

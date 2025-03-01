package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestURLValidator(t *testing.T) {
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
		"valid-http": {
			val: types.StringValue("http://example.com"),
		},
		"valid-https": {
			val: types.StringValue("https://example.com"),
		},
		"valid-with-path": {
			val: types.StringValue("https://example.com/path"),
		},
		"valid-with-query": {
			val: types.StringValue("https://example.com/path?query=value"),
		},
		"valid-with-port": {
			val: types.StringValue("https://example.com:8443"),
		},
		"invalid-no-scheme": {
			val:         types.StringValue("example.com"),
			expectError: true,
		},
		"invalid-no-host": {
			val:         types.StringValue("https://"),
			expectError: true,
		},
		"invalid-malformed": {
			val:         types.StringValue("htt ps://example.com"),
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
			URL().ValidateString(context.Background(), request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}

func TestHTTPSURLValidator(t *testing.T) {
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
		"valid-https": {
			val: types.StringValue("https://example.com"),
		},
		"valid-with-path": {
			val: types.StringValue("https://example.com/path"),
		},
		"invalid-http": {
			val:         types.StringValue("http://example.com"),
			expectError: true,
		},
		"invalid-no-scheme": {
			val:         types.StringValue("example.com"),
			expectError: true,
		},
		"invalid-no-host": {
			val:         types.StringValue("https://"),
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
			HTTPSUrl().ValidateString(context.Background(), request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}

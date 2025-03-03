package validators

import (
	"context"
	"testing"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
)

func TestControllerVersionValidator_Description(t *testing.T) {
	tests := []struct {
		name        string
		validator   ControllerVersionValidator
		expected    string
		description string
	}{
		{
			name: "exact version",
			validator: ControllerVersionValidator{
				exactVersion: base.AsVersion("7.0.0"),
			},
			expected:    "Validates that the controller version is exactly 7.0.0",
			description: "Should describe exact version check",
		},
		{
			name: "min version",
			validator: ControllerVersionValidator{
				minVersion: base.AsVersion("7.0.0"),
			},
			expected:    "Validates that the controller version is at least 7.0.0",
			description: "Should describe minimum version check",
		},
		{
			name: "max version",
			validator: ControllerVersionValidator{
				maxVersion: base.AsVersion("7.0.0"),
			},
			expected:    "Validates that the controller version is at most 7.0.0",
			description: "Should describe maximum version check",
		},
		{
			name: "version range",
			validator: ControllerVersionValidator{
				minVersion: base.AsVersion("7.0.0"),
				maxVersion: base.AsVersion("8.0.0"),
			},
			expected:    "Validates that the controller version is between 7.0.0 and 8.0.0",
			description: "Should describe version range check",
		},
		{
			name:        "no constraint",
			validator:   ControllerVersionValidator{},
			expected:    "Validates the controller version",
			description: "Should provide generic description when no constraints",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			actual := test.validator.Description(ctx)
			assert.Equal(t, test.expected, actual, test.description)
		})
	}
}

func TestControllerVersionValidator_ValidateResource(t *testing.T) {
	tests := []struct {
		name              string
		controllerVersion string
		validator         ControllerVersionValidator
		expectError       bool
		description       string
	}{
		{
			name:              "exact version match",
			controllerVersion: "7.0.0",
			validator: ControllerVersionValidator{
				exactVersion: base.AsVersion("7.0.0"),
			},
			expectError: false,
			description: "Should pass when exact version matches",
		},
		{
			name:              "exact version mismatch",
			controllerVersion: "7.0.0",
			validator: ControllerVersionValidator{
				exactVersion: base.AsVersion("7.1.0"),
			},
			expectError: true,
			description: "Should fail when exact version doesn't match",
		},
		{
			name:              "min version satisfied",
			controllerVersion: "7.5.0",
			validator: ControllerVersionValidator{
				minVersion: base.AsVersion("7.0.0"),
			},
			expectError: false,
			description: "Should pass when version meets minimum",
		},
		{
			name:              "min version not satisfied",
			controllerVersion: "6.5.0",
			validator: ControllerVersionValidator{
				minVersion: base.AsVersion("7.0.0"),
			},
			expectError: true,
			description: "Should fail when version doesn't meet minimum",
		},
		{
			name:              "max version satisfied",
			controllerVersion: "6.5.0",
			validator: ControllerVersionValidator{
				maxVersion: base.AsVersion("7.0.0"),
			},
			expectError: false,
			description: "Should pass when version is below maximum",
		},
		{
			name:              "max version not satisfied",
			controllerVersion: "7.5.0",
			validator: ControllerVersionValidator{
				maxVersion: base.AsVersion("7.0.0"),
			},
			expectError: true,
			description: "Should fail when version exceeds maximum",
		},
		{
			name:              "version in range",
			controllerVersion: "7.5.0",
			validator: ControllerVersionValidator{
				minVersion: base.AsVersion("7.0.0"),
				maxVersion: base.AsVersion("8.0.0"),
			},
			expectError: false,
			description: "Should pass when version is in range",
		},
		{
			name:              "version below range",
			controllerVersion: "6.5.0",
			validator: ControllerVersionValidator{
				minVersion: base.AsVersion("7.0.0"),
				maxVersion: base.AsVersion("8.0.0"),
			},
			expectError: true,
			description: "Should fail when version is below range",
		},
		{
			name:              "version above range",
			controllerVersion: "8.5.0",
			validator: ControllerVersionValidator{
				minVersion: base.AsVersion("7.0.0"),
				maxVersion: base.AsVersion("8.0.0"),
			},
			expectError: true,
			description: "Should fail when version is above range",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()

			// Create a mock client with the specified version
			mockClient := &base.Client{
				Version: version.Must(version.NewVersion(test.controllerVersion)),
			}

			// Update the validator with the mock client
			test.validator.client = mockClient

			// Create request and response objects
			req := resource.ValidateConfigRequest{}
			resp := resource.ValidateConfigResponse{
				Diagnostics: diag.Diagnostics{},
			}

			// Call the validator
			test.validator.ValidateResource(ctx, req, &resp)

			// Check if the result matches expectations
			if test.expectError {
				assert.True(t, resp.Diagnostics.HasError(), test.description)
			} else {
				assert.False(t, resp.Diagnostics.HasError(), test.description)
			}
		})
	}
}

func TestResourceHelperFunctions(t *testing.T) {
	mockClient := &base.Client{
		Version: base.AsVersion("7.5.0"),
	}

	tests := []struct {
		name        string
		validator   resource.ConfigValidator
		expectError bool
		description string
	}{
		{
			name:        "ResourceRequireMinVersion passing",
			validator:   ResourceRequireMinVersion(mockClient, "7.0.0", ""),
			expectError: false,
			description: "ResourceRequireMinVersion should pass with sufficient version",
		},
		{
			name:        "ResourceRequireMinVersion failing",
			validator:   ResourceRequireMinVersion(mockClient, "8.0.0", ""),
			expectError: true,
			description: "ResourceRequireMinVersion should fail with insufficient version",
		},
		{
			name:        "ResourceRequireMaxVersion passing",
			validator:   ResourceRequireMaxVersion(mockClient, "8.0.0", ""),
			expectError: false,
			description: "ResourceRequireMaxVersion should pass with acceptable version",
		},
		{
			name:        "ResourceRequireMaxVersion failing",
			validator:   ResourceRequireMaxVersion(mockClient, "7.0.0", ""),
			expectError: true,
			description: "ResourceRequireMaxVersion should fail with too high version",
		},
		{
			name:        "ResourceRequireVersionRange passing",
			validator:   ResourceRequireVersionRange(mockClient, "7.0.0", "8.0.0", ""),
			expectError: false,
			description: "ResourceRequireVersionRange should pass with version in range",
		},
		{
			name:        "ResourceRequireVersionRange failing (below)",
			validator:   ResourceRequireVersionRange(mockClient, "7.6.0", "8.0.0", ""),
			expectError: true,
			description: "ResourceRequireVersionRange should fail with version below range",
		},
		{
			name:        "ResourceRequireVersionRange failing (above)",
			validator:   ResourceRequireVersionRange(mockClient, "6.0.0", "7.0.0", ""),
			expectError: true,
			description: "ResourceRequireVersionRange should fail with version above range",
		},
		{
			name:        "ResourceRequireExactVersion passing",
			validator:   ResourceRequireExactVersion(mockClient, "7.5.0", ""),
			expectError: false,
			description: "ResourceRequireExactVersion should pass with exact version match",
		},
		{
			name:        "ResourceRequireExactVersion failing",
			validator:   ResourceRequireExactVersion(mockClient, "7.5.1", ""),
			expectError: true,
			description: "ResourceRequireExactVersion should fail with version mismatch",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()

			// Create request and response objects
			req := resource.ValidateConfigRequest{}
			resp := resource.ValidateConfigResponse{
				Diagnostics: diag.Diagnostics{},
			}

			// Call the validator
			test.validator.(ControllerVersionValidator).ValidateResource(ctx, req, &resp)

			// Check if the result matches expectations
			if test.expectError {
				assert.True(t, resp.Diagnostics.HasError(), test.description)
			} else {
				assert.False(t, resp.Diagnostics.HasError(), test.description)
			}
		})
	}
}

func TestDatasourceHelperFunctions(t *testing.T) {
	mockClient := &base.Client{
		Version: base.AsVersion("7.5.0"),
	}

	tests := []struct {
		name        string
		validator   datasource.ConfigValidator
		expectError bool
		description string
	}{
		{
			name:        "DatasourceRequireMinVersion passing",
			validator:   DatasourceRequireMinVersion(mockClient, "7.0.0", ""),
			expectError: false,
			description: "DatasourceRequireMinVersion should pass with sufficient version",
		},
		{
			name:        "DatasourceRequireMinVersion failing",
			validator:   DatasourceRequireMinVersion(mockClient, "8.0.0", ""),
			expectError: true,
			description: "DatasourceRequireMinVersion should fail with insufficient version",
		},
		{
			name:        "DatasourceRequireMaxVersion passing",
			validator:   DatasourceRequireMaxVersion(mockClient, "8.0.0", ""),
			expectError: false,
			description: "DatasourceRequireMaxVersion should pass with acceptable version",
		},
		{
			name:        "DatasourceRequireMaxVersion failing",
			validator:   DatasourceRequireMaxVersion(mockClient, "7.0.0", ""),
			expectError: true,
			description: "DatasourceRequireMaxVersion should fail with too high version",
		},
		{
			name:        "DatasourceRequireVersionRange passing",
			validator:   DatasourceRequireVersionRange(mockClient, "7.0.0", "8.0.0", ""),
			expectError: false,
			description: "DatasourceRequireVersionRange should pass with version in range",
		},
		{
			name:        "DatasourceRequireVersionRange failing (below)",
			validator:   DatasourceRequireVersionRange(mockClient, "7.6.0", "8.0.0", ""),
			expectError: true,
			description: "DatasourceRequireVersionRange should fail with version below range",
		},
		{
			name:        "DatasourceRequireVersionRange failing (above)",
			validator:   DatasourceRequireVersionRange(mockClient, "6.0.0", "7.0.0", ""),
			expectError: true,
			description: "DatasourceRequireVersionRange should fail with version above range",
		},
		{
			name:        "DatasourceRequireExactVersion passing",
			validator:   DatasourceRequireExactVersion(mockClient, "7.5.0", ""),
			expectError: false,
			description: "DatasourceRequireExactVersion should pass with exact version match",
		},
		{
			name:        "DatasourceRequireExactVersion failing",
			validator:   DatasourceRequireExactVersion(mockClient, "7.5.1", ""),
			expectError: true,
			description: "DatasourceRequireExactVersion should fail with version mismatch",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()

			// Create request and response objects
			req := datasource.ValidateConfigRequest{}
			resp := datasource.ValidateConfigResponse{
				Diagnostics: diag.Diagnostics{},
			}

			// Call the validator
			test.validator.(ControllerVersionValidator).ValidateDataSource(ctx, req, &resp)

			// Check if the result matches expectations
			if test.expectError {
				assert.True(t, resp.Diagnostics.HasError(), test.description)
			} else {
				assert.False(t, resp.Diagnostics.HasError(), test.description)
			}
		})
	}
}

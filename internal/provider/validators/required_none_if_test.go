package validators_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
)

// Common test case structure for string conditions
type requiredNoneIfTestCase struct {
	condition       types.String
	field1          types.String
	field2          types.String
	expectError     bool
	expectErrorText string
}

// Common test case structure for bool conditions
type requiredNoneIfBoolTestCase struct {
	condition       types.Bool
	field1          types.String
	field2          types.String
	expectError     bool
	expectErrorText string
}

// Function to create a schema object with string condition
func createRequiredNoneIfSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"condition": schema.StringAttribute{
				Optional: true,
			},
			"field1": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"field2": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Function to create a schema object with bool condition
func createRequiredNoneIfBoolSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"condition": schema.BoolAttribute{
				Optional: true,
			},
			"field1": schema.StringAttribute{
				Optional: true,
			},
			"field2": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

// Function to create a config with string condition
func createRequiredNoneIfConfig(schema schema.Schema, testCase requiredNoneIfTestCase) tfsdk.Config {
	return tfsdk.Config{
		Schema: schema,
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"condition": tftypes.String,
					"field1":    tftypes.String,
					"field2":    tftypes.String,
				},
			},
			map[string]tftypes.Value{
				"condition": stringToTfValue(testCase.condition),
				"field1":    stringToTfValue(testCase.field1),
				"field2":    stringToTfValue(testCase.field2),
			},
		),
	}
}

// Function to create a config with bool condition
func createRequiredNoneIfBoolConfig(schema schema.Schema, testCase requiredNoneIfBoolTestCase) tfsdk.Config {
	return tfsdk.Config{
		Schema: schema,
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"condition": tftypes.Bool,
					"field1":    tftypes.String,
					"field2":    tftypes.String,
				},
			},
			map[string]tftypes.Value{
				"condition": boolToTfValue(testCase.condition),
				"field1":    stringToTfValue(testCase.field1),
				"field2":    stringToTfValue(testCase.field2),
			},
		),
	}
}

// Function to verify validation results
func verifyRequiredNoneIfResults(t *testing.T, diagnostics error, testCase requiredNoneIfTestCase) {
	if testCase.expectError {
		assert.Error(t, diagnostics)
	} else {
		assert.NoError(t, diagnostics)
	}
}

// Test RequiredNoneIf with string condition
func TestRequiredNoneIf(t *testing.T) {
	testCases := map[string]requiredNoneIfTestCase{
		"matching_condition_all_configured": {
			condition:       types.StringValue("test"),
			field1:          types.StringValue("value1"),
			field2:          types.StringValue("value2"),
			expectError:     true,
			expectErrorText: "If \"condition\" equals \"test\", any of those attributes must not be configured: [field1,field2]",
		},
		"matching_condition_one_configured": {
			condition:       types.StringValue("test"),
			field1:          types.StringValue("value1"),
			field2:          types.StringNull(),
			expectError:     true,
			expectErrorText: "If \"condition\" equals \"test\", any of those attributes must not be configured: [field1,field2]",
		},
		"matching_condition_none_configured": {
			condition:   types.StringValue("test"),
			field1:      types.StringNull(),
			field2:      types.StringNull(),
			expectError: false,
		},
		"non_matching_condition_all_configured": {
			condition:   types.StringValue("non-test"),
			field1:      types.StringValue("value1"),
			field2:      types.StringValue("value2"),
			expectError: false,
		},
		"matching_condition_unknown_values": {
			condition:   types.StringValue("test"),
			field1:      types.StringUnknown(),
			field2:      types.StringValue("value2"),
			expectError: false, // Unknown values should skip validation
		},
		"null_condition_all_configured": {
			condition:   types.StringNull(),
			field1:      types.StringValue("value1"),
			field2:      types.StringValue("value2"),
			expectError: false,
		},
		"unknown_condition_all_configured": {
			condition:   types.StringUnknown(),
			field1:      types.StringValue("value1"),
			field2:      types.StringValue("value2"),
			expectError: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			schema := createRequiredNoneIfSchema()
			config := createRequiredNoneIfConfig(schema, testCase)

			validator := validators.RequiredNoneIf(
				path.MatchRoot("condition"),
				types.StringValue("test"),
				path.MatchRoot("field1"),
				path.MatchRoot("field2"),
			)

			diagnostics := validator.Validate(ctx, config)

			if testCase.expectError {
				assert.True(t, diagnostics.HasError())
				if testCase.expectErrorText != "" {
					found := false
					for _, diag := range diagnostics {
						if diag.Detail() != "" && diag.Detail() == testCase.expectErrorText {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected error text not found")
				}
			} else {
				assert.False(t, diagnostics.HasError())
			}
		})
	}
}

// Test RequiredNoneIfSet with string condition
func TestRequiredNoneIfSet(t *testing.T) {
	testCases := map[string]requiredNoneIfTestCase{
		"condition_set_all_configured": {
			condition:       types.StringValue("any-value"),
			field1:          types.StringValue("value1"),
			field2:          types.StringValue("value2"),
			expectError:     true,
			expectErrorText: "If \"condition\" is set, any of those attributes must not be configured: [field1,field2]",
		},
		"condition_set_one_configured": {
			condition:       types.StringValue("any-value"),
			field1:          types.StringValue("value1"),
			field2:          types.StringNull(),
			expectError:     true,
			expectErrorText: "If \"condition\" is set, any of those attributes must not be configured: [field1,field2]",
		},
		"condition_set_none_configured": {
			condition:   types.StringValue("any-value"),
			field1:      types.StringNull(),
			field2:      types.StringNull(),
			expectError: false,
		},
		"condition_null_all_configured": {
			condition:   types.StringNull(),
			field1:      types.StringValue("value1"),
			field2:      types.StringValue("value2"),
			expectError: false,
		},
		"condition_unknown_all_configured": {
			condition:   types.StringUnknown(),
			field1:      types.StringValue("value1"),
			field2:      types.StringValue("value2"),
			expectError: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			schema := createRequiredNoneIfSchema()
			config := createRequiredNoneIfConfig(schema, testCase)

			validator := validators.RequiredNoneIfSet(
				path.MatchRoot("condition"),
				path.MatchRoot("field1"),
				path.MatchRoot("field2"),
			)

			diagnostics := validator.Validate(ctx, config)

			if testCase.expectError {
				assert.True(t, diagnostics.HasError())
				if testCase.expectErrorText != "" {
					found := false
					for _, diag := range diagnostics {
						if diag.Detail() != "" && diag.Detail() == testCase.expectErrorText {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected error text not found")
				}
			} else {
				assert.False(t, diagnostics.HasError())
			}
		})
	}
}

// Test RequiredNoneIf with boolean condition
func TestRequiredNoneIfWithBoolCondition(t *testing.T) {
	testCases := map[string]requiredNoneIfBoolTestCase{
		"matching_true_condition_all_configured": {
			condition:       types.BoolValue(true),
			field1:          types.StringValue("value1"),
			field2:          types.StringValue("value2"),
			expectError:     true,
			expectErrorText: "If \"condition\" equals true, any of those attributes must not be configured: [field1,field2]",
		},
		"matching_true_condition_none_configured": {
			condition:   types.BoolValue(true),
			field1:      types.StringNull(),
			field2:      types.StringNull(),
			expectError: false,
		},
		"non_matching_false_condition_all_configured": {
			condition:   types.BoolValue(false),
			field1:      types.StringValue("value1"),
			field2:      types.StringValue("value2"),
			expectError: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			schema := createRequiredNoneIfBoolSchema()
			config := createRequiredNoneIfBoolConfig(schema, testCase)

			validator := validators.RequiredNoneIf(
				path.MatchRoot("condition"),
				types.BoolValue(true),
				path.MatchRoot("field1"),
				path.MatchRoot("field2"),
			)

			diagnostics := validator.Validate(ctx, config)

			if testCase.expectError {
				assert.True(t, diagnostics.HasError())
				if testCase.expectErrorText != "" {
					found := false
					for _, diag := range diagnostics {
						if diag.Detail() != "" && diag.Detail() == testCase.expectErrorText {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected error text not found")
				}
			} else {
				assert.False(t, diagnostics.HasError())
			}
		})
	}
}

// Test ValidateDataSource method
func TestRequiredNoneIfValidateDataSource(t *testing.T) {
	testCases := map[string]requiredNoneIfTestCase{
		"matching_condition_all_configured": {
			condition:   types.StringValue("test"),
			field1:      types.StringValue("value1"),
			field2:      types.StringValue("value2"),
			expectError: true,
		},
		"matching_condition_none_configured": {
			condition:   types.StringValue("test"),
			field1:      types.StringNull(),
			field2:      types.StringNull(),
			expectError: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			schema := createRequiredNoneIfSchema()
			config := createRequiredNoneIfConfig(schema, testCase)

			validator := validators.RequiredNoneIf(
				path.MatchRoot("condition"),
				types.StringValue("test"),
				path.MatchRoot("field1"),
				path.MatchRoot("field2"),
			)

			request := datasource.ValidateConfigRequest{
				Config: config,
			}

			response := &datasource.ValidateConfigResponse{}

			validator.ValidateDataSource(ctx, request, response)

			if testCase.expectError {
				assert.True(t, response.Diagnostics.HasError())
			} else {
				assert.False(t, response.Diagnostics.HasError())
			}
		})
	}
}

// Test ValidateProvider method
func TestRequiredNoneIfValidateProvider(t *testing.T) {
	testCases := map[string]requiredNoneIfTestCase{
		"matching_condition_all_configured": {
			condition:   types.StringValue("test"),
			field1:      types.StringValue("value1"),
			field2:      types.StringValue("value2"),
			expectError: true,
		},
		"matching_condition_none_configured": {
			condition:   types.StringValue("test"),
			field1:      types.StringNull(),
			field2:      types.StringNull(),
			expectError: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			schema := createRequiredNoneIfSchema()
			config := createRequiredNoneIfConfig(schema, testCase)

			validator := validators.RequiredNoneIf(
				path.MatchRoot("condition"),
				types.StringValue("test"),
				path.MatchRoot("field1"),
				path.MatchRoot("field2"),
			)

			request := provider.ValidateConfigRequest{
				Config: config,
			}

			response := &provider.ValidateConfigResponse{}

			validator.ValidateProvider(ctx, request, response)

			if testCase.expectError {
				assert.True(t, response.Diagnostics.HasError())
			} else {
				assert.False(t, response.Diagnostics.HasError())
			}
		})
	}
}

// Test ValidateResource method
func TestRequiredNoneIfValidateResource(t *testing.T) {
	testCases := map[string]requiredNoneIfTestCase{
		"matching_condition_all_configured": {
			condition:   types.StringValue("test"),
			field1:      types.StringValue("value1"),
			field2:      types.StringValue("value2"),
			expectError: true,
		},
		"matching_condition_none_configured": {
			condition:   types.StringValue("test"),
			field1:      types.StringNull(),
			field2:      types.StringNull(),
			expectError: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			schema := createRequiredNoneIfSchema()
			config := createRequiredNoneIfConfig(schema, testCase)

			validator := validators.RequiredNoneIf(
				path.MatchRoot("condition"),
				types.StringValue("test"),
				path.MatchRoot("field1"),
				path.MatchRoot("field2"),
			)

			request := resource.ValidateConfigRequest{
				Config: config,
			}

			response := &resource.ValidateConfigResponse{}

			validator.ValidateResource(ctx, request, response)

			if testCase.expectError {
				assert.True(t, response.Diagnostics.HasError())
			} else {
				assert.False(t, response.Diagnostics.HasError())
			}
		})
	}
}

// Test the Description and MarkdownDescription methods for both variants
func TestRequiredNoneIfDescription(t *testing.T) {
	t.Run("RequiredNoneIf description", func(t *testing.T) {
		validator := validators.RequiredNoneIf(
			path.MatchRoot("condition"),
			types.StringValue("test"),
			path.MatchRoot("field1"),
			path.MatchRoot("field2"),
		)

		ctx := context.Background()
		desc := validator.Description(ctx)
		mdDesc := validator.MarkdownDescription(ctx)

		expectedDesc := `If "condition" equals "test", any of those attributes must not be configured: [field1,field2]`
		assert.Equal(t, expectedDesc, desc)
		assert.Equal(t, expectedDesc, mdDesc)
	})

	t.Run("RequiredNoneIfSet description", func(t *testing.T) {
		validator := validators.RequiredNoneIfSet(
			path.MatchRoot("condition"),
			path.MatchRoot("field1"),
			path.MatchRoot("field2"),
		)

		ctx := context.Background()
		desc := validator.Description(ctx)
		mdDesc := validator.MarkdownDescription(ctx)

		expectedDesc := `If "condition" is set, any of those attributes must not be configured: [field1,field2]`
		assert.Equal(t, expectedDesc, desc)
		assert.Equal(t, expectedDesc, mdDesc)
	})
}

// Test with missing path
func TestRequiredNoneIfWithMissingPath(t *testing.T) {
	ctx := context.Background()
	schema := createRequiredNoneIfSchema()
	testCase := requiredNoneIfTestCase{
		condition: types.StringValue("test"),
		field1:    types.StringValue("value1"),
		field2:    types.StringValue("value2"),
	}
	config := createRequiredNoneIfConfig(schema, testCase)

	validator := validators.RequiredNoneIf(
		path.MatchRoot("non_existent"),
		types.StringValue("test"),
		path.MatchRoot("field1"),
		path.MatchRoot("field2"),
	)

	diagnostics := validator.Validate(ctx, config)

	// Should not get an error because the condition path doesn't match
	assert.False(t, diagnostics.HasError())
}

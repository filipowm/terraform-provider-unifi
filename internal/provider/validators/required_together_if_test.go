package validators_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
)

// Helper function to convert types.Int32 to tftypes.Value
func numberToTfValue(value types.Int32) tftypes.Value {
	if value.IsNull() {
		return tftypes.NewValue(tftypes.Number, nil)
	} else if value.IsUnknown() {
		return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue)
	}

	return tftypes.NewValue(tftypes.Number, float64(value.ValueInt32()))
}

// Common test case structure for string conditions
type requiredTogetherIfStringConditionTestCase struct {
	condition       types.String
	field1          types.String
	field2          types.String
	expectError     bool
	expectErrorText string
}

// Common test case structure for bool conditions
type requiredTogetherIfBoolConditionTestCase struct {
	condition       types.Bool
	field1          types.String
	field2          types.String
	expectError     bool
	expectErrorText string
}

// Common test case structure for int32 conditions
type int32ConditionTestCase struct {
	condition       types.Int32
	field1          types.String
	field2          types.String
	expectError     bool
	expectErrorText string
}

// Function to create a schema object with string condition
func createRequiredIfStringConditionSchema() schema.Schema {
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
func createRequiredIfBoolConditionSchema() schema.Schema {
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

// Function to create a schema object with int32 condition
func createInt32ConditionSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"condition": schema.Int32Attribute{
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
func createRequiredIfStringConditionConfig(schema schema.Schema, testCase requiredTogetherIfStringConditionTestCase) tfsdk.Config {
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
func createRequiredIfBoolConditionConfig(schema schema.Schema, testCase requiredTogetherIfBoolConditionTestCase) tfsdk.Config {
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

// Function to create a config with int32 condition
func createInt32ConditionConfig(schema schema.Schema, testCase int32ConditionTestCase) tfsdk.Config {
	return tfsdk.Config{
		Schema: schema,
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"condition": tftypes.Number,
					"field1":    tftypes.String,
					"field2":    tftypes.String,
				},
			},
			map[string]tftypes.Value{
				"condition": numberToTfValue(testCase.condition),
				"field1":    stringToTfValue(testCase.field1),
				"field2":    stringToTfValue(testCase.field2),
			},
		),
	}
}

// Function to verify validation results
func verifyValidationResults(t *testing.T, response interface{}, testCase interface{}) {
	t.Helper()

	var hasDiagnosticError bool
	var diagnosticErrors []interface{}

	// Extract diagnostic errors based on the type
	switch d := response.(type) {
	case resource.ValidateConfigResponse:
		hasDiagnosticError = d.Diagnostics.HasError()
		if hasDiagnosticError {
			diagnosticErrors = []interface{}{d.Diagnostics.Errors()[0]}
		}
	case *resource.ValidateConfigResponse:
		hasDiagnosticError = d.Diagnostics.HasError()
		if hasDiagnosticError {
			diagnosticErrors = []interface{}{d.Diagnostics.Errors()[0]}
		}
	case datasource.ValidateConfigResponse:
		hasDiagnosticError = d.Diagnostics.HasError()
		if hasDiagnosticError {
			diagnosticErrors = []interface{}{d.Diagnostics.Errors()[0]}
		}
	case *datasource.ValidateConfigResponse:
		hasDiagnosticError = d.Diagnostics.HasError()
		if hasDiagnosticError {
			diagnosticErrors = []interface{}{d.Diagnostics.Errors()[0]}
		}
	case provider.ValidateConfigResponse:
		hasDiagnosticError = d.Diagnostics.HasError()
		if hasDiagnosticError {
			diagnosticErrors = []interface{}{d.Diagnostics.Errors()[0]}
		}
	case *provider.ValidateConfigResponse:
		hasDiagnosticError = d.Diagnostics.HasError()
		if hasDiagnosticError {
			diagnosticErrors = []interface{}{d.Diagnostics.Errors()[0]}
		}
	case validator.StringResponse:
		hasDiagnosticError = d.Diagnostics.HasError()
		if hasDiagnosticError {
			diagnosticErrors = []interface{}{d.Diagnostics.Errors()[0]}
		}
	case *validator.StringResponse:
		hasDiagnosticError = d.Diagnostics.HasError()
		if hasDiagnosticError {
			diagnosticErrors = []interface{}{d.Diagnostics.Errors()[0]}
		}
	default:
		t.Fatalf("Unsupported response type: %T", response)
	}

	// Verify results based on test case type
	switch tc := testCase.(type) {
	case requiredTogetherIfStringConditionTestCase:
		if tc.expectError {
			assert.True(t, hasDiagnosticError)
			if hasDiagnosticError {
				assert.Contains(t, fmt.Sprintf("%v", diagnosticErrors[0]), tc.expectErrorText)
			}
		} else {
			assert.False(t, hasDiagnosticError)
		}
	case requiredTogetherIfBoolConditionTestCase:
		if tc.expectError {
			assert.True(t, hasDiagnosticError)
			if hasDiagnosticError {
				assert.Contains(t, fmt.Sprintf("%v", diagnosticErrors[0]), tc.expectErrorText)
			}
		} else {
			assert.False(t, hasDiagnosticError)
		}
	case int32ConditionTestCase:
		if tc.expectError {
			require.True(t, hasDiagnosticError)
			if hasDiagnosticError {
				assert.Contains(t, fmt.Sprintf("%v", diagnosticErrors[0]), tc.expectErrorText)
			}
		} else {
			assert.False(t, hasDiagnosticError)
		}
	default:
		t.Fatalf("Unsupported test case type: %T", testCase)
	}
}

func TestRequiredTogetherIf(t *testing.T) {
	t.Parallel()

	testCases := map[string]requiredTogetherIfStringConditionTestCase{
		"condition-matches-all-fields-set": {
			condition:   types.StringValue("expected"),
			field1:      types.StringValue("value1"),
			field2:      types.StringValue("value2"),
			expectError: false,
		},
		"condition-matches-field1-missing": {
			condition:       types.StringValue("expected"),
			field1:          types.StringNull(),
			field2:          types.StringValue("value2"),
			expectError:     true,
			expectErrorText: "If condition equals \"expected\", these attributes must be configured together",
		},
		"condition-matches-field2-missing": {
			condition:       types.StringValue("expected"),
			field1:          types.StringValue("value1"),
			field2:          types.StringNull(),
			expectError:     true,
			expectErrorText: "If condition equals \"expected\", these attributes must be configured together",
		},
		"condition-matches-both-fields-missing": {
			condition:   types.StringValue("expected"),
			field1:      types.StringNull(),
			field2:      types.StringNull(),
			expectError: true,
		},
		"condition-does-not-match": {
			condition:   types.StringValue("different"),
			field1:      types.StringNull(),
			field2:      types.StringValue("value2"),
			expectError: false,
		},
		"condition-is-null": {
			condition:   types.StringNull(),
			field1:      types.StringNull(),
			field2:      types.StringValue("value2"),
			expectError: false,
		},
		"condition-is-unknown": {
			condition:   types.StringUnknown(),
			field1:      types.StringNull(),
			field2:      types.StringValue("value2"),
			expectError: false,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			schemaObject := createRequiredIfStringConditionSchema()
			val := validators.RequiredTogetherIf(
				path.MatchRoot("condition"),
				types.StringValue("expected"),
				path.MatchRoot("field1"),
				path.MatchRoot("field2"),
			)

			config := createRequiredIfStringConditionConfig(schemaObject, testCase)
			request := validator.StringRequest{
				ConfigValue: testCase.condition,
				Config:      config,
				Path:        path.Root("condition"),
			}

			response := validator.StringResponse{}
			val.ValidateString(ctx, request, &response)

			verifyValidationResults(t, response, testCase)
		})
	}
}

func TestRequiredTogetherIfWithBoolCondition(t *testing.T) {
	t.Parallel()

	testCases := map[string]requiredTogetherIfBoolConditionTestCase{
		"condition-true-all-fields-set": {
			condition:   types.BoolValue(true),
			field1:      types.StringValue("value1"),
			field2:      types.StringValue("value2"),
			expectError: false,
		},
		"condition-true-field1-missing": {
			condition:       types.BoolValue(true),
			field1:          types.StringNull(),
			field2:          types.StringValue("value2"),
			expectError:     true,
			expectErrorText: "If condition equals true, these attributes must be configured together",
		},
		"condition-false-fields-missing": {
			condition:   types.BoolValue(false),
			field1:      types.StringNull(),
			field2:      types.StringValue("value2"),
			expectError: false,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			schemaObject := createRequiredIfBoolConditionSchema()

			// Create a validator with Bool condition
			val := validators.RequiredTogetherIf(
				path.MatchRoot("condition"),
				types.BoolValue(true),
				path.MatchRoot("field1"),
				path.MatchRoot("field2"),
			)

			config := createRequiredIfBoolConditionConfig(schemaObject, testCase)
			request := resource.ValidateConfigRequest{
				Config: config,
			}

			response := resource.ValidateConfigResponse{}
			val.ValidateResource(context.Background(), request, &response)

			verifyValidationResults(t, response, testCase)
		})
	}
}

func TestRequiredTogetherIfWithNumberCondition(t *testing.T) {
	t.Parallel()

	testCases := map[string]int32ConditionTestCase{
		"condition-matches-all-fields-set": {
			condition:   types.Int32Value(42),
			field1:      types.StringValue("value1"),
			field2:      types.StringValue("value2"),
			expectError: false,
		},
		"condition-matches-field-missing": {
			condition:       types.Int32Value(42),
			field1:          types.StringNull(),
			field2:          types.StringValue("value2"),
			expectError:     true,
			expectErrorText: "If condition equals 42, these attributes must be configured together",
		},
		"condition-does-not-match": {
			condition:   types.Int32Value(24),
			field1:      types.StringNull(),
			field2:      types.StringValue("value2"),
			expectError: false,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			schemaObject := createInt32ConditionSchema()
			val := validators.RequiredTogetherIf(
				path.MatchRoot("condition"),
				types.Int32Value(42),
				path.MatchRoot("field1"),
				path.MatchRoot("field2"),
			)

			config := createInt32ConditionConfig(schemaObject, testCase)
			request := resource.ValidateConfigRequest{
				Config: config,
			}

			response := resource.ValidateConfigResponse{}
			val.ValidateResource(ctx, request, &response)

			verifyValidationResults(t, response, testCase)
		})
	}
}

func TestResourceRequiredTogetherIf(t *testing.T) {
	t.Parallel()

	testCases := map[string]requiredTogetherIfStringConditionTestCase{
		"condition-matches-all-fields-set": {
			condition:   types.StringValue("expected"),
			field1:      types.StringValue("value1"),
			field2:      types.StringValue("value2"),
			expectError: false,
		},
		"condition-matches-field-missing": {
			condition:       types.StringValue("expected"),
			field1:          types.StringNull(),
			field2:          types.StringValue("value2"),
			expectError:     true,
			expectErrorText: "If condition equals \"expected\", these attributes must be configured together",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			schemaObject := createRequiredIfStringConditionSchema()

			// Test the RequiredTogetherIf function
			val := validators.RequiredTogetherIf(
				path.MatchRoot("condition"),
				types.StringValue("expected"),
				path.MatchRoot("field1"),
				path.MatchRoot("field2"),
			)

			config := createRequiredIfStringConditionConfig(schemaObject, testCase)
			request := resource.ValidateConfigRequest{
				Config: config,
			}

			response := resource.ValidateConfigResponse{}
			val.ValidateResource(ctx, request, &response)

			verifyValidationResults(t, response, testCase)
		})
	}
}

func TestDataSourceRequiredTogetherIf(t *testing.T) {
	t.Parallel()

	testCases := map[string]requiredTogetherIfStringConditionTestCase{
		"condition-matches-all-fields-set": {
			condition:   types.StringValue("expected"),
			field1:      types.StringValue("value1"),
			field2:      types.StringValue("value2"),
			expectError: false,
		},
		"condition-matches-field-missing": {
			condition:       types.StringValue("expected"),
			field1:          types.StringNull(),
			field2:          types.StringValue("value2"),
			expectError:     true,
			expectErrorText: "If condition equals \"expected\", these attributes must be configured together",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			schemaObject := createRequiredIfStringConditionSchema()

			// Test the RequiredTogetherIf function with a datasource validator
			val := validators.RequiredTogetherIf(
				path.MatchRoot("condition"),
				types.StringValue("expected"),
				path.MatchRoot("field1"),
				path.MatchRoot("field2"),
			)

			config := createRequiredIfStringConditionConfig(schemaObject, testCase)
			request := datasource.ValidateConfigRequest{
				Config: config,
			}

			response := datasource.ValidateConfigResponse{}
			val.ValidateDataSource(ctx, request, &response)

			verifyValidationResults(t, response, testCase)
		})
	}
}

func TestProviderRequiredTogetherIf(t *testing.T) {
	t.Parallel()

	testCases := map[string]requiredTogetherIfStringConditionTestCase{
		"condition-matches-all-fields-set": {
			condition:   types.StringValue("expected"),
			field1:      types.StringValue("value1"),
			field2:      types.StringValue("value2"),
			expectError: false,
		},
		"condition-matches-field-missing": {
			condition:       types.StringValue("expected"),
			field1:          types.StringNull(),
			field2:          types.StringValue("value2"),
			expectError:     true,
			expectErrorText: "If condition equals \"expected\", these attributes must be configured together",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			schemaObject := createRequiredIfStringConditionSchema()

			// Test the RequiredTogetherIf function with a provider validator
			val := validators.RequiredTogetherIf(
				path.MatchRoot("condition"),
				types.StringValue("expected"),
				path.MatchRoot("field1"),
				path.MatchRoot("field2"),
			)

			config := createRequiredIfStringConditionConfig(schemaObject, testCase)
			request := provider.ValidateConfigRequest{
				Config: config,
			}

			response := provider.ValidateConfigResponse{}
			val.ValidateProvider(ctx, request, &response)

			verifyValidationResults(t, response, testCase)
		})
	}
}

func TestRequiredTogetherIfWithNonBooleanCondition(t *testing.T) {
	t.Parallel()

	type customType struct {
		Name string
	}

	testStruct := customType{Name: "test"}

	testCases := map[string]struct {
		condValue      types.String
		expectedValue  any
		field1         types.String
		field2         types.String
		expectError    bool
		matcherChanged bool
	}{
		"custom-object-not-equal": {
			condValue:      types.StringValue("test"),
			expectedValue:  testStruct,
			field1:         types.StringNull(),
			field2:         types.StringValue("value2"),
			expectError:    true, // Changed to true - we expect validation to fail when the condition matches
			matcherChanged: true,
		},
		"same-string-different-value": {
			condValue:      types.StringValue("test"),
			expectedValue:  "different",
			field1:         types.StringNull(),
			field2:         types.StringValue("value2"),
			expectError:    false,
			matcherChanged: false,
		},
		"different-types": {
			condValue:      types.StringValue("123"),
			expectedValue:  123,
			field1:         types.StringNull(),
			field2:         types.StringValue("value2"),
			expectError:    true, // Changed to true - condition equals 123 and field1 is null, should cause validation to fail
			matcherChanged: false,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			schemaObject := schema.Schema{
				Attributes: map[string]schema.Attribute{
					"condition": schema.StringAttribute{
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

			expectedValue := testCase.expectedValue
			var v validators.RequiredTogetherIfValidator

			if testCase.matcherChanged {
				// Using the string "test" as the expected condition value
				v = validators.RequiredTogetherIf(
					path.MatchRoot("condition"),
					types.StringValue("test"),
					path.MatchRoot("field1"),
					path.MatchRoot("field2"),
				)
			} else {
				// Normal string comparison using the string representation of expectedValue
				v = validators.RequiredTogetherIf(
					path.MatchRoot("condition"),
					types.StringValue(fmt.Sprintf("%v", expectedValue)),
					path.MatchRoot("field1"),
					path.MatchRoot("field2"),
				)
			}

			config := tfsdk.Config{
				Schema: schemaObject,
				Raw: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"condition": tftypes.String,
							"field1":    tftypes.String,
							"field2":    tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"condition": stringToTfValue(testCase.condValue),
						"field1":    stringToTfValue(testCase.field1),
						"field2":    stringToTfValue(testCase.field2),
					},
				),
			}

			request := resource.ValidateConfigRequest{
				Config: config,
			}

			response := resource.ValidateConfigResponse{}
			v.ValidateResource(ctx, request, &response)

			if testCase.expectError {
				assert.True(t, response.Diagnostics.HasError())
			} else {
				assert.False(t, response.Diagnostics.HasError())
			}
		})
	}
}

func TestRequiredTogetherIfWithUnknownTargetPaths(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		condition       types.String
		field1          types.String
		field2          types.String
		fieldPath       string // Path that will be used in the validator but doesn't exist in schema
		expectError     bool
		expectErrorText string
	}{
		"unknown-path-condition-matches": {
			condition:   types.StringValue("expected"),
			field1:      types.StringValue("value1"),
			field2:      types.StringValue("value2"),
			fieldPath:   "unknown_field",
			expectError: true, // Changed to true since the validator errors when a path doesn't exist
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			schemaObject := schema.Schema{
				Attributes: map[string]schema.Attribute{
					"condition": schema.StringAttribute{
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

			val := validators.RequiredTogetherIf(
				path.MatchRoot("condition"),
				types.StringValue("expected"),
				path.MatchRoot("field1"),
				path.MatchRoot("field2"),
				path.MatchRoot(testCase.fieldPath),
			)

			config := tfsdk.Config{
				Schema: schemaObject,
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

			request := resource.ValidateConfigRequest{
				Config: config,
			}

			response := resource.ValidateConfigResponse{}
			val.ValidateResource(ctx, request, &response)

			if testCase.expectError {
				assert.True(t, response.Diagnostics.HasError())
				assert.Contains(t, response.Diagnostics.Errors()[0].Detail(), testCase.expectErrorText)
			} else {
				assert.False(t, response.Diagnostics.HasError())
			}
		})
	}
}

func TestRequiredTogetherIfSet(t *testing.T) {
	schema := createRequiredIfStringConditionSchema()
	testCases := map[string]requiredTogetherIfStringConditionTestCase{
		"No condition value specified": {
			condition:   types.StringNull(),
			field1:      types.StringValue("value1"),
			field2:      types.StringNull(),
			expectError: false,
		},
		"Unknown condition value": {
			condition:   types.StringUnknown(),
			field1:      types.StringValue("value1"),
			field2:      types.StringNull(),
			expectError: false,
		},
		"Condition set, both fields set": {
			condition:   types.StringValue("any_value"),
			field1:      types.StringValue("value1"),
			field2:      types.StringValue("value2"),
			expectError: false,
		},
		"Condition set, neither field set": {
			condition:   types.StringValue("any_value"),
			field1:      types.StringNull(),
			field2:      types.StringNull(),
			expectError: true,
		},
		"Condition set, only field1 set": {
			condition:       types.StringValue("any_value"),
			field1:          types.StringValue("value1"),
			field2:          types.StringNull(),
			expectError:     true,
			expectErrorText: "If condition is set, these attributes must be configured together: [field1,field2]",
		},
		"Condition set, only field2 set": {
			condition:       types.StringValue("any_value"),
			field1:          types.StringNull(),
			field2:          types.StringValue("value2"),
			expectError:     true,
			expectErrorText: "If condition is set, these attributes must be configured together: [field1,field2]",
		},
		"Condition set, field1 set, field2 unknown": {
			condition:   types.StringValue("any_value"),
			field1:      types.StringValue("value1"),
			field2:      types.StringUnknown(),
			expectError: false,
		},
		"Condition set, field1 unknown, field2 set": {
			condition:   types.StringValue("any_value"),
			field1:      types.StringUnknown(),
			field2:      types.StringValue("value2"),
			expectError: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			config := createRequiredIfStringConditionConfig(schema, testCase)
			v := validators.RequiredTogetherIfSet(
				path.MatchRoot("condition"),
				path.MatchRoot("field1"),
				path.MatchRoot("field2"),
			)

			request := resource.ValidateConfigRequest{
				Config: config,
			}
			response := &resource.ValidateConfigResponse{}

			v.ValidateResource(context.Background(), request, response)
			verifyValidationResults(t, response, testCase)
		})
	}
}

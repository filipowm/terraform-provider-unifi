package validators_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
type stringConditionTestCase struct {
	condition   types.String
	field1      types.String
	expectError bool
}

// Common test case structure for bool conditions
type boolConditionTestCase struct {
	condition   types.Bool
	field1      types.String
	expectError bool
}

// Function to create a schema object with string condition
func createStringConditionSchema() schema.Schema {
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
		},
	}
}

// Function to create a schema object with bool condition
func createBoolConditionSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"condition": schema.BoolAttribute{
				Optional: true,
			},
			"field1": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

// Function to create a config with string condition
func createStringConditionConfig(schema schema.Schema, testCase stringConditionTestCase) tfsdk.Config {
	return tfsdk.Config{
		Schema: schema,
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"condition": tftypes.String,
					"field1":    tftypes.String,
				},
			},
			map[string]tftypes.Value{
				"condition": stringToTfValue(testCase.condition),
				"field1":    stringToTfValue(testCase.field1),
			},
		),
	}
}

// Function to create a config with bool condition
func createBoolConditionConfig(schema schema.Schema, testCase boolConditionTestCase) tfsdk.Config {
	return tfsdk.Config{
		Schema: schema,
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"condition": tftypes.Bool,
					"field1":    tftypes.String,
				},
			},
			map[string]tftypes.Value{
				"condition": boolToTfValue(testCase.condition),
				"field1":    stringToTfValue(testCase.field1),
			},
		),
	}
}

// Mock validators
type mockResourceValidator struct {
	called bool
}

func (v *mockResourceValidator) Description(ctx context.Context) string {
	return "Mock Resource Validator"
}

func (v *mockResourceValidator) MarkdownDescription(ctx context.Context) string {
	return "Mock Resource Validator"
}

func (v *mockResourceValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	v.called = true
}

type mockProviderValidator struct {
	called bool
}

func (v *mockProviderValidator) Description(ctx context.Context) string {
	return "Mock Provider Validator"
}

func (v *mockProviderValidator) MarkdownDescription(ctx context.Context) string {
	return "Mock Provider Validator"
}

func (v *mockProviderValidator) ValidateProvider(ctx context.Context, req provider.ValidateConfigRequest, resp *provider.ValidateConfigResponse) {
	v.called = true
}

type mockDatasourceValidator struct {
	called bool
}

func (v *mockDatasourceValidator) Description(ctx context.Context) string {
	return "Mock Datasource Validator"
}

func (v *mockDatasourceValidator) MarkdownDescription(ctx context.Context) string {
	return "Mock Datasource Validator"
}

func (v *mockDatasourceValidator) ValidateDataSource(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	v.called = true
}

// Test ResourceIf with string condition
func TestResourceIf(t *testing.T) {
	testCases := map[string]struct {
		createValidator func(mock *mockResourceValidator) validators.IfValidator
		conditionValue  string
		testCase        stringConditionTestCase
		expectedCalled  bool
	}{
		"matching_condition": {
			createValidator: func(mock *mockResourceValidator) validators.IfValidator {
				return validators.ResourceIf(
					path.MatchRoot("condition"),
					types.StringValue("test"),
					mock,
				)
			},
			conditionValue: "test",
			testCase: stringConditionTestCase{
				condition: types.StringValue("test"),
				field1:    types.StringValue("value1"),
			},
			expectedCalled: true,
		},
		"non_matching_condition": {
			createValidator: func(mock *mockResourceValidator) validators.IfValidator {
				return validators.ResourceIf(
					path.MatchRoot("condition"),
					types.StringValue("test"),
					mock,
				)
			},
			conditionValue: "test",
			testCase: stringConditionTestCase{
				condition: types.StringValue("not-test"),
				field1:    types.StringValue("value1"),
			},
			expectedCalled: false,
		},
		"null_condition": {
			createValidator: func(mock *mockResourceValidator) validators.IfValidator {
				return validators.ResourceIf(
					path.MatchRoot("condition"),
					types.StringValue("test"),
					mock,
				)
			},
			conditionValue: "test",
			testCase: stringConditionTestCase{
				condition: types.StringNull(),
				field1:    types.StringValue("value1"),
			},
			expectedCalled: false,
		},
		"unknown_condition": {
			createValidator: func(mock *mockResourceValidator) validators.IfValidator {
				return validators.ResourceIf(
					path.MatchRoot("condition"),
					types.StringValue("test"),
					mock,
				)
			},
			conditionValue: "test",
			testCase: stringConditionTestCase{
				condition: types.StringUnknown(),
				field1:    types.StringValue("value1"),
			},
			expectedCalled: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			mock := &mockResourceValidator{}
			validator := testCase.createValidator(mock)

			ctx := context.Background()
			schema := createStringConditionSchema()
			config := createStringConditionConfig(schema, testCase.testCase)

			request := resource.ValidateConfigRequest{
				Config: config,
			}

			response := &resource.ValidateConfigResponse{
				Diagnostics: diag.Diagnostics{},
			}

			validator.ValidateResource(ctx, request, response)

			assert.Equal(t, testCase.expectedCalled, mock.called)
		})
	}
}

// Test ResourceIfSet with string condition
func TestResourceIfSet(t *testing.T) {
	testCases := map[string]struct {
		testCase       stringConditionTestCase
		expectedCalled bool
	}{
		"condition_set": {
			testCase: stringConditionTestCase{
				condition: types.StringValue("any-value"),
				field1:    types.StringValue("value1"),
			},
			expectedCalled: true,
		},
		"condition_null": {
			testCase: stringConditionTestCase{
				condition: types.StringNull(),
				field1:    types.StringValue("value1"),
			},
			expectedCalled: false,
		},
		"condition_unknown": {
			testCase: stringConditionTestCase{
				condition: types.StringUnknown(),
				field1:    types.StringValue("value1"),
			},
			expectedCalled: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			mock := &mockResourceValidator{}
			validator := validators.ResourceIfSet(
				path.MatchRoot("condition"),
				mock,
			)

			ctx := context.Background()
			schema := createStringConditionSchema()
			config := createStringConditionConfig(schema, testCase.testCase)

			request := resource.ValidateConfigRequest{
				Config: config,
			}

			response := &resource.ValidateConfigResponse{
				Diagnostics: diag.Diagnostics{},
			}

			validator.ValidateResource(ctx, request, response)

			assert.Equal(t, testCase.expectedCalled, mock.called)
		})
	}
}

// Test ProviderIfSet with string condition
func TestProviderIfSet(t *testing.T) {
	testCases := map[string]struct {
		testCase       stringConditionTestCase
		expectedCalled bool
	}{
		"condition_set": {
			testCase: stringConditionTestCase{
				condition: types.StringValue("any-value"),
				field1:    types.StringValue("value1"),
			},
			expectedCalled: true,
		},
		"condition_null": {
			testCase: stringConditionTestCase{
				condition: types.StringNull(),
				field1:    types.StringValue("value1"),
			},
			expectedCalled: false,
		},
		"condition_unknown": {
			testCase: stringConditionTestCase{
				condition: types.StringUnknown(),
				field1:    types.StringValue("value1"),
			},
			expectedCalled: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			mock := &mockProviderValidator{}
			validator := validators.ProviderIfSet(
				path.MatchRoot("condition"),
				mock,
			)

			ctx := context.Background()
			schema := createStringConditionSchema()
			config := createStringConditionConfig(schema, testCase.testCase)

			request := provider.ValidateConfigRequest{
				Config: config,
			}

			response := &provider.ValidateConfigResponse{
				Diagnostics: diag.Diagnostics{},
			}

			validator.ValidateProvider(ctx, request, response)

			assert.Equal(t, testCase.expectedCalled, mock.called)
		})
	}
}

// Test DatasourceIfSet with string condition
func TestDatasourceIfSet(t *testing.T) {
	testCases := map[string]struct {
		testCase       stringConditionTestCase
		expectedCalled bool
	}{
		"condition_set": {
			testCase: stringConditionTestCase{
				condition: types.StringValue("any-value"),
				field1:    types.StringValue("value1"),
			},
			expectedCalled: true,
		},
		"condition_null": {
			testCase: stringConditionTestCase{
				condition: types.StringNull(),
				field1:    types.StringValue("value1"),
			},
			expectedCalled: false,
		},
		"condition_unknown": {
			testCase: stringConditionTestCase{
				condition: types.StringUnknown(),
				field1:    types.StringValue("value1"),
			},
			expectedCalled: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			mock := &mockDatasourceValidator{}
			validator := validators.DatasourceIfSet(
				path.MatchRoot("condition"),
				mock,
			)

			ctx := context.Background()
			schema := createStringConditionSchema()
			config := createStringConditionConfig(schema, testCase.testCase)

			request := datasource.ValidateConfigRequest{
				Config: config,
			}

			response := &datasource.ValidateConfigResponse{
				Diagnostics: diag.Diagnostics{},
			}

			validator.ValidateDataSource(ctx, request, response)

			assert.Equal(t, testCase.expectedCalled, mock.called)
		})
	}
}

// Test the Description and MarkdownDescription methods for both variants
func TestIfValidatorDescription(t *testing.T) {
	t.Run("ResourceIf description", func(t *testing.T) {
		mock := &mockResourceValidator{}
		validator := validators.ResourceIf(
			path.MatchRoot("condition"),
			types.StringValue("test"),
			mock,
		)

		ctx := context.Background()
		desc := validator.Description(ctx)
		mdDesc := validator.MarkdownDescription(ctx)

		expectedDesc := `If "condition" equals "test", then check validators`
		assert.Equal(t, expectedDesc, desc)
		assert.Equal(t, expectedDesc, mdDesc)
	})

	t.Run("ResourceIfSet description", func(t *testing.T) {
		mock := &mockResourceValidator{}
		validator := validators.ResourceIfSet(
			path.MatchRoot("condition"),
			mock,
		)

		ctx := context.Background()
		desc := validator.Description(ctx)
		mdDesc := validator.MarkdownDescription(ctx)

		expectedDesc := `If "condition" is set, then check validators`
		assert.Equal(t, expectedDesc, desc)
		assert.Equal(t, expectedDesc, mdDesc)
	})
}

// Test with bool condition
func TestResourceIfWithBoolCondition(t *testing.T) {
	testCases := map[string]struct {
		testCase       boolConditionTestCase
		expectedCalled bool
	}{
		"matching_true_condition": {
			testCase: boolConditionTestCase{
				condition: types.BoolValue(true),
				field1:    types.StringValue("value1"),
			},
			expectedCalled: true,
		},
		"non_matching_false_condition": {
			testCase: boolConditionTestCase{
				condition: types.BoolValue(false),
				field1:    types.StringValue("value1"),
			},
			expectedCalled: false,
		},
		"null_condition": {
			testCase: boolConditionTestCase{
				condition: types.BoolNull(),
				field1:    types.StringValue("value1"),
			},
			expectedCalled: false,
		},
		"unknown_condition": {
			testCase: boolConditionTestCase{
				condition: types.BoolUnknown(),
				field1:    types.StringValue("value1"),
			},
			expectedCalled: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			mock := &mockResourceValidator{}
			validator := validators.ResourceIf(
				path.MatchRoot("condition"),
				types.BoolValue(true),
				mock,
			)

			ctx := context.Background()
			schema := createBoolConditionSchema()
			config := createBoolConditionConfig(schema, testCase.testCase)

			request := resource.ValidateConfigRequest{
				Config: config,
			}

			response := &resource.ValidateConfigResponse{
				Diagnostics: diag.Diagnostics{},
			}

			validator.ValidateResource(ctx, request, response)

			assert.Equal(t, testCase.expectedCalled, mock.called)
		})
	}
}

// Test with missing path
func TestIfValidatorWithMissingPath(t *testing.T) {
	mock := &mockResourceValidator{}
	validator := validators.ResourceIf(
		path.MatchRoot("non_existent"),
		types.StringValue("test"),
		mock,
	)

	ctx := context.Background()
	schema := createStringConditionSchema()
	testCase := stringConditionTestCase{
		condition: types.StringValue("test"),
		field1:    types.StringValue("value1"),
	}
	config := createStringConditionConfig(schema, testCase)

	request := resource.ValidateConfigRequest{
		Config: config,
	}

	response := &resource.ValidateConfigResponse{
		Diagnostics: diag.Diagnostics{},
	}

	validator.ValidateResource(ctx, request, response)

	// The validator should not be called because the path doesn't exist
	assert.False(t, mock.called)
}

// Test with multiple validators
func TestIfValidatorWithMultipleValidators(t *testing.T) {
	mock1 := &mockResourceValidator{}
	mock2 := &mockResourceValidator{}
	mock3 := &mockResourceValidator{}

	validator := validators.ResourceIf(
		path.MatchRoot("condition"),
		types.StringValue("test"),
		mock1, mock2, mock3,
	)

	ctx := context.Background()
	schema := createStringConditionSchema()
	testCase := stringConditionTestCase{
		condition: types.StringValue("test"),
		field1:    types.StringValue("value1"),
	}
	config := createStringConditionConfig(schema, testCase)

	request := resource.ValidateConfigRequest{
		Config: config,
	}

	response := &resource.ValidateConfigResponse{
		Diagnostics: diag.Diagnostics{},
	}

	validator.ValidateResource(ctx, request, response)

	// All validators should be called
	assert.True(t, mock1.called)
	assert.True(t, mock2.called)
	assert.True(t, mock3.called)
}

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
type requiredValueIfTestCase struct {
	conditionValue types.String
	targetValue    types.String
	expectError    bool
}

// Function to create a schema object for RequiredValueIf tests
func createRequiredValueIfSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"condition_attr": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"target_attr": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Function to create a config for RequiredValueIf tests
func createRequiredValueIfConfig(schema schema.Schema, testCase requiredValueIfTestCase) tfsdk.Config {
	var conditionValue, targetValue tftypes.Value

	if testCase.conditionValue.IsNull() {
		conditionValue = tftypes.NewValue(tftypes.String, nil)
	} else if testCase.conditionValue.IsUnknown() {
		conditionValue = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	} else {
		conditionValue = tftypes.NewValue(tftypes.String, testCase.conditionValue.ValueString())
	}

	if testCase.targetValue.IsNull() {
		targetValue = tftypes.NewValue(tftypes.String, nil)
	} else if testCase.targetValue.IsUnknown() {
		targetValue = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	} else {
		targetValue = tftypes.NewValue(tftypes.String, testCase.targetValue.ValueString())
	}

	return tfsdk.Config{
		Schema: schema,
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"condition_attr": tftypes.String,
				"target_attr":    tftypes.String,
			},
		}, map[string]tftypes.Value{
			"condition_attr": conditionValue,
			"target_attr":    targetValue,
		}),
	}
}

func TestRequiredValueIf(t *testing.T) {
	schema := createRequiredValueIfSchema()
	testCases := map[string]requiredValueIfTestCase{
		"condition match target match": {
			conditionValue: types.StringValue("active"),
			targetValue:    types.StringValue("enabled"),
			expectError:    false,
		},
		"condition match target mismatch": {
			conditionValue: types.StringValue("active"),
			targetValue:    types.StringValue("disabled"),
			expectError:    true,
		},
		"condition match target null": {
			conditionValue: types.StringValue("active"),
			targetValue:    types.StringNull(),
			expectError:    true,
		},
		"condition match target unknown": {
			conditionValue: types.StringValue("active"),
			targetValue:    types.StringUnknown(),
			expectError:    false,
		},
		"condition mismatch": {
			conditionValue: types.StringValue("inactive"),
			targetValue:    types.StringValue("disabled"),
			expectError:    false,
		},
		"condition null": {
			conditionValue: types.StringNull(),
			targetValue:    types.StringNull(),
			expectError:    false,
		},
		"condition unknown": {
			conditionValue: types.StringUnknown(),
			targetValue:    types.StringNull(),
			expectError:    false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			config := createRequiredValueIfConfig(schema, testCase)
			validator := validators.RequiredValueIf(
				path.MatchRoot("condition_attr"),
				types.StringValue("active"),
				path.MatchRoot("target_attr"),
				types.StringValue("enabled"),
			)
			diags := validator.Validate(context.Background(), config)

			if testCase.expectError {
				assert.True(t, diags.HasError(), "expected error, but got none")
			} else {
				assert.False(t, diags.HasError(), "expected no error, but got: %v", diags)
			}
		})
	}
}

func TestResourceRequiredValueIf(t *testing.T) {
	schema := createRequiredValueIfSchema()
	testCase := requiredValueIfTestCase{
		conditionValue: types.StringValue("active"),
		targetValue:    types.StringValue("disabled"),
		expectError:    true,
	}
	config := createRequiredValueIfConfig(schema, testCase)
	validator := validators.RequiredValueIf(
		path.MatchRoot("condition_attr"),
		types.StringValue("active"),
		path.MatchRoot("target_attr"),
		types.StringValue("enabled"),
	)

	resp := &resource.ValidateConfigResponse{}
	validator.ValidateResource(context.Background(), resource.ValidateConfigRequest{Config: config}, resp)
	assert.True(t, resp.Diagnostics.HasError(), "expected error, but got none")
}

func TestDataSourceRequiredValueIf(t *testing.T) {
	schema := createRequiredValueIfSchema()
	testCase := requiredValueIfTestCase{
		conditionValue: types.StringValue("active"),
		targetValue:    types.StringValue("disabled"),
		expectError:    true,
	}
	config := createRequiredValueIfConfig(schema, testCase)
	validator := validators.RequiredValueIf(
		path.MatchRoot("condition_attr"),
		types.StringValue("active"),
		path.MatchRoot("target_attr"),
		types.StringValue("enabled"),
	)

	resp := &datasource.ValidateConfigResponse{}
	validator.ValidateDataSource(context.Background(), datasource.ValidateConfigRequest{Config: config}, resp)
	assert.True(t, resp.Diagnostics.HasError(), "expected error, but got none")
}

func TestProviderRequiredValueIf(t *testing.T) {
	schema := createRequiredValueIfSchema()
	testCase := requiredValueIfTestCase{
		conditionValue: types.StringValue("active"),
		targetValue:    types.StringValue("disabled"),
		expectError:    true,
	}
	config := createRequiredValueIfConfig(schema, testCase)
	validator := validators.RequiredValueIf(
		path.MatchRoot("condition_attr"),
		types.StringValue("active"),
		path.MatchRoot("target_attr"),
		types.StringValue("enabled"),
	)

	resp := &provider.ValidateConfigResponse{}
	validator.ValidateProvider(context.Background(), provider.ValidateConfigRequest{Config: config}, resp)
	assert.True(t, resp.Diagnostics.HasError(), "expected error, but got none")
}

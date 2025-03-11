package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var (
	_ datasource.ConfigValidator = &RequiredValueIfValidator{}
	_ provider.ConfigValidator   = &RequiredValueIfValidator{}
	_ resource.ConfigValidator   = &RequiredValueIfValidator{}
)

// RequiredValueIfValidator validates that if a condition attribute is set to a specific value,
// then a target attribute must be set to a specific value.
type RequiredValueIfValidator struct {
	ConditionPath  path.Expression
	ConditionValue attr.Value
	TargetPath     path.Expression
	TargetValue    attr.Value
}

func (v RequiredValueIfValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v RequiredValueIfValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("If %s equals %v, then %s must equal %v", v.ConditionPath, v.ConditionValue, v.TargetPath, v.TargetValue)
}

func (v RequiredValueIfValidator) ValidateDataSource(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	resp.Diagnostics = v.Validate(ctx, req.Config)
}

func (v RequiredValueIfValidator) ValidateProvider(ctx context.Context, req provider.ValidateConfigRequest, resp *provider.ValidateConfigResponse) {
	resp.Diagnostics = v.Validate(ctx, req.Config)
}

func (v RequiredValueIfValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	resp.Diagnostics = v.Validate(ctx, req.Config)
}

func (v RequiredValueIfValidator) ValidateEphemeralResource(ctx context.Context, req ephemeral.ValidateConfigRequest, resp *ephemeral.ValidateConfigResponse) {
	resp.Diagnostics = v.Validate(ctx, req.Config)
}

func (v RequiredValueIfValidator) shouldValidate(ctx context.Context, config tfsdk.Config) bool {
	// First check the condition attribute's value
	matchedPaths, matchedPathsDiags := config.PathMatches(ctx, v.ConditionPath)
	if matchedPathsDiags.HasError() || len(matchedPaths) == 0 {
		return false
	}

	// Get the value of the condition attribute
	var conditionValue attr.Value
	getConditionDiags := config.GetAttribute(ctx, matchedPaths[0], &conditionValue)
	if getConditionDiags.HasError() {
		return false
	}

	// If the condition attribute is null or unknown, skip validation
	if conditionValue.IsNull() || conditionValue.IsUnknown() {
		return false
	}

	// Check if the condition matches
	return conditionValueMatches(ctx, conditionValue, v.ConditionValue)
}

func (v RequiredValueIfValidator) Validate(ctx context.Context, config tfsdk.Config) diag.Diagnostics {
	diags := diag.Diagnostics{}
	if !v.shouldValidate(ctx, config) {
		return diags
	}

	// Condition matched, now validate that the target attribute has the required value
	matchedPaths, matchedPathsDiags := config.PathMatches(ctx, v.TargetPath)
	diags.Append(matchedPathsDiags...)
	if diags.HasError() || len(matchedPaths) == 0 {
		return diags
	}

	// Get the value of the target attribute
	var targetValue attr.Value
	getTargetDiags := config.GetAttribute(ctx, matchedPaths[0], &targetValue)
	diags.Append(getTargetDiags...)
	if diags.HasError() {
		return diags
	}

	// Skip validation if the target value is unknown
	if targetValue.IsUnknown() {
		return diags
	}

	// If the target value is null or doesn't match the required value, add a diagnostic
	if targetValue.IsNull() || !conditionValueMatches(ctx, targetValue, v.TargetValue) {
		diags.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			matchedPaths[0],
			fmt.Sprintf("When %s is set to %v, %s must be set to %v", v.ConditionPath, v.ConditionValue, v.TargetPath, v.TargetValue),
		))
	}

	return diags
}

// RequiredValueIf creates a validator that ensures if a condition attribute equals a specific value,
// then a target attribute must equal a specific value.
func RequiredValueIf(conditionPath path.Expression, conditionValue attr.Value, targetPath path.Expression, targetValue attr.Value) RequiredValueIfValidator {
	return RequiredValueIfValidator{
		ConditionPath:  conditionPath,
		ConditionValue: conditionValue,
		TargetPath:     targetPath,
		TargetValue:    targetValue,
	}
}

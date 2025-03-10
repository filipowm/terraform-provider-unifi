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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var (
	_ datasource.ConfigValidator = &RequiredTogetherIfValidator{}
	_ provider.ConfigValidator   = &RequiredTogetherIfValidator{}
	_ resource.ConfigValidator   = &RequiredTogetherIfValidator{}
	_ validator.Object           = &RequiredTogetherIfValidator{}
	_ validator.String           = &RequiredTogetherIfValidator{}
	_ validator.Bool             = &RequiredTogetherIfValidator{}
)

type RequiredTogetherIfValidator struct {
	ConditionPath     path.Expression
	ConditionValue    attr.Value
	TargetExpressions path.Expressions
	CheckOnlyIfSet    bool // When true, only checks if the condition value is set (not null), not its actual value
}

func (v RequiredTogetherIfValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	resp.Diagnostics = v.Validate(ctx, req.Config)
}

func (v RequiredTogetherIfValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v RequiredTogetherIfValidator) MarkdownDescription(_ context.Context) string {
	if v.CheckOnlyIfSet {
		return fmt.Sprintf("If %q is set, these attributes must be configured together: %s", v.ConditionPath, v.TargetExpressions)
	}
	return fmt.Sprintf("If %q equals %s, these attributes must be configured together: %s", v.ConditionPath, v.ConditionValue, v.TargetExpressions)
}

func (v RequiredTogetherIfValidator) ValidateDataSource(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	resp.Diagnostics = v.Validate(ctx, req.Config)
}

func (v RequiredTogetherIfValidator) ValidateProvider(ctx context.Context, req provider.ValidateConfigRequest, resp *provider.ValidateConfigResponse) {
	resp.Diagnostics = v.Validate(ctx, req.Config)
}

func (v RequiredTogetherIfValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	resp.Diagnostics = v.Validate(ctx, req.Config)
}

func (v RequiredTogetherIfValidator) ValidateEphemeralResource(ctx context.Context, req ephemeral.ValidateConfigRequest, resp *ephemeral.ValidateConfigResponse) {
	resp.Diagnostics = v.Validate(ctx, req.Config)
}

func (v RequiredTogetherIfValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	resp.Diagnostics.Append(v.Validate(ctx, req.Config)...)
}

func (v RequiredTogetherIfValidator) shouldValidate(ctx context.Context, config tfsdk.Config) bool {
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
	if v.CheckOnlyIfSet {
		return !conditionValue.IsNull()
	}
	return conditionValueMatches(ctx, conditionValue, v.ConditionValue)
}

func (v RequiredTogetherIfValidator) Validate(ctx context.Context, config tfsdk.Config) diag.Diagnostics {
	diags := diag.Diagnostics{}
	if !v.shouldValidate(ctx, config) {
		return diags
	}

	// Condition matched, now apply the RequiredTogether validation
	configuredPaths := path.Paths{}
	foundPaths := path.Paths{}
	unknownPaths := path.Paths{}

	// Check that all target attributes are present
	for _, expression := range v.TargetExpressions {
		matchedPaths, matchedPathsDiags := config.PathMatches(ctx, expression)
		diags.Append(matchedPathsDiags...)

		// Collect all errors
		if matchedPathsDiags.HasError() {
			continue
		}

		// Capture all matched paths
		foundPaths.Append(matchedPaths...)

		for _, matchedPath := range matchedPaths {
			var value attr.Value
			getAttributeDiags := config.GetAttribute(ctx, matchedPath, &value)
			diags.Append(getAttributeDiags...)

			// Collect all errors
			if getAttributeDiags.HasError() {
				continue
			}

			// If value is unknown, collect the path to skip validation later
			if value.IsUnknown() {
				unknownPaths.Append(matchedPath)
				continue
			}

			// If value is null, move onto the next one
			if value.IsNull() {
				continue
			}

			// Value is known and not null, it is configured
			configuredPaths.Append(matchedPath)
		}
	}

	// Return early if all paths were null
	//if len(configuredPaths) == 0 {
	//	return diags
	//}

	// If there are unknown values, we cannot know if the validator should
	// succeed or not
	if len(unknownPaths) > 0 {
		return diags
	}

	// If configured paths does not equal all matched paths, then something
	// was missing
	if len(configuredPaths) != len(foundPaths) {
		diags.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			foundPaths[0],
			v.Description(ctx),
		))
	}

	return diags
}

// ValidateString method to implement the validator.String interface
func (v RequiredTogetherIfValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	resp.Diagnostics.Append(v.Validate(ctx, req.Config)...)
}

// RequiredTogetherIf creates a validator for string type attributes that ensures
// a set of target attributes are configured together if a condition attribute equals a specific value.
func RequiredTogetherIf(conditionPath path.Expression, conditionValue attr.Value, targetExpressions ...path.Expression) RequiredTogetherIfValidator {
	return RequiredTogetherIfValidator{
		ConditionPath:     conditionPath,
		ConditionValue:    conditionValue,
		TargetExpressions: targetExpressions,
		CheckOnlyIfSet:    false,
	}
}

// RequiredTogetherIfSet creates a validator that ensures a set of target attributes
// are configured together if a condition attribute is set (not null), regardless of its value.
func RequiredTogetherIfSet(conditionPath path.Expression, targetExpressions ...path.Expression) RequiredTogetherIfValidator {
	return RequiredTogetherIfValidator{
		ConditionPath:     conditionPath,
		ConditionValue:    nil, // Not used for this validator
		TargetExpressions: targetExpressions,
		CheckOnlyIfSet:    true,
	}
}

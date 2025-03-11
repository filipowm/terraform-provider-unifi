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
	_ datasource.ConfigValidator = &RequiredNoneIfValidator{}
	_ provider.ConfigValidator   = &RequiredNoneIfValidator{}
	_ resource.ConfigValidator   = &RequiredNoneIfValidator{}
	_ validator.Object           = &RequiredNoneIfValidator{}
)

type RequiredNoneIfValidator struct {
	ifValidatorBase
	TargetExpressions path.Expressions
}

func (v RequiredNoneIfValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	resp.Diagnostics = v.Validate(ctx, req.Config)
}

func (v RequiredNoneIfValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v RequiredNoneIfValidator) MarkdownDescription(_ context.Context) string {
	if v.CheckOnlyIfSet {
		return fmt.Sprintf("If %q is set, any of those attributes must not be configured: %s", v.ConditionPath, v.TargetExpressions)
	}
	return fmt.Sprintf("If %q equals %s, any of those attributes must not be configured: %s", v.ConditionPath, v.ConditionValue, v.TargetExpressions)
}

func (v RequiredNoneIfValidator) ValidateDataSource(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	resp.Diagnostics = v.Validate(ctx, req.Config)
}

func (v RequiredNoneIfValidator) ValidateProvider(ctx context.Context, req provider.ValidateConfigRequest, resp *provider.ValidateConfigResponse) {
	resp.Diagnostics = v.Validate(ctx, req.Config)
}

func (v RequiredNoneIfValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	resp.Diagnostics = v.Validate(ctx, req.Config)
}

func (v RequiredNoneIfValidator) ValidateEphemeralResource(ctx context.Context, req ephemeral.ValidateConfigRequest, resp *ephemeral.ValidateConfigResponse) {
	resp.Diagnostics = v.Validate(ctx, req.Config)
}

func (v RequiredNoneIfValidator) Validate(ctx context.Context, config tfsdk.Config) diag.Diagnostics {
	diags := diag.Diagnostics{}
	if !v.shouldValidate(ctx, config) {
		return diags
	}

	// Condition matched, now apply the RequiredNone validation
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

	if len(unknownPaths) > 0 {
		return diags
	}

	// If configured paths does not equal all matched paths, then something
	// was missing
	if len(configuredPaths) > 0 {
		diags.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			foundPaths[0],
			v.Description(ctx),
		))
	}

	return diags
}

// ValidateString method to implement the validator.String interface
func (v RequiredNoneIfValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	resp.Diagnostics.Append(v.Validate(ctx, req.Config)...)
}

// RequiredNoneIf creates a validator for attributes that ensures
// a set of target attributes are not configured together if a condition attribute equals a specific value.
func RequiredNoneIf(conditionPath path.Expression, conditionValue attr.Value, targetExpressions ...path.Expression) RequiredNoneIfValidator {
	return RequiredNoneIfValidator{
		ifValidatorBase: ifValidatorBase{
			ConditionPath:  conditionPath,
			ConditionValue: conditionValue,
			CheckOnlyIfSet: false,
		},
		TargetExpressions: targetExpressions,
	}
}

// RequiredNoneIfSet creates a validator that ensures a set of target attributes
// are configured not together if a condition attribute is set (not null), regardless of its value.
func RequiredNoneIfSet(conditionPath path.Expression, targetExpressions ...path.Expression) RequiredNoneIfValidator {
	return RequiredNoneIfValidator{
		ifValidatorBase: ifValidatorBase{
			ConditionPath:  conditionPath,
			CheckOnlyIfSet: true,
		},
		TargetExpressions: targetExpressions,
	}
}

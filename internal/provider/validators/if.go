package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var (
	_ datasource.ConfigValidator = &IfValidator{}
	_ provider.ConfigValidator   = &IfValidator{}
	_ resource.ConfigValidator   = &IfValidator{}
)

type ifValidatorBase struct {
	ConditionPath  path.Expression
	ConditionValue attr.Value
	CheckOnlyIfSet bool // When true, only checks if the condition value is set (not null), not its actual value
}

func (v ifValidatorBase) shouldValidate(ctx context.Context, config tfsdk.Config) bool {
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

type IfValidator struct {
	ifValidatorBase
	resourceValidators   []resource.ConfigValidator
	providerValidators   []provider.ConfigValidator
	datasourceValidators []datasource.ConfigValidator
}

func (v IfValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v IfValidator) MarkdownDescription(_ context.Context) string {
	if v.CheckOnlyIfSet {
		return fmt.Sprintf("If %q is set, then check validators", v.ConditionPath)
	}
	return fmt.Sprintf("If %q equals %s, then check validators", v.ConditionPath, v.ConditionValue)
}

func (v IfValidator) ValidateDataSource(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	if !v.shouldValidate(ctx, req.Config) {
		return
	}
	for _, v := range v.datasourceValidators {
		v.ValidateDataSource(ctx, req, resp)
	}
}

func (v IfValidator) ValidateProvider(ctx context.Context, req provider.ValidateConfigRequest, resp *provider.ValidateConfigResponse) {
	if !v.shouldValidate(ctx, req.Config) {
		return
	}
	for _, v := range v.providerValidators {
		v.ValidateProvider(ctx, req, resp)
	}
}

func (v IfValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if !v.shouldValidate(ctx, req.Config) {
		return
	}

	for _, v := range v.resourceValidators {
		v.ValidateResource(ctx, req, resp)
	}
}

func ResourceIf(conditionPath path.Expression, conditionValue attr.Value, validators ...resource.ConfigValidator) IfValidator {
	return IfValidator{
		ifValidatorBase: ifValidatorBase{
			ConditionPath:  conditionPath,
			ConditionValue: conditionValue,
			CheckOnlyIfSet: false,
		},
		resourceValidators: validators,
	}
}

func ResourceIfSet(conditionPath path.Expression, validators ...resource.ConfigValidator) IfValidator {
	return IfValidator{
		ifValidatorBase: ifValidatorBase{
			ConditionPath:  conditionPath,
			ConditionValue: nil,
			CheckOnlyIfSet: true,
		},
		resourceValidators: validators,
	}
}

func ProviderIfSet(conditionPath path.Expression, validators ...provider.ConfigValidator) IfValidator {
	return IfValidator{
		ifValidatorBase: ifValidatorBase{
			ConditionPath:  conditionPath,
			ConditionValue: nil,
			CheckOnlyIfSet: true,
		},
		providerValidators: validators,
	}
}
func DatasourceIfSet(conditionPath path.Expression, validators ...datasource.ConfigValidator) IfValidator {
	return IfValidator{
		ifValidatorBase: ifValidatorBase{
			ConditionPath:  conditionPath,
			ConditionValue: nil,
			CheckOnlyIfSet: true,
		},
		datasourceValidators: validators,
	}
}

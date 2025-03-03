package validators

import (
	"context"
	"fmt"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ resource.ConfigValidator   = &ControllerVersionValidator{}
	_ datasource.ConfigValidator = &ControllerVersionValidator{}
	_ validator.String           = &ControllerVersionValidator{}
	_ validator.Bool             = &ControllerVersionValidator{}
	_ validator.Int64            = &ControllerVersionValidator{}
	_ validator.Float64          = &ControllerVersionValidator{}
	_ validator.List             = &ControllerVersionValidator{}
	_ validator.Map              = &ControllerVersionValidator{}
	_ validator.Object           = &ControllerVersionValidator{}
	_ validator.Set              = &ControllerVersionValidator{}
)

// ControllerVersionValidator is a validator that checks if the UniFi controller version
// matches the specified constraints.
type ControllerVersionValidator struct {
	client           *base.Client
	minVersion       *version.Version
	maxVersion       *version.Version
	exactVersion     *version.Version
	conditionMessage string
}

// Description returns a description of the validator.
func (v ControllerVersionValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

// MarkdownDescription returns a markdown description of the validator.
func (v ControllerVersionValidator) MarkdownDescription(_ context.Context) string {
	if v.exactVersion != nil {
		return fmt.Sprintf("Validates that the controller version is exactly %s", v.exactVersion)
	}
	if v.minVersion != nil && v.maxVersion != nil {
		return fmt.Sprintf("Validates that the controller version is between %s and %s", v.minVersion, v.maxVersion)
	}
	if v.minVersion != nil {
		return fmt.Sprintf("Validates that the controller version is at least %s", v.minVersion)
	}
	if v.maxVersion != nil {
		return fmt.Sprintf("Validates that the controller version is at most %s", v.maxVersion)
	}
	return "Validates the controller version"
}

// ValidateResource validates the resource configuration.
func (v ControllerVersionValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if v.client == nil || v.client.Version == nil {
		resp.Diagnostics.AddWarning("Controller version not available", "Provider was not initialized properly. UniFi client or controller version is not available")
		return
	}

	v.validateVersion(ctx, &resp.Diagnostics)
}

// ValidateDataSource validates the datasource configuration.
func (v ControllerVersionValidator) ValidateDataSource(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	if v.client == nil || v.client.Version == nil {
		resp.Diagnostics.AddWarning("Controller version not available", "Provider was not initialized properly. UniFi client or controller version is not available")
		return
	}

	v.validateVersion(ctx, &resp.Diagnostics)
}

// validateVersion checks if the controller version meets the constraints
func (v ControllerVersionValidator) validateVersion(_ context.Context, diags *diag.Diagnostics) {
	controllerVersion := v.client.Version

	message := v.conditionMessage
	if message == "" {
		message = "Controller version does not meet requirements"
	}

	if v.exactVersion != nil && !controllerVersion.Equal(v.exactVersion) {
		diags.AddError(
			message,
			fmt.Sprintf("Controller version %s does not match required version %s", controllerVersion, v.exactVersion),
		)
		return
	}

	if v.minVersion != nil && controllerVersion.LessThan(v.minVersion) {
		diags.AddError(
			message,
			fmt.Sprintf("Controller version %s is less than minimum required version %s", controllerVersion, v.minVersion),
		)
		return
	}

	if v.maxVersion != nil && controllerVersion.GreaterThan(v.maxVersion) {
		diags.AddError(
			message,
			fmt.Sprintf("Controller version %s is greater than maximum allowed version %s", controllerVersion, v.maxVersion),
		)
		return
	}
}

// validateAttributeVersion is a helper function for attribute validators
func (v ControllerVersionValidator) validateAttributeVersion(ctx context.Context, req path.Path) diag.Diagnostics {
	diags := diag.Diagnostics{}

	if v.client == nil || v.client.Version == nil {
		diags.AddWarning("Controller version not available", "Provider was not initialized properly. UniFi client or controller version is not available")
		return diags
	}

	controllerVersion := v.client.Version

	message := v.conditionMessage
	if message == "" {
		message = "Controller version does not meet requirements"
	}

	if v.exactVersion != nil && !controllerVersion.Equal(v.exactVersion) {
		diags.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req,
			v.Description(ctx),
			fmt.Sprintf("Controller version %s does not match required version %s to use given attribute", controllerVersion, v.exactVersion),
		))
		return diags
	}

	if v.minVersion != nil && controllerVersion.LessThan(v.minVersion) {
		diags.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req,
			v.Description(ctx),
			fmt.Sprintf("Controller version %s is less than minimum required version %s to use given attribute", controllerVersion, v.minVersion),
		))
		return diags
	}

	if v.maxVersion != nil && controllerVersion.GreaterThan(v.maxVersion) {
		diags.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req,
			v.Description(ctx),
			fmt.Sprintf("Controller version %s is greater than maximum allowed version %s to use given attribute", controllerVersion, v.maxVersion),
		))
		return diags
	}

	return diags
}

// ValidateString implements validator.String
func (v ControllerVersionValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	resp.Diagnostics.Append(v.validateAttributeVersion(ctx, req.Path)...)
}

// ValidateBool implements validator.Bool
func (v ControllerVersionValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	resp.Diagnostics.Append(v.validateAttributeVersion(ctx, req.Path)...)
}

// ValidateInt64 implements validator.Int64
func (v ControllerVersionValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	resp.Diagnostics.Append(v.validateAttributeVersion(ctx, req.Path)...)
}

// ValidateFloat64 implements validator.Float64
func (v ControllerVersionValidator) ValidateFloat64(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	resp.Diagnostics.Append(v.validateAttributeVersion(ctx, req.Path)...)
}

// ValidateList implements validator.List
func (v ControllerVersionValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	resp.Diagnostics.Append(v.validateAttributeVersion(ctx, req.Path)...)
}

// ValidateMap implements validator.Map
func (v ControllerVersionValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	resp.Diagnostics.Append(v.validateAttributeVersion(ctx, req.Path)...)
}

// ValidateObject implements validator.Object
func (v ControllerVersionValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	resp.Diagnostics.Append(v.validateAttributeVersion(ctx, req.Path)...)
}

// ValidateSet implements validator.Set
func (v ControllerVersionValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	resp.Diagnostics.Append(v.validateAttributeVersion(ctx, req.Path)...)
}

// ResourceRequireMinVersion returns a resource validator that checks if the controller version
// is at least the specified version.
func ResourceRequireMinVersion(client *base.Client, minVersion string, conditionMessage string) resource.ConfigValidator {
	return ControllerVersionValidator{
		client:           client,
		minVersion:       base.AsVersion(minVersion),
		conditionMessage: conditionMessage,
	}
}

// ResourceRequireMaxVersion returns a resource validator that checks if the controller version
// is at most the specified version.
func ResourceRequireMaxVersion(client *base.Client, maxVersion string, conditionMessage string) resource.ConfigValidator {
	return ControllerVersionValidator{
		client:           client,
		maxVersion:       base.AsVersion(maxVersion),
		conditionMessage: conditionMessage,
	}
}

// ResourceRequireVersionRange returns a resource validator that checks if the controller version
// is within the specified range (inclusive).
func ResourceRequireVersionRange(client *base.Client, minVersion, maxVersion string, conditionMessage string) resource.ConfigValidator {
	return ControllerVersionValidator{
		client:           client,
		minVersion:       base.AsVersion(minVersion),
		maxVersion:       base.AsVersion(maxVersion),
		conditionMessage: conditionMessage,
	}
}

// ResourceRequireExactVersion returns a resource validator that checks if the controller version
// matches the specified version exactly.
func ResourceRequireExactVersion(client *base.Client, exactVersion string, conditionMessage string) resource.ConfigValidator {
	return ControllerVersionValidator{
		client:           client,
		exactVersion:     base.AsVersion(exactVersion),
		conditionMessage: conditionMessage,
	}
}

// DatasourceRequireMinVersion returns a datasource validator that checks if the controller version
// is at least the specified version.
func DatasourceRequireMinVersion(client *base.Client, minVersion string, conditionMessage string) datasource.ConfigValidator {
	return ControllerVersionValidator{
		client:           client,
		minVersion:       base.AsVersion(minVersion),
		conditionMessage: conditionMessage,
	}
}

// DatasourceRequireMaxVersion returns a datasource validator that checks if the controller version
// is at most the specified version.
func DatasourceRequireMaxVersion(client *base.Client, maxVersion string, conditionMessage string) datasource.ConfigValidator {
	return ControllerVersionValidator{
		client:           client,
		maxVersion:       base.AsVersion(maxVersion),
		conditionMessage: conditionMessage,
	}
}

// DatasourceRequireVersionRange returns a datasource validator that checks if the controller version
// is within the specified range (inclusive).
func DatasourceRequireVersionRange(client *base.Client, minVersion, maxVersion string, conditionMessage string) datasource.ConfigValidator {
	return ControllerVersionValidator{
		client:           client,
		minVersion:       base.AsVersion(minVersion),
		maxVersion:       base.AsVersion(maxVersion),
		conditionMessage: conditionMessage,
	}
}

// DatasourceRequireExactVersion returns a datasource validator that checks if the controller version
// matches the specified version exactly.
func DatasourceRequireExactVersion(client *base.Client, exactVersion string, conditionMessage string) datasource.ConfigValidator {
	return ControllerVersionValidator{
		client:           client,
		exactVersion:     base.AsVersion(exactVersion),
		conditionMessage: conditionMessage,
	}
}

// StringRequireMinVersion returns a string validator that checks if the controller version
// is at least the specified version.
func StringRequireMinVersion(client *base.Client, minVersion string, conditionMessage string) validator.String {
	return ControllerVersionValidator{
		client:           client,
		minVersion:       base.AsVersion(minVersion),
		conditionMessage: conditionMessage,
	}
}

// StringRequireMaxVersion returns a string validator that checks if the controller version
// is at most the specified version.
func StringRequireMaxVersion(client *base.Client, maxVersion string, conditionMessage string) validator.String {
	return ControllerVersionValidator{
		client:           client,
		maxVersion:       base.AsVersion(maxVersion),
		conditionMessage: conditionMessage,
	}
}

// StringRequireVersionRange returns a string validator that checks if the controller version
// is within the specified range (inclusive).
func StringRequireVersionRange(client *base.Client, minVersion, maxVersion string, conditionMessage string) validator.String {
	return ControllerVersionValidator{
		client:           client,
		minVersion:       base.AsVersion(minVersion),
		maxVersion:       base.AsVersion(maxVersion),
		conditionMessage: conditionMessage,
	}
}

// StringRequireExactVersion returns a string validator that checks if the controller version
// matches the specified version exactly.
func StringRequireExactVersion(client *base.Client, exactVersion string, conditionMessage string) validator.String {
	return ControllerVersionValidator{
		client:           client,
		exactVersion:     base.AsVersion(exactVersion),
		conditionMessage: conditionMessage,
	}
}

// BoolRequireMinVersion returns a bool validator that checks if the controller version
// is at least the specified version.
func BoolRequireMinVersion(client *base.Client, minVersion string, conditionMessage string) validator.Bool {
	return ControllerVersionValidator{
		client:           client,
		minVersion:       base.AsVersion(minVersion),
		conditionMessage: conditionMessage,
	}
}

// BoolRequireMaxVersion returns a bool validator that checks if the controller version
// is at most the specified version.
func BoolRequireMaxVersion(client *base.Client, maxVersion string, conditionMessage string) validator.Bool {
	return ControllerVersionValidator{
		client:           client,
		maxVersion:       base.AsVersion(maxVersion),
		conditionMessage: conditionMessage,
	}
}

// BoolRequireVersionRange returns a bool validator that checks if the controller version
// is within the specified range (inclusive).
func BoolRequireVersionRange(client *base.Client, minVersion, maxVersion string, conditionMessage string) validator.Bool {
	return ControllerVersionValidator{
		client:           client,
		minVersion:       base.AsVersion(minVersion),
		maxVersion:       base.AsVersion(maxVersion),
		conditionMessage: conditionMessage,
	}
}

// BoolRequireExactVersion returns a bool validator that checks if the controller version
// matches the specified version exactly.
func BoolRequireExactVersion(client *base.Client, exactVersion string, conditionMessage string) validator.Bool {
	return ControllerVersionValidator{
		client:           client,
		exactVersion:     base.AsVersion(exactVersion),
		conditionMessage: conditionMessage,
	}
}

// Int64RequireMinVersion returns an int64 validator that checks if the controller version
// is at least the specified version.
func Int64RequireMinVersion(client *base.Client, minVersion string, conditionMessage string) validator.Int64 {
	return ControllerVersionValidator{
		client:           client,
		minVersion:       base.AsVersion(minVersion),
		conditionMessage: conditionMessage,
	}
}

// Int64RequireMaxVersion returns an int64 validator that checks if the controller version
// is at most the specified version.
func Int64RequireMaxVersion(client *base.Client, maxVersion string, conditionMessage string) validator.Int64 {
	return ControllerVersionValidator{
		client:           client,
		maxVersion:       base.AsVersion(maxVersion),
		conditionMessage: conditionMessage,
	}
}

// Int64RequireVersionRange returns an int64 validator that checks if the controller version
// is within the specified range (inclusive).
func Int64RequireVersionRange(client *base.Client, minVersion, maxVersion string, conditionMessage string) validator.Int64 {
	return ControllerVersionValidator{
		client:           client,
		minVersion:       base.AsVersion(minVersion),
		maxVersion:       base.AsVersion(maxVersion),
		conditionMessage: conditionMessage,
	}
}

// Int64RequireExactVersion returns an int64 validator that checks if the controller version
// matches the specified version exactly.
func Int64RequireExactVersion(client *base.Client, exactVersion string, conditionMessage string) validator.Int64 {
	return ControllerVersionValidator{
		client:           client,
		exactVersion:     base.AsVersion(exactVersion),
		conditionMessage: conditionMessage,
	}
}

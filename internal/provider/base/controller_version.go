package base

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/types"
)

func AsVersion(versionString string) *version.Version {
	return version.Must(version.NewVersion(versionString))
}

// TODO remove this legacy.
var (
	ControllerV6                = AsVersion("6.0.0")
	ControllerV7                = AsVersion("7.0.0")
	ControllerV8                = AsVersion("8.0.0")
	ControllerV9                = AsVersion("9.0.0")
	ControllerVersionAPIKeyAuth = AsVersion("9.0.108")
	// https://community.ui.com/releases/UniFi-Network-Application-8-2-93/fce86dc6-897a-4944-9c53-1eec7e37e738
	ControllerVersionDNSRecords = AsVersion("8.2.93")

	// https://community.ui.com/releases/UniFi-Network-Controller-6-1-61/62f1ad38-1ac5-430c-94b0-becbb8f71d7d
	ControllerVersionWPA3 = AsVersion("6.1.61")
)

func (c *Client) SupportsAPIKeyAuthentication() bool {
	return c.Version.GreaterThanOrEqual(ControllerVersionAPIKeyAuth)
}

func (c *Client) SupportsWPA3() bool {
	return c.Version.GreaterThanOrEqual(ControllerVersionWPA3)
}

func (c *Client) SupportsDNSRecords() bool {
	return c.Version.GreaterThanOrEqual(ControllerVersionDNSRecords)
}

func CheckMinimumControllerVersion(versionString string) error {
	v, err := version.NewVersion(versionString)
	if err != nil {
		return err
	}
	if v.LessThan(ControllerV6) {
		return fmt.Errorf("controller version %q or greater is required to use the provider, found %q", ControllerV6, v)
	}
	return nil
}

// TODO remove until here

// ControllerVersionValidator is a validator that checks if the UniFi controller version
// matches the specified constraints.
type ControllerVersionValidator interface {
	RequireMinVersion(minVersion string) diag.Diagnostics
	RequireMaxVersion(maxVersion string) diag.Diagnostics
	RequireVersionBetween(minVersion, maxVersion string) diag.Diagnostics
	RequireMinVersionForPath(minVersion string, attrPath path.Path, config tfsdk.Config) diag.Diagnostics
	RequireMaxVersionForPath(maxVersion string, attrPath path.Path, config tfsdk.Config) diag.Diagnostics
	RequireVersionBetweenForPath(minVersion, maxVersion string, attrPath path.Path, config tfsdk.Config) diag.Diagnostics
}

var _ ControllerVersionValidator = &controllerVersionValidator{}

func NewControllerVersionValidator(client *Client) ControllerVersionValidator {
	return &controllerVersionValidator{client: client}
}

type controllerVersionValidator struct {
	client *Client
}

func (v controllerVersionValidator) RequireMinVersion(minVersion string) diag.Diagnostics {
	return v.requireVersion(minVersionRequirement(minVersion), nil)
}

func (v controllerVersionValidator) RequireMaxVersion(maxVersion string) diag.Diagnostics {
	return v.requireVersion(maxVersionRequirement(maxVersion), nil)
}

func (v controllerVersionValidator) RequireVersionBetween(minVersion, maxVersion string) diag.Diagnostics {
	return v.requireVersion(versionBetweenRequirement(minVersion, maxVersion), nil)
}

func (v controllerVersionValidator) RequireMinVersionForPath(minVersion string, attrPath path.Path, config tfsdk.Config) diag.Diagnostics {
	return v.requireVersionForPath(minVersionRequirement(minVersion), attrPath, config)
}

func (v controllerVersionValidator) RequireMaxVersionForPath(maxVersion string, attrPath path.Path, config tfsdk.Config) diag.Diagnostics {
	return v.requireVersionForPath(maxVersionRequirement(maxVersion), attrPath, config)
}

func (v controllerVersionValidator) RequireVersionBetweenForPath(minVersion, maxVersion string, attrPath path.Path, config tfsdk.Config) diag.Diagnostics {
	return v.requireVersionForPath(versionBetweenRequirement(minVersion, maxVersion), attrPath, config)
}

func minVersionRequirement(minVersion string) versionRequirement {
	return versionRequirement{minVersion: AsVersion(minVersion)}
}

func maxVersionRequirement(maxVersion string) versionRequirement {
	return versionRequirement{maxVersion: AsVersion(maxVersion)}
}

func versionBetweenRequirement(minVersion, maxVersion string) versionRequirement {
	return versionRequirement{minVersion: AsVersion(minVersion), maxVersion: AsVersion(maxVersion)}
}

type versionRequirement struct {
	minVersion *version.Version
	maxVersion *version.Version
}

func (r versionRequirement) isBetweenRequirement() bool {
	return r.minVersion != nil && r.maxVersion != nil
}

func (r versionRequirement) isMinRequirement() bool {
	return r.minVersion != nil && r.maxVersion == nil
}

func (r versionRequirement) isMaxRequirement() bool {
	return r.minVersion == nil && r.maxVersion != nil
}

const controllerVersionErrorMessage = "Controller version does not meet requirements"

func (v controllerVersionValidator) requireVersionForPath(req versionRequirement, attrPath path.Path, config tfsdk.Config) diag.Diagnostics {
	diags := diag.Diagnostics{}
	var val attr.Value
	diags.Append(config.GetAttribute(context.Background(), attrPath, &val)...)
	if diags.HasError() {
		return diags
	}
	if !types.IsDefined(val) {
		return diags
	}
	diags.Append(v.requireVersion(req, &attrPath)...)
	return diags
}

// requireVersion checks if the controller version meets the constraints.
func (v controllerVersionValidator) requireVersion(req versionRequirement, attrPath *path.Path) diag.Diagnostics {
	diags := diag.Diagnostics{}
	if v.client == nil || v.client.Version == nil {
		diags.AddError("Controller version not available", "Provider was not initialized properly. UniFi client or controller version is not available")
		return diags
	}

	controllerVersion := v.client.Version
	errorBuilder := strings.Builder{}
	if attrPath != nil {
		fmt.Fprintf(&errorBuilder, "%s is not supported. ", attrPath.String())
	}
	fmt.Fprintf(&errorBuilder, "Controller version %s", controllerVersion)
	failed := false

	if req.isBetweenRequirement() && (controllerVersion.LessThan(req.minVersion) || controllerVersion.GreaterThan(req.maxVersion)) {
		failed = true
		fmt.Fprintf(&errorBuilder, " is not between required %s and %s", req.minVersion, req.maxVersion)
	} else if req.isMinRequirement() && controllerVersion.LessThan(req.minVersion) {
		failed = true
		fmt.Fprintf(&errorBuilder, " is less than minimum required version %s", req.minVersion)
	} else if req.isMaxRequirement() && controllerVersion.GreaterThan(req.maxVersion) {
		failed = true
		fmt.Fprintf(&errorBuilder, " is greater than maximum required version %s", req.maxVersion)
	}
	if failed {
		diags.AddError(controllerVersionErrorMessage, errorBuilder.String())
	}
	return diags
}

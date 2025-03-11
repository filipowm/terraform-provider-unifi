package settings

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dpiModel struct {
	base.Model
	Enabled               types.Bool `tfsdk:"enabled"`
	FingerprintingEnabled types.Bool `tfsdk:"fingerprinting_enabled"`
}

func (d *dpiModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingDpi{
		ID:                    d.ID.ValueString(),
		Enabled:               d.Enabled.ValueBool(),
		FingerprintingEnabled: d.FingerprintingEnabled.ValueBool(),
	}

	return model, diags
}

func (d *dpiModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingDpi)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingDpi")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.Enabled = types.BoolValue(model.Enabled)
	d.FingerprintingEnabled = types.BoolValue(model.FingerprintingEnabled)

	return diags
}

var (
	_ base.ResourceModel               = &dpiModel{}
	_ resource.Resource                = &dpiResource{}
	_ resource.ResourceWithConfigure   = &dpiResource{}
	_ resource.ResourceWithImportState = &dpiResource{}
)

type dpiResource struct {
	*BaseSettingResource[*dpiModel]
}

func (r *dpiResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Deep Packet Inspection (DPI) settings for a UniFi site. DPI is a feature that allows the UniFi controller to analyze network traffic and identify applications and services being used on the network.",
		Attributes: map[string]schema.Attribute{
			"id":   base.ID(),
			"site": base.SiteAttribute(),
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether Deep Packet Inspection is enabled.",
				Required:            true,
			},
			"fingerprinting_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether DPI fingerprinting is enabled. Fingerprinting allows the controller to identify applications and services based on traffic patterns.",
				Required:            true,
			},
		},
	}
}

func NewDpiResource() resource.Resource {
	r := &dpiResource{}
	r.BaseSettingResource = NewBaseSettingResource(
		"unifi_setting_dpi",
		func() *dpiModel { return &dpiModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingDpi(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			return client.UpdateSettingDpi(ctx, site, body.(*unifi.SettingDpi))
		},
	)
	return r
}

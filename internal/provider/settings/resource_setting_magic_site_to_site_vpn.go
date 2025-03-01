package settings

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type magicSiteToSiteVpnModel struct {
	base.Model
	Enabled types.Bool `tfsdk:"enabled"`
}

func (d *magicSiteToSiteVpnModel) AsUnifiModel() (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingMagicSiteToSiteVpn{
		ID:      d.ID.ValueString(),
		Enabled: d.Enabled.ValueBool(),
	}

	return model, diags
}

func (d *magicSiteToSiteVpnModel) Merge(other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingMagicSiteToSiteVpn)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingMagicSiteToSiteVpn")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.Enabled = types.BoolValue(model.Enabled)

	return diags
}

var (
	_ base.ResourceModel               = &magicSiteToSiteVpnModel{}
	_ resource.Resource                = &magicSiteToSiteVpnResource{}
	_ resource.ResourceWithConfigure   = &magicSiteToSiteVpnResource{}
	_ resource.ResourceWithImportState = &magicSiteToSiteVpnResource{}
)

type magicSiteToSiteVpnResource struct {
	*BaseSettingResource[*magicSiteToSiteVpnModel]
}

func (r *magicSiteToSiteVpnResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Magic Site to Site VPN settings for a UniFi site.",
		Attributes: map[string]schema.Attribute{
			"id":   base.ID(),
			"site": base.SiteAttribute(),
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the Magic Site to Site VPN is enabled.",
				Required:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func NewMagicSiteToSiteVpnResource() resource.Resource {
	r := &magicSiteToSiteVpnResource{}
	r.BaseSettingResource = NewBaseSettingResource(
		"unifi_setting_magic_site_to_site_vpn",
		func() *magicSiteToSiteVpnModel { return &magicSiteToSiteVpnModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingMagicSiteToSiteVpn(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			return client.UpdateSettingMagicSiteToSiteVpn(ctx, site, body.(*unifi.SettingMagicSiteToSiteVpn))
		},
	)
	return r
}

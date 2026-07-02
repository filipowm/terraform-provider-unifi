package settings

import (
	"context"

	ut "github.com/filipowm/terraform-provider-unifi/internal/provider/types"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
)

type connectivityModel struct {
	base.Model
	Enabled types.Bool `tfsdk:"enabled"`
}

func (d *connectivityModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingConnectivity{
		ID:      d.ID.ValueString(),
		Enabled: d.Enabled.ValueBool(),
	}

	return model, diags
}

func (d *connectivityModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingConnectivity)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingConnectivity")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.Enabled = types.BoolValue(model.Enabled)

	return diags
}

var (
	_ base.ResourceModel               = &connectivityModel{}
	_ resource.Resource                = &connectivityResource{}
	_ resource.ResourceWithConfigure   = &connectivityResource{}
	_ resource.ResourceWithImportState = &connectivityResource{}
)

type connectivityResource struct {
	*base.GenericResource[*connectivityModel]
}

func (r *connectivityResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the site Connectivity settings for a UniFi site. The `enabled` flag controls **wireless meshing** " +
			"(shown as \"Wireless Meshing\" in the controller UI): when on, access points can uplink to the network over a hidden " +
			"wireless backhaul instead of Ethernet, and each AP keeps a standby mesh radio running for failover. Disable it on a " +
			"fully wired site to reclaim that radio/airtime; APs that lose their wired uplink then go offline rather than re-joining " +
			"over mesh. The controller-generated mesh SSID and PSK are intentionally not managed by this resource.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether wireless meshing is enabled for the site.",
				Required:            true,
			},
		},
	}
}

func NewConnectivityResource() resource.Resource {
	r := &connectivityResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_connectivity",
		func() *connectivityModel { return &connectivityModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingConnectivity(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingConnectivity)
			return client.UpdateSettingConnectivity(ctx, site, b)
		},
	)
	return r
}

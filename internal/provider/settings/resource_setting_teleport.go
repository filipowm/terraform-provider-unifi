package settings

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type teleportModel struct {
	base.Model
	Enabled types.Bool   `tfsdk:"enabled"`
	Subnet  types.String `tfsdk:"subnet"`
}

func (d *teleportModel) AsUnifiModel() (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingTeleport{
		ID:         d.ID.ValueString(),
		Enabled:    d.Enabled.ValueBool(),
		SubnetCidr: d.Subnet.ValueString(),
	}

	return model, diags
}

func (d *teleportModel) Merge(other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingTeleport)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingTeleport")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.Enabled = types.BoolValue(model.Enabled)
	d.Subnet = types.StringValue(model.SubnetCidr)

	return diags
}

var (
	_ base.ResourceModel               = &teleportModel{}
	_ resource.Resource                = &teleportResource{}
	_ resource.ResourceWithConfigure   = &teleportResource{}
	_ resource.ResourceWithImportState = &teleportResource{}
)

type teleportResource struct {
	*BaseSettingResource[*teleportModel]
}

func (r *teleportResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Teleport settings for a UniFi site. Teleport is a secure remote access technology that allows authorized users to connect to UniFi devices from anywhere.",
		Attributes: map[string]schema.Attribute{
			"id":   base.ID(),
			"site": base.SiteAttribute(),
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether Teleport is enabled.",
				Required:            true,
			},
			"subnet": schema.StringAttribute{
				MarkdownDescription: "The subnet CIDR for Teleport (e.g., `192.168.1.0/24`). Can be empty but must be set explicitly.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validators.CIDROrEmpty(),
				},
			},
		},
	}
}

func NewTeleportResource() resource.Resource {
	r := &teleportResource{}
	r.BaseSettingResource = NewBaseSettingResource(
		"unifi_setting_teleport",
		func() *teleportModel { return &teleportModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingTeleport(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			return client.UpdateSettingTeleport(ctx, site, body.(*unifi.SettingTeleport))
		},
	)
	return r
}

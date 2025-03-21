package settings

import (
	"context"
	ut "github.com/filipowm/terraform-provider-unifi/internal/provider/types"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type uswModel struct {
	base.Model
	DHCPSnoop types.Bool `tfsdk:"dhcp_snoop"`
}

func (d *uswModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingUsw{
		ID:        d.ID.ValueString(),
		DHCPSnoop: d.DHCPSnoop.ValueBool(),
	}

	return model, diags
}

func (d *uswModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingUsw)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingUsw")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.DHCPSnoop = types.BoolValue(model.DHCPSnoop)

	return diags
}

var (
	_ base.ResourceModel               = &uswModel{}
	_ resource.Resource                = &uswResource{}
	_ resource.ResourceWithConfigure   = &uswResource{}
	_ resource.ResourceWithImportState = &uswResource{}
)

type uswResource struct {
	*base.GenericResource[*uswModel]
}

func (r *uswResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages UniFi Switch (USW) settings for a UniFi site. These settings control global switch behaviors such as DHCP snooping.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"dhcp_snoop": schema.BoolAttribute{
				MarkdownDescription: "Whether DHCP snooping is enabled. DHCP snooping is a security feature that filters untrusted DHCP messages and builds a binding database of valid hosts.",
				Required:            true,
			},
		},
	}
}

func NewUswResource() resource.Resource {
	r := &uswResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_usw",
		func() *uswModel { return &uswModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingUsw(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			return client.UpdateSettingUsw(ctx, site, body.(*unifi.SettingUsw))
		},
	)
	return r
}

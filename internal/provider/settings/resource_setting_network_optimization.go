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

type networkOptimizationModel struct {
	base.Model
	Enabled types.Bool `tfsdk:"enabled"`
}

func (d *networkOptimizationModel) AsUnifiModel() (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingNetworkOptimization{
		ID:      d.ID.ValueString(),
		Enabled: d.Enabled.ValueBool(),
	}

	return model, diags
}

func (d *networkOptimizationModel) Merge(other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingNetworkOptimization)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingNetworkOptimization")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.Enabled = types.BoolValue(model.Enabled)

	return diags
}

var (
	_ base.ResourceModel               = &networkOptimizationModel{}
	_ resource.Resource                = &networkOptimizationResource{}
	_ resource.ResourceWithConfigure   = &networkOptimizationResource{}
	_ resource.ResourceWithImportState = &networkOptimizationResource{}
)

type networkOptimizationResource struct {
	*BaseSettingResource[*networkOptimizationModel]
}

func (r *networkOptimizationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Network Optimization settings for a UniFi site. UniFi network optimization is a feature designed to automatically enhance the performance of a UniFi network" +
			" by making automatic adjustments to various settings such as channel selection, transmit power, or frequency usage",
		Attributes: map[string]schema.Attribute{
			"id":   base.ID(),
			"site": base.SiteAttribute(),
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the Network Optimization is enabled.",
				Required:            true,
			},
		},
	}
}

func NewNetworkOptimizationResource() resource.Resource {
	r := &networkOptimizationResource{}
	r.BaseSettingResource = NewBaseSettingResource(
		"unifi_setting_network_optimization",
		func() *networkOptimizationModel { return &networkOptimizationModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingNetworkOptimization(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			return client.UpdateSettingNetworkOptimization(ctx, site, body.(*unifi.SettingNetworkOptimization))
		},
	)
	return r
}

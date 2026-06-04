package settings

import (
	"context"
	"regexp"

	ut "github.com/filipowm/terraform-provider-unifi/internal/provider/types"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var colorHexRegexp = regexp.MustCompile(`^[0-9A-Fa-f]{6}$`)

type etherLightingNetworkOverrideModel struct {
	NetworkID types.String `tfsdk:"network_id"`
	ColorHex  types.String `tfsdk:"color_hex"`
}

func (m *etherLightingNetworkOverrideModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"network_id": types.StringType,
		"color_hex":  types.StringType,
	}
}

type etherLightingSpeedOverrideModel struct {
	Speed    types.String `tfsdk:"speed"`
	ColorHex types.String `tfsdk:"color_hex"`
}

func (m *etherLightingSpeedOverrideModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"speed":     types.StringType,
		"color_hex": types.StringType,
	}
}

type etherLightingModel struct {
	base.Model
	NetworkOverrides types.Set `tfsdk:"network_overrides"`
	SpeedOverrides   types.Set `tfsdk:"speed_overrides"`
}

func (d *etherLightingModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingEtherLighting{
		ID: d.ID.ValueString(),
	}

	if ut.IsDefined(d.NetworkOverrides) {
		var overrides []etherLightingNetworkOverrideModel
		diags.Append(d.NetworkOverrides.ElementsAs(ctx, &overrides, false)...)
		if diags.HasError() {
			return nil, diags
		}
		model.NetworkOverrides = make([]unifi.SettingEtherLightingNetworkOverrides, 0, len(overrides))
		for _, o := range overrides {
			model.NetworkOverrides = append(model.NetworkOverrides, unifi.SettingEtherLightingNetworkOverrides{
				Key:         o.NetworkID.ValueString(),
				RawColorHex: o.ColorHex.ValueString(),
			})
		}
	}

	if ut.IsDefined(d.SpeedOverrides) {
		var overrides []etherLightingSpeedOverrideModel
		diags.Append(d.SpeedOverrides.ElementsAs(ctx, &overrides, false)...)
		if diags.HasError() {
			return nil, diags
		}
		model.SpeedOverrides = make([]unifi.SettingEtherLightingSpeedOverrides, 0, len(overrides))
		for _, o := range overrides {
			model.SpeedOverrides = append(model.SpeedOverrides, unifi.SettingEtherLightingSpeedOverrides{
				Key:         o.Speed.ValueString(),
				RawColorHex: o.ColorHex.ValueString(),
			})
		}
	}

	return model, diags
}

func (d *etherLightingModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingEtherLighting)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingEtherLighting")
		return diags
	}

	d.ID = types.StringValue(model.ID)

	networkModels := make([]etherLightingNetworkOverrideModel, 0, len(model.NetworkOverrides))
	for _, o := range model.NetworkOverrides {
		networkModels = append(networkModels, etherLightingNetworkOverrideModel{
			NetworkID: types.StringValue(o.Key),
			ColorHex:  types.StringValue(o.RawColorHex),
		})
	}
	networkList, networkDiags := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: (&etherLightingNetworkOverrideModel{}).AttributeTypes()}, networkModels)
	diags.Append(networkDiags...)
	if diags.HasError() {
		return diags
	}
	d.NetworkOverrides = networkList

	speedModels := make([]etherLightingSpeedOverrideModel, 0, len(model.SpeedOverrides))
	for _, o := range model.SpeedOverrides {
		speedModels = append(speedModels, etherLightingSpeedOverrideModel{
			Speed:    types.StringValue(o.Key),
			ColorHex: types.StringValue(o.RawColorHex),
		})
	}
	speedList, speedDiags := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: (&etherLightingSpeedOverrideModel{}).AttributeTypes()}, speedModels)
	diags.Append(speedDiags...)
	if diags.HasError() {
		return diags
	}
	d.SpeedOverrides = speedList

	return diags
}

var (
	_ base.ResourceModel               = &etherLightingModel{}
	_ resource.Resource                = &etherLightingResource{}
	_ resource.ResourceWithConfigure   = &etherLightingResource{}
	_ resource.ResourceWithImportState = &etherLightingResource{}
)

type etherLightingResource struct {
	*base.GenericResource[*etherLightingModel]
}

func (r *etherLightingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the site-level Etherlighting palette — the colors that switches with per-port LEDs " +
			"(USW Pro Max line) use when a device's `ether_lighting` block selects a scheme. `network_overrides` assigns a " +
			"color per network/VLAN (used by `mode = \"network\"`); `speed_overrides` assigns a color per link-speed class " +
			"(used by `mode = \"speed\"`). Overrides take precedence over the controller's built-in default palette; " +
			"networks or speeds without an override keep their default color. NOTE: the controller silently drops any override whose color equals that network's built-in default — declare only colors that actually differ from the defaults, or the entry will not round-trip.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"network_overrides": schema.SetNestedAttribute{
				MarkdownDescription: "Per-network LED colors, used when a device's Etherlighting `mode` is `network`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"network_id": schema.StringAttribute{
							MarkdownDescription: "ID of the network/VLAN this color applies to (e.g. `unifi_network.iot.id`).",
							Required:            true,
						},
						"color_hex": schema.StringAttribute{
							MarkdownDescription: "LED color as a 6-digit RGB hex string without `#` (e.g. `ff6c14`).",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(colorHexRegexp, "must be a 6-digit RGB hex string without '#'"),
							},
						},
					},
				},
			},
			"speed_overrides": schema.SetNestedAttribute{
				MarkdownDescription: "Per-link-speed LED colors, used when a device's Etherlighting `mode` is `speed`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"speed": schema.StringAttribute{
							MarkdownDescription: "Link-speed class this color applies to.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("FE", "GbE", "2.5GbE", "5GbE", "10GbE", "25GbE", "40GbE", "100GbE"),
							},
						},
						"color_hex": schema.StringAttribute{
							MarkdownDescription: "LED color as a 6-digit RGB hex string without `#` (e.g. `ffc107`).",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(colorHexRegexp, "must be a 6-digit RGB hex string without '#'"),
							},
						},
					},
				},
			},
		},
	}
}

func NewEtherLightingResource() resource.Resource {
	r := &etherLightingResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_ether_lighting",
		func() *etherLightingModel { return &etherLightingModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingEtherLighting(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			return client.UpdateSettingEtherLighting(ctx, site, body.(*unifi.SettingEtherLighting))
		},
	)
	return r
}

package settings

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
	"github.com/filipowm/terraform-provider-unifi/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// GeoIPFilteringModel represents the GeoIP filtering configuration
type GeoIPFilteringModel struct {
	Enabled          types.Bool   `tfsdk:"enabled"`
	Block            types.String `tfsdk:"block"`
	Countries        types.List   `tfsdk:"countries"`
	TrafficDirection types.String `tfsdk:"traffic_direction"`
}

func (m *GeoIPFilteringModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled": types.BoolType,
		"block":   types.StringType,
		"countries": types.ListType{
			ElemType: types.StringType,
		},
		"traffic_direction": types.StringType,
	}
}

// UpnpModel represents the UPNP configuration
type UpnpModel struct {
	Enabled       types.Bool   `tfsdk:"enabled"`
	NatPmpEnabled types.Bool   `tfsdk:"nat_pmp_enabled"`
	SecureMode    types.Bool   `tfsdk:"secure_mode"`
	WANInterface  types.String `tfsdk:"wan_interface"`
}

func (m *UpnpModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled":         types.BoolType,
		"nat_pmp_enabled": types.BoolType,
		"secure_mode":     types.BoolType,
		"wan_interface":   types.StringType,
	}
}

// usgModel represents the data model for USG (UniFi Security Gateway) settings.
// It defines how USG features like mDNS and DHCP relay are configured for a UniFi site.
type usgModel struct {
	base.Model
	MulticastDnsEnabled types.Bool `tfsdk:"multicast_dns_enabled"`
	DhcpRelayServers    types.List `tfsdk:"dhcp_relay_servers"`

	// Geo IP filtering
	GeoIPFiltering types.Object `tfsdk:"geo_ip_filtering"`

	// UPNP configuration
	Upnp types.Object `tfsdk:"upnp"`
}

func (d *usgModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingUsg{
		ID:          d.ID.ValueString(),
		MdnsEnabled: d.MulticastDnsEnabled.ValueBool(),
	}

	// Extract DHCP relay servers from the list
	var dhcpRelayServers []string
	diags.Append(utils.ListElementsAs(d.DhcpRelayServers, &dhcpRelayServers)...)
	if diags.HasError() {
		return nil, diags
	}

	// Assign DHCP relay servers to the model (up to 5)
	// Map each server by index to appropriate field
	serverFields := []struct {
		index    int
		fieldPtr *string
	}{
		{0, &model.DHCPRelayServer1},
		{1, &model.DHCPRelayServer2},
		{2, &model.DHCPRelayServer3},
		{3, &model.DHCPRelayServer4},
		{4, &model.DHCPRelayServer5},
	}

	for _, sf := range serverFields {
		if sf.index < len(dhcpRelayServers) {
			*sf.fieldPtr = dhcpRelayServers[sf.index]
		}
	}

	// Assign Geo IP filtering attributes
	if base.IsDefined(d.GeoIPFiltering) {
		var geoIPFiltering *GeoIPFilteringModel
		diags.Append(d.GeoIPFiltering.As(ctx, &geoIPFiltering, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}

		model.GeoIPFilteringEnabled = geoIPFiltering.Enabled.ValueBool()
		model.GeoIPFilteringBlock = geoIPFiltering.Block.ValueString()
		model.GeoIPFilteringTrafficDirection = geoIPFiltering.TrafficDirection.ValueString()
		countries, diags := utils.ListElementsToString(ctx, geoIPFiltering.Countries)
		if diags.HasError() {
			return nil, diags
		}
		model.GeoIPFilteringCountries = countries
	} else {
		model.GeoIPFilteringEnabled = false
	}

	// Assign UPNP attributes
	if base.IsDefined(d.Upnp) {
		var upnp *UpnpModel
		diags.Append(d.Upnp.As(ctx, &upnp, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}

		model.UpnpEnabled = upnp.Enabled.ValueBool()
		model.UpnpNATPmpEnabled = upnp.NatPmpEnabled.ValueBool()
		model.UpnpSecureMode = upnp.SecureMode.ValueBool()
		model.UpnpWANInterface = upnp.WANInterface.ValueString()
	} else {
		model.UpnpEnabled = false
	}

	return model, diags
}

func (d *usgModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingUsg)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingUsg")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.MulticastDnsEnabled = types.BoolValue(model.MdnsEnabled)

	// Extract non-empty DHCP relay servers
	dhcpRelay := []string{}
	for _, s := range []string{
		model.DHCPRelayServer1,
		model.DHCPRelayServer2,
		model.DHCPRelayServer3,
		model.DHCPRelayServer4,
		model.DHCPRelayServer5,
	} {
		if s == "" {
			continue
		}
		dhcpRelay = append(dhcpRelay, s)
	}

	// Set the DHCP relay servers list
	dhcpRelayServers, diags := types.ListValueFrom(ctx, types.StringType, dhcpRelay)
	if diags.HasError() {
		return diags
	}
	d.DhcpRelayServers = dhcpRelayServers

	// Set Geo IP filtering attributes
	geoIPFiltering := &GeoIPFilteringModel{
		Enabled:          types.BoolValue(model.GeoIPFilteringEnabled),
		Block:            types.StringValue(model.GeoIPFilteringBlock),
		TrafficDirection: types.StringValue(model.GeoIPFilteringTrafficDirection),
	}

	countries, diags := utils.StringToListElements(ctx, model.GeoIPFilteringCountries)
	if diags.HasError() {
		return diags
	}
	geoIPFiltering.Countries = countries

	// Create object value from attributes
	geoIPObject, diags := types.ObjectValueFrom(ctx, geoIPFiltering.AttributeTypes(), geoIPFiltering)
	if diags.HasError() {
		return diags
	}
	d.GeoIPFiltering = geoIPObject

	// Set UPNP attributes
	upnp := &UpnpModel{
		Enabled:       types.BoolValue(model.UpnpEnabled),
		NatPmpEnabled: types.BoolValue(model.UpnpNATPmpEnabled),
		SecureMode:    types.BoolValue(model.UpnpSecureMode),
		WANInterface:  types.StringValue(model.UpnpWANInterface),
	}

	// Create object value from attributes
	upnpObject, diags := types.ObjectValueFrom(ctx, upnp.AttributeTypes(), upnp)
	if diags.HasError() {
		return diags
	}
	d.Upnp = upnpObject
	return diags
}

func (r *usgResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_usg` resource manages advanced settings for UniFi Security Gateways (USG) and UniFi Dream Machines (UDM/UDM-Pro).\n\n" +
				"This resource allows you to configure gateway-specific features including:\n" +
				"  * Multicast DNS (mDNS) for service discovery\n" +
				"  * DHCP relay for forwarding DHCP requests to external servers\n" +
				"  * Geo IP filtering for country-based traffic control\n" +
				"  * UPNP for automatic port forwarding\n\n" +
				"Note: Some settings may not be available on all controller versions. For example, multicast_dns_enabled is not supported on UniFi OS v7+.",
		Attributes: map[string]schema.Attribute{
			"id":   base.ID(),
			"site": base.SiteAttribute(),
			"multicast_dns_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable multicast DNS (mDNS/Bonjour/Avahi) forwarding across VLANs. This allows devices to discover services " +
						"(like printers, Chromecasts, etc.) even when they are on different networks. Note: Not supported on UniFi OS v7+.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"dhcp_relay_servers": schema.ListAttribute{
				MarkdownDescription: "List of up to 5 DHCP relay servers (specified by IP address) that will receive forwarded DHCP requests. " +
						"This is useful when you want to use external DHCP servers instead of the built-in DHCP server. " +
						"Example: ['192.168.1.5', '192.168.2.5']",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Default: listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				Validators: []validator.List{
					listvalidator.SizeAtMost(5),
					listvalidator.ValueStringsAre(validators.IPv4()),
				},
			},
			"geo_ip_filtering": schema.SingleNestedAttribute{
				MarkdownDescription: "Geographic IP filtering configuration. Allows blocking or allowing traffic based on country.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Object{
					validators.RequiredTogetherIf(path.MatchRoot("enabled"), types.BoolValue(true), path.MatchRoot("countries")),
				},
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable geographic IP filtering. When enabled, traffic from specified countries will be blocked or allowed " +
								"according to the configured rules.",
						Required: true,
					},
					"block": schema.StringAttribute{
						MarkdownDescription: "Specifies whether the selected countries should be blocked or allowed. Valid values are `block` (default) or `allow`. " +
								"When set to `block`, traffic from the specified countries will be blocked. When set to `allow`, only traffic from the " +
								"specified countries will be allowed.",
						Optional: true,
						Computed: true,
						Default:  stringdefault.StaticString("block"),
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("block", "allow"),
						},
					},
					"countries": schema.ListAttribute{
						MarkdownDescription: "List of two-letter following ISO 3166-1 alpha-2 country codes to block or allow. " +
								"Example: `['US', 'CA', 'MX']` for United States, Canada, and Mexico.",
						Optional:    true,
						ElementType: types.StringType,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
							listvalidator.ValueStringsAre(validators.CountryCodeAlpha2()),
						},
					},
					"traffic_direction": schema.StringAttribute{
						MarkdownDescription: "Specifies which traffic direction the geo IP filtering applies to. Valid values are `both`, `ingress`, or `egress`. " +
								"`both` (default) filters traffic in both directions, `ingress` filters only incoming traffic, and `egress` filters only outgoing traffic.",
						Optional: true,
						Computed: true,
						Default:  stringdefault.StaticString("both"),
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("both", "ingress", "egress"),
						},
					},
				},
			},
			"upnp": schema.SingleNestedAttribute{
				MarkdownDescription: "UPNP (Universal Plug and Play) configuration. Enables automatic port forwarding for applications that support it.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable UPNP functionality. When enabled, applications can automatically " +
								"request port forwarding rules from the gateway.",
						Required: true,
					},
					"nat_pmp_enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable NAT-PMP (NAT Port Mapping Protocol) support alongside UPNP. NAT-PMP is " +
								"Apple's alternative to UPNP, providing similar automatic port mapping capabilities.",
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(false),
					},
					"secure_mode": schema.BoolAttribute{
						MarkdownDescription: "Enable secure mode for UPNP. In secure mode, the gateway only forwards ports " +
								"to the device that specifically requested them, enhancing security.",
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(false),
					},
					"wan_interface": schema.StringAttribute{
						MarkdownDescription: "Specify which WAN interface to use for UPNP service. Valid values are " +
								"`WAN` (primary interface) or `WAN2` (secondary interface, if available).",
						Optional: true,
						Computed: true,
						Default:  stringdefault.StaticString("WAN"),
						Validators: []validator.String{
							stringvalidator.OneOf("WAN", "WAN2"),
						},
					},
				},
			},
		},
	}
}

// NewUsgResource creates a new instance of the USG resource.
func NewUsgResource() resource.Resource {
	r := &usgResource{}
	r.BaseSettingResource = NewBaseSettingResource(
		"unifi_setting_usg",
		func() *usgModel { return &usgModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingUsg(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			return client.UpdateSettingUsg(ctx, site, body.(*unifi.SettingUsg))
		},
	)
	return r
}

var (
	_ base.ResourceModel                    = &usgModel{}
	_ resource.Resource                     = &usgResource{}
	_ resource.ResourceWithConfigure        = &usgResource{}
	_ resource.ResourceWithImportState      = &usgResource{}
	_ resource.ResourceWithModifyPlan       = &usgResource{}
	_ resource.ResourceWithConfigValidators = &usgResource{}
)

type usgResource struct {
	*BaseSettingResource[*usgModel]
}

func (r *usgResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{}
}

func (r *usgResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	resp.Diagnostics.Append(r.RequireMaxVersionForPath("7.0", path.Root("multicast_dns_enabled"), req.Config)...)
}

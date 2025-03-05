package settings

import (
	"context"
	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
	"github.com/filipowm/terraform-provider-unifi/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// usgModel represents the data model for USG (UniFi Security Gateway) settings.
// It defines how USG features like mDNS and DHCP relay are configured for a UniFi site.
type usgModel struct {
	base.Model
	MulticastDnsEnabled types.Bool `tfsdk:"multicast_dns_enabled"`
	DhcpRelayServers    types.List `tfsdk:"dhcp_relay_servers"`
}

func (d *usgModel) AsUnifiModel() (interface{}, diag.Diagnostics) {
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
	model.DHCPRelayServer1 = append(dhcpRelayServers, "")[0]
	model.DHCPRelayServer2 = append(dhcpRelayServers, "", "")[1]
	model.DHCPRelayServer3 = append(dhcpRelayServers, "", "", "")[2]
	model.DHCPRelayServer4 = append(dhcpRelayServers, "", "", "", "")[3]
	model.DHCPRelayServer5 = append(dhcpRelayServers, "", "", "", "", "")[4]

	return model, diags
}

func (d *usgModel) Merge(other interface{}) diag.Diagnostics {
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
	dhcpRelayServers, diags := types.ListValueFrom(context.Background(), types.StringType, dhcpRelay)
	if diags.HasError() {
		return diags
	}
	d.DhcpRelayServers = dhcpRelayServers

	return diags
}

var (
	_ base.ResourceModel               = &usgModel{}
	_ resource.Resource                = &usgResource{}
	_ resource.ResourceWithConfigure   = &usgResource{}
	_ resource.ResourceWithImportState = &usgResource{}
	_ resource.ResourceWithModifyPlan  = &usgResource{}
)

type usgResource struct {
	*BaseSettingResource[*usgModel]
}

func (r *usgResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	resp.Diagnostics.Append(r.RequireMaxVersionForPath("7.0", path.Root("multicast_dns_enabled"), req.Config)...)
}

func (r *usgResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_usg` resource manages advanced settings for UniFi Security Gateways (USG) and UniFi Dream Machines (UDM/UDM-Pro).\n\n" +
			"This resource allows you to configure gateway-specific features including:\n" +
			"  * Multicast DNS (mDNS) for service discovery\n" +
			"  * DHCP relay for forwarding DHCP requests to external servers\n\n" +
			"These settings are particularly useful for:\n" +
			"  * Enabling device discovery across VLANs (using mDNS)\n" +
			"  * Centralizing DHCP management in enterprise environments\n" +
			"  * Integration with existing network infrastructure\n\n" +
			"Note: Some settings may not be available on all controller versions. For example, multicast_dns_enabled is not supported on UniFi OS v7+.",
		Attributes: map[string]schema.Attribute{
			"id":   base.ID(),
			"site": base.SiteAttribute(),
			"multicast_dns_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable multicast DNS (mDNS/Bonjour/Avahi) forwarding across VLANs. This allows devices to discover services " +
					"(like printers, Chromecasts, etc.) even when they are on different networks. Note: Not supported on UniFi OS v7+.",
				Optional: true,
				Computed: true,
			},
			"dhcp_relay_servers": schema.ListAttribute{
				MarkdownDescription: "List of up to 5 DHCP relay servers (specified by IP address) that will receive forwarded DHCP requests. " +
					"This is useful when you want to use external DHCP servers instead of the built-in DHCP server. " +
					"Example: ['192.168.1.5', '192.168.2.5']",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				Validators: []validator.List{
					listvalidator.SizeAtMost(5),
					listvalidator.ValueStringsAre(validators.IPv4()),
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

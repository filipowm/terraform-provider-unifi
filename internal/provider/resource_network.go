package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	wanUsernameRegexp   = regexp.MustCompile("[^\"' ]+|^$")
	validateWANUsername = validation.StringMatch(wanUsernameRegexp, "invalid WAN username")

	wanTypeRegexp   = regexp.MustCompile("disabled|dhcp|static|pppoe")
	validateWANType = validation.StringMatch(wanTypeRegexp, "invalid WAN connection type")

	wanTypeV6Regexp   = regexp.MustCompile("disabled|dhcpv6|static")
	validateWANTypeV6 = validation.StringMatch(wanTypeV6Regexp, "invalid WANv6 connection type")

	wanPasswordRegexp   = regexp.MustCompile("[^\"' ]+")
	validateWANPassword = validation.StringMatch(wanPasswordRegexp, "invalid WAN password")

	wanNetworkGroupRegexp   = regexp.MustCompile("WAN[2]?|WAN_LTE_FAILOVER")
	validateWANNetworkGroup = validation.StringMatch(wanNetworkGroupRegexp, "invalid WAN network group")

	wanV6NetworkGroupRegexp   = regexp.MustCompile("wan[2]?")
	validateWANV6NetworkGroup = validation.StringMatch(wanV6NetworkGroupRegexp, "invalid WANv6 network group")

	ipV6InterfaceTypeRegexp   = regexp.MustCompile("none|pd|static")
	validateIpV6InterfaceType = validation.StringMatch(ipV6InterfaceTypeRegexp, "invalid IPv6 interface type")

	// This is a slightly larger range than the UI, it includes some reserved ones, so could be tightened up.
	validateVLANID = validation.IntBetween(0, 4096)

	ipV6RAPriorityRegexp   = regexp.MustCompile("high|medium|low")
	validateIpV6RAPriority = validation.StringMatch(ipV6RAPriorityRegexp, "invalid IPv6 RA priority")
)

var _ resource.Resource = &NetworkResource{}
var _ resource.ResourceWithImportState = &NetworkResource{}

func resourceNetwork() resource.Resource {
	return &NetworkResource{}
}

type NetworkResource struct {
	client *client
}

type NetworkResourceModel struct {
	ID                      types.String   `tfsdk:"id"`
	Site                    types.String   `tfsdk:"site"`
	Name                    types.String   `tfsdk:"name"`
	Purpose                 types.String   `tfsdk:"purpose"`
	VLANID                  types.Int64    `tfsdk:"vlan_id"`
	Subnet                  types.String   `tfsdk:"subnet"`
	NetworkGroup            types.String   `tfsdk:"network_group"`
	DHCPStart               types.String   `tfsdk:"dhcp_start"`
	DHCPStop                types.String   `tfsdk:"dhcp_stop"`
	DHCPEnabled             types.Bool     `tfsdk:"dhcp_enabled"`
	DHCPLease               types.Int64    `tfsdk:"dhcp_lease"`
	DHCPDNS                 []types.String `tfsdk:"dhcp_dns"`
	DHCPDBootEnabled        types.Bool     `tfsdk:"dhcpd_boot_enabled"`
	DHCPDBootServer         types.String   `tfsdk:"dhcpd_boot_server"`
	DHCPDBootFilename       types.String   `tfsdk:"dhcpd_boot_filename"`
	DHCPRelayEnabled        types.Bool     `tfsdk:"dhcp_relay_enabled"`
	DomainName              types.String   `tfsdk:"domain_name"`
	IGMPSnooping            types.Bool     `tfsdk:"igmp_snooping"`
	MulticastDNS            types.Bool     `tfsdk:"multicast_dns"`
	Enabled                 types.Bool     `tfsdk:"enabled"`
	InternetAccessEnabled   types.Bool     `tfsdk:"internet_access_enabled"`
	NetworkIsolationEnabled types.Bool     `tfsdk:"network_isolation_enabled"`
	// IPv6 settings
	IPV6InterfaceType       types.String `tfsdk:"ipv6_interface_type"`
	IPV6StaticSubnet        types.String `tfsdk:"ipv6_static_subnet"`
	IPV6PDInterface         types.String `tfsdk:"ipv6_pd_interface"`
	IPV6PDPrefixid          types.String `tfsdk:"ipv6_pd_prefixid"`
	IPV6PDStart             types.String `tfsdk:"ipv6_pd_start"`
	IPV6PDStop              types.String `tfsdk:"ipv6_pd_stop"`
	IPV6RaEnabled           types.Bool   `tfsdk:"ipv6_ra_enable"`
	IPV6RaPreferredLifetime types.Int64  `tfsdk:"ipv6_ra_preferred_lifetime"`
	IPV6RaPriority          types.String `tfsdk:"ipv6_ra_priority"`
	IPV6RaValidLifetime     types.Int64  `tfsdk:"ipv6_ra_valid_lifetime"`
	// DHCPv6 settings
	DHCPV6DNS     []types.String `tfsdk:"dhcp_v6_dns"`
	DHCPV6DNSAuto types.Bool     `tfsdk:"dhcp_v6_dns_auto"`
	DHCPV6Enabled types.Bool     `tfsdk:"dhcp_v6_enabled"`
	DHCPV6Lease   types.Int64    `tfsdk:"dhcp_v6_lease"`
	DHCPV6Start   types.String   `tfsdk:"dhcp_v6_start"`
	DHCPV6Stop    types.String   `tfsdk:"dhcp_v6_stop"`
	// WAN settings
	WANIP           types.String   `tfsdk:"wan_ip"`
	WANType         types.String   `tfsdk:"wan_type"`
	WANNetmask      types.String   `tfsdk:"wan_netmask"`
	WANGateway      types.String   `tfsdk:"wan_gateway"`
	WANNetworkGroup types.String   `tfsdk:"wan_networkgroup"`
	WANEgressQOS    types.Int64    `tfsdk:"wan_egress_qos"`
	WANUsername     types.String   `tfsdk:"wan_username"`
	XWANPassword    types.String   `tfsdk:"x_wan_password"`
	WANDNS          []types.String `tfsdk:"wan_dns"`
	WANTypeV6       types.String   `tfsdk:"wan_type_v6"`
	WANDHCPv6PDSize types.Int64    `tfsdk:"wan_dhcp_v6_pd_size"`
	WANIPV6         types.String   `tfsdk:"wan_ipv6"`
	WANGatewayV6    types.String   `tfsdk:"wan_gateway_v6"`
	WANPrefixlen    types.Int64    `tfsdk:"wan_prefixlen"`
}

func (r *NetworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (r *NetworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "`unifi_network` manages WAN/LAN/VLAN networks.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the network.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site": schema.StringAttribute{
				Description: "The name of the site to associate the network with.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the network.",
				Required:    true,
			},
			"purpose": schema.StringAttribute{
				Description: "The purpose of the network. Must be one of `corporate`, `guest`, `wan`, or `vlan-only`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("corporate", "guest", "wan", "vlan-only"),
				},
			},
			"vlan_id": schema.Int64Attribute{
				Description: "The VLAN ID of the network.",
				Optional:    true,
			},
			"subnet": schema.StringAttribute{
				Description: "The subnet of the network. Must be a valid CIDR address.",
				Optional:    true,
			},
			"network_group": schema.StringAttribute{
				Description: "The group of the network.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("LAN"),
			},
			"dhcp_start": schema.StringAttribute{
				Description: "The IPv4 address where the DHCP range of addresses starts.",
				Optional:    true,
			},
			"dhcp_stop": schema.StringAttribute{
				Description: "The IPv4 address where the DHCP range of addresses stops.",
				Optional:    true,
			},
			"dhcp_enabled": schema.BoolAttribute{
				Description: "Specifies whether DHCP is enabled or not on this network.",
				Optional:    true,
			},
			"dhcp_lease": schema.Int64Attribute{
				Description: "Specifies the lease time for DHCP addresses in seconds.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(86400),
			},
			"dhcp_dns": schema.ListAttribute{
				Description: "Specifies the IPv4 addresses for the DNS server to be returned from the DHCP server.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"dhcpd_boot_enabled": schema.BoolAttribute{
				Description: "Toggles on the DHCP boot options.",
				Optional:    true,
			},
			"dhcpd_boot_server": schema.StringAttribute{
				Description: "Specifies the IPv4 address of a TFTP server to network boot from.",
				Optional:    true,
			},
			"dhcpd_boot_filename": schema.StringAttribute{
				Description: "Specifies the file to PXE boot from on the dhcpd_boot_server.",
				Optional:    true,
			},
			"dhcp_relay_enabled": schema.BoolAttribute{
				Description: "Specifies whether DHCP relay is enabled or not on this network.",
				Optional:    true,
			},
			"domain_name": schema.StringAttribute{
				Description: "The domain name of this network.",
				Optional:    true,
			},
			"igmp_snooping": schema.BoolAttribute{
				Description: "Specifies whether IGMP snooping is enabled or not.",
				Optional:    true,
			},
			"multicast_dns": schema.BoolAttribute{
				Description: "Specifies whether Multicast DNS (mDNS) is enabled or not on the network.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Specifies whether this network is enabled or not.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"internet_access_enabled": schema.BoolAttribute{
				Description: "Specifies whether this network should be allowed to access the internet or not.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"network_isolation_enabled": schema.BoolAttribute{
				Description: "Specifies whether this network should be isolated from other networks or not.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			// IPv6 settings
			"ipv6_interface_type": schema.StringAttribute{
				Description: "Specifies which type of IPv6 connection to use.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("none"),
				Validators: []validator.String{
					stringvalidator.RegexMatches(ipV6InterfaceTypeRegexp, "must be one of none, pd, or static"),
				},
			},
			"ipv6_static_subnet": schema.StringAttribute{
				Description: "Specifies the static IPv6 subnet when `ipv6_interface_type` is 'static'.",
				Optional:    true,
			},
			"ipv6_pd_interface": schema.StringAttribute{
				Description: "Specifies which WAN interface to use for IPv6 PD.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(wanV6NetworkGroupRegexp, "must be one of wan or wan2"),
				},
			},
			"ipv6_pd_prefixid": schema.StringAttribute{
				Description: "Specifies the IPv6 Prefix ID.",
				Optional:    true,
			},
			"ipv6_pd_start": schema.StringAttribute{
				Description: "Start address of the DHCPv6 range. Used if `ipv6_interface_type` is set to `pd`.",
				Optional:    true,
			},
			"ipv6_pd_stop": schema.StringAttribute{
				Description: "End address of the DHCPv6 range. Used if `ipv6_interface_type` is set to `pd`.",
				Optional:    true,
			},
			"ipv6_ra_enable": schema.BoolAttribute{
				Description: "Specifies whether to enable router advertisements or not.",
				Optional:    true,
			},
			"ipv6_ra_preferred_lifetime": schema.Int64Attribute{
				Description: "Lifetime in which the address can be used.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(14400),
			},
			"ipv6_ra_priority": schema.StringAttribute{
				Description: "IPv6 router advertisement priority.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(ipV6RaPriorityRegexp, "must be one of high, medium, or low"),
				},
			},
			"ipv6_ra_valid_lifetime": schema.Int64Attribute{
				Description: "Total lifetime in which the address can be used.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(86400),
			},
			// DHCPv6 settings
			"dhcp_v6_dns": schema.ListAttribute{
				Description: "Specifies the IPv6 addresses for the DNS server.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"dhcp_v6_dns_auto": schema.BoolAttribute{
				Description: "Specifies DNS source to propagate.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"dhcp_v6_enabled": schema.BoolAttribute{
				Description: "Enable stateful DHCPv6 for static configuration.",
				Optional:    true,
			},
			"dhcp_v6_lease": schema.Int64Attribute{
				Description: "Specifies the lease time for DHCPv6 addresses in seconds.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(86400),
			},
			"dhcp_v6_start": schema.StringAttribute{
				Description: "Start address of the DHCPv6 range.",
				Optional:    true,
			},
			"dhcp_v6_stop": schema.StringAttribute{
				Description: "End address of the DHCPv6 range.",
				Optional:    true,
			},
			// WAN settings
			"wan_ip": schema.StringAttribute{
				Description: "The IPv4 address of the WAN.",
				Optional:    true,
			},
			"wan_type": schema.StringAttribute{
				Description: "Specifies the IPV4 WAN connection type.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(wanTypeRegexp, "must be one of disabled, dhcp, static, or pppoe"),
				},
			},
			"wan_netmask": schema.StringAttribute{
				Description: "The IPv4 netmask of the WAN.",
				Optional:    true,
			},
			"wan_gateway": schema.StringAttribute{
				Description: "The IPv4 gateway of the WAN.",
				Optional:    true,
			},
			"wan_networkgroup": schema.StringAttribute{
				Description: "Specifies the WAN network group.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(wanNetworkGroupRegexp, "must be one of WAN, WAN2, or WAN_LTE_FAILOVER"),
				},
			},
			"wan_egress_qos": schema.Int64Attribute{
				Description: "Specifies the WAN egress quality of service.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"wan_username": schema.StringAttribute{
				Description: "Specifies the IPV4 WAN username.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(wanUsernameRegexp, "invalid WAN username"),
				},
			},
			"x_wan_password": schema.StringAttribute{
				Description: "Specifies the IPV4 WAN password.",
				Optional:    true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(wanPasswordRegexp, "invalid WAN password"),
				},
			},
			"wan_dns": schema.ListAttribute{
				Description: "DNS servers IPs of the WAN.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"wan_type_v6": schema.StringAttribute{
				Description: "Specifies the IPV6 WAN connection type.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(wanTypeV6Regexp, "must be one of disabled, static, or dhcpv6"),
				},
			},
			"wan_dhcp_v6_pd_size": schema.Int64Attribute{
				Description: "Specifies the IPv6 prefix size to request from ISP.",
				Optional:    true,
			},
			"wan_ipv6": schema.StringAttribute{
				Description: "The IPv6 address of the WAN.",
				Optional:    true,
			},
			"wan_gateway_v6": schema.StringAttribute{
				Description: "The IPv6 gateway of the WAN.",
				Optional:    true,
			},
			"wan_prefixlen": schema.Int64Attribute{
				Description: "The IPv6 prefix length of the WAN.",
				Optional:    true,
			},
		},
	}
}

func (r *NetworkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *NetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NetworkResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	network := &unifi.Network{
		Name:              plan.Name.ValueString(),
		Purpose:           plan.Purpose.ValueString(),
		VLAN:              int(plan.VLANID.ValueInt64()),
		IPSubnet:          plan.Subnet.ValueString(),
		NetworkGroup:      plan.NetworkGroup.ValueString(),
		DHCPDStart:        plan.DHCPStart.ValueString(),
		DHCPDStop:         plan.DHCPStop.ValueString(),
		DHCPDEnabled:      plan.DHCPEnabled.ValueBool(),
		DHCPDLeaseTime:    int(plan.DHCPLease.ValueInt64()),
		DHCPDBootEnabled:  plan.DHCPDBootEnabled.ValueBool(),
		DHCPDBootServer:   plan.DHCPDBootServer.ValueString(),
		DHCPDBootFilename: plan.DHCPDBootFilename.ValueString(),
		DHCPRelayEnabled:  plan.DHCPRelayEnabled.ValueBool(),
		DomainName:        plan.DomainName.ValueString(),
		IGMPSnooping:      plan.IGMPSnooping.ValueBool(),
		MdnsEnabled:       plan.MulticastDNS.ValueBool(),
		Enabled:           plan.Enabled.ValueBool(),
		// Add remaining field mappings...
	}

	site := plan.Site.ValueString()
	if site == "" {
		site = r.client.site
	}

	created, err := r.client.c.CreateNetwork(ctx, site, network)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating network",
			fmt.Sprintf("Could not create network: %s", err),
		)
		return
	}

	plan.ID = types.StringValue(created.ID)
	// Add remaining field updates...

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *NetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NetworkResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	site := state.Site.ValueString()
	if site == "" {
		site = r.client.site
	}

	network, err := r.client.c.GetNetwork(ctx, site, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading network",
			fmt.Sprintf("Could not read network: %s", err),
		)
		return
	}

	// Update state with the current network data
	state.Name = types.StringValue(network.Name)
	state.Purpose = types.StringValue(network.Purpose)
	state.VLANID = types.Int64Value(int64(network.VLAN))
	// Add remaining field updates...

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *NetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NetworkResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	network := &unifi.Network{
		ID:      plan.ID.ValueString(),
		Name:    plan.Name.ValueString(),
		Purpose: plan.Purpose.ValueString(),
		// Add remaining field mappings...
	}

	site := plan.Site.ValueString()
	if site == "" {
		site = r.client.site
	}

	updated, err := r.client.c.UpdateNetwork(ctx, site, network)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating network",
			fmt.Sprintf("Could not update network: %s", err),
		)
		return
	}

	// Update state with the updated network data
	plan.ID = types.StringValue(updated.ID)
	// Add remaining field updates...

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *NetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NetworkResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	site := state.Site.ValueString()
	if site == "" {
		site = r.client.site
	}

	err := r.client.c.DeleteNetwork(ctx, site, state.ID.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting network",
			fmt.Sprintf("Could not delete network: %s", err),
		)
		return
	}
}

func (r *NetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

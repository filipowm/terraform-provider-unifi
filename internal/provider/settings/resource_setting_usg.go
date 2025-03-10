package settings

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
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

// TCPTimeoutModel represents the TCP timeout configuration
type TCPTimeoutModel struct {
	CloseTimeout       types.Int64 `tfsdk:"close_timeout"`
	CloseWaitTimeout   types.Int64 `tfsdk:"close_wait_timeout"`
	EstablishedTimeout types.Int64 `tfsdk:"established_timeout"`
	FinWaitTimeout     types.Int64 `tfsdk:"fin_wait_timeout"`
	LastAckTimeout     types.Int64 `tfsdk:"last_ack_timeout"`
	SynRecvTimeout     types.Int64 `tfsdk:"syn_recv_timeout"`
	SynSentTimeout     types.Int64 `tfsdk:"syn_sent_timeout"`
	TimeWaitTimeout    types.Int64 `tfsdk:"time_wait_timeout"`
}

func (m *TCPTimeoutModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"close_timeout":       types.Int64Type,
		"close_wait_timeout":  types.Int64Type,
		"established_timeout": types.Int64Type,
		"fin_wait_timeout":    types.Int64Type,
		"last_ack_timeout":    types.Int64Type,
		"syn_recv_timeout":    types.Int64Type,
		"syn_sent_timeout":    types.Int64Type,
		"time_wait_timeout":   types.Int64Type,
	}
}

// DNSVerificationModel represents the DNS Verification configuration
type DNSVerificationModel struct {
	Domain             types.String `tfsdk:"domain"`
	PrimaryDNSServer   types.String `tfsdk:"primary_dns_server"`
	SecondaryDNSServer types.String `tfsdk:"secondary_dns_server"`
	SettingPreference  types.String `tfsdk:"setting_preference"`
}

func (m *DNSVerificationModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"domain":               types.StringType,
		"primary_dns_server":   types.StringType,
		"secondary_dns_server": types.StringType,
		"setting_preference":   types.StringType,
	}
}

// DNSVerificationModel represents the DNS Verification configuration
type DHCPRelayModel struct {
	AgentsPackets types.String `tfsdk:"agents_packets"`
	HopCount      types.Int64  `tfsdk:"hop_count"`
	MaxSize       types.Int64  `tfsdk:"max_size"`
	Port          types.Int64  `tfsdk:"port"`
}

func (m *DHCPRelayModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"agents_packets": types.StringType,
		"hop_count":      types.Int64Type,
		"max_size":       types.Int64Type,
		"port":           types.Int64Type,
	}
}

// usgModel represents the data model for USG (UniFi Security Gateway) settings.
// It defines how USG features like mDNS and DHCP relay are configured for a UniFi site.
type usgModel struct {
	base.Model
	MulticastDnsEnabled types.Bool `tfsdk:"multicast_dns_enabled"`

	// Geo IP filtering
	GeoIPFiltering types.Object `tfsdk:"geo_ip_filtering"`

	// UPNP configuration
	Upnp types.Object `tfsdk:"upnp"`

	// ARP Cache Configuration
	ArpCacheBaseReachable types.Int64  `tfsdk:"arp_cache_base_reachable"`
	ArpCacheTimeout       types.String `tfsdk:"arp_cache_timeout"`

	// DHCP Configuration
	BroadcastPing       types.Bool   `tfsdk:"broadcast_ping"`
	DhcpdHostfileUpdate types.Bool   `tfsdk:"dhcpd_hostfile_update"`
	DhcpdUseDnsmasq     types.Bool   `tfsdk:"dhcpd_use_dnsmasq"`
	DnsmasqAllServers   types.Bool   `tfsdk:"dnsmasq_all_servers"`
	DhcpRelayServers    types.List   `tfsdk:"dhcp_relay_servers"` // TODO deprecated
	DhcpRelay           types.Object `tfsdk:"dhcp_relay"`

	// DNS Verification
	DnsVerification types.Object `tfsdk:"dns_verification"`

	// Network Tools
	EchoServer types.String `tfsdk:"echo_server"`

	// Protocol Modules
	FtpModule  types.Bool `tfsdk:"ftp_module"`
	GreModule  types.Bool `tfsdk:"gre_module"`
	H323Module types.Bool `tfsdk:"h323_module"`
	PptpModule types.Bool `tfsdk:"pptp_module"`
	SipModule  types.Bool `tfsdk:"sip_module"`
	TftpModule types.Bool `tfsdk:"tftp_module"`

	// ICMP Settings
	IcmpTimeout types.Int64 `tfsdk:"icmp_timeout"`

	// LLDP Settings
	LldpEnableAll types.Bool `tfsdk:"lldp_enable_all"`

	// MSS Clamp Settings
	MssClamp    types.String `tfsdk:"mss_clamp"`
	MssClampMss types.Int64  `tfsdk:"mss_clamp_mss"`

	// Offload Settings
	OffloadAccounting types.Bool `tfsdk:"offload_accounting"`
	OffloadL2Blocking types.Bool `tfsdk:"offload_l2_blocking"`
	OffloadSch        types.Bool `tfsdk:"offload_sch"`

	// Timeout Settings
	OtherTimeout             types.Int64  `tfsdk:"other_timeout"`
	TimeoutSettingPreference types.String `tfsdk:"timeout_setting_preference"`

	// TCP Settings (nested)
	TcpTimeouts types.Object `tfsdk:"tcp_timeouts"`

	// Redirects
	ReceiveRedirects types.Bool `tfsdk:"receive_redirects"`
	SendRedirects    types.Bool `tfsdk:"send_redirects"`

	// Security Settings
	SynCookies types.Bool `tfsdk:"syn_cookies"`

	// UDP Settings
	UdpOtherTimeout  types.Int64 `tfsdk:"udp_other_timeout"`
	UdpStreamTimeout types.Int64 `tfsdk:"udp_stream_timeout"`

	// WAN Settings
	UnbindWanMonitors types.Bool `tfsdk:"unbind_wan_monitors"`
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

	// TODO deprecated
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
	// TODO end of deprecated

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

	if base.IsDefined(d.TcpTimeouts) {
		var tcpTimeouts *TCPTimeoutModel
		diags.Append(d.TcpTimeouts.As(ctx, &tcpTimeouts, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}

		model.TCPCloseTimeout = int(tcpTimeouts.CloseTimeout.ValueInt64())
		model.TCPCloseWaitTimeout = int(tcpTimeouts.CloseWaitTimeout.ValueInt64())
		model.TCPEstablishedTimeout = int(tcpTimeouts.EstablishedTimeout.ValueInt64())
		model.TCPFinWaitTimeout = int(tcpTimeouts.FinWaitTimeout.ValueInt64())
		model.TCPLastAckTimeout = int(tcpTimeouts.LastAckTimeout.ValueInt64())
		model.TCPSynRecvTimeout = int(tcpTimeouts.SynRecvTimeout.ValueInt64())
		model.TCPSynSentTimeout = int(tcpTimeouts.SynSentTimeout.ValueInt64())
		model.TCPTimeWaitTimeout = int(tcpTimeouts.TimeWaitTimeout.ValueInt64())
	}

	// Assign DNS Verification attributes
	if base.IsDefined(d.DnsVerification) {
		var dnsVerification *DNSVerificationModel
		diags.Append(d.DnsVerification.As(ctx, &dnsVerification, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}

		model.DNSVerification = unifi.SettingUsgDNSVerification{
			Domain:             dnsVerification.Domain.ValueString(),
			PrimaryDNSServer:   dnsVerification.PrimaryDNSServer.ValueString(),
			SecondaryDNSServer: dnsVerification.SecondaryDNSServer.ValueString(),
			SettingPreference:  dnsVerification.SettingPreference.ValueString(),
		}
	}

	if base.IsDefined(d.DhcpRelay) {
		var dhcpRelay *DHCPRelayModel
		diags.Append(d.DhcpRelay.As(ctx, &dhcpRelay, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}

		model.DHCPRelayAgentsPackets = dhcpRelay.AgentsPackets.ValueString()
		model.DHCPRelayHopCount = int(dhcpRelay.HopCount.ValueInt64())
		model.DHCPRelayMaxSize = int(dhcpRelay.MaxSize.ValueInt64())
		model.DHCPRelayPort = int(dhcpRelay.Port.ValueInt64())
	}

	model.ArpCacheBaseReachable = int(d.ArpCacheBaseReachable.ValueInt64())
	model.ArpCacheTimeout = d.ArpCacheTimeout.ValueString()
	model.BroadcastPing = d.BroadcastPing.ValueBool()
	model.DHCPDHostfileUpdate = d.DhcpdHostfileUpdate.ValueBool()
	model.DHCPDUseDNSmasq = d.DhcpdUseDnsmasq.ValueBool()
	model.DNSmasqAllServers = d.DnsmasqAllServers.ValueBool()
	//model.DHCPRelayAgentsPackets = d.DhcpRelayAgentsPackets.ValueString()
	//model.DHCPRelayHopCount = int(d.DhcpRelayHopCount.ValueInt64())
	//model.DHCPRelayMaxSize = int(d.DhcpRelayMaxSize.ValueInt64())
	//model.DHCPRelayPort = int(d.DhcpRelayPort.ValueInt64())
	model.EchoServer = d.EchoServer.ValueString()
	model.FtpModule = d.FtpModule.ValueBool()
	model.GreModule = d.GreModule.ValueBool()
	model.H323Module = d.H323Module.ValueBool()
	model.PptpModule = d.PptpModule.ValueBool()
	model.SipModule = d.SipModule.ValueBool()
	model.TFTPModule = d.TftpModule.ValueBool()
	model.ICMPTimeout = int(d.IcmpTimeout.ValueInt64())
	model.LldpEnableAll = d.LldpEnableAll.ValueBool()
	model.MssClamp = d.MssClamp.ValueString()
	model.MssClampMss = int(d.MssClampMss.ValueInt64())
	model.OffloadAccounting = d.OffloadAccounting.ValueBool()
	model.OffloadL2Blocking = d.OffloadL2Blocking.ValueBool()
	model.OffloadSch = d.OffloadSch.ValueBool()
	model.OtherTimeout = int(d.OtherTimeout.ValueInt64())
	model.TimeoutSettingPreference = d.TimeoutSettingPreference.ValueString()
	model.ReceiveRedirects = d.ReceiveRedirects.ValueBool()
	model.SendRedirects = d.SendRedirects.ValueBool()
	model.SynCookies = d.SynCookies.ValueBool()
	model.UDPOtherTimeout = int(d.UdpOtherTimeout.ValueInt64())
	model.UDPStreamTimeout = int(d.UdpStreamTimeout.ValueInt64())
	model.UnbindWANMonitors = d.UnbindWanMonitors.ValueBool()
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

	// Convert DNS Verification settings
	dnsVerificationModel := DNSVerificationModel{
		Domain:             types.StringValue(model.DNSVerification.Domain),
		PrimaryDNSServer:   types.StringValue(model.DNSVerification.PrimaryDNSServer),
		SecondaryDNSServer: types.StringValue(model.DNSVerification.SecondaryDNSServer),
		SettingPreference:  types.StringValue(model.DNSVerification.SettingPreference),
	}
	dnsVerificationObj, dnsVerificationObjDiags := types.ObjectValueFrom(ctx, dnsVerificationModel.AttributeTypes(), &dnsVerificationModel)
	diags.Append(dnsVerificationObjDiags...)

	d.DnsVerification = dnsVerificationObj
	// Convert TCP Timeout settings
	tcpTimeoutModel := TCPTimeoutModel{
		CloseTimeout:       types.Int64Value(int64(model.TCPCloseTimeout)),
		CloseWaitTimeout:   types.Int64Value(int64(model.TCPCloseWaitTimeout)),
		EstablishedTimeout: types.Int64Value(int64(model.TCPEstablishedTimeout)),
		FinWaitTimeout:     types.Int64Value(int64(model.TCPFinWaitTimeout)),
		LastAckTimeout:     types.Int64Value(int64(model.TCPLastAckTimeout)),
		SynRecvTimeout:     types.Int64Value(int64(model.TCPSynRecvTimeout)),
		SynSentTimeout:     types.Int64Value(int64(model.TCPSynSentTimeout)),
		TimeWaitTimeout:    types.Int64Value(int64(model.TCPTimeWaitTimeout)),
	}

	tcpTimeoutObj, tcpTimeoutObjDiags := types.ObjectValueFrom(ctx, tcpTimeoutModel.AttributeTypes(), &tcpTimeoutModel)
	diags.Append(tcpTimeoutObjDiags...)
	d.TcpTimeouts = tcpTimeoutObj

	// Convert DHCP Relay settings
	dhcpRelayModel := DHCPRelayModel{
		AgentsPackets: types.StringValue(model.DHCPRelayAgentsPackets),
		HopCount:      types.Int64Value(int64(model.DHCPRelayHopCount)),
		MaxSize:       types.Int64Value(int64(model.DHCPRelayMaxSize)),
		Port:          types.Int64Value(int64(model.DHCPRelayPort)),
	}

	// TODO deprecated

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
	// TODO end of deprecated
	dhcpRelayObj, dhcpRelayObjDiags := types.ObjectValueFrom(ctx, dhcpRelayModel.AttributeTypes(), &dhcpRelayModel)
	diags.Append(dhcpRelayObjDiags...)
	d.DhcpRelay = dhcpRelayObj

	// Set all flat attributes
	d.ArpCacheBaseReachable = types.Int64Value(int64(model.ArpCacheBaseReachable))
	d.ArpCacheTimeout = types.StringValue(model.ArpCacheTimeout)
	d.BroadcastPing = types.BoolValue(model.BroadcastPing)
	d.DhcpdHostfileUpdate = types.BoolValue(model.DHCPDHostfileUpdate)
	d.DhcpdUseDnsmasq = types.BoolValue(model.DHCPDUseDNSmasq)
	d.DnsmasqAllServers = types.BoolValue(model.DNSmasqAllServers)
	d.EchoServer = types.StringValue(model.EchoServer)
	d.FtpModule = types.BoolValue(model.FtpModule)
	d.GreModule = types.BoolValue(model.GreModule)
	d.H323Module = types.BoolValue(model.H323Module)
	d.PptpModule = types.BoolValue(model.PptpModule)
	d.SipModule = types.BoolValue(model.SipModule)
	d.TftpModule = types.BoolValue(model.TFTPModule)
	d.IcmpTimeout = types.Int64Value(int64(model.ICMPTimeout))
	d.LldpEnableAll = types.BoolValue(model.LldpEnableAll)
	d.MssClamp = types.StringValue(model.MssClamp)
	d.MssClampMss = types.Int64Value(int64(model.MssClampMss))
	d.OffloadAccounting = types.BoolValue(model.OffloadAccounting)
	d.OffloadL2Blocking = types.BoolValue(model.OffloadL2Blocking)
	d.OffloadSch = types.BoolValue(model.OffloadSch)
	d.OtherTimeout = types.Int64Value(int64(model.OtherTimeout))
	d.TimeoutSettingPreference = types.StringValue(model.TimeoutSettingPreference)
	d.ReceiveRedirects = types.BoolValue(model.ReceiveRedirects)
	d.SendRedirects = types.BoolValue(model.SendRedirects)
	d.SynCookies = types.BoolValue(model.SynCookies)
	d.UdpOtherTimeout = types.Int64Value(int64(model.UDPOtherTimeout))
	d.UdpStreamTimeout = types.Int64Value(int64(model.UDPStreamTimeout))
	d.UnbindWanMonitors = types.BoolValue(model.UnbindWANMonitors)
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
				DeprecationMessage: "This attribute is deprecated and will be removed in a future release. `dhcp_relay.servers` attribute will be introduced as a replacement.",
				ElementType:        types.StringType,
				Optional:           true,
				Computed:           true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Default: utils.DefaultEmptyList(types.StringType),
				Validators: []validator.List{
					listvalidator.SizeAtMost(5),
					listvalidator.ValueStringsAre(validators.IPv4()),
				},
			},
			"dhcp_relay": schema.SingleNestedAttribute{
				MarkdownDescription: "DHCP relay configuration. Forwards DHCP requests to external servers.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"agents_packets": schema.StringAttribute{
						MarkdownDescription: "Specify the DHCP relay agent's packets. Valid values are `append`, `discard`, `forward`, or `replace`. ",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("append", "discard", "forward", "replace"),
						},
					},
					"hop_count": schema.Int64Attribute{
						MarkdownDescription: "Maximum number of hops for DHCP relay packets.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Int64{
							int64validator.Between(1, 255),
						},
					},
					"max_size": schema.Int64Attribute{
						MarkdownDescription: "Maximum size of DHCP relay packets. Requires value between 64 and 1400.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Int64{
							int64validator.Between(64, 1400),
						},
					},
					"port": schema.Int64Attribute{
						MarkdownDescription: "Port for DHCP relay to listen on.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Int64{
							int64validator.Between(1, 65535),
						},
					},
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
			// ARP Cache Configuration
			"arp_cache_base_reachable": schema.Int64Attribute{
				MarkdownDescription: "The base reachable timeout for ARP cache entries in seconds.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"arp_cache_timeout": schema.StringAttribute{
				MarkdownDescription: "The timeout strategy for ARP cache. Valid values are 'normal', 'min-dhcp-lease', or 'custom'.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// DHCP Configuration
			"broadcast_ping": schema.BoolAttribute{
				MarkdownDescription: "Enable responding to broadcast ping requests.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"dhcpd_hostfile_update": schema.BoolAttribute{
				MarkdownDescription: "Enable updating hostfiles with DHCP client information.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"dhcpd_use_dnsmasq": schema.BoolAttribute{
				MarkdownDescription: "Use dnsmasq for DHCP services.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"dnsmasq_all_servers": schema.BoolAttribute{
				MarkdownDescription: "Query all DNS servers.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			// DNS Verification
			"dns_verification": schema.SingleNestedAttribute{
				MarkdownDescription: "DNS verification settings for validating DNS responses.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Object{
					validators.RequiredTogetherIf(path.MatchRoot("setting_preference"), types.StringValue("manual"), path.MatchRoot("primary_dns_server"), path.MatchRoot("domain")),
					validators.RequiredNoneIf(path.MatchRoot("setting_preference"), types.StringValue("auto"), path.MatchRoot("primary_dns_server"), path.MatchRoot("secondary_dns_server"), path.MatchRoot("domain")),
				},
				Attributes: map[string]schema.Attribute{
					"domain": schema.StringAttribute{
						MarkdownDescription: "Domain for DNS verification.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"primary_dns_server": schema.StringAttribute{
						MarkdownDescription: "Primary DNS server for verification.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							validators.IPv4(),
						},
					},
					"secondary_dns_server": schema.StringAttribute{
						MarkdownDescription: "Secondary DNS server for verification.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							validators.IPv4(),
						},
					},
					"setting_preference": schema.StringAttribute{
						MarkdownDescription: "Preference setting for DNS verification (auto, manual).",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("auto", "manual"),
						},
					},
				},
			},

			// Network Tools
			"echo_server": schema.StringAttribute{
				MarkdownDescription: "Server for echo tests.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Protocol Modules
			"ftp_module": schema.BoolAttribute{
				MarkdownDescription: "Enable FTP protocol helper module.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"gre_module": schema.BoolAttribute{
				MarkdownDescription: "Enable GRE protocol helper module.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"h323_module": schema.BoolAttribute{
				MarkdownDescription: "Enable H.323 protocol helper module.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"pptp_module": schema.BoolAttribute{
				MarkdownDescription: "Enable PPTP protocol helper module. Requires GRE module (`gre_module`) to be enabled.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"sip_module": schema.BoolAttribute{
				MarkdownDescription: "Enable SIP protocol helper module.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"tftp_module": schema.BoolAttribute{
				MarkdownDescription: "Enable TFTP protocol helper module.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			// ICMP Settings
			"icmp_timeout": schema.Int64Attribute{
				MarkdownDescription: "ICMP timeout in seconds for connection tracking.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},

			// LLDP Settings
			"lldp_enable_all": schema.BoolAttribute{
				MarkdownDescription: "Enable Link Layer Discovery Protocol (LLDP) on all interfaces.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			// MSS Clamp Settings
			"mss_clamp": schema.StringAttribute{
				MarkdownDescription: "TCP Maximum Segment Size clamping mode.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"mss_clamp_mss": schema.Int64Attribute{
				MarkdownDescription: "TCP Maximum Segment Size value.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(100, 9999),
				},
			},

			// Offload Settings
			"offload_accounting": schema.BoolAttribute{
				MarkdownDescription: "Enable accounting offload.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"offload_l2_blocking": schema.BoolAttribute{
				MarkdownDescription: "Enable L2 blocking offload.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"offload_sch": schema.BoolAttribute{
				MarkdownDescription: "Enable scheduling offload.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			// Timeout Settings
			"other_timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout for other protocols in seconds.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"timeout_setting_preference": schema.StringAttribute{
				MarkdownDescription: "Preference for timeout settings (auto, manual).",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("auto", "manual"),
				},
			},

			// TCP Settings (nested)
			"tcp_timeouts": schema.SingleNestedAttribute{
				MarkdownDescription: "TCP timeouts for various connection states.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"close_timeout": schema.Int64Attribute{
						MarkdownDescription: "Timeout for TCP connections in CLOSE state.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"close_wait_timeout": schema.Int64Attribute{
						MarkdownDescription: "Timeout for TCP connections in CLOSE_WAIT state.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"established_timeout": schema.Int64Attribute{
						MarkdownDescription: "Timeout for TCP connections in ESTABLISHED state.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"fin_wait_timeout": schema.Int64Attribute{
						MarkdownDescription: "Timeout for TCP connections in FIN_WAIT state.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"last_ack_timeout": schema.Int64Attribute{
						MarkdownDescription: "Timeout for TCP connections in LAST_ACK state.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"syn_recv_timeout": schema.Int64Attribute{
						MarkdownDescription: "Timeout for TCP connections in SYN_RECV state.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"syn_sent_timeout": schema.Int64Attribute{
						MarkdownDescription: "Timeout for TCP connections in SYN_SENT state.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"time_wait_timeout": schema.Int64Attribute{
						MarkdownDescription: "Timeout for TCP connections in TIME_WAIT state.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
				},
			},

			// Redirects
			"receive_redirects": schema.BoolAttribute{
				MarkdownDescription: "Accept ICMP redirect messages.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"send_redirects": schema.BoolAttribute{
				MarkdownDescription: "Allow sending ICMP redirect messages.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			// Security Settings
			"syn_cookies": schema.BoolAttribute{
				MarkdownDescription: "Enable SYN cookies to protect against SYN flood attacks.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			// UDP Settings
			"udp_other_timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout for other UDP connections in seconds.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"udp_stream_timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout for UDP stream connections in seconds.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},

			// WAN Settings
			"unbind_wan_monitors": schema.BoolAttribute{
				MarkdownDescription: "Unbind WAN monitors to prevent unnecessary traffic.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
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
	return []resource.ConfigValidator{
		validators.RequiredValueIf(path.MatchRoot("pptp_module"), types.BoolValue(true), path.MatchRoot("gre_module"), types.BoolValue(true)),
	}
}

func (r *usgResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	resp.Diagnostics.Append(r.RequireMaxVersionForPath("7.0", path.Root("multicast_dns_enabled"), req.Config)...)
}

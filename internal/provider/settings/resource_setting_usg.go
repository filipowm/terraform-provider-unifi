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
			"  * Multicast DNS (mDNS) for cross-VLAN service discovery\n" +
			"  * DHCP relay for forwarding DHCP requests to external servers\n" +
			"  * Geo IP filtering for country-based traffic control\n" +
			"  * UPNP/NAT-PMP for automatic port forwarding\n" +
			"  * Protocol helpers for FTP, GRE, H323, PPTP, SIP, and TFTP\n" +
			"  * TCP/UDP timeout settings for connection tracking\n" +
			"  * Security features like SYN cookies and ICMP redirect controls\n" +
			"  * MSS clamping for optimizing MTU issues\n\n" +
			"Note: Some settings may not be available on all controller versions. For example, multicast_dns_enabled is not supported on UniFi OS v7+. Changes to certain attributes may not be reflected in the plan unless explicitly modified in the configuration.",
		Attributes: map[string]schema.Attribute{
			"id":   base.ID(),
			"site": base.SiteAttribute(),
			"multicast_dns_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable multicast DNS (mDNS/Bonjour/Avahi) forwarding across VLANs. This allows devices to discover services " +
					"(like printers, Chromecasts, Apple devices, etc.) even when they are on different networks or VLANs. " +
					"When enabled, the gateway will forward mDNS packets between networks, facilitating cross-VLAN service discovery. " +
					"Note: This setting is not supported on UniFi OS v7+ as it has been replaced by mDNS settings in the network configuration.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"dhcp_relay_servers": schema.ListAttribute{
				MarkdownDescription: "List of up to 5 DHCP relay servers (specified by IP address) that will receive forwarded DHCP requests. " +
					"This is useful when you want to use external DHCP servers instead of the built-in DHCP server on the USG/UDM. " +
					"When configured, the gateway will forward DHCP discovery packets from clients to these external servers, allowing " +
					"centralized IP address management across multiple networks. " +
					"Example: `['192.168.1.5', '192.168.2.5']`",
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
				MarkdownDescription: "Advanced DHCP relay configuration settings. Controls how the gateway forwards DHCP requests to external servers " +
					"and manages DHCP relay agent behavior. Use this block to fine-tune DHCP relay functionality beyond simply specifying relay servers.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"agents_packets": schema.StringAttribute{
						MarkdownDescription: "Specifies how to handle DHCP relay agent information in packets. Valid values are:\n" +
							"  * `append` - Add relay agent information to packets that may already contain it\n" +
							"  * `discard` - Drop packets that already contain relay agent information\n" +
							"  * `forward` - Forward packets regardless of relay agent information\n" +
							"  * `replace` - Replace existing relay agent information with the gateway's information",
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("append", "discard", "forward", "replace"),
						},
					},
					"hop_count": schema.Int64Attribute{
						MarkdownDescription: "Maximum number of relay agents that can forward the DHCP packet before it is discarded. " +
							"This prevents DHCP packets from being forwarded indefinitely in complex network topologies. " +
							"Valid values range from 1 to 255, with lower values recommended for simpler networks.",
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Int64{
							int64validator.Between(1, 255),
						},
					},
					"max_size": schema.Int64Attribute{
						MarkdownDescription: "Maximum size (in bytes) of DHCP relay packets that will be forwarded. " +
							"Packets exceeding this size will be truncated or dropped. Valid values range from 64 to 1400 bytes. " +
							"The default is typically sufficient for most DHCP implementations, but may need adjustment if using " +
							"extensive DHCP options or vendor-specific information.",
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Int64{
							int64validator.Between(64, 1400),
						},
					},
					"port": schema.Int64Attribute{
						MarkdownDescription: "UDP port number for the DHCP relay service to listen on. The standard DHCP server port is 67, " +
							"but this can be customized if needed for specific network configurations. Valid values range from 1 to 65535. " +
							"Ensure this doesn't conflict with other services running on the gateway.",
						Optional: true,
						Computed: true,
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
				MarkdownDescription: "Geographic IP filtering configuration that allows blocking or allowing traffic based on country of origin. " +
					"This feature uses IP geolocation databases to identify the country associated with IP addresses and apply filtering rules. " +
					"Useful for implementing country-specific access policies or blocking traffic from high-risk regions.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Object{
					validators.RequiredTogetherIf(path.MatchRoot("enabled"), types.BoolValue(true), path.MatchRoot("countries")),
				},
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable geographic IP filtering. When enabled, traffic from specified countries will be blocked or allowed " +
							"according to the configured rules. When set to `true`, you must also specify the `countries` list. " +
							"Setting this to `false` disables all country-based filtering regardless of other settings.",
						Required: true,
					},
					"block": schema.StringAttribute{
						MarkdownDescription: "Specifies whether the selected countries should be blocked or allowed. Valid values are:\n" +
							"  * `block` (default) - Traffic from the specified countries will be blocked, while traffic from all other countries will be allowed\n" +
							"  * `allow` - Only traffic from the specified countries will be allowed, while traffic from all other countries will be blocked\n\n" +
							"This setting effectively determines whether the `countries` list functions as a blocklist or an allowlist.",
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
						MarkdownDescription: "List of two-letter ISO 3166-1 alpha-2 country codes to block or allow, depending on the `block` setting. " +
							"Must contain at least one country code when geo IP filtering is enabled. Country codes are case-insensitive but are typically " +
							"written in uppercase.\n\n" +
							"Examples:\n" +
							"  * `['US', 'CA', 'MX']` - United States, Canada, and Mexico\n" +
							"  * `['CN', 'RU', 'IR']` - China, Russia, and Iran\n" +
							"  * `['GB', 'DE', 'FR']` - United Kingdom, Germany, and France",
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
						MarkdownDescription: "Specifies which traffic direction the geo IP filtering applies to. Valid values are:\n" +
							"  * `both` (default) - Filters traffic in both directions (incoming and outgoing)\n" +
							"  * `ingress` - Filters only incoming traffic (from WAN to LAN)\n" +
							"  * `egress` - Filters only outgoing traffic (from LAN to WAN)\n\n" +
							"This setting is useful for creating more granular filtering policies. For example, you might want to block incoming traffic " +
							"from certain countries while still allowing outgoing connections to those same countries.",
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
				MarkdownDescription: "UPNP (Universal Plug and Play) configuration settings. UPNP allows compatible applications and devices to automatically " +
					"configure port forwarding rules on the gateway without manual intervention. This is commonly used by gaming consoles, " +
					"media servers, VoIP applications, and other network services that require incoming connections.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable UPNP functionality. When enabled, applications and devices on the local network can automatically " +
							"request port forwarding rules from the gateway without manual configuration. This simplifies the use of applications " +
							"that require inbound connections, but may present security risks if not properly configured with `secure_mode`.",
						Required: true,
					},
					"nat_pmp_enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable NAT-PMP (NAT Port Mapping Protocol) support alongside UPNP. NAT-PMP is " +
							"Apple's alternative to UPNP, providing similar automatic port mapping capabilities. When enabled, Apple devices " +
							"like Macs, iPhones, and iPads can automatically configure port forwarding for services like AirPlay, FaceTime, " +
							"iMessage, and other Apple services. Defaults to `false`.",
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(false),
					},
					"secure_mode": schema.BoolAttribute{
						MarkdownDescription: "Enable secure mode for UPNP. In secure mode, the gateway only forwards ports " +
							"to the device that specifically requested them, enhancing security. This prevents malicious applications from " +
							"redirecting ports to different devices than intended. It's strongly recommended to enable this setting when using UPNP " +
							"to minimize security risks. Defaults to `false`.",
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(false),
					},
					"wan_interface": schema.StringAttribute{
						MarkdownDescription: "Specify which WAN interface to use for UPNP service. Valid values are:\n" +
							"  * `WAN` (default) - Use the primary WAN interface for UPNP port forwarding\n" +
							"  * `WAN2` - Use the secondary WAN interface for UPNP port forwarding (if available)\n\n" +
							"This setting is particularly relevant for dual-WAN setups where you may want to direct UPNP traffic through " +
							"a specific WAN connection. If your gateway only has a single WAN interface, use the default `WAN` setting.",
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
				MarkdownDescription: "The base reachable timeout (in seconds) for ARP cache entries. This controls how long the gateway considers " +
					"a MAC-to-IP mapping valid without needing to refresh it. Higher values reduce network traffic but may cause stale " +
					"entries if devices change IP addresses frequently.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"arp_cache_timeout": schema.StringAttribute{
				MarkdownDescription: "The timeout strategy for ARP cache entries. Valid values are:\n" +
					"  * `normal` - Use system default timeouts\n" +
					"  * `min-dhcp-lease` - Set ARP timeout to match the minimum DHCP lease time\n" +
					"  * `custom` - Use the custom timeout value specified in `arp_cache_base_reachable`\n\n" +
					"This setting determines how long MAC-to-IP mappings are stored in the ARP cache before being refreshed.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// DHCP Configuration
			"broadcast_ping": schema.BoolAttribute{
				MarkdownDescription: "Enable responding to broadcast ping requests (ICMP echo requests sent to the broadcast address). " +
					"When enabled, the gateway will respond to pings sent to the broadcast address of the network (e.g., 192.168.1.255). " +
					"This can be useful for network diagnostics but may also be used in certain denial-of-service attacks.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"dhcpd_hostfile_update": schema.BoolAttribute{
				MarkdownDescription: "Enable updating the gateway's host files with DHCP client information. When enabled, the gateway will " +
					"automatically add entries to its host file for each DHCP client, allowing hostname resolution for devices " +
					"that receive IP addresses via DHCP. This improves name resolution on the local network.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"dhcpd_use_dnsmasq": schema.BoolAttribute{
				MarkdownDescription: "Use dnsmasq for DHCP services instead of the default DHCP server. Dnsmasq provides integrated DNS and DHCP " +
					"functionality with additional features like DNS caching, DHCP static leases, and local domain name resolution. " +
					"This can improve DNS resolution performance and provide more flexible DHCP options.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"dnsmasq_all_servers": schema.BoolAttribute{
				MarkdownDescription: "When enabled, dnsmasq will query all configured DNS servers simultaneously and use the fastest response. " +
					"This can improve DNS resolution speed but may increase DNS traffic. By default, dnsmasq queries servers " +
					"sequentially, only trying the next server if the current one fails to respond.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			// DNS Verification
			"dns_verification": schema.SingleNestedAttribute{
				MarkdownDescription: "DNS verification settings for validating DNS responses. This feature helps detect and prevent DNS spoofing " +
					"attacks by verifying DNS responses against trusted DNS servers. When configured, the gateway can compare DNS " +
					"responses with those from known trusted servers to identify potential tampering or poisoning attempts.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Object{
					validators.RequiredTogetherIf(path.MatchRoot("setting_preference"), types.StringValue("manual"), path.MatchRoot("primary_dns_server"), path.MatchRoot("domain")),
					validators.RequiredNoneIf(path.MatchRoot("setting_preference"), types.StringValue("auto"), path.MatchRoot("primary_dns_server"), path.MatchRoot("secondary_dns_server"), path.MatchRoot("domain")),
				},
				Attributes: map[string]schema.Attribute{
					"domain": schema.StringAttribute{
						MarkdownDescription: "The domain name to use for DNS verification tests. The gateway will query this domain when testing DNS " +
							"server responses. This should be a reliable domain that is unlikely to change frequently. " +
							"Required when `setting_preference` is set to `manual`.",
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"primary_dns_server": schema.StringAttribute{
						MarkdownDescription: "The IP address of the primary trusted DNS server to use for verification. DNS responses will be compared " +
							"against responses from this server to detect potential DNS spoofing. Required when `setting_preference` is " +
							"set to `manual`. Must be a valid IPv4 address.",
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							validators.IPv4(),
						},
					},
					"secondary_dns_server": schema.StringAttribute{
						MarkdownDescription: "The IP address of the secondary trusted DNS server to use for verification. This server will be used " +
							"if the primary server is unavailable. Optional even when `setting_preference` is set to `manual`. " +
							"Must be a valid IPv4 address if specified.",
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							validators.IPv4(),
						},
					},
					"setting_preference": schema.StringAttribute{
						MarkdownDescription: "Determines how DNS verification servers are configured. Valid values are:\n" +
							"  * `auto` - The gateway will automatically select DNS servers for verification\n" +
							"  * `manual` - Use the manually specified `primary_dns_server` and optionally `secondary_dns_server`\n\n" +
							"When set to `manual`, you must also specify `primary_dns_server` and `domain` values.",
						Optional: true,
						Computed: true,
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
				MarkdownDescription: "The hostname or IP address of a server to use for network echo tests. Echo tests send packets to this server " +
					"and measure response times to evaluate network connectivity and performance. This can be used for network " +
					"diagnostics and monitoring.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Protocol Modules
			"ftp_module": schema.BoolAttribute{
				MarkdownDescription: "Enable the FTP (File Transfer Protocol) helper module. This module allows the gateway to properly handle " +
					"FTP connections through NAT by tracking the control channel and dynamically opening required data ports. " +
					"Without this helper, passive FTP connections may fail when clients are behind NAT.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"gre_module": schema.BoolAttribute{
				MarkdownDescription: "Enable the GRE (Generic Routing Encapsulation) protocol helper module. This module allows proper handling " +
					"of GRE tunneling protocol through the gateway's firewall. GRE is commonly used for VPN tunnels and other " +
					"encapsulation needs. Required if you plan to use PPTP VPNs (see `pptp_module`).",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"h323_module": schema.BoolAttribute{
				MarkdownDescription: "Enable the H.323 protocol helper module. H.323 is a standard for multimedia communications (audio, video, " +
					"and data) over packet-based networks. This helper allows H.323-based applications like video conferencing " +
					"systems to work properly through NAT by tracking connection details and opening required ports.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"pptp_module": schema.BoolAttribute{
				MarkdownDescription: "Enable the PPTP (Point-to-Point Tunneling Protocol) helper module. This module allows PPTP VPN connections " +
					"to work properly through the gateway's firewall and NAT. PPTP uses GRE for tunneling, so the `gre_module` " +
					"must also be enabled for PPTP to function correctly. Note that PPTP has known security vulnerabilities and " +
					"more secure VPN protocols are generally recommended.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"sip_module": schema.BoolAttribute{
				MarkdownDescription: "Enable the SIP (Session Initiation Protocol) helper module. SIP is used for initiating, maintaining, and " +
					"terminating real-time sessions for voice, video, and messaging applications (VoIP, video conferencing). " +
					"This helper allows SIP-based applications to work correctly through NAT by tracking SIP connections and " +
					"dynamically opening the necessary ports for media streams.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"tftp_module": schema.BoolAttribute{
				MarkdownDescription: "Enable the TFTP (Trivial File Transfer Protocol) helper module. This module allows TFTP connections to work " +
					"properly through the gateway's firewall and NAT. TFTP is commonly used for firmware updates, configuration " +
					"file transfers, and network booting of devices. The helper tracks TFTP connections and ensures return traffic " +
					"is properly handled.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			// ICMP Settings
			"icmp_timeout": schema.Int64Attribute{
				MarkdownDescription: "ICMP timeout in seconds for connection tracking. This controls how long the gateway maintains state " +
					"information for ICMP (ping) packets in its connection tracking table. Higher values maintain ICMP connection " +
					"state longer, while lower values reclaim resources more quickly but may affect some diagnostic tools.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},

			// LLDP Settings
			"lldp_enable_all": schema.BoolAttribute{
				MarkdownDescription: "Enable Link Layer Discovery Protocol (LLDP) on all interfaces. LLDP is a vendor-neutral protocol that " +
					"allows network devices to advertise their identity, capabilities, and neighbors on a local network. When enabled, " +
					"the gateway will both send and receive LLDP packets, facilitating network discovery and management tools.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			// MSS Clamp Settings
			"mss_clamp": schema.StringAttribute{
				MarkdownDescription: "TCP Maximum Segment Size (MSS) clamping mode. MSS clamping adjusts the maximum segment size of TCP packets " +
					"to prevent fragmentation issues when packets traverse networks with different MTU sizes. Valid values include:\n" +
					"  * `auto` - Automatically determine appropriate MSS values based on interface MTUs\n" +
					"  * `custom` - Use the custom MSS value specified in `mss_clamp_mss`\n" +
					"  * `disabled` - Do not perform MSS clamping\n\n" +
					"This setting is particularly important for VPN connections and networks with non-standard MTU sizes.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"mss_clamp_mss": schema.Int64Attribute{
				MarkdownDescription: "Custom TCP Maximum Segment Size (MSS) value in bytes. This value is used when `mss_clamp` is set to `custom`. " +
					"The MSS value should typically be set to the path MTU minus 40 bytes (for IPv4) or minus 60 bytes (for IPv6) to account " +
					"for TCP/IP header overhead. Valid values range from 100 to 9999, with common values being 1460 (for standard 1500 MTU) " +
					"or 1400 (for VPN tunnels).",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(100, 9999),
				},
			},

			// Offload Settings
			"offload_accounting": schema.BoolAttribute{
				MarkdownDescription: "Enable hardware accounting offload. When enabled, the gateway will use hardware acceleration for traffic " +
					"accounting functions, reducing CPU load and potentially improving throughput for high-traffic environments. " +
					"This setting may not be supported on all hardware models.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"offload_l2_blocking": schema.BoolAttribute{
				MarkdownDescription: "Enable hardware offload for Layer 2 (L2) blocking functions. When enabled, the gateway will use hardware " +
					"acceleration for blocking traffic at the data link layer (MAC address level), which can improve performance " +
					"when implementing MAC-based filtering or isolation. This setting may not be supported on all hardware models.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"offload_sch": schema.BoolAttribute{
				MarkdownDescription: "Enable hardware scheduling offload. When enabled, the gateway will use hardware acceleration for packet " +
					"scheduling functions, which can improve QoS (Quality of Service) performance and throughput for prioritized traffic. " +
					"This setting may not be supported on all hardware models and may affect other hardware offload capabilities.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			// Timeout Settings
			"other_timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout (in seconds) for connection tracking of protocols other than TCP, UDP, and ICMP. This controls how long " +
					"the gateway maintains state information for connections using other protocols. Higher values maintain connection state " +
					"longer, while lower values reclaim resources more quickly but may affect some applications using non-standard protocols.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"timeout_setting_preference": schema.StringAttribute{
				MarkdownDescription: "Determines how connection timeout values are configured. Valid values are:\n" +
					"  * `auto` - The gateway will automatically determine appropriate timeout values based on system defaults\n" +
					"  * `manual` - Use the manually specified timeout values for various connection types\n\n" +
					"When set to `manual`, you should specify values for the various timeout settings like `tcp_timeouts`, " +
					"`udp_stream_timeout`, `udp_other_timeout`, `icmp_timeout`, and `other_timeout`.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("auto", "manual"),
				},
			},

			// TCP Settings (nested)
			"tcp_timeouts": schema.SingleNestedAttribute{
				MarkdownDescription: "TCP connection timeout settings for various TCP connection states. These settings control how long the gateway " +
					"maintains state information for TCP connections in different states before removing them from the connection tracking " +
					"table. Proper timeout values balance resource usage with connection reliability. These settings are particularly " +
					"relevant when `timeout_setting_preference` is set to `manual`.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"close_timeout": schema.Int64Attribute{
						MarkdownDescription: "Timeout (in seconds) for TCP connections in the CLOSE state. The CLOSE state occurs when a connection is " +
							"being terminated but may still have packets in transit. Lower values reclaim resources more quickly, while higher " +
							"values ensure all packets are properly processed during connection termination.",
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"close_wait_timeout": schema.Int64Attribute{
						MarkdownDescription: "Timeout (in seconds) for TCP connections in the CLOSE_WAIT state. The CLOSE_WAIT state occurs when the remote " +
							"end has initiated connection termination, but the local application hasn't closed the connection yet. This timeout " +
							"prevents resources from being held indefinitely if a local application fails to properly close its connection.",
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"established_timeout": schema.Int64Attribute{
						MarkdownDescription: "Timeout (in seconds) for TCP connections in the ESTABLISHED state. This is the most important TCP timeout as it " +
							"determines how long idle but established connections are maintained in the connection tracking table. Higher values " +
							"(e.g., 86400 = 24 hours) are suitable for long-lived connections, while lower values conserve resources but may cause " +
							"issues with applications that maintain idle connections.",
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"fin_wait_timeout": schema.Int64Attribute{
						MarkdownDescription: "Timeout (in seconds) for TCP connections in the FIN_WAIT state. The FIN_WAIT states occur during the normal " +
							"TCP connection termination process after a FIN packet has been sent. This timeout prevents resources from being held " +
							"if the connection termination process doesn't complete properly.",
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"last_ack_timeout": schema.Int64Attribute{
						MarkdownDescription: "Timeout (in seconds) for TCP connections in the LAST_ACK state. The LAST_ACK state occurs during connection " +
							"termination when the remote end has sent a FIN, the local end has responded with a FIN and ACK, and is waiting for " +
							"the final ACK from the remote end to complete the connection termination.",
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"syn_recv_timeout": schema.Int64Attribute{
						MarkdownDescription: "Timeout (in seconds) for TCP connections in the SYN_RECV state. This state occurs during connection establishment " +
							"after receiving a SYN packet and sending a SYN-ACK, but before receiving the final ACK to complete the three-way " +
							"handshake. A lower timeout helps mitigate SYN flood attacks by releasing resources for incomplete connections more quickly.",
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"syn_sent_timeout": schema.Int64Attribute{
						MarkdownDescription: "Timeout (in seconds) for TCP connections in the SYN_SENT state. This state occurs during connection establishment " +
							"after sending a SYN packet but before receiving a SYN-ACK response. This timeout determines how long the system will " +
							"wait for a response to connection attempts before giving up.",
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"time_wait_timeout": schema.Int64Attribute{
						MarkdownDescription: "Timeout (in seconds) for TCP connections in the TIME_WAIT state. The TIME_WAIT state occurs after a connection " +
							"has been closed but is maintained to ensure any delayed packets are properly handled. The standard recommendation is " +
							"2 minutes (120 seconds), but can be reduced in high-connection environments to free resources more quickly at the " +
							"risk of potential connection issues if delayed packets arrive.",
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
				},
			},

			// Redirects
			"receive_redirects": schema.BoolAttribute{
				MarkdownDescription: "Enable accepting ICMP redirect messages. ICMP redirects are messages sent by routers to inform hosts of better " +
					"routes to specific destinations. When enabled, the gateway will update its routing table based on these messages. " +
					"While useful for route optimization, this can potentially be exploited for man-in-the-middle attacks, so it's often " +
					"disabled in security-sensitive environments.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"send_redirects": schema.BoolAttribute{
				MarkdownDescription: "Enable sending ICMP redirect messages. When enabled, the gateway will send ICMP redirect messages to hosts on the " +
					"local network to inform them of better routes to specific destinations. This can help optimize network traffic but " +
					"is typically only needed when the gateway has multiple interfaces on the same subnet or in complex routing scenarios.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			// Security Settings
			"syn_cookies": schema.BoolAttribute{
				MarkdownDescription: "Enable SYN cookies to protect against SYN flood attacks. SYN cookies are a technique that helps mitigate TCP SYN " +
					"flood attacks by avoiding the need to track incomplete connections in a backlog queue. When enabled, the gateway can " +
					"continue to establish legitimate connections even when under a SYN flood attack. This is a recommended security setting " +
					"for internet-facing gateways.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			// UDP Settings
			"udp_other_timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout (in seconds) for general UDP connections. Since UDP is connectionless, this timeout determines how long the " +
					"gateway maintains state information for UDP packets that don't match the criteria for stream connections. This applies " +
					"to most short-lived UDP communications like DNS queries. Lower values free resources more quickly but may affect some " +
					"applications that expect longer session persistence.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"udp_stream_timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout (in seconds) for UDP stream connections. This applies to UDP traffic patterns that resemble ongoing streams, " +
					"such as VoIP calls, video streaming, or online gaming. The gateway identifies these based on traffic patterns and " +
					"maintains state information longer than for regular UDP traffic. Higher values improve reliability for streaming " +
					"applications but consume more connection tracking resources.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},

			// WAN Settings
			"unbind_wan_monitors": schema.BoolAttribute{
				MarkdownDescription: "Unbind WAN monitors to prevent unnecessary traffic. When enabled, the gateway will stop certain monitoring processes " +
					"that periodically check WAN connectivity. This can reduce unnecessary traffic on metered connections or in environments " +
					"where the monitoring traffic might trigger security alerts. However, disabling these monitors may affect the gateway's " +
					"ability to detect and respond to WAN connectivity issues.",
				Optional: true,
				Computed: true,
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
	resp.Diagnostics.Append(r.RequireMaxVersionForPath("9.0", path.Root("unbind_wan_monitors"), req.Config)...)
}

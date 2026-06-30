package network

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/utils"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"golang.org/x/crypto/curve25519"
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

	// Anchored alternation: the surrounding ^(...)$ is mandatory. Without it, RE2 parses
	// "none|pd|static|single_network" as (^none)|(pd)|(static)|(single_network$), leaving the
	// middle branches as unanchored substring matches (e.g. "xstaticy" would pass). The grouped,
	// fully anchored form rejects anything that is not exactly one of the four accepted values.
	ipV6InterfaceTypeRegexp   = regexp.MustCompile("^(none|pd|static|single_network)$")
	validateIpV6InterfaceType = validation.StringMatch(ipV6InterfaceTypeRegexp, "invalid IPv6 interface type")

	// This is a slightly larger range than the UI, it includes some reserved ones, so could be tightened up.
	validateVLANID = validation.IntBetween(0, 4096)

	ipV6RAPriorityRegexp   = regexp.MustCompile("high|medium|low")
	validateIpV6RAPriority = validation.StringMatch(ipV6RAPriorityRegexp, "invalid IPv6 RA priority")
)

func ResourceNetwork() *schema.Resource {
	return &schema.Resource{
		Description: "The `unifi_network` resource manages networks in your UniFi environment, including WAN, LAN, and VLAN networks. " +
			"This resource enables you to:\n\n" +
			"* Create and manage different types of networks (corporate, guest, WAN, VLAN-only)\n" +
			"* Configure network addressing and DHCP settings\n" +
			"* Set up IPv6 networking features\n" +
			"* Manage DHCP relay and DNS settings\n" +
			"* Configure network groups and VLANs\n\n" +
			"Common use cases include:\n" +
			"* Setting up corporate and guest networks with different security policies\n" +
			"* Configuring WAN connectivity with various authentication methods\n" +
			"* Creating VLANs for network segmentation\n" +
			"* Managing DHCP and DNS services for network clients",

		CreateContext: resourceNetworkCreate,
		ReadContext:   resourceNetworkRead,
		UpdateContext: resourceNetworkUpdate,
		DeleteContext: resourceNetworkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: importNetwork,
		},

		// Cross-field validation the per-attribute schema can't express: a
		// vpn-client requires its companion fields and rejects wireguard_* on other
		// purposes, DHCP Guarding requires at least one trusted server, and the
		// default-gateway override must keep its enable toggle and gateway IP
		// consistent. Catches misconfigurations at plan time instead of as an opaque
		// controller 400.
		CustomizeDiff: customdiff.All(
			customizeNetworkVPNClient,
			customizeNetworkDHCPGuarding,
			customizeNetworkDefaultGateway,
		),

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the network.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"site": {
				Description: "The name of the site to associate the network with.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The name of the network. This should be a descriptive name that helps identify the network's purpose, " +
					"such as 'Corporate-Main', 'Guest-Network', or 'IoT-VLAN'.",
				Type:     schema.TypeString,
				Required: true,
			},
			"purpose": {
				Description: "The purpose/type of the network. Must be one of:\n" +
					"* `corporate` - Standard network for corporate use with full access\n" +
					"* `guest` - Isolated network for guest access with limited permissions\n" +
					"* `wan` - External network connection (WAN uplink)\n" +
					"* `vlan-only` - VLAN network without DHCP services\n" +
					"* `vpn-client` - Site-to-site VPN client connection (see the `vpn_type` and " +
					"`wireguard_client_*` arguments to configure a WireGuard VPN client)",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"corporate", "guest", "wan", "vlan-only", "vpn-client"}, false),
			},
			"vlan_id": {
				Description: "The VLAN ID for this network. Valid range is 0-4096. Common uses:\n" +
					"* 1-4094: Standard VLAN range for network segmentation\n" +
					"* 0: Untagged/native VLAN\n" +
					"* >4094: Reserved for special purposes",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validateVLANID,
			},
			"subnet": {
				Description: "The IPv4 subnet for this network in CIDR notation (e.g., '192.168.1.0/24'). " +
					"This defines the network's address space and determines the range of IP addresses available for DHCP.",
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: utils.CidrDiffSuppress,
				ValidateFunc:     utils.CidrValidate,
			},
			"network_group": {
				Description: "The network group for this network. Default is 'LAN'. For WAN networks, use 'WAN' or 'WAN2'. " +
					"Network groups help organize and apply policies to multiple networks.",
				Type:     schema.TypeString,
				Optional: true,
				Default:  "LAN",
			},
			"firewall_zone_id": {
				Description: "The ID of the Zone-Based Firewall (ZBF) zone this network belongs to. This is only " +
					"meaningful on UniFi OS 9.x controllers with Zone-Based Firewall enabled. The zone ID is " +
					"**site-scoped**: an ID from a different site is rejected or silently dropped by the controller.\n\n" +
					"This attribute is `Optional` + `Computed`:\n" +
					"* Leave it **unset** to preserve whatever zone the controller (or a `unifi_firewall_zone` " +
					"resource) has assigned. The provider never sends the field when it is not configured, so it " +
					"cannot clobber a zone managed elsewhere.\n" +
					"* **Set** it to explicitly pin or move this network to a specific zone — choose the zone " +
					"appropriate for the network's purpose (e.g. Internal, External, Guest).\n\n" +
					"On read the controller-assigned zone is always populated, so drift is detectable and " +
					"`terraform import` round-trips cleanly. Note the standard `Optional`+`Computed` \"sticky " +
					"value\" semantics: once set and later removed from configuration the value persists in state " +
					"rather than reverting, and removing it does **not** un-zone the network.\n\n" +
					"To manage zone membership from the zone side instead, use `unifi_firewall_zone.networks`. " +
					"Do not manage the same network-to-zone association from both sides.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"dhcp_start": {
				Description: "The starting IPv4 address of the DHCP range. Examples:\n" +
					"* For subnet 192.168.1.0/24, typical start: '192.168.1.100'\n" +
					"* For subnet 10.0.0.0/24, typical start: '10.0.0.100'\n" +
					"Ensure this address is within the network's subnet.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"dhcp_stop": {
				Description: "The ending IPv4 address of the DHCP range. Examples:\n" +
					"* For subnet 192.168.1.0/24, typical stop: '192.168.1.254'\n" +
					"* For subnet 10.0.0.0/24, typical stop: '10.0.0.254'\n" +
					"Must be greater than dhcp_start and within the network's subnet.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"dhcp_enabled": {
				Description: "Controls whether DHCP server is enabled for this network. When enabled:\n" +
					"* The network will automatically assign IP addresses to clients\n" +
					"* DHCP options (DNS, lease time) will be provided to clients\n" +
					"* Static IP assignments can still be made outside the DHCP range",
				Type:     schema.TypeBool,
				Optional: true,
			},
			"dhcp_lease": {
				Description: "The DHCP lease time in seconds. Common values:\n" +
					"* 86400 (1 day) - Default, suitable for most networks\n" +
					"* 3600 (1 hour) - For testing or temporary networks\n" +
					"* 604800 (1 week) - For stable networks with static clients\n" +
					"* 2592000 (30 days) - For very stable networks",
				Type:     schema.TypeInt,
				Optional: true,
				Default:  86400,
			},
			"dhcp_dns": {
				Description: "List of IPv4 DNS server addresses to be provided to DHCP clients. Examples:\n" +
					"* Use ['8.8.8.8', '8.8.4.4'] for Google DNS\n" +
					"* Use ['1.1.1.1', '1.0.0.1'] for Cloudflare DNS\n" +
					"* Use internal DNS servers for corporate networks\n" +
					"Maximum 4 servers can be specified.",
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 4,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.IsIPv4Address,
						validation.StringLenBetween(1, 50),
					),
				},
			},
			"dhcpd_boot_enabled": {
				Description: "Enables DHCP boot options for PXE boot or network boot configurations. When enabled:\n" +
					"* Allows network devices to boot from a TFTP server\n" +
					"* Requires dhcpd_boot_server and dhcpd_boot_filename to be set\n" +
					"* Commonly used for diskless workstations or network installations",
				Type:     schema.TypeBool,
				Optional: true,
			},
			"dhcpd_boot_server": {
				Description: "The IPv4 address of the TFTP server for network boot. This setting:\n" +
					"* Is required when dhcpd_boot_enabled is true\n" +
					"* Should be a reliable, always-on server\n" +
					"* Must be accessible to all clients that need to boot",
				Type: schema.TypeString,
				// TODO: IPv4 validation?
				Optional: true,
			},
			"dhcpd_boot_filename": {
				Description: "The boot filename to be loaded from the TFTP server. Examples:\n" +
					"* 'pxelinux.0' - Standard PXE boot loader\n" +
					"* 'undionly.kpxe' - iPXE boot loader\n" +
					"* Custom paths for specific boot images",
				Type:     schema.TypeString,
				Optional: true,
			},
			"dhcpd_gateway_enabled": {
				Description: "Controls whether the default gateway advertised to this network's DHCP " +
					"clients is selected automatically or set manually — equivalent to switching the " +
					"network's default gateway from automatic to a manually specified address in the " +
					"UniFi UI (the exact control label and location vary across controller versions). " +
					"When `false` (automatic, the default) the controller advertises the network's own " +
					"interface IP as the gateway via DHCP option 3. Set this to `true` to advertise the " +
					"address in `dhcpd_gateway` instead — useful for pointing clients at a custom next " +
					"hop such as a VPN/subnet-router node (e.g. Tailscale).\n\n" +
					"This attribute is `Optional` and `Computed`: when omitted from configuration it " +
					"inherits the current value reported by the controller (so a value set in the UI " +
					"is preserved) rather than being reset. When `true`, `dhcpd_gateway` is required.\n\n" +
					"Only meaningful when this network runs the UniFi DHCP server (`dhcp_enabled = true` " +
					"and `dhcp_relay_enabled = false`); it has no effect on `wan` or `vlan-only` networks. " +
					"Note: on some controller versions the network must also be in manual configuration " +
					"mode (toggled in the UniFi UI) before a manually-specified gateway is honored.",
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"dhcpd_gateway": {
				Description: "The IPv4 default gateway to advertise to this network's DHCP clients (DHCP " +
					"option 3) when `dhcpd_gateway_enabled` is `true`. Typically an address inside this " +
					"network's `subnet`; an off-subnet address (e.g. a 100.64.0.0/10 Tailscale CGNAT " +
					"address) passes validation here but may be rejected by the controller at apply. " +
					"IPv4 only — there is no IPv6 default-gateway override.\n\n" +
					"This attribute is `Optional` and `Computed`: when omitted it inherits the current " +
					"value reported by the controller (so a manually-set gateway, or a value the " +
					"controller echoes in auto mode, does not show as drift). Set it together with " +
					"`dhcpd_gateway_enabled = true` to manage the override from Terraform.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"dhcp_relay_enabled": {
				Description: "Enables DHCP relay for this network. When enabled:\n" +
					"* DHCP requests are forwarded to an external DHCP server\n" +
					"* Local DHCP server is disabled\n" +
					"* Useful for centralized DHCP management",
				Type:     schema.TypeBool,
				Optional: true,
			},
			"dhcp_v6_dns": {
				Description: "List of IPv6 DNS server addresses for DHCPv6 clients. Examples:\n" +
					"* Use ['2001:4860:4860::8888', '2001:4860:4860::8844'] for Google DNS\n" +
					"* Use ['2606:4700:4700::1111', '2606:4700:4700::1001'] for Cloudflare DNS\n" +
					"Only used when dhcp_v6_dns_auto is false. Maximum of 4 addresses are allowed.",
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 4,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPv6Address,
					// TODO: should this ensure blank can't get through?
				},
			},
			"dhcp_v6_dns_auto": {
				Description: "Controls DNS server source for DHCPv6 clients:\n" +
					"* true - Use upstream DNS servers (recommended)\n" +
					"* false - Use manually specified servers from dhcp_v6_dns\n" +
					"Default is true for easier management.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"dhcp_v6_enabled": {
				Description: "Enables stateful DHCPv6 for IPv6 address assignment. When enabled:\n" +
					"* Provides IPv6 addresses to clients\n" +
					"* Works alongside SLAAC if configured\n" +
					"* Allows for more controlled IPv6 addressing",
				Type:     schema.TypeBool,
				Optional: true,
			},
			"dhcp_v6_lease": {
				Description: "The DHCPv6 lease time in seconds. Common values:\n" +
					"* 86400 (1 day) - Default setting\n" +
					"* 3600 (1 hour) - For testing\n" +
					"* 604800 (1 week) - For stable networks\n" +
					"Typically longer than IPv4 DHCP leases.",
				Type:     schema.TypeInt,
				Optional: true,
				Default:  86400,
			},
			"dhcp_v6_start": {
				Description: "The starting IPv6 address for the DHCPv6 range. Used in static DHCPv6 configuration.\n" +
					"Must be a valid IPv6 address within your allocated IPv6 subnet.\n\n" +
					"This attribute is `Optional` + `Computed`: when omitted from configuration it inherits the " +
					"current value reported by the controller, so a value configured in the UI (or read in via " +
					"`terraform import`) is preserved rather than planned for removal. Note the standard " +
					"`Optional`+`Computed` \"sticky value\" semantics — once the controller has a value, removing " +
					"the attribute from configuration leaves that value in place rather than clearing it (the " +
					"provider serializes this field with `omitempty`, so an empty value is never sent). There is " +
					"therefore no way to clear it by deleting it from configuration; turn DHCPv6 off with " +
					"`dhcp_v6_enabled`/`ipv6_interface_type` instead. Set it explicitly to manage the value from Terraform.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IsIPv6Address,
			},
			"dhcp_v6_stop": {
				Description: "The ending IPv6 address for the DHCPv6 range. Used in static DHCPv6 configuration.\n" +
					"Must be after dhcp_v6_start in the IPv6 address space.\n\n" +
					"This attribute is `Optional` + `Computed`: when omitted from configuration it inherits the " +
					"current value reported by the controller, so a value configured in the UI (or read in via " +
					"`terraform import`) is preserved rather than planned for removal. Note the standard " +
					"`Optional`+`Computed` \"sticky value\" semantics — once the controller has a value, removing " +
					"the attribute from configuration leaves that value in place rather than clearing it (the " +
					"provider serializes this field with `omitempty`, so an empty value is never sent). There is " +
					"therefore no way to clear it by deleting it from configuration; turn DHCPv6 off with " +
					"`dhcp_v6_enabled`/`ipv6_interface_type` instead. Set it explicitly to manage the value from Terraform.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IsIPv6Address,
			},
			"domain_name": {
				Description: "The domain name for this network. Examples:\n" +
					"* 'corp.example.com' - For corporate networks\n" +
					"* 'guest.example.com' - For guest networks\n" +
					"* 'iot.example.com' - For IoT networks\n" +
					"Used for internal DNS resolution and DHCP options.",
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Description: "Controls whether this network is active. When disabled:\n" +
					"* Network will not be available to clients\n" +
					"* DHCP services will be stopped\n" +
					"* Existing clients will be disconnected\n" +
					"Useful for temporary network maintenance or security measures.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"igmp_snooping": {
				Description: "Enables IGMP (Internet Group Management Protocol) snooping. When enabled:\n" +
					"* Optimizes multicast traffic flow\n" +
					"* Reduces network congestion\n" +
					"* Improves performance for multicast applications (e.g., IPTV)\n" +
					"Recommended for networks with multicast traffic.",
				Type:     schema.TypeBool,
				Optional: true,
			},
			"dhcp_guarding": {
				Description: "Enables DHCP Guarding for this network, blocking DHCP server responses from " +
					"untrusted/rogue sources so only the trusted DHCP server can hand out leases. When enabled:\n" +
					"* Drops DHCP offers/acknowledgements from servers other than the trusted one\n" +
					"* Protects clients from rogue or misconfigured DHCP servers\n\n" +
					"This attribute is `Optional` and `Computed`: when omitted from configuration it inherits the " +
					"current value reported by the controller (so a value enabled in the UI is preserved), rather than " +
					"being reset. Set it explicitly to manage the value from Terraform.",
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"dhcp_guarding_trusted_servers": {
				Description: "List of trusted DHCP server IPv4 addresses for DHCP Guarding. When `dhcp_guarding` " +
					"is enabled the controller drops DHCP offers from every server except those listed here, so at " +
					"least one address is required whenever guarding is on (for a network served by the UniFi gateway's " +
					"own DHCP server this is typically the network's gateway IP). Maximum 3 servers can be specified.\n\n" +
					"Like `dhcp_guarding`, this attribute is `Optional` and `Computed`: when omitted it inherits the " +
					"current value reported by the controller (so a list configured in the UI is preserved rather than " +
					"cleared). Set it explicitly to manage the trusted servers from Terraform.",
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 3,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.IsIPv4Address,
						validation.StringLenBetween(1, 50),
					),
				},
			},
			"upnp_lan_enabled": {
				Description: "Whether clients on THIS network are allowed to request UPnP/NAT-PMP port mappings. " +
					"Per-network opt-in that complements the gateway-global UPnP toggle " +
					"(`unifi_setting_usg.upnp_enabled`): UPnP must be enabled globally AND on a given network for that " +
					"network's devices to self-map WAN ports. Leave false on untrusted networks (IoT, Guest, …) so a " +
					"compromised device cannot open inbound holes in the firewall; enable only on networks whose devices " +
					"you trust to manage their own port mappings.",
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"ipv6_interface_type": {
				Description: "Specifies the IPv6 connection type. Must be one of:\n" +
					"* `none` - IPv6 disabled (default)\n" +
					"* `static` - Static IPv6 addressing\n" +
					"* `pd` - Prefix Delegation from upstream\n" +
					"* `single_network` - Share a delegated IPv6 prefix with a single LAN\n\n" +
					"Choose based on your IPv6 deployment strategy and ISP capabilities. " +
					"Note: `single_network` has companion controller settings (the single-network " +
					"interface/LAN binding) that this provider does not yet expose, so a bare " +
					"`single_network` network may not be fully configurable.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "none",
				ValidateFunc: validateIpV6InterfaceType,
			},
			"ipv6_static_subnet": {
				Description: "The static IPv6 subnet in CIDR notation (e.g., '2001:db8::/64') when using static IPv6.\n" +
					"Only applicable when `ipv6_interface_type` is 'static'.\n" +
					"Must be a valid IPv6 subnet allocated to your organization.\n\n" +
					"This attribute is `Optional` + `Computed`: when omitted from configuration it inherits the " +
					"current value reported by the controller, so a value configured in the UI (or read in via " +
					"`terraform import`) is preserved rather than planned for removal. Note the standard " +
					"`Optional`+`Computed` \"sticky value\" semantics — once the controller has a value, removing " +
					"the attribute from configuration leaves that value in place rather than clearing it (the " +
					"provider serializes this field with `omitempty`, so an empty value is never sent). There is " +
					"therefore no way to clear it by deleting it from configuration; switch `ipv6_interface_type` " +
					"away from 'static' to disable static IPv6 instead. Set it explicitly to manage the value from Terraform.",
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ipv6_pd_interface": {
				Description: "The WAN interface to use for IPv6 Prefix Delegation. Options:\n" +
					"* `wan` - Primary WAN interface\n" +
					"* `wan2` - Secondary WAN interface\n" +
					"Only applicable when `ipv6_interface_type` is 'pd'.\n\n" +
					"This attribute is `Optional` + `Computed`: when omitted from configuration it inherits the " +
					"current value reported by the controller, so a value configured in the UI (or read in via " +
					"`terraform import`) is preserved rather than planned for removal. Note the standard " +
					"`Optional`+`Computed` \"sticky value\" semantics — once the controller has a value, removing " +
					"the attribute from configuration leaves that value in place rather than clearing it (the " +
					"provider serializes this field with `omitempty`, so an empty value is never sent). There is " +
					"therefore no way to clear it by deleting it from configuration; switch `ipv6_interface_type` " +
					"away from 'pd' to disable Prefix Delegation instead. Set it explicitly to manage the value from Terraform.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateWANV6NetworkGroup,
			},
			"ipv6_pd_prefixid": {
				Description: "The IPv6 Prefix ID for Prefix Delegation. Used to:\n" +
					"* Differentiate multiple delegated prefixes\n" +
					"* Create unique subnets from the delegated prefix\n" +
					"Typically a hexadecimal value (e.g., '0', '1', 'a1').",
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipv6_pd_start": {
				Description: "The starting IPv6 address for Prefix Delegation range.\n" +
					"Only used when `ipv6_interface_type` is 'pd'.\n" +
					"Must be within the delegated prefix range.\n\n" +
					"This attribute is `Optional` + `Computed`: when omitted from configuration it inherits the " +
					"current value reported by the controller, so a value configured in the UI (or read in via " +
					"`terraform import`) is preserved rather than planned for removal. Note the standard " +
					"`Optional`+`Computed` \"sticky value\" semantics — once the controller has a value, removing " +
					"the attribute from configuration leaves that value in place rather than clearing it (the " +
					"provider serializes this field with `omitempty`, so an empty value is never sent). There is " +
					"therefore no way to clear it by deleting it from configuration; switch `ipv6_interface_type` " +
					"away from 'pd' to disable Prefix Delegation instead. Set it explicitly to manage the value from Terraform.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IsIPv6Address,
			},
			"ipv6_pd_stop": {
				Description: "The ending IPv6 address for Prefix Delegation range.\n" +
					"Only used when `ipv6_interface_type` is 'pd'.\n" +
					"Must be after `ipv6_pd_start` within the delegated prefix.\n\n" +
					"This attribute is `Optional` + `Computed`: when omitted from configuration it inherits the " +
					"current value reported by the controller, so a value configured in the UI (or read in via " +
					"`terraform import`) is preserved rather than planned for removal. Note the standard " +
					"`Optional`+`Computed` \"sticky value\" semantics — once the controller has a value, removing " +
					"the attribute from configuration leaves that value in place rather than clearing it (the " +
					"provider serializes this field with `omitempty`, so an empty value is never sent). There is " +
					"therefore no way to clear it by deleting it from configuration; switch `ipv6_interface_type` " +
					"away from 'pd' to disable Prefix Delegation instead. Set it explicitly to manage the value from Terraform.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IsIPv6Address,
			},
			"ipv6_ra_enable": {
				Description: "Enables IPv6 Router Advertisements (RA). When enabled:\n" +
					"* Announces IPv6 prefix information to clients\n" +
					"* Enables SLAAC address configuration\n" +
					"* Required for most IPv6 deployments",
				Type:     schema.TypeBool,
				Optional: true,
			},
			"internet_access_enabled": {
				Description: "Controls internet access for this network. When disabled:\n" +
					"* Clients cannot access external networks\n" +
					"* Internal network access remains available\n" +
					"* Useful for creating isolated or secure networks",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"network_isolation_enabled": {
				Description: "Isolates this network from other local networks/VLANs on the site. When enabled:\n" +
					"* Hosts on this network cannot route to or from other local networks on the site\n" +
					"* Gateway and internet access are retained (internet access is subject to `internet_access_enabled`)\n" +
					"* This is a routing/firewall option for network-to-network isolation, distinct from per-client (WLAN) isolation",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"ipv6_ra_preferred_lifetime": {
				Description: "The preferred lifetime (in seconds) for IPv6 addresses in Router Advertisements.\n" +
					"* Must be less than or equal to `ipv6_ra_valid_lifetime`\n" +
					"* Default: 14400 (4 hours)\n" +
					"* After this time, addresses become deprecated but still usable",
				Type:     schema.TypeInt,
				Optional: true,
				Default:  14400,
			},
			"ipv6_ra_priority": {
				Description: "Sets the priority for IPv6 Router Advertisements. Options:\n" +
					"* `high` - Preferred for primary networks\n" +
					"* `medium` - Standard priority\n" +
					"* `low` - For backup or secondary networks\n" +
					"Affects router selection when multiple IPv6 routers exist.\n\n" +
					"This attribute is `Optional` + `Computed`: when omitted from configuration it inherits the " +
					"current value reported by the controller, so a value configured in the UI (or read in via " +
					"`terraform import`) is preserved rather than planned for removal. Note the standard " +
					"`Optional`+`Computed` \"sticky value\" semantics — once the controller has a value, removing " +
					"the attribute from configuration leaves that value in place rather than clearing it (the " +
					"provider serializes this field with `omitempty`, so an empty value is never sent). There is " +
					"therefore no way to clear it by deleting it from configuration; turn Router Advertisements " +
					"off with `ipv6_ra_enable` instead. Set it explicitly to manage the value from Terraform.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateIpV6RAPriority,
			},
			"ipv6_ra_valid_lifetime": {
				Description: "The valid lifetime (in seconds) for IPv6 addresses in Router Advertisements.\n" +
					"* Must be greater than or equal to `ipv6_ra_preferred_lifetime`\n" +
					"* Default: 86400 (24 hours)\n" +
					"* After this time, addresses become invalid",
				Type:     schema.TypeInt,
				Optional: true,
				Default:  86400,
			},
			"multicast_dns": {
				Description: "Enables Multicast DNS (mDNS/Bonjour/Avahi) on the network. When enabled:\n" +
					"* Allows device discovery (e.g., printers, Chromecasts)\n" +
					"* Supports zero-configuration networking\n" +
					"* Available on Controller version 7 and later",
				Type:     schema.TypeBool,
				Optional: true,
			},
			"wan_ip": {
				Description: "The static IPv4 address for WAN interface.\n" +
					"Required when `wan_type` is 'static'.\n" +
					"Must be a valid public IP address assigned by your ISP.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"wan_netmask": {
				Description: "The IPv4 netmask for WAN interface (e.g., '255.255.255.0').\n" +
					"Required when `wan_type` is 'static'.\n" +
					"Must match the subnet mask provided by your ISP.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"wan_gateway": {
				Description: "The IPv4 gateway address for WAN interface.\n" +
					"Required when `wan_type` is 'static'.\n" +
					"Typically the ISP's router IP address.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"wan_dns": {
				Description: "List of IPv4 DNS servers for WAN interface. Examples:\n" +
					"* ISP provided DNS servers\n" +
					"* Public DNS services (e.g., 8.8.8.8, 1.1.1.1)\n" +
					"* Maximum 4 servers can be specified",
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 4,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPv4Address,
				},
			},
			"wan_type": {
				Description: "The IPv4 WAN connection type. Options:\n" +
					"* `disabled` - WAN interface disabled\n" +
					"* `static` - Static IP configuration\n" +
					"* `dhcp` - Dynamic IP from ISP\n" +
					"* `pppoe` - PPPoE connection (common for DSL)\n" +
					"Choose based on your ISP's requirements.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateWANType,
			},
			"wan_networkgroup": {
				Description: "The WAN interface group assignment. Options:\n" +
					"* `WAN` - Primary WAN interface\n" +
					"* `WAN2` - Secondary WAN interface\n" +
					"* `WAN_LTE_FAILOVER` - LTE backup connection\n" +
					"Used for dual WAN and failover configurations.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateWANNetworkGroup,
			},
			"wan_egress_qos": {
				Description: "Quality of Service (QoS) priority for WAN egress traffic (0-7).\n" +
					"* 0 (default) - Best effort\n" +
					"* 1-4 - Increasing priority\n" +
					"* 5-7 - Highest priority, use sparingly\n" +
					"Higher values get preferential treatment.",
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"wan_username": {
				Description: "Username for WAN authentication.\n" +
					"* Required for PPPoE connections\n" +
					"* May be needed for some ISP configurations\n" +
					"* Cannot contain spaces or special characters",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateWANUsername,
			},
			"x_wan_password": {
				Description: "Password for WAN authentication.\n" +
					"* Required for PPPoE connections\n" +
					"* May be needed for some ISP configurations\n" +
					"* Must be kept secret",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateWANPassword,
			},
			"wan_type_v6": {
				Description: "The IPv6 WAN connection type. Options:\n" +
					"* `disabled` - IPv6 disabled\n" +
					"* `static` - Static IPv6 configuration\n" +
					"* `dhcpv6` - Dynamic IPv6 from ISP\n" +
					"Choose based on your ISP's requirements.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateWANTypeV6,
			},
			"wan_dhcp_v6_pd_size": {
				Description: "The IPv6 prefix size to request from ISP. Must be between 48 and 64.\n" +
					"Only applicable when `wan_type_v6` is 'dhcpv6'.",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(48, 64),
			},
			"wan_ipv6": {
				Description: "The static IPv6 address for WAN interface.\n" +
					"Required when `wan_type_v6` is 'static'.\n" +
					"Must be a valid public IPv6 address assigned by your ISP.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPv6Address,
			},
			"wan_gateway_v6": {
				Description: "The IPv6 gateway address for WAN interface.\n" +
					"Required when `wan_type_v6` is 'static'.\n" +
					"Typically the ISP's router IPv6 address.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPv6Address,
			},
			"wan_prefixlen": {
				Description: "The IPv6 prefix length for WAN interface. Must be between 1 and 128.\n" +
					"Only applicable when `wan_type_v6` is 'static'.",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 128),
			},
			"vpn_type": {
				Description: "The VPN type for a `vpn-client` network. Currently `wireguard-client` is supported, " +
					"which connects the gateway to a remote WireGuard server. Only applicable when `purpose` is " +
					"'vpn-client'. A `wireguard-client` network also requires `subnet` (the tunnel interface address, " +
					"e.g. `10.0.0.2/32`) and `dhcp_dns` (interface DNS); the controller rejects the create without them.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"wireguard-client"}, false),
			},
			"wireguard_interface": {
				Description: "The WAN interface the WireGuard tunnel egresses from. One of `wan` or `wan2`. " +
					"Only applicable when `vpn_type` is 'wireguard-client'.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"wan", "wan2"}, false),
			},
			"wireguard_client_mode": {
				Description: "How the WireGuard VPN client peer is configured. Currently only `manual` is supported, " +
					"configuring the peer with the individual `wireguard_client_*` arguments. " +
					"Only applicable when `vpn_type` is 'wireguard-client'.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"manual"}, false),
			},
			"wireguard_client_peer_ip": {
				Description: "The remote WireGuard server's endpoint host or IP address that the gateway dials. " +
					"Only applicable when `vpn_type` is 'wireguard-client'.",
				Type:     schema.TypeString,
				Optional: true,
			},
			"wireguard_client_peer_port": {
				Description: "The remote WireGuard server's listen port (e.g. 51820). " +
					"Only applicable when `vpn_type` is 'wireguard-client'.",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IsPortNumber,
			},
			"wireguard_client_peer_public_key": {
				Description: "The remote WireGuard server's public key (the peer the gateway connects to). " +
					"Only applicable when `vpn_type` is 'wireguard-client'.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: utils.WireguardKeyValidate,
			},
			"wireguard_client_preshared_key": {
				Description: "An optional WireGuard pre-shared key (PSK) for an additional layer of symmetric-key " +
					"security with the peer. Keep this value secret. The controller may not return this value on read, " +
					"so it is computed to avoid spurious drift. Only applicable when `vpn_type` is 'wireguard-client'.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Sensitive:    true,
				ValidateFunc: utils.WireguardKeyValidate,
			},
			"wireguard_client_preshared_key_enabled": {
				Description: "Whether a WireGuard pre-shared key is used with the peer. " +
					"Only applicable when `vpn_type` is 'wireguard-client'.",
				Type:     schema.TypeBool,
				Optional: true,
			},
			"x_wireguard_private_key": {
				Description: "The gateway's own WireGuard private key for this VPN client. If omitted, a key pair is " +
					"generated for you and the public key is exposed via `wireguard_public_key`. Keep this value secret. " +
					"Only applicable when `vpn_type` is 'wireguard-client'.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Sensitive:    true,
				ValidateFunc: utils.WireguardKeyValidate,
			},
			"wireguard_public_key": {
				Description: "The gateway's own WireGuard public key for this VPN client. The controller does not " +
					"return it, so the provider derives it from the private key (Curve25519). Add this key as a peer " +
					"on the remote WireGuard server. Only set when `vpn_type` is 'wireguard-client'.",
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpn_client_default_route": {
				Description: "When true, route all of the gateway's internet traffic through the VPN client tunnel. " +
					"When false (default), only the destinations in `uid_vpn_custom_routing` are routed through the tunnel. " +
					"Only applicable when `purpose` is 'vpn-client'.",
				Type:     schema.TypeBool,
				Optional: true,
			},
			"vpn_client_pull_dns": {
				Description: "When true, use DNS servers advertised by the VPN peer for traffic on the tunnel. " +
					"Only applicable when `purpose` is 'vpn-client'.",
				Type:     schema.TypeBool,
				Optional: true,
			},
			"uid_vpn_custom_routing": {
				Description: "The list of destination subnets (CIDR notation) routed through the VPN client tunnel " +
					"when `vpn_client_default_route` is false. Values are canonicalized to their network address " +
					"(e.g. `10.0.0.1/16` becomes `10.0.0.0/16`). Only applicable when `purpose` is 'vpn-client'.",
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateFunc:     utils.CidrValidate,
					DiffSuppressFunc: utils.CidrDiffSuppress,
				},
			},
		},
	}
}

func resourceNetworkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	// Create-only: mint the gateway's WireGuard private key when the user omits it.
	// The controller requires it in the create payload (it generates none) and never
	// returns it on read, so generating it here rather than in the shared request
	// builder avoids silently rotating an imported network's key on a later update
	// (after import the key is empty in state; on update omitempty drops it and the
	// controller keeps the one it already has).
	if d.Get("vpn_type").(string) == "wireguard-client" && d.Get("x_wireguard_private_key").(string) == "" {
		key, err := generateWireguardPrivateKey()
		if err != nil {
			return diag.FromErr(fmt.Errorf("unable to generate WireGuard private key: %w", err))
		}
		if err := d.Set("x_wireguard_private_key", key); err != nil {
			return diag.FromErr(err)
		}
	}

	req, err := resourceNetworkGetResourceData(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.CreateNetwork(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return resourceNetworkSetResourceData(resp, d, site)
}

func resourceNetworkGetResourceData(d *schema.ResourceData, meta interface{}) (*unifi.Network, error) {
	// c := meta.(*provider.Client)

	vlan := d.Get("vlan_id").(int)
	dhcpDNS, err := utils.ListToStringSlice(d.Get("dhcp_dns").([]interface{}))
	if err != nil {
		return nil, fmt.Errorf("unable to convert dhcp_dns to string slice: %w", err)
	}
	dhcpGuardServers, err := utils.ListToStringSlice(d.Get("dhcp_guarding_trusted_servers").([]interface{}))
	if err != nil {
		return nil, fmt.Errorf("unable to convert dhcp_guarding_trusted_servers to string slice: %w", err)
	}
	dhcpV6DNS, err := utils.ListToStringSlice(d.Get("dhcp_v6_dns").([]interface{}))
	if err != nil {
		return nil, fmt.Errorf("unable to convert dhcp_v6_dns to string slice: %w", err)
	}
	wanDNS, err := utils.ListToStringSlice(d.Get("wan_dns").([]interface{}))
	if err != nil {
		return nil, fmt.Errorf("unable to convert wan_dns to string slice: %w", err)
	}
	uidVPNCustomRouting, err := utils.ListToStringSlice(d.Get("uid_vpn_custom_routing").([]interface{}))
	if err != nil {
		return nil, fmt.Errorf("unable to convert uid_vpn_custom_routing to string slice: %w", err)
	}
	// Canonicalize each route to its network address so config and controller-returned
	// values stay consistent (matches the diff suppression on the schema).
	uidVPNCustomRouting = utils.CidrListZeroBased(uidVPNCustomRouting)

	vpnType := d.Get("vpn_type").(string)

	// For a LAN the `subnet` is the gateway address, so CidrOneBased applies the +1
	// host offset. A vpn-client's subnet is the tunnel interface address; CidrZeroBased
	// canonicalizes it to its network address (host bits drop below /32, which the
	// CustomizeDiff rejects), so the /32 host address round-trips intact.
	subnet := d.Get("subnet").(string)
	ipSubnet := utils.CidrOneBased(subnet)
	if d.Get("purpose").(string) == "vpn-client" {
		ipSubnet = utils.CidrZeroBased(subnet)
	}

	n := &unifi.Network{
		Name:              d.Get("name").(string),
		Purpose:           d.Get("purpose").(string),
		VLAN:              vlan,
		IPSubnet:          ipSubnet,
		NetworkGroup:      d.Get("network_group").(string),
		DHCPDStart:        d.Get("dhcp_start").(string),
		DHCPDStop:         d.Get("dhcp_stop").(string),
		DHCPDEnabled:      d.Get("dhcp_enabled").(bool),
		DHCPDLeaseTime:    d.Get("dhcp_lease").(int),
		DHCPDBootEnabled:  d.Get("dhcpd_boot_enabled").(bool),
		DHCPDBootServer:   d.Get("dhcpd_boot_server").(string),
		DHCPDBootFilename: d.Get("dhcpd_boot_filename").(string),
		// DHCP default-gateway override (UI "Default Gateway" Auto/Manual). Both
		// fields lack omitempty so they serialize on every PUT; Optional+Computed
		// means an omitted attribute re-sends the controller's read-back value
		// rather than clobbering a UI-set gateway.
		DHCPDGatewayEnabled: d.Get("dhcpd_gateway_enabled").(bool),
		DHCPDGateway:        d.Get("dhcpd_gateway").(string),
		DHCPRelayEnabled:    d.Get("dhcp_relay_enabled").(bool),
		DHCPguardEnabled:    d.Get("dhcp_guarding").(bool),
		// Trusted DHCP servers for DHCP Guarding. Same hackish positional fan-out as
		// DHCPDDNS{x}; an empty list maps to "" entries. ¯\_(ツ)_/¯
		DHCPDIP1:       append(dhcpGuardServers, "")[0],
		DHCPDIP2:       append(dhcpGuardServers, "", "")[1],
		DHCPDIP3:       append(dhcpGuardServers, "", "", "")[2],
		DomainName:     d.Get("domain_name").(string),
		IGMPSnooping:   d.Get("igmp_snooping").(bool),
		UpnpLanEnabled: d.Get("upnp_lan_enabled").(bool),
		MdnsEnabled:    d.Get("multicast_dns").(bool),
		Enabled:        d.Get("enabled").(bool),

		DHCPDDNSEnabled: len(dhcpDNS) > 0,
		// this is kinda hacky but ¯\_(ツ)_/¯
		DHCPDDNS1: append(dhcpDNS, "")[0],
		DHCPDDNS2: append(dhcpDNS, "", "")[1],
		DHCPDDNS3: append(dhcpDNS, "", "", "")[2],
		DHCPDDNS4: append(dhcpDNS, "", "", "", "")[3],

		VLANEnabled: vlan != 0 && vlan != 1,

		// Same hackish code as for DHCPv4 ¯\_(ツ)_/¯
		DHCPDV6DNS1: append(dhcpV6DNS, "")[0],
		DHCPDV6DNS2: append(dhcpV6DNS, "", "")[1],
		DHCPDV6DNS3: append(dhcpV6DNS, "", "", "")[2],
		DHCPDV6DNS4: append(dhcpV6DNS, "", "", "", "")[3],

		DHCPDV6DNSAuto:   d.Get("dhcp_v6_dns_auto").(bool),
		DHCPDV6Enabled:   d.Get("dhcp_v6_enabled").(bool),
		DHCPDV6LeaseTime: d.Get("dhcp_v6_lease").(int),
		DHCPDV6Start:     d.Get("dhcp_v6_start").(string),
		DHCPDV6Stop:      d.Get("dhcp_v6_stop").(string),

		IPV6InterfaceType:       d.Get("ipv6_interface_type").(string),
		IPV6Subnet:              d.Get("ipv6_static_subnet").(string),
		IPV6PDInterface:         d.Get("ipv6_pd_interface").(string),
		IPV6PDPrefixid:          d.Get("ipv6_pd_prefixid").(string),
		IPV6PDStart:             d.Get("ipv6_pd_start").(string),
		IPV6PDStop:              d.Get("ipv6_pd_stop").(string),
		IPV6RaEnabled:           d.Get("ipv6_ra_enable").(bool),
		IPV6RaPreferredLifetime: d.Get("ipv6_ra_preferred_lifetime").(int),
		IPV6RaPriority:          d.Get("ipv6_ra_priority").(string),
		IPV6RaValidLifetime:     d.Get("ipv6_ra_valid_lifetime").(int),

		InternetAccessEnabled:   d.Get("internet_access_enabled").(bool),
		NetworkIsolationEnabled: d.Get("network_isolation_enabled").(bool),

		WANIP:           d.Get("wan_ip").(string),
		WANType:         d.Get("wan_type").(string),
		WANNetmask:      d.Get("wan_netmask").(string),
		WANGateway:      d.Get("wan_gateway").(string),
		WANNetworkGroup: d.Get("wan_networkgroup").(string),
		WANEgressQOS:    d.Get("wan_egress_qos").(int),
		WANUsername:     d.Get("wan_username").(string),
		XWANPassword:    d.Get("x_wan_password").(string),

		WANTypeV6:       d.Get("wan_type_v6").(string),
		WANDHCPv6PDSize: d.Get("wan_dhcp_v6_pd_size").(int),
		WANIPV6:         d.Get("wan_ipv6").(string),
		WANGatewayV6:    d.Get("wan_gateway_v6").(string),
		WANPrefixlen:    d.Get("wan_prefixlen").(int),

		// this is kinda hacky but ¯\_(ツ)_/¯
		WANDNS1: append(wanDNS, "")[0],
		WANDNS2: append(wanDNS, "", "")[1],
		WANDNS3: append(wanDNS, "", "", "")[2],
		WANDNS4: append(wanDNS, "", "", "", "")[3],

		// WireGuard VPN client (purpose = "vpn-client", vpn_type = "wireguard-client").
		// wireguard_public_key is computed (the provider derives it on read), so it is not sent here.
		VPNType:                            vpnType,
		WireguardInterface:                 d.Get("wireguard_interface").(string),
		WireguardClientMode:                d.Get("wireguard_client_mode").(string),
		WireguardClientPeerIP:              d.Get("wireguard_client_peer_ip").(string),
		WireguardClientPeerPort:            d.Get("wireguard_client_peer_port").(int),
		WireguardClientPeerPublicKey:       d.Get("wireguard_client_peer_public_key").(string),
		WireguardClientPresharedKey:        d.Get("wireguard_client_preshared_key").(string),
		WireguardClientPresharedKeyEnabled: d.Get("wireguard_client_preshared_key_enabled").(bool),
		XWireguardPrivateKey:               d.Get("x_wireguard_private_key").(string),
		VPNClientDefaultRoute:              d.Get("vpn_client_default_route").(bool),
		VPNClientPullDNS:                   d.Get("vpn_client_pull_dns").(bool),
		UidVPNCustomRouting:                uidVPNCustomRouting,
	}

	// Zone-Based Firewall (UniFi OS 9.x) zone membership. Only send firewall_zone_id
	// when the user explicitly configured it. If it is omitted (null/unknown) leave it
	// empty so omitempty drops it from the payload — preserving today's behavior and not
	// clobbering a zone managed via unifi_firewall_zone.networks. Plain d.Get is
	// insufficient here: for an Optional+Computed string it returns the stale state
	// value when config is null, which would re-send (and fight) an externally-managed
	// zone. utils.IsRawConfigSet inspects d.GetRawConfig() and treats null and empty-string as
	// "not set" (the StringIsNotEmpty validator already rejects an explicit "").
	if raw := d.GetRawConfig(); utils.IsRawConfigSet(raw, "firewall_zone_id") {
		n.FirewallZoneID = d.Get("firewall_zone_id").(string)
	}

	return n, nil
}

// customizeNetworkVPNClient enforces the cross-field rules for vpn-client networks
// at plan time. Presence is read from GetRawConfig so a field supplied through
// interpolation (e.g. var.x, unknown at plan) counts as set instead of tripping a
// false "required" error.
func customizeNetworkVPNClient(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	raw := d.GetRawConfig()
	if raw.IsNull() {
		return nil // no config (e.g. on destroy) — nothing to validate
	}

	// Every field that only belongs on a vpn-client network.
	vpnFields := []string{
		"vpn_type", "wireguard_interface", "wireguard_client_mode",
		"wireguard_client_peer_ip", "wireguard_client_peer_public_key",
		"x_wireguard_private_key", "wireguard_client_preshared_key",
		"wireguard_client_peer_port", "uid_vpn_custom_routing",
	}

	if d.Get("purpose").(string) != "vpn-client" {
		for _, k := range vpnFields {
			if utils.IsRawConfigSet(raw, k) {
				return fmt.Errorf("%q is only valid when purpose = %q", k, "vpn-client")
			}
		}
		return nil
	}

	if d.Get("vpn_type").(string) != "wireguard-client" {
		return fmt.Errorf("%q is required when purpose = %q (only %q is supported)", "vpn_type", "vpn-client", "wireguard-client")
	}
	if !utils.IsRawConfigSet(raw, "subnet") {
		return fmt.Errorf("%q (the tunnel interface address, e.g. 10.0.0.2/32) is required for a wireguard-client network", "subnet")
	}
	// The tunnel address must be a /32 host address: CidrZeroBased zeroes the host
	// bits below /32, silently corrupting a shorter prefix. Skip when unknown.
	if subnet := d.Get("subnet").(string); subnet != "" {
		ip, ipNet, err := net.ParseCIDR(subnet)
		if err != nil || ip.To4() == nil {
			return fmt.Errorf("%q must be an IPv4 CIDR for a wireguard-client network", "subnet")
		}
		if ones, _ := ipNet.Mask.Size(); ones != 32 {
			return fmt.Errorf("%q must be a /32 tunnel interface address (e.g. 10.0.0.2/32)", "subnet")
		}
	}
	if !utils.IsRawConfigSet(raw, "dhcp_dns") {
		return fmt.Errorf("%q (interface DNS) is required for a wireguard-client network", "dhcp_dns")
	}
	for _, k := range []string{"wireguard_client_peer_ip", "wireguard_client_peer_public_key", "wireguard_client_peer_port"} {
		if !utils.IsRawConfigSet(raw, k) {
			return fmt.Errorf("%q is required when vpn_type = %q", k, "wireguard-client")
		}
	}
	return nil
}

// customizeNetworkDHCPGuarding enforces that DHCP Guarding has at least one trusted
// DHCP server: the controller rejects guarding with no trusted server (api.err.
// MissingIPAddress). The check is driven off the *raw config*, not d.Get. Both
// attributes are Optional+Computed, and in a ResourceDiff a Computed list reads back
// empty even when its prior value is being inherited — while the scalar dhcp_guarding
// still surfaces the inherited true. Gating on d.Get would therefore wrongly fire on
// an unrelated Update that merely inherits a previously-enabled guarding plus its
// trusted servers (the exact issue #123 regression: omitting dhcp_guarding while
// changing some other attribute). So only enforce when the user explicitly enables
// guarding in *this* configuration; an inherited value was already validated when it
// was first set, and apply preserves the inherited trusted servers.
func customizeNetworkDHCPGuarding(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	return validateDHCPGuardingRawConfig(d.GetRawConfig())
}

// validateDHCPGuardingRawConfig holds the pure raw-config logic so it is unit-testable
// without constructing a ResourceDiff. See customizeNetworkDHCPGuarding for the why.
func validateDHCPGuardingRawConfig(raw cty.Value) error {
	if raw.IsNull() || !raw.Type().HasAttribute("dhcp_guarding") {
		return nil
	}
	// dhcp_guarding is a bool, so read it from raw config directly — IsRawConfigSet
	// is for strings/numbers/collections and would panic on a bool. Skip unless it
	// is explicitly, known-true in config (null = omitted, unknown = interpolated).
	guarding := raw.GetAttr("dhcp_guarding")
	if guarding.IsNull() || !guarding.IsKnown() || guarding.False() {
		return nil
	}
	if !utils.IsRawConfigSet(raw, "dhcp_guarding_trusted_servers") {
		return fmt.Errorf("%q is required when %q is enabled: DHCP Guarding needs at least one trusted DHCP server IP address", "dhcp_guarding_trusted_servers", "dhcp_guarding")
	}
	return nil
}

// customizeNetworkDefaultGateway enforces that the DHCP default-gateway override
// keeps its enable toggle and gateway IP consistent at plan time. The controller
// would otherwise either silently ignore a gateway written with the override off,
// or reject an override turned on with no gateway. Driven off the *raw config* (not
// d.Get) for the same reason as customizeNetworkDHCPGuarding: both attributes are
// Optional+Computed, so on an Update that omits them they inherit the controller's
// values, and gating on the inherited value would wrongly fire on an unrelated
// change. Each rule therefore keys on the *explicit* config of its counterpart.
func customizeNetworkDefaultGateway(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	return validateDefaultGatewayRawConfig(d.GetRawConfig())
}

// validateDefaultGatewayRawConfig holds the pure raw-config logic so it is
// unit-testable without constructing a ResourceDiff. See customizeNetworkDefaultGateway.
func validateDefaultGatewayRawConfig(raw cty.Value) error {
	if raw.IsNull() || !raw.Type().HasAttribute("dhcpd_gateway_enabled") {
		return nil
	}
	// dhcpd_gateway_enabled is a bool, so read it from raw config directly
	// (IsRawConfigSet is for strings/numbers/collections and would panic on a bool).
	// null = omitted (inherits), unknown = interpolated (can't validate at plan).
	enabled := raw.GetAttr("dhcpd_gateway_enabled")
	enabledKnown := !enabled.IsNull() && enabled.IsKnown()
	gatewaySet := utils.IsRawConfigSet(raw, "dhcpd_gateway")

	// Override explicitly off but a gateway IP supplied: the controller would ignore
	// the gateway. Only fire on an explicit false so an omitted (inherited) toggle
	// that may already be true is not tripped.
	if gatewaySet && enabledKnown && enabled.False() {
		return fmt.Errorf("%q must be true when %q is set: enable the default-gateway override, or remove %q", "dhcpd_gateway_enabled", "dhcpd_gateway", "dhcpd_gateway")
	}
	// Override explicitly on but no gateway IP: mirrors the DHCP Guarding gate. Only
	// fire when the user enables it in *this* config, so an inherited true on an
	// unrelated Update (gateway preserved via Optional+Computed) is not tripped.
	if enabledKnown && enabled.True() && !gatewaySet {
		return fmt.Errorf("%q is required when %q is true: set the IPv4 gateway to advertise to DHCP clients, or set %q = false", "dhcpd_gateway", "dhcpd_gateway_enabled", "dhcpd_gateway_enabled")
	}
	return nil
}

// generateWireguardPrivateKey returns a base64-encoded Curve25519 private key in the
// same format as `wg genkey` and the UniFi UI: 32 random bytes, clamped per the
// Curve25519 requirements. Used to mint the gateway's own key when the user omits
// x_wireguard_private_key, since the controller will not generate one itself.
func generateWireguardPrivateKey() (string, error) {
	var key [32]byte
	if _, err := rand.Read(key[:]); err != nil {
		return "", err
	}
	key[0] &= 248
	key[31] &= 127
	key[31] |= 64
	return base64.StdEncoding.EncodeToString(key[:]), nil
}

// wireguardPublicKey derives the base64 WireGuard public key from a base64
// private key via Curve25519 scalar-base multiplication (matching `wg pubkey`).
// The controller stores the private key but returns a null public key, so the
// provider computes it to populate wireguard_public_key, the value you add as a
// peer on the remote WireGuard server.
func wireguardPublicKey(privateKeyB64 string) (string, error) {
	priv, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		return "", err
	}
	if len(priv) != 32 {
		return "", fmt.Errorf("WireGuard private key must be 32 bytes, got %d", len(priv))
	}
	pub, err := curve25519.X25519(priv, curve25519.Basepoint)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(pub), nil
}

func resourceNetworkSetResourceData(resp *unifi.Network, d *schema.ResourceData, site string) diag.Diagnostics {
	wanType := ""
	wanDNS := []string{}
	wanIP := ""
	wanNetmask := ""
	wanGateway := ""

	if resp.Purpose == "wan" {
		wanType = resp.WANType

		for _, dns := range []string{
			resp.WANDNS1,
			resp.WANDNS2,
			resp.WANDNS3,
			resp.WANDNS4,
		} {
			if dns == "" {
				continue
			}
			wanDNS = append(wanDNS, dns)
		}

		if wanType != "dhcp" {
			wanIP = resp.WANIP
			wanNetmask = resp.WANNetmask
			wanGateway = resp.WANGateway
		}

		// TODO: set other wan only fields here?
	}

	vlan := 0
	if resp.VLANEnabled {
		vlan = resp.VLAN
	}

	dhcpLease := resp.DHCPDLeaseTime
	if resp.DHCPDEnabled && dhcpLease == 0 {
		dhcpLease = 86400
	}

	dhcpDNS := []string{}
	if resp.DHCPDDNSEnabled {
		for _, dns := range []string{
			resp.DHCPDDNS1,
			resp.DHCPDDNS2,
			resp.DHCPDDNS3,
			resp.DHCPDDNS4,
		} {
			if dns == "" {
				continue
			}
			dhcpDNS = append(dhcpDNS, dns)
		}
	}

	dhcpGuardServers := []string{}
	for _, ip := range []string{
		resp.DHCPDIP1,
		resp.DHCPDIP2,
		resp.DHCPDIP3,
	} {
		if ip == "" {
			continue
		}
		dhcpGuardServers = append(dhcpGuardServers, ip)
	}

	dhcpV6DNS := []string{}
	for _, dns := range []string{
		resp.DHCPDV6DNS1,
		resp.DHCPDV6DNS2,
		resp.DHCPDV6DNS3,
		resp.DHCPDV6DNS4,
	} {
		if dns == "" {
			continue
		}
		dhcpV6DNS = append(dhcpV6DNS, dns)
	}

	d.Set("site", site)
	d.Set("name", resp.Name)
	d.Set("purpose", resp.Purpose)
	d.Set("vlan_id", vlan)
	d.Set("subnet", utils.CidrZeroBased(resp.IPSubnet))

	networkGroup := resp.NetworkGroup
	if resp.Purpose == "wan" && networkGroup == "" {
		networkGroup = "LAN"
	}
	d.Set("network_group", networkGroup)

	// Always read back the firewall zone so drift is detectable and imports round-trip.
	d.Set("firewall_zone_id", resp.FirewallZoneID)

	d.Set("dhcp_dns", dhcpDNS)
	d.Set("dhcp_enabled", resp.DHCPDEnabled)
	d.Set("dhcp_lease", dhcpLease)
	d.Set("dhcp_relay_enabled", resp.DHCPRelayEnabled)
	d.Set("dhcp_guarding", resp.DHCPguardEnabled)
	d.Set("dhcp_guarding_trusted_servers", dhcpGuardServers)
	d.Set("dhcp_start", resp.DHCPDStart)
	d.Set("dhcp_stop", resp.DHCPDStop)
	d.Set("dhcp_v6_dns_auto", resp.DHCPDV6DNSAuto)
	d.Set("dhcp_v6_dns", dhcpV6DNS)
	d.Set("dhcp_v6_enabled", resp.DHCPDV6Enabled)
	d.Set("dhcp_v6_lease", resp.DHCPDV6LeaseTime)
	d.Set("dhcp_v6_start", resp.DHCPDV6Start)
	d.Set("dhcp_v6_stop", resp.DHCPDV6Stop)
	d.Set("dhcpd_boot_enabled", resp.DHCPDBootEnabled)
	d.Set("dhcpd_boot_filename", resp.DHCPDBootFilename)
	d.Set("dhcpd_boot_server", resp.DHCPDBootServer)
	d.Set("dhcpd_gateway_enabled", resp.DHCPDGatewayEnabled)
	d.Set("dhcpd_gateway", resp.DHCPDGateway)
	d.Set("domain_name", resp.DomainName)
	d.Set("enabled", resp.Enabled)
	d.Set("igmp_snooping", resp.IGMPSnooping)
	d.Set("upnp_lan_enabled", resp.UpnpLanEnabled)
	d.Set("internet_access_enabled", resp.InternetAccessEnabled)
	d.Set("network_isolation_enabled", resp.NetworkIsolationEnabled)

	ipv6InterfaceType := resp.IPV6InterfaceType
	if resp.Purpose == "wan" && ipv6InterfaceType == "" {
		ipv6InterfaceType = "none"
	}
	d.Set("ipv6_interface_type", ipv6InterfaceType)
	d.Set("ipv6_pd_interface", resp.IPV6PDInterface)
	d.Set("ipv6_pd_prefixid", resp.IPV6PDPrefixid)
	d.Set("ipv6_pd_start", resp.IPV6PDStart)
	d.Set("ipv6_pd_stop", resp.IPV6PDStop)
	d.Set("ipv6_ra_enable", resp.IPV6RaEnabled)
	d.Set("ipv6_ra_preferred_lifetime", resp.IPV6RaPreferredLifetime)
	d.Set("ipv6_ra_priority", resp.IPV6RaPriority)
	d.Set("ipv6_ra_valid_lifetime", resp.IPV6RaValidLifetime)
	d.Set("ipv6_static_subnet", resp.IPV6Subnet)
	d.Set("multicast_dns", resp.MdnsEnabled)
	d.Set("wan_dhcp_v6_pd_size", resp.WANDHCPv6PDSize)
	d.Set("wan_dns", wanDNS)
	d.Set("wan_egress_qos", resp.WANEgressQOS)
	d.Set("wan_gateway_v6", resp.WANGatewayV6)
	d.Set("wan_gateway", wanGateway)
	d.Set("wan_ip", wanIP)
	d.Set("wan_ipv6", resp.WANIPV6)
	d.Set("wan_netmask", wanNetmask)
	d.Set("wan_networkgroup", resp.WANNetworkGroup)
	d.Set("wan_prefixlen", resp.WANPrefixlen)
	d.Set("wan_type_v6", resp.WANTypeV6)
	d.Set("wan_type", wanType)
	d.Set("wan_username", resp.WANUsername)
	d.Set("x_wan_password", resp.XWANPassword)

	d.Set("vpn_type", resp.VPNType)
	d.Set("wireguard_interface", resp.WireguardInterface)
	d.Set("wireguard_client_mode", resp.WireguardClientMode)
	d.Set("wireguard_client_peer_ip", resp.WireguardClientPeerIP)
	d.Set("wireguard_client_peer_port", resp.WireguardClientPeerPort)
	d.Set("wireguard_client_peer_public_key", resp.WireguardClientPeerPublicKey)
	// Write-only secrets: the controller may omit these on read. Only overwrite state when a
	// non-empty value is returned, otherwise the configured/generated secret would be blanked
	// and the resource would drift on every refresh.
	if resp.WireguardClientPresharedKey != "" {
		d.Set("wireguard_client_preshared_key", resp.WireguardClientPresharedKey)
	}
	d.Set("wireguard_client_preshared_key_enabled", resp.WireguardClientPresharedKeyEnabled)
	if resp.XWireguardPrivateKey != "" {
		d.Set("x_wireguard_private_key", resp.XWireguardPrivateKey)
	}
	// The controller returns a null public key, so derive it from the private key
	// (the response value, or the one we generated and stored in state).
	// The controller returns a null public key. wireguardPublicKey errors on an
	// empty/short key, so the err==nil check already skips the no-key case.
	wgPublicKey := resp.WireguardPublicKey
	if wgPublicKey == "" {
		wgPrivateKey := resp.XWireguardPrivateKey
		if wgPrivateKey == "" {
			wgPrivateKey = d.Get("x_wireguard_private_key").(string)
		}
		if derived, err := wireguardPublicKey(wgPrivateKey); err == nil {
			wgPublicKey = derived
		}
	}
	d.Set("wireguard_public_key", wgPublicKey)
	d.Set("vpn_client_default_route", resp.VPNClientDefaultRoute)
	d.Set("vpn_client_pull_dns", resp.VPNClientPullDNS)
	d.Set("uid_vpn_custom_routing", utils.CidrListZeroBased(resp.UidVPNCustomRouting))

	return nil
}

func resourceNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	id := d.Id()

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.GetNetwork(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceNetworkSetResourceData(resp, d, site)
}

func resourceNetworkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	req, err := resourceNetworkGetResourceData(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	req.ID = d.Id()
	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	req.SiteID = site

	// go-unifi v1.9.3's updateNetwork converts a successful-but-empty PUT response
	// into unifi.ErrNotFound (see utils.ReReadOnUpdateNotFound / issue #98); re-read
	// to tell a spurious error from a genuine out-of-band deletion.
	resp, err := c.UpdateNetwork(ctx, site, req)
	resp, found, err := utils.ReReadOnUpdateNotFound(resp, err, func() (*unifi.Network, error) {
		return c.GetNetwork(ctx, site, req.ID)
	})
	if err != nil {
		return diag.FromErr(err)
	}
	if !found {
		// The network is genuinely gone; clear state so it is recreated on the next
		// apply (mirrors resourceNetworkRead and resourceNetworkDelete).
		d.SetId("")
		return nil
	}

	return resourceNetworkSetResourceData(resp, d, site)
}

func resourceNetworkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	id := d.Id()

	err := c.DeleteNetwork(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		return nil
	}
	return diag.FromErr(err)
}

func importNetwork(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	c := meta.(*base.Client)
	id := d.Id()
	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	if strings.Contains(id, ":") {
		importParts := strings.SplitN(id, ":", 2)
		site = importParts[0]
		id = importParts[1]
	}

	if strings.HasPrefix(id, "name=") {
		targetName := strings.TrimPrefix(id, "name=")
		var err error
		if id, err = getNetworkIDByName(ctx, c.Client, targetName, site); err != nil {
			return nil, err
		}
	}

	if id != "" {
		d.SetId(id)
	}
	if site != "" {
		d.Set("site", site)
	}

	return []*schema.ResourceData{d}, nil
}

func getNetworkIDByName(ctx context.Context, client unifi.Client, networkName, site string) (string, error) {
	networks, err := client.ListNetwork(ctx, site)
	if err != nil {
		return "", err
	}

	idMatchingName := ""
	var allNames []string
	for _, network := range networks {
		allNames = append(allNames, network.Name)
		if network.Name != networkName {
			continue
		}
		if idMatchingName != "" {
			return "", fmt.Errorf("found multiple networks with name '%s'", networkName)
		}
		idMatchingName = network.ID
	}
	if idMatchingName == "" {
		return "", fmt.Errorf("found no networks with name '%s', found: %s", networkName, strings.Join(allNames, ", "))
	}
	return idMatchingName, nil
}

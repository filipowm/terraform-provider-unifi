package network

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/utils"
	"regexp"
	"strings"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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

	ipV6InterfaceTypeRegexp   = regexp.MustCompile("none|pd|static")
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
					"Must be a valid IPv6 address within your allocated IPv6 subnet.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPv6Address,
			},
			"dhcp_v6_stop": {
				Description: "The ending IPv6 address for the DHCPv6 range. Used in static DHCPv6 configuration.\n" +
					"Must be after dhcp_v6_start in the IPv6 address space.",
				Type:         schema.TypeString,
				Optional:     true,
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
					"* `pd` - Prefix Delegation from upstream\n\n" +
					"Choose based on your IPv6 deployment strategy and ISP capabilities.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "none",
				ValidateFunc: validateIpV6InterfaceType,
			},
			"ipv6_static_subnet": {
				Description: "The static IPv6 subnet in CIDR notation (e.g., '2001:db8::/64') when using static IPv6.\n" +
					"Only applicable when `ipv6_interface_type` is 'static'.\n" +
					"Must be a valid IPv6 subnet allocated to your organization.",
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipv6_pd_interface": {
				Description: "The WAN interface to use for IPv6 Prefix Delegation. Options:\n" +
					"* `wan` - Primary WAN interface\n" +
					"* `wan2` - Secondary WAN interface\n" +
					"Only applicable when `ipv6_interface_type` is 'pd'.",
				Type:         schema.TypeString,
				Optional:     true,
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
					"Must be within the delegated prefix range.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPv6Address,
			},
			"ipv6_pd_stop": {
				Description: "The ending IPv6 address for Prefix Delegation range.\n" +
					"Only used when `ipv6_interface_type` is 'pd'.\n" +
					"Must be after `ipv6_pd_start` within the delegated prefix.",
				Type:         schema.TypeString,
				Optional:     true,
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
				Description: "Enables network isolation. When enabled:\n" +
					"* Prevents communication between clients on this network\n" +
					"* Each client can only communicate with the gateway\n" +
					"* Commonly used for guest networks or IoT devices",
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
					"Affects router selection when multiple IPv6 routers exist.",
				Type:         schema.TypeString,
				Optional:     true,
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
					"e.g. `10.0.0.2/32`) and `dhcp_dns` (interface DNS) — the controller rejects the create without them.",
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
				Type:     schema.TypeString,
				Optional: true,
			},
			"wireguard_client_preshared_key": {
				Description: "An optional WireGuard pre-shared key (PSK) for an additional layer of symmetric-key " +
					"security with the peer. Keep this value secret. The controller may not return this value on read, " +
					"so it is computed to avoid spurious drift. Only applicable when `vpn_type` is 'wireguard-client'.",
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
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
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
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
	for i, cidr := range uidVPNCustomRouting {
		uidVPNCustomRouting[i] = utils.CidrZeroBased(cidr)
	}

	// The UniFi controller requires the gateway's own WireGuard private key in the
	// create payload — it does NOT generate one server-side, and an omitted key is
	// rejected with api.err.WireguardMissingPrivateKey. Mirror the UI, which generates
	// the key client-side: when x_wireguard_private_key is left unset for a
	// wireguard-client network, generate one here and persist it to state so it stays
	// stable across refreshes (the controller does not return it on read).
	vpnType := d.Get("vpn_type").(string)
	xWireguardPrivateKey := d.Get("x_wireguard_private_key").(string)
	if vpnType == "wireguard-client" && xWireguardPrivateKey == "" {
		generated, genErr := generateWireguardPrivateKey()
		if genErr != nil {
			return nil, fmt.Errorf("unable to generate WireGuard private key: %w", genErr)
		}
		xWireguardPrivateKey = generated
		if setErr := d.Set("x_wireguard_private_key", generated); setErr != nil {
			return nil, fmt.Errorf("unable to persist generated WireGuard private key: %w", setErr)
		}
	}

	// For a LAN the `subnet` is the gateway address, so CidrOneBased applies the
	// +1 host offset. A vpn-client's subnet is the tunnel interface address, which
	// must be sent verbatim — the +1 would corrupt it and cause perpetual drift.
	ipSubnet := utils.CidrOneBased(d.Get("subnet").(string))
	if d.Get("purpose").(string) == "vpn-client" {
		ipSubnet = utils.CidrZeroBased(d.Get("subnet").(string))
	}

	return &unifi.Network{
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
		DHCPRelayEnabled:  d.Get("dhcp_relay_enabled").(bool),
		DomainName:        d.Get("domain_name").(string),
		IGMPSnooping:      d.Get("igmp_snooping").(bool),
		UpnpLanEnabled:    d.Get("upnp_lan_enabled").(bool),
		MdnsEnabled:       d.Get("multicast_dns").(bool),
		Enabled:           d.Get("enabled").(bool),

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
		// wireguard_public_key is computed (derived by the controller), so it is not sent here.
		VPNType:                            vpnType,
		WireguardInterface:                 d.Get("wireguard_interface").(string),
		WireguardClientMode:                d.Get("wireguard_client_mode").(string),
		WireguardClientPeerIP:              d.Get("wireguard_client_peer_ip").(string),
		WireguardClientPeerPort:            d.Get("wireguard_client_peer_port").(int),
		WireguardClientPeerPublicKey:       d.Get("wireguard_client_peer_public_key").(string),
		WireguardClientPresharedKey:        d.Get("wireguard_client_preshared_key").(string),
		WireguardClientPresharedKeyEnabled: d.Get("wireguard_client_preshared_key_enabled").(bool),
		XWireguardPrivateKey:               xWireguardPrivateKey,
		VPNClientDefaultRoute:              d.Get("vpn_client_default_route").(bool),
		VPNClientPullDNS:                   d.Get("vpn_client_pull_dns").(bool),
		UidVPNCustomRouting:                uidVPNCustomRouting,
	}, nil
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
// provider computes it to populate wireguard_public_key — the value you add as a
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

	d.Set("dhcp_dns", dhcpDNS)
	d.Set("dhcp_enabled", resp.DHCPDEnabled)
	d.Set("dhcp_lease", dhcpLease)
	d.Set("dhcp_relay_enabled", resp.DHCPRelayEnabled)
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
	wgPublicKey := resp.WireguardPublicKey
	if wgPublicKey == "" {
		wgPrivateKey := resp.XWireguardPrivateKey
		if wgPrivateKey == "" {
			wgPrivateKey = d.Get("x_wireguard_private_key").(string)
		}
		if wgPrivateKey != "" {
			if derived, err := wireguardPublicKey(wgPrivateKey); err == nil {
				wgPublicKey = derived
			}
		}
	}
	d.Set("wireguard_public_key", wgPublicKey)
	d.Set("vpn_client_default_route", resp.VPNClientDefaultRoute)
	d.Set("vpn_client_pull_dns", resp.VPNClientPullDNS)
	customRouting := make([]string, len(resp.UidVPNCustomRouting))
	for i, cidr := range resp.UidVPNCustomRouting {
		if canonical := utils.CidrZeroBased(cidr); canonical != "" {
			customRouting[i] = canonical
		} else {
			customRouting[i] = cidr
		}
	}
	d.Set("uid_vpn_custom_routing", customRouting)

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

	resp, err := c.UpdateNetwork(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
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

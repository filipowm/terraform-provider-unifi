package network

import (
	"context"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
)

func DataNetwork() *schema.Resource {
	return &schema.Resource{
		Description: "`unifi_network` data source can be used to retrieve settings for a network by name or ID.",

		ReadContext: dataNetworkRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description:   "The ID of the network.",
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"name"},
			},
			"site": {
				Description: "The name of the site to associate the network with.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
			},
			"name": {
				Description:   "The name of the network.",
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"id"},
			},

			// read-only / computed
			"purpose": {
				Description: "The purpose of the network. One of `corporate`, `guest`, `wan`, or `vlan-only`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vlan_id": {
				Description: "The VLAN ID of the network.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"subnet": {
				Description: "The subnet of the network (CIDR address).",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"network_group": {
				Description: "The group of the network.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"firewall_zone_id": {
				Description: "The ID of the Zone-Based Firewall (ZBF) zone this network belongs to. Only " +
					"meaningful on UniFi OS 9.x controllers with Zone-Based Firewall enabled; empty otherwise. " +
					"The zone ID is site-scoped.",
				Type:     schema.TypeString,
				Computed: true,
			},
			"dhcp_start": {
				Description: "The IPv4 address where the DHCP range of addresses starts.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"dhcp_stop": {
				Description: "The IPv4 address where the DHCP range of addresses stops.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"dhcp_enabled": {
				Description: "whether DHCP is enabled or not on this network.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"dhcp_lease": {
				Description: "lease time for DHCP addresses.",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"dhcp_dns": {
				Description: "IPv4 addresses for the DNS server to be returned from the DHCP " +
					"server.",
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"dhcpd_boot_enabled": {
				Description: "Toggles on the DHCP boot options. will be set to true if you have dhcpd_boot_filename, and dhcpd_boot_server set.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"dhcpd_boot_server": {
				Description: "IPv4 address of a TFTP server to network boot from.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"dhcpd_boot_filename": {
				Description: "the file to PXE boot from on the dhcpd_boot_server.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"dhcpd_gateway_enabled": {
				Description: "Whether the DHCP default gateway is manually overridden (true) or auto (false).",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"dhcpd_gateway": {
				Description: "The IPv4 default gateway advertised to DHCP clients when the override is enabled.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"dhcp_v6_dns": {
				Description: "Specifies the IPv6 addresses for the DNS server to be returned from the DHCP " +
					"server. Used if `dhcp_v6_dns_auto` is set to `false`.",
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"dhcp_v6_dns_auto": {
				Description: "Specifies DNS source to propagate. If set `false` the entries in `dhcp_v6_dns` are used, the upstream entries otherwise",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"dhcp_v6_enabled": {
				Description: "Enable stateful DHCPv6 for static configuration.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"dhcp_v6_lease": {
				Description: "Specifies the lease time for DHCPv6 addresses.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"dhcp_v6_start": {
				Description: "start address of the DHCPv6 range. Used in static DHCPv6 configuration.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"dhcp_v6_stop": {
				Description: "End address of the DHCPv6 range. Used in static DHCPv6 configuration.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"domain_name": {
				Description: "The domain name of this network.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"igmp_snooping": {
				Description: "Specifies whether IGMP snooping is enabled or not.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"dhcp_guarding": {
				Description: "Specifies whether DHCP Guarding (rogue/untrusted DHCP server protection) is enabled or not.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"ipv6_interface_type": {
				Description: "Specifies which type of IPv6 connection to use. Must be one of either `static`, `pd`, `single_network`, or `none`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ipv6_static_subnet": {
				Description: "Specifies the static IPv6 subnet (when ipv6_interface_type is 'static').",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ipv6_pd_interface": {
				Description: "Specifies which WAN interface to use for IPv6 PD. Must be one of either `wan` or `wan2`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ipv6_pd_prefixid": {
				Description: "Specifies the IPv6 Prefix ID.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ipv6_pd_start": {
				Description: "start address of the DHCPv6 range. Used if `ipv6_interface_type` is set to `pd`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ipv6_pd_stop": {
				Description: "End address of the DHCPv6 range. Used if `ipv6_interface_type` is set to `pd`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ipv6_ra_enable": {
				Description: "Specifies whether to enable router advertisements or not.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"ipv6_ra_preferred_lifetime": {
				Description: "Lifetime in which the address can be used. Address becomes deprecated afterwards. Must be lower than or equal to `ipv6_ra_valid_lifetime`",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"ipv6_ra_priority": {
				Description: "IPv6 router advertisement priority. Must be one of either `high`, `medium`, or `low`",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ipv6_ra_valid_lifetime": {
				Description: "Total lifetime in which the address can be used. Must be equal to or greater than `ipv6_ra_preferred_lifetime`.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"multicast_dns": {
				Description: "Specifies whether Multicast DNS (mDNS) is enabled or not on the network (Controller >=v7).",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"wan_ip": {
				Description: "The IPv4 address of the WAN.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"wan_netmask": {
				Description: "The IPv4 netmask of the WAN.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"wan_gateway": {
				Description: "The IPv4 gateway of the WAN.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"wan_dns": {
				Description: "DNS servers IPs of the WAN.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"wan_type": {
				Description: "Specifies the IPV4 WAN connection type. One of either `disabled`, `static`, `dhcp`, or `pppoe`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"wan_networkgroup": {
				Description: "Specifies the WAN network group. One of either `WAN`, `WAN2` or `WAN_LTE_FAILOVER`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"wan_egress_qos": {
				Description: "Specifies the WAN egress quality of service.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"wan_username": {
				Description: "Specifies the IPV4 WAN username.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"x_wan_password": {
				Description: "Specifies the IPV4 WAN password.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"wan_type_v6": {
				Description: "Specifies the IPV6 WAN connection type. Must be one of either `disabled`, `static`, or `dhcpv6`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"wan_dhcp_v6_pd_size": {
				Description: "Specifies the IPv6 prefix size to request from ISP. Must be a number between 48 and 64.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"wan_ipv6": {
				Description: "The IPv6 address of the WAN.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"wan_gateway_v6": {
				Description: "The IPv6 gateway of the WAN.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"wan_prefixlen": {
				Description: "The IPv6 prefix length of the WAN. Must be between 1 and 128.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func dataNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	name, _ := d.Get("name").(string)
	site, _ := d.Get("site").(string)
	id, _ := d.Get("id").(string)
	if site == "" {
		site = c.Site
	}
	if (name == "" && id == "") || (name != "" && id != "") {
		return diag.Errorf("One of 'name' OR 'id' is required")
	}

	networks, err := c.ListNetwork(ctx, site)
	if err != nil {
		return diag.FromErr(err)
	}
	for _, n := range networks {
		if (name != "" && n.Name == name) || (id != "" && n.ID == id) {
			dhcpDNS := []string{}
			for _, dns := range []string{
				n.DHCPDDNS1,
				n.DHCPDDNS2,
				n.DHCPDDNS3,
				n.DHCPDDNS4,
			} {
				if dns == "" {
					continue
				}
				dhcpDNS = append(dhcpDNS, dns)
			}
			wanDNS := []string{}
			for _, dns := range []string{
				n.WANDNS1,
				n.WANDNS2,
				n.WANDNS3,
				n.WANDNS4,
			} {
				if dns == "" {
					continue
				}
				wanDNS = append(wanDNS, dns)
			}

			d.SetId(n.ID)
			for key, value := range map[string]interface{}{
				"site":                  site,
				"name":                  n.Name,
				"purpose":               n.Purpose,
				"vlan_id":               n.VLAN,
				"subnet":                utils.CidrZeroBased(n.IPSubnet),
				"network_group":         n.NetworkGroup,
				"firewall_zone_id":      n.FirewallZoneID,
				"dhcp_dns":              dhcpDNS,
				"dhcp_start":            n.DHCPDStart,
				"dhcp_stop":             n.DHCPDStop,
				"dhcp_enabled":          n.DHCPDEnabled,
				"dhcp_lease":            n.DHCPDLeaseTime,
				"dhcpd_boot_enabled":    n.DHCPDBootEnabled,
				"dhcpd_boot_server":     n.DHCPDBootServer,
				"dhcpd_boot_filename":   n.DHCPDBootFilename,
				"dhcpd_gateway_enabled": n.DHCPDGatewayEnabled,
				"dhcpd_gateway":         n.DHCPDGateway,
				"domain_name":           n.DomainName,
				"igmp_snooping":         n.IGMPSnooping,
				"dhcp_guarding":         n.DHCPguardEnabled,
				"ipv6_interface_type":   n.IPV6InterfaceType,
				"ipv6_static_subnet":    n.IPV6Subnet,
				"ipv6_pd_interface":     n.IPV6PDInterface,
				"ipv6_pd_prefixid":      n.IPV6PDPrefixid,
				"ipv6_ra_enable":        n.IPV6RaEnabled,
				"multicast_dns":         n.MdnsEnabled,
				"wan_ip":                n.WANIP,
				"wan_netmask":           n.WANNetmask,
				"wan_gateway":           n.WANGateway,
				"wan_type":              n.WANType,
				"wan_dns":               wanDNS,
				"wan_networkgroup":      n.WANNetworkGroup,
				"wan_egress_qos":        n.WANEgressQOS,
				"wan_username":          n.WANUsername,
				"x_wan_password":        n.XWANPassword,
				"wan_type_v6":           n.WANTypeV6,
				"wan_dhcp_v6_pd_size":   n.WANDHCPv6PDSize,
				"wan_ipv6":              n.WANIPV6,
				"wan_gateway_v6":        n.WANGatewayV6,
				"wan_prefixlen":         n.WANPrefixlen,
			} {
				if err := d.Set(key, value); err != nil {
					return diag.FromErr(err)
				}
			}

			return nil
		}
	}

	return diag.Errorf("network not found with name %s", name)
}

package firewall

import (
	"context"
	"errors"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/utils"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
	"regexp"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var firewallRuleProtocolRegexp = regexp.MustCompile("^$|all|([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])|tcp_udp|ah|ax.25|dccp|ddp|egp|eigrp|encap|esp|etherip|fc|ggp|gre|hip|hmp|icmp|idpr-cmtp|idrp|igmp|igp|ip|ipcomp|ipencap|ipip|ipv6|ipv6-frag|ipv6-icmp|ipv6-nonxt|ipv6-opts|ipv6-route|isis|iso-tp4|l2tp|manet|mobility-header|mpls-in-ip|ospf|pim|pup|rdp|rohc|rspf|rsvp|sctp|shim6|skip|st|tcp|udp|udplite|vmtp|vrrp|wesp|xns-idp|xtp")
var firewallRuleProtocolV6Regexp = regexp.MustCompile("^$|([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])|ah|all|dccp|eigrp|esp|gre|icmpv6|ipcomp|ipv6|ipv6-frag|ipv6-icmp|ipv6-nonxt|ipv6-opts|ipv6-route|isis|l2tp|manet|mobility-header|mpls-in-ip|ospf|pim|rsvp|sctp|shim6|tcp|tcp_udp|udp|vrrp")
var firewallRuleICMPv6TypenameRegexp = regexp.MustCompile("^$|address-unreachable|bad-header|beyond-scope|communication-prohibited|destination-unreachable|echo-reply|echo-request|failed-policy|neighbor-advertisement|neighbor-solicitation|no-route|packet-too-big|parameter-problem|port-unreachable|redirect|reject-route|router-advertisement|router-solicitation|time-exceeded|ttl-zero-during-reassembly|ttl-zero-during-transit|unknown-header-type|unknown-option")

func ResourceFirewallRule() *schema.Resource {
	return &schema.Resource{
		Description: "The `unifi_firewall_rule` resource manages firewall rules.\n\n" +
			"This resource allows you to create and manage firewall rules that control traffic flow between different network segments (WAN, LAN, Guest) " +
			"for both IPv4 and IPv6 traffic. Rules can be configured to allow, drop, or reject traffic based on various criteria including protocols, " +
			"ports, and IP addresses.\n\n" +
			"Rules are processed in order based on their `rule_index`, with lower numbers being processed first. Custom rules should use indices between " +
			"2000-2999 or 4000-4999 to avoid conflicts with system rules.",

		CreateContext: resourceFirewallRuleCreate,
		ReadContext:   resourceFirewallRuleRead,
		UpdateContext: resourceFirewallRuleUpdate,
		DeleteContext: resourceFirewallRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: base.ImportSiteAndID,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the firewall rule in the UniFi controller.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"site": {
				Description: "The name of the UniFi site where the firewall rule should be created. If not specified, the default site will be used.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "A friendly name for the firewall rule. This helps identify the rule's purpose in the UniFi controller UI.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"action": {
				Description: "The action to take when traffic matches this rule. Valid values are:\n" +
					"  * `accept` - Allow the traffic\n" +
					"  * `drop` - Silently block the traffic\n" +
					"  * `reject` - Block the traffic and send an ICMP rejection message",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"drop", "accept", "reject"}, false),
			},
			"ruleset": {
				Description: "Defines which traffic flow this rule applies to. The format is [NETWORK]_[DIRECTION], where:\n" +
					"  * NETWORK can be: WAN, LAN, GUEST (or their IPv6 variants WANv6, LANv6, GUESTv6)\n" +
					"  * DIRECTION can be:\n" +
					"    * IN - Traffic entering the network\n" +
					"    * OUT - Traffic leaving the network\n" +
					"    * LOCAL - Traffic destined for the USG/UDM itself\n\n" +
					"Examples: WAN_IN (incoming WAN traffic), LAN_OUT (outgoing LAN traffic), GUEST_LOCAL (traffic to Controller from guest network)",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"WAN_IN", "WAN_OUT", "WAN_LOCAL", "LAN_IN", "LAN_OUT", "LAN_LOCAL", "GUEST_IN", "GUEST_OUT", "GUEST_LOCAL", "WANv6_IN", "WANv6_OUT", "WANv6_LOCAL", "LANv6_IN", "LANv6_OUT", "LANv6_LOCAL", "GUESTv6_IN", "GUESTv6_OUT", "GUESTv6_LOCAL"}, false),
			},
			"rule_index": {
				Description: "The processing order for this rule. Lower numbers are processed first. Custom rules should use:\n" +
					"  * 2000-2999 for rules processed before auto-generated rules\n" +
					"  * 4000-4999 for rules processed after auto-generated rules",
				Type:     schema.TypeInt,
				Required: true,
				// 2[0-9]{3}|4[0-9]{3}
			},
			"protocol": {
				Description: "The IPv4 protocol this rule applies to. Common values (not all are listed) include:\n" +
					"  * `all` - Match all protocols\n" +
					"  * `tcp` - TCP traffic only (e.g., web, email)\n" +
					"  * `udp` - UDP traffic only (e.g., DNS, VoIP)\n" +
					"  * `tcp_udp` - Both TCP and UDP\n" +
					"  * `icmp` - ICMP traffic (ping, traceroute)\n" +
					"  * Protocol numbers (1-255) for other protocols\n\n" +
					"Examples:\n" +
					"  * Use 'tcp' for web server rules (ports 80, 443)\n" +
					"  * Use 'udp' for VoIP or gaming traffic\n" +
					"  * Use 'all' for general network access rules",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(firewallRuleProtocolRegexp, "must be a valid IPv4 protocol"),
			},
			"protocol_v6": {
				Description: "The IPv6 protocol this rule applies to. Similar to 'protocol' but for IPv6 traffic. Common values (not all are listed) include:\n" +
					"  * `all` - Match all protocols\n" +
					"  * `tcp` - TCP traffic only\n" +
					"  * `udp` - UDP traffic only\n" +
					"  * `tcp_udp` - Both TCP and UDP traffic\n" +
					"  * `ipv6-icmp` - ICMPv6 traffic",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(firewallRuleProtocolV6Regexp, "must be a valid IPv6 protocol"),
			},
			"icmp_typename": {
				Description: "The ICMP type name when protocol is set to 'icmp'. Common values include:\n" +
					"  * `echo-request` - ICMP ping requests\n" +
					"  * `echo-reply` - ICMP ping replies\n" +
					"  * `destination-unreachable` - Host/network unreachable messages\n" +
					"  * `time-exceeded` - TTL exceeded messages (traceroute)",
				Type:     schema.TypeString,
				Optional: true,
			},
			"icmp_v6_typename": {
				Description: "The ICMPv6 type name when protocol_v6 is set to 'ipv6-icmp'. Common values (not all are listed) include:\n" +
					"  * `echo-request` - IPv6 ping requests\n" +
					"  * `echo-reply` - IPv6 ping replies\n" +
					"  * `neighbor-solicitation` - IPv6 neighbor discovery\n" +
					"  * `neighbor-advertisement` - IPv6 neighbor announcements\n" +
					"  * `destination-unreachable` - Host/network unreachable messages\n" +
					"  * `packet-too-big` - Path MTU discovery messages",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(firewallRuleICMPv6TypenameRegexp, "must be a ICMPv6 type"),
			},
			"enabled": {
				Description: "Whether this firewall rule is active (true) or disabled (false). Defaults to true.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},

			// sources
			"src_network_id": {
				Description: "The ID of the source network this rule applies to. This can be found in the URL when viewing the network " +
					"in the UniFi controller, or by using the network's name in the form `[site]/[network_name]`.",
				Type:     schema.TypeString,
				Optional: true,
			},
			"src_network_type": {
				Description: "The type of source network address. Valid values are:\n" +
					"  * `ADDRv4` - Single IPv4 address\n" +
					"  * `NETv4` - IPv4 network in CIDR notation",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "NETv4",
				ValidateFunc: validation.StringInSlice([]string{"ADDRv4", "NETv4"}, false),
			},
			"src_firewall_group_ids": {
				Description: "A list of firewall group IDs to use as sources. Groups can contain:\n" +
					"  * IP Address Groups - For matching specific IP addresses\n" +
					"  * Network Groups - For matching entire subnets\n" +
					"  * Port Groups - For matching specific port numbers\n\n" +
					"Example uses:\n" +
					"  * Group of trusted admin IPs for remote access\n" +
					"  * Group of IoT device networks for isolation\n" +
					"  * Group of common service ports for allowing specific applications",
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"src_address": {
				Description: "The source IPv4 address for the firewall rule.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"src_address_ipv6": {
				Description: "The source IPv6 address or network in CIDR notation (e.g., '2001:db8::1' or '2001:db8::/64'). " +
					"Used for IPv6 firewall rules.",
				Type:     schema.TypeString,
				Optional: true,
			},
			"src_port": {
				Description: "The source port(s) for this rule. Can be:\n" +
					"  * A single port number (e.g., '80')\n" +
					"  * A port range (e.g., '8000:8080')\n" +
					"  * A list of ports/ranges separated by commas",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validators.PortRangeV2,
			},
			"src_mac": {
				Description: "The source MAC address this rule applies to. Use this to create rules that match specific devices " +
					"regardless of their IP address. Format: 'XX:XX:XX:XX:XX:XX'. MAC addresses are case-insensitive.",
				Type:     schema.TypeString,
				Optional: true,
			},

			// destinations
			"dst_network_id": {
				Description: "The ID of the destination network this rule applies to. This can be found in the URL when viewing the network " +
					"in the UniFi controller.",
				Type:     schema.TypeString,
				Optional: true,
			},
			"dst_network_type": {
				Description: "The type of destination network address. Valid values are:\n" +
					"  * `ADDRv4` - Single IPv4 address\n" +
					"  * `NETv4` - IPv4 network in CIDR notation",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "NETv4",
				ValidateFunc: validation.StringInSlice([]string{"ADDRv4", "NETv4"}, false),
			},
			"dst_firewall_group_ids": {
				Description: "A list of firewall group IDs to use as destinations. Groups can contain IP addresses, networks, or port numbers. " +
					"This allows you to create reusable sets of addresses/ports and reference them in multiple rules.",
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"dst_address": {
				Description: "The destination IPv4 address or network in CIDR notation (e.g., '192.168.1.10' or '192.168.0.0/24'). " +
					"The format must match dst_network_type - use a single IP for ADDRv4 or CIDR for NETv4.",
				Type:     schema.TypeString,
				Optional: true,
			},
			"dst_address_ipv6": {
				Description: "The destination IPv6 address or network in CIDR notation (e.g., '2001:db8::1' or '2001:db8::/64'). " +
					"Used for IPv6 firewall rules.",
				Type:     schema.TypeString,
				Optional: true,
			},
			"dst_port": {
				Description: "The destination port(s) for this rule. Can be:\n" +
					"  * A single port number (e.g., '80')\n" +
					"  * A port range (e.g., '8000:8080')\n" +
					"  * A list of ports/ranges separated by commas",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validators.PortRangeV2,
			},

			// advanced
			"logging": {
				Description: "Enable logging for the firewall rule.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"state_established": {
				Description: "Match established connections. When enabled:\n" +
					"  * Rule only applies to packets that are part of an existing connection\n" +
					"  * Useful for allowing return traffic without creating separate rules\n" +
					"  * Common in WAN_IN rules to allow responses to outbound connections\n\n" +
					"Example: Allow established connections from WAN while blocking new incoming connections",
				Type:     schema.TypeBool,
				Optional: true,
			},
			"state_invalid": {
				Description: "Match where the state is invalid.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"state_new": {
				Description: "Match where the state is new.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"state_related": {
				Description: "Match where the state is related.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"ip_sec": {
				Description:  "Specify whether the rule matches on IPsec packets. Can be one of `match-ipsec` or `match-none`.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"match-ipsec", "match-none"}, false),
			},
		},
	}
}

func resourceFirewallRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	req, err := resourceFirewallRuleGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.CreateFirewallRule(ctx, site, req)
	if err != nil {
		if utils.IsServerErrorContains(err, "api.err.FirewallGroupTypeExists") {
			return diag.Errorf("firewall rule groups must be of different group types (ie. a port group and address group): %s", err)
		}

		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return resourceFirewallRuleSetResourceData(resp, d, site)
}

func resourceFirewallRuleGetResourceData(d *schema.ResourceData) (*unifi.FirewallRule, error) {
	srcFirewallGroupIDs, err := utils.SetToStringSlice(d.Get("src_firewall_group_ids").(*schema.Set))
	if err != nil {
		return nil, err
	}

	dstFirewallGroupIDs, err := utils.SetToStringSlice(d.Get("dst_firewall_group_ids").(*schema.Set))
	if err != nil {
		return nil, err
	}

	return &unifi.FirewallRule{
		Enabled:          d.Get("enabled").(bool),
		Name:             d.Get("name").(string),
		Action:           d.Get("action").(string),
		Ruleset:          d.Get("ruleset").(string),
		RuleIndex:        d.Get("rule_index").(int),
		Protocol:         d.Get("protocol").(string),
		ProtocolV6:       d.Get("protocol_v6").(string),
		ICMPTypename:     d.Get("icmp_typename").(string),
		ICMPv6Typename:   d.Get("icmp_v6_typename").(string),
		Logging:          d.Get("logging").(bool),
		IPSec:            d.Get("ip_sec").(string),
		StateEstablished: d.Get("state_established").(bool),
		StateInvalid:     d.Get("state_invalid").(bool),
		StateNew:         d.Get("state_new").(bool),
		StateRelated:     d.Get("state_related").(bool),

		SrcNetworkType:      d.Get("src_network_type").(string),
		SrcMACAddress:       d.Get("src_mac").(string),
		SrcAddress:          d.Get("src_address").(string),
		SrcAddressIPV6:      d.Get("src_address_ipv6").(string),
		SrcPort:             d.Get("src_port").(string),
		SrcNetworkID:        d.Get("src_network_id").(string),
		SrcFirewallGroupIDs: srcFirewallGroupIDs,

		DstNetworkType:      d.Get("dst_network_type").(string),
		DstAddress:          d.Get("dst_address").(string),
		DstAddressIPV6:      d.Get("dst_address_ipv6").(string),
		DstPort:             d.Get("dst_port").(string),
		DstNetworkID:        d.Get("dst_network_id").(string),
		DstFirewallGroupIDs: dstFirewallGroupIDs,
	}, nil
}

func resourceFirewallRuleSetResourceData(resp *unifi.FirewallRule, d *schema.ResourceData, site string) diag.Diagnostics {
	d.Set("site", site)
	d.Set("name", resp.Name)
	d.Set("enabled", resp.Enabled)
	d.Set("action", resp.Action)
	d.Set("ruleset", resp.Ruleset)
	d.Set("rule_index", resp.RuleIndex)
	d.Set("protocol", resp.Protocol)
	d.Set("protocol_v6", resp.ProtocolV6)
	d.Set("icmp_typename", resp.ICMPTypename)
	d.Set("icmp_v6_typename", resp.ICMPv6Typename)
	d.Set("logging", resp.Logging)
	d.Set("ip_sec", resp.IPSec)
	d.Set("state_established", resp.StateEstablished)
	d.Set("state_invalid", resp.StateInvalid)
	d.Set("state_new", resp.StateNew)
	d.Set("state_related", resp.StateRelated)

	d.Set("src_network_type", resp.SrcNetworkType)
	d.Set("src_firewall_group_ids", utils.StringSliceToSet(resp.SrcFirewallGroupIDs))
	d.Set("src_mac", resp.SrcMACAddress)
	d.Set("src_address", resp.SrcAddress)
	d.Set("src_address_ipv6", resp.SrcAddressIPV6)
	d.Set("src_network_id", resp.SrcNetworkID)
	d.Set("src_port", resp.SrcPort)

	d.Set("dst_network_type", resp.DstNetworkType)
	d.Set("dst_firewall_group_ids", utils.StringSliceToSet(resp.DstFirewallGroupIDs))
	d.Set("dst_address", resp.DstAddress)
	d.Set("dst_address_ipv6", resp.DstAddressIPV6)
	d.Set("dst_network_id", resp.DstNetworkID)
	d.Set("dst_port", resp.DstPort)

	return nil
}

func resourceFirewallRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	id := d.Id()

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.GetFirewallRule(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFirewallRuleSetResourceData(resp, d, site)
}

func resourceFirewallRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	req, err := resourceFirewallRuleGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req.ID = d.Id()

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	req.SiteID = site

	resp, err := c.UpdateFirewallRule(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFirewallRuleSetResourceData(resp, d, site)
}

func resourceFirewallRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	id := d.Id()

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	err := c.DeleteFirewallRule(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		return nil
	}
	return diag.FromErr(err)
}

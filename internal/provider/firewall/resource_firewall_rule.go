package firewall

import (
	"context"
	"errors"
	"regexp"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/utils"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
)

var (
	firewallRuleProtocolRegexp       = regexp.MustCompile("^$|all|([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])|tcp_udp|ah|ax.25|dccp|ddp|egp|eigrp|encap|esp|etherip|fc|ggp|gre|hip|hmp|icmp|idpr-cmtp|idrp|igmp|igp|ip|ipcomp|ipencap|ipip|ipv6|ipv6-frag|ipv6-icmp|ipv6-nonxt|ipv6-opts|ipv6-route|isis|iso-tp4|l2tp|manet|mobility-header|mpls-in-ip|ospf|pim|pup|rdp|rohc|rspf|rsvp|sctp|shim6|skip|st|tcp|udp|udplite|vmtp|vrrp|wesp|xns-idp|xtp")
	firewallRuleProtocolV6Regexp     = regexp.MustCompile("^$|([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])|ah|all|dccp|eigrp|esp|gre|icmpv6|ipcomp|ipv6|ipv6-frag|ipv6-icmp|ipv6-nonxt|ipv6-opts|ipv6-route|isis|l2tp|manet|mobility-header|mpls-in-ip|ospf|pim|rsvp|sctp|shim6|tcp|tcp_udp|udp|vrrp")
	firewallRuleICMPv6TypenameRegexp = regexp.MustCompile("^$|address-unreachable|bad-header|beyond-scope|communication-prohibited|destination-unreachable|echo-reply|echo-request|failed-policy|neighbor-advertisement|neighbor-solicitation|no-route|packet-too-big|parameter-problem|port-unreachable|redirect|reject-route|router-advertisement|router-solicitation|time-exceeded|ttl-zero-during-reassembly|ttl-zero-during-transit|unknown-header-type|unknown-option")
)

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
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	req, err := resourceFirewallRuleGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	site, _ := d.Get("site").(string)
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
	srcGroupSet, _ := d.Get("src_firewall_group_ids").(*schema.Set)
	srcFirewallGroupIDs, err := utils.SetToStringSlice(srcGroupSet)
	if err != nil {
		return nil, err
	}

	dstGroupSet, _ := d.Get("dst_firewall_group_ids").(*schema.Set)
	dstFirewallGroupIDs, err := utils.SetToStringSlice(dstGroupSet)
	if err != nil {
		return nil, err
	}

	enabled, _ := d.Get("enabled").(bool)
	name, _ := d.Get("name").(string)
	action, _ := d.Get("action").(string)
	ruleset, _ := d.Get("ruleset").(string)
	ruleIndex, _ := d.Get("rule_index").(int)
	protocol, _ := d.Get("protocol").(string)
	protocolV6, _ := d.Get("protocol_v6").(string)
	icmpTypename, _ := d.Get("icmp_typename").(string)
	icmpV6Typename, _ := d.Get("icmp_v6_typename").(string)
	logging, _ := d.Get("logging").(bool)
	ipSec, _ := d.Get("ip_sec").(string)
	stateEstablished, _ := d.Get("state_established").(bool)
	stateInvalid, _ := d.Get("state_invalid").(bool)
	stateNew, _ := d.Get("state_new").(bool)
	stateRelated, _ := d.Get("state_related").(bool)

	srcNetworkType, _ := d.Get("src_network_type").(string)
	srcMAC, _ := d.Get("src_mac").(string)
	srcAddress, _ := d.Get("src_address").(string)
	srcAddressIPV6, _ := d.Get("src_address_ipv6").(string)
	srcPort, _ := d.Get("src_port").(string)
	srcNetworkID, _ := d.Get("src_network_id").(string)

	dstNetworkType, _ := d.Get("dst_network_type").(string)
	dstAddress, _ := d.Get("dst_address").(string)
	dstAddressIPV6, _ := d.Get("dst_address_ipv6").(string)
	dstPort, _ := d.Get("dst_port").(string)
	dstNetworkID, _ := d.Get("dst_network_id").(string)

	return &unifi.FirewallRule{
		Enabled:          enabled,
		Name:             name,
		Action:           action,
		Ruleset:          ruleset,
		RuleIndex:        ruleIndex,
		Protocol:         protocol,
		ProtocolV6:       protocolV6,
		ICMPTypename:     icmpTypename,
		ICMPv6Typename:   icmpV6Typename,
		Logging:          logging,
		IPSec:            ipSec,
		StateEstablished: stateEstablished,
		StateInvalid:     stateInvalid,
		StateNew:         stateNew,
		StateRelated:     stateRelated,

		SrcNetworkType:      srcNetworkType,
		SrcMACAddress:       srcMAC,
		SrcAddress:          srcAddress,
		SrcAddressIPV6:      srcAddressIPV6,
		SrcPort:             srcPort,
		SrcNetworkID:        srcNetworkID,
		SrcFirewallGroupIDs: srcFirewallGroupIDs,

		DstNetworkType:      dstNetworkType,
		DstAddress:          dstAddress,
		DstAddressIPV6:      dstAddressIPV6,
		DstPort:             dstPort,
		DstNetworkID:        dstNetworkID,
		DstFirewallGroupIDs: dstFirewallGroupIDs,
	}, nil
}

func resourceFirewallRuleSetResourceData(resp *unifi.FirewallRule, d *schema.ResourceData, site string) diag.Diagnostics {
	setters := []struct {
		key   string
		value interface{}
	}{
		{"site", site},
		{"name", resp.Name},
		{"enabled", resp.Enabled},
		{"action", resp.Action},
		{"ruleset", resp.Ruleset},
		{"rule_index", resp.RuleIndex},
		{"protocol", resp.Protocol},
		{"protocol_v6", resp.ProtocolV6},
		{"icmp_typename", resp.ICMPTypename},
		{"icmp_v6_typename", resp.ICMPv6Typename},
		{"logging", resp.Logging},
		{"ip_sec", resp.IPSec},
		{"state_established", resp.StateEstablished},
		{"state_invalid", resp.StateInvalid},
		{"state_new", resp.StateNew},
		{"state_related", resp.StateRelated},

		{"src_network_type", resp.SrcNetworkType},
		{"src_firewall_group_ids", utils.StringSliceToSet(resp.SrcFirewallGroupIDs)},
		{"src_mac", resp.SrcMACAddress},
		{"src_address", resp.SrcAddress},
		{"src_address_ipv6", resp.SrcAddressIPV6},
		{"src_network_id", resp.SrcNetworkID},
		{"src_port", resp.SrcPort},

		{"dst_network_type", resp.DstNetworkType},
		{"dst_firewall_group_ids", utils.StringSliceToSet(resp.DstFirewallGroupIDs)},
		{"dst_address", resp.DstAddress},
		{"dst_address_ipv6", resp.DstAddressIPV6},
		{"dst_network_id", resp.DstNetworkID},
		{"dst_port", resp.DstPort},
	}
	for _, s := range setters {
		if err := d.Set(s.key, s.value); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceFirewallRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	id := d.Id()

	site, _ := d.Get("site").(string)
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
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	req, err := resourceFirewallRuleGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req.ID = d.Id()

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	req.SiteID = site

	// go-unifi v1.9.2's updateFirewallRule converts a successful-but-empty PUT
	// response into unifi.ErrNotFound (see utils.ReReadOnUpdateNotFound / issue #98);
	// re-read to tell a spurious error from a genuine out-of-band deletion.
	resp, err := c.UpdateFirewallRule(ctx, site, req)
	resp, found, err := utils.ReReadOnUpdateNotFound(resp, err, func() (*unifi.FirewallRule, error) {
		return c.GetFirewallRule(ctx, site, req.ID)
	})
	if err != nil {
		return diag.FromErr(err)
	}
	if !found {
		d.SetId("")
		return nil
	}

	return resourceFirewallRuleSetResourceData(resp, d, site)
}

func resourceFirewallRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	id := d.Id()

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	err := c.DeleteFirewallRule(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		return nil
	}
	return diag.FromErr(err)
}

package device

import (
	"context"
	"errors"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/utils"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourcePortProfile() *schema.Resource {
	return &schema.Resource{
		Description: "The `unifi_port_profile` resource manages port profiles that can be applied to UniFi switch ports.\n\n" +
			"Port profiles define a collection of settings that can be applied to one or more switch ports, including:\n" +
			"  * Network and VLAN settings\n" +
			"  * Port speed and duplex settings\n" +
			"  * Security features like 802.1X authentication and port isolation\n" +
			"  * Rate limiting and QoS settings\n" +
			"  * Network protocols like LLDP and STP\n\n" +
			"Creating port profiles allows for consistent configuration across multiple switch ports and easier management of port settings.",

		CreateContext: resourcePortProfileCreate,
		ReadContext:   resourcePortProfileRead,
		UpdateContext: resourcePortProfileUpdate,
		DeleteContext: resourcePortProfileDelete,
		Importer: &schema.ResourceImporter{
			StateContext: base.ImportSiteAndID,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the port profile in the UniFi controller.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"site": {
				Description: "The name of the UniFi site where the port profile should be created. If not specified, the default site will be used.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},
			"autoneg": {
				Description: "Enable automatic negotiation of port speed and duplex settings. When enabled, this overrides manual speed and duplex settings. Recommended for most use cases.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"dot1x_ctrl": {
				Description: "802.1X port-based network access control (PNAC) mode. Valid values are:\n" +
					"  * `force_authorized` - Port allows all traffic, no authentication required (default)\n" +
					"  * `force_unauthorized` - Port blocks all traffic regardless of authentication\n" +
					"  * `auto` - Standard 802.1X authentication required before port access is granted\n" +
					"  * `mac_based` - Authentication based on client MAC address, useful for devices that don't support 802.1X\n" +
					"  * `multi_host` - Allows multiple devices after first successful authentication, common in VoIP phone setups\n\n" +
					"Use 'auto' for highest security, 'mac_based' for legacy devices, and 'multi_host' when daisy-chaining devices.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "force_authorized",
				ValidateFunc: validation.StringInSlice([]string{"auto", "force_authorized", "force_unauthorized", "mac_based", "multi_host"}, false),
			},
			"dot1x_idle_timeout": {
				Description:  "The number of seconds before an inactive authenticated MAC address is removed when using MAC-based 802.1X control. Range: 0-65535 seconds.",
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      300,
				ValidateFunc: validation.IntBetween(0, 65535),
			},
			"egress_rate_limit_kbps": {
				Description:  "The maximum outbound bandwidth allowed on the port in kilobits per second. Range: 64-9999999 kbps. Only applied when egress_rate_limit_kbps_enabled is true.",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(64, 9999999),
			},
			"egress_rate_limit_kbps_enabled": {
				Description: "Enable outbound bandwidth rate limiting on the port. When enabled, traffic will be limited to the rate specified in egress_rate_limit_kbps.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"excluded_network_ids": {
				Description: "List of network IDs to exclude when forward is set to 'customize'. This allows you to prevent specific networks from being accessible on ports using this profile.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"forward": {
				Description: "VLAN forwarding mode for the port. Valid values are:\n" +
					"  * `all` - Forward all VLANs (trunk port)\n" +
					"  * `native` - Only forward untagged traffic (access port)\n" +
					"  * `customize` - Forward selected VLANs (use with `excluded_network_ids`)\n" +
					"  * `disabled` - Disable VLAN forwarding\n\n" +
					"Examples:\n" +
					"  * Use 'all' for uplink ports or connections to VLAN-aware devices\n" +
					"  * Use 'native' for end-user devices or simple network connections\n" +
					"  * Use 'customize' to create a selective trunk port (e.g., for a server needing access to specific VLANs)",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "native",
				ValidateFunc: validation.StringInSlice([]string{"all", "native", "customize", "disabled"}, false),
			},
			"full_duplex": {
				Description: "Enable full-duplex mode when auto-negotiation is disabled. Full duplex allows simultaneous two-way communication.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"isolation": {
				Description: "Enable port isolation. When enabled, devices connected to ports with this profile cannot communicate with each other, providing enhanced security.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"lldpmed_enabled": {
				Description: "Enable Link Layer Discovery Protocol-Media Endpoint Discovery (LLDP-MED). This allows for automatic discovery and configuration of devices like VoIP phones.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"lldpmed_notify_enabled": {
				Description: "Enable LLDP-MED topology change notifications. When enabled:\n" +
					"* Network devices will be notified of topology changes\n" +
					"* Useful for VoIP phones and other LLDP-MED capable devices\n" +
					"* Helps maintain accurate network topology information\n" +
					"* Facilitates faster device configuration and provisioning",
				Type:     schema.TypeBool,
				Optional: true,
			},
			// TODO: rename to native_network_id
			"native_networkconf_id": {
				Description: "The ID of the network to use as the native (untagged) network on ports using this profile. " +
					"This is typically used for:\n" +
					"* Access ports where devices need untagged access\n" +
					"* Trunk ports to specify the native VLAN\n" +
					"* Management networks for network devices",
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Description: "A descriptive name for the port profile. Examples:\n" +
					"* 'AP-Trunk-Port' - For access point uplinks\n" +
					"* 'VoIP-Phone-Port' - For VoIP phone connections\n" +
					"* 'User-Access-Port' - For standard user connections\n" +
					"* 'IoT-Device-Port' - For IoT device connections",
				Type:     schema.TypeString,
				Optional: true,
			},
			"op_mode": {
				Description:  "The operation mode for the port profile. Can only be `switch`",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "switch",
				ValidateFunc: validation.StringInSlice([]string{"switch"}, false),
			},
			"poe_mode": {
				Description:  "The POE mode for the port profile. Can be one of `auto`, `passv24`, `passthrough` or `off`.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"auto", "passv24", "passthrough", "off"}, false),
			},
			"port_security_enabled": {
				Description: "Enable MAC address-based port security. When enabled:\n" +
					"* Only devices with specified MAC addresses can connect\n" +
					"* Unauthorized devices will be blocked\n" +
					"* Provides protection against unauthorized network access\n" +
					"* Must be used with port_security_mac_address list",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"port_security_mac_address": {
				Description: "List of allowed MAC addresses when port security is enabled. Each address should be:\n" +
					"* In standard format (e.g., 'aa:bb:cc:dd:ee:ff')\n" +
					"* Unique per device\n" +
					"* Verified to belong to authorized devices\n" +
					"Only effective when port_security_enabled is true",
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"priority_queue1_level": {
				Description: "Priority queue 1 level (0-100) for Quality of Service (QoS). Used for:\n" +
					"* Low-priority background traffic\n" +
					"* Bulk data transfers\n" +
					"* Non-time-sensitive applications\n" +
					"Higher values give more bandwidth to this queue",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 100),
			},
			"priority_queue2_level": {
				Description: "Priority queue 2 level (0-100) for Quality of Service (QoS). Used for:\n" +
					"* Standard user traffic\n" +
					"* Web browsing and email\n" +
					"* General business applications\n" +
					"Higher values give more bandwidth to this queue",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 100),
			},
			"priority_queue3_level": {
				Description: "Priority queue 3 level (0-100) for Quality of Service (QoS). Used for:\n" +
					"* High-priority traffic\n" +
					"* Voice and video conferencing\n" +
					"* Time-sensitive applications\n" +
					"Higher values give more bandwidth to this queue",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 100),
			},
			"priority_queue4_level": {
				Description: "Priority queue 4 level (0-100) for Quality of Service (QoS). Used for:\n" +
					"* Highest priority traffic\n" +
					"* Critical real-time applications\n" +
					"* Emergency communications\n" +
					"Higher values give more bandwidth to this queue",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 100),
			},
			"speed": {
				Description: "Port speed in Mbps when auto-negotiation is disabled. Common values:\n" +
					"* 10 - 10 Mbps (legacy devices)\n" +
					"* 100 - 100 Mbps (Fast Ethernet)\n" +
					"* 1000 - 1 Gbps (Gigabit Ethernet)\n" +
					"* 2500 - 2.5 Gbps (Multi-Gigabit)\n" +
					"* 5000 - 5 Gbps (Multi-Gigabit)\n" +
					"* 10000 - 10 Gbps (10 Gigabit)\n" +
					"Only used when autoneg is false",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntInSlice([]int{10, 100, 1000, 2500, 5000, 10000, 20000, 25000, 40000, 50000, 100000}),
			},
			"stormctrl_bcast_enabled": {
				Description: "Enable broadcast storm control. When enabled:\n" +
					"* Limits broadcast traffic to prevent network flooding\n" +
					"* Protects against broadcast storms\n" +
					"* Helps maintain network stability\n" +
					"Use with stormctrl_bcast_rate to set threshold",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"stormctrl_bcast_level": {
				Description:   "The broadcast Storm Control level for the port profile. Can be between 0 and 100.",
				Type:          schema.TypeInt,
				Optional:      true,
				ValidateFunc:  validation.IntBetween(0, 100),
				ConflictsWith: []string{"stormctrl_bcast_rate"},
			},
			"stormctrl_bcast_rate": {
				Description: "Maximum broadcast traffic rate in packets per second (0 - 14880000). Used to:\n" +
					"* Control broadcast traffic levels\n" +
					"* Prevent network congestion\n" +
					"* Balance between necessary broadcasts and network protection\n" +
					"Only effective when `stormctrl_bcast_enabled` is true",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 14880000),
			},
			"stormctrl_mcast_enabled": {
				Description: "Enable multicast storm control. When enabled:\n" +
					"* Limits multicast traffic to prevent network flooding\n" +
					"* Important for networks with multicast applications\n" +
					"* Helps maintain quality of service\n" +
					"Use with `stormctrl_mcast_rate` to set threshold",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"stormctrl_mcast_level": {
				Description:   "The multicast Storm Control level for the port profile. Can be between 0 and 100.",
				Type:          schema.TypeInt,
				Optional:      true,
				ValidateFunc:  validation.IntBetween(0, 100),
				ConflictsWith: []string{"stormctrl_mcast_rate"},
			},
			"stormctrl_mcast_rate": {
				Description: "Maximum multicast traffic rate in packets per second (0 - 14880000). Used to:\n" +
					"* Control multicast traffic levels\n" +
					"* Ensure bandwidth for critical multicast services\n" +
					"* Prevent multicast traffic from overwhelming the network\n" +
					"Only effective when stormctrl_mcast_enabled is true",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 14880000),
			},
			"stormctrl_type": {
				Description:  "The type of Storm Control to use for the port profile. Can be one of `level` or `rate`.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"level", "rate"}, false),
			},
			"stormctrl_ucast_enabled": {
				Description: "Enable unknown unicast storm control. When enabled:\n" +
					"* Limits unknown unicast traffic to prevent flooding\n" +
					"* Protects against MAC spoofing attacks\n" +
					"* Helps maintain network performance\n" +
					"Use with stormctrl_ucast_rate to set threshold",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"stormctrl_ucast_level": {
				Description:   "The unknown unicast Storm Control level for the port profile. Can be between 0 and 100.",
				Type:          schema.TypeInt,
				Optional:      true,
				ValidateFunc:  validation.IntBetween(0, 100),
				ConflictsWith: []string{"stormctrl_ucast_rate"},
			},
			"stormctrl_ucast_rate": {
				Description: "Maximum unknown unicast traffic rate in packets per second (0 - 14880000). Used to:\n" +
					"* Control unknown unicast traffic levels\n" +
					"* Prevent network saturation from unknown destinations\n" +
					"* Balance security with network usability\n" +
					"Only effective when stormctrl_ucast_enabled is true",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 14880000),
			},
			"stp_port_mode": {
				Description: "Spanning Tree Protocol (STP) configuration for the port. When enabled:\n" +
					"* Prevents network loops in switch-to-switch connections\n" +
					"* Provides automatic failover in redundant topologies\n" +
					"* Helps maintain network stability\n\n" +
					"Best practices:\n" +
					"* Enable on switch uplink ports\n" +
					"* Enable on ports connecting to other switches\n" +
					"* Can be disabled on end-device ports for faster initialization",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"tagged_vlan_mgmt": {
				Description: "VLAN tagging behavior for the port. Valid values are:\n" +
					"* `auto` - Automatically handle VLAN tags (recommended)\n" +
					"    - Intelligently manages tagged and untagged traffic\n" +
					"    - Best for most deployments\n" +
					"* `block_all` - Block all VLAN tagged traffic\n" +
					"    - Use for security-sensitive ports\n" +
					"    - Prevents VLAN hopping attacks\n" +
					"* `custom` - Custom VLAN configuration\n" +
					"    - Manual control over VLAN behavior\n" +
					"    - For specific VLAN requirements",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"auto", "block_all", "custom"}, false),
			},
			// TODO: rename to voice_network_id
			"voice_networkconf_id": {
				Description: "The ID of the network to use for Voice over IP (VoIP) traffic. Used for:\n" +
					"* Automatic VoIP VLAN configuration\n" +
					"* Voice traffic prioritization\n" +
					"* QoS settings for voice packets\n\n" +
					"Common scenarios:\n" +
					"* IP phone deployments with separate voice VLAN\n" +
					"* Unified communications systems\n" +
					"* Converged voice/data networks\n\n" +
					"Works in conjunction with LLDP-MED for automatic phone provisioning.",
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourcePortProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	req, err := resourcePortProfileGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	resp, err := c.CreatePortProfile(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return resourcePortProfileSetResourceData(resp, d, site)
}

func resourcePortProfileGetResourceData(d *schema.ResourceData) (*unifi.PortProfile, error) {
	portSecurityMacAddress, err := utils.SetToStringSlice(d.Get("port_security_mac_address").(*schema.Set))
	if err != nil {
		return nil, err
	}

	excludedNetworkIDs, err := utils.SetToStringSlice(d.Get("excluded_network_ids").(*schema.Set))
	if err != nil {
		return nil, err
	}

	return &unifi.PortProfile{
		Autoneg:                      d.Get("autoneg").(bool),
		Dot1XCtrl:                    d.Get("dot1x_ctrl").(string),
		Dot1XIDleTimeout:             d.Get("dot1x_idle_timeout").(int),
		EgressRateLimitKbps:          d.Get("egress_rate_limit_kbps").(int),
		EgressRateLimitKbpsEnabled:   d.Get("egress_rate_limit_kbps_enabled").(bool),
		ExcludedNetworkIDs:           excludedNetworkIDs,
		Forward:                      d.Get("forward").(string),
		FullDuplex:                   d.Get("full_duplex").(bool),
		Isolation:                    d.Get("isolation").(bool),
		LldpmedEnabled:               d.Get("lldpmed_enabled").(bool),
		LldpmedNotifyEnabled:         d.Get("lldpmed_notify_enabled").(bool),
		NATiveNetworkID:              d.Get("native_networkconf_id").(string),
		Name:                         d.Get("name").(string),
		OpMode:                       d.Get("op_mode").(string),
		PoeMode:                      d.Get("poe_mode").(string),
		PortSecurityEnabled:          d.Get("port_security_enabled").(bool),
		PortSecurityMACAddress:       portSecurityMacAddress,
		PriorityQueue1Level:          d.Get("priority_queue1_level").(int),
		PriorityQueue2Level:          d.Get("priority_queue2_level").(int),
		PriorityQueue3Level:          d.Get("priority_queue3_level").(int),
		PriorityQueue4Level:          d.Get("priority_queue4_level").(int),
		Speed:                        d.Get("speed").(int),
		StormctrlBroadcastastEnabled: d.Get("stormctrl_bcast_enabled").(bool),
		StormctrlBroadcastastLevel:   d.Get("stormctrl_bcast_level").(int),
		StormctrlBroadcastastRate:    d.Get("stormctrl_bcast_rate").(int),
		StormctrlMcastEnabled:        d.Get("stormctrl_mcast_enabled").(bool),
		StormctrlMcastLevel:          d.Get("stormctrl_mcast_level").(int),
		StormctrlMcastRate:           d.Get("stormctrl_mcast_rate").(int),
		StormctrlType:                d.Get("stormctrl_type").(string),
		StormctrlUcastEnabled:        d.Get("stormctrl_ucast_enabled").(bool),
		StormctrlUcastLevel:          d.Get("stormctrl_ucast_level").(int),
		StormctrlUcastRate:           d.Get("stormctrl_ucast_rate").(int),
		StpPortMode:                  d.Get("stp_port_mode").(bool),
		TaggedVLANMgmt:               d.Get("tagged_vlan_mgmt").(string),
		VoiceNetworkID:               d.Get("voice_networkconf_id").(string),
	}, nil
}

func resourcePortProfileSetResourceData(resp *unifi.PortProfile, d *schema.ResourceData, site string) diag.Diagnostics {
	d.Set("site", site)
	d.Set("autoneg", resp.Autoneg)
	d.Set("dot1x_ctrl", resp.Dot1XCtrl)
	d.Set("dot1x_idle_timeout", resp.Dot1XIDleTimeout)
	d.Set("egress_rate_limit_kbps", resp.EgressRateLimitKbps)
	d.Set("egress_rate_limit_kbps_enabled", resp.EgressRateLimitKbpsEnabled)
	d.Set("excluded_network_ids", utils.StringSliceToSet(resp.ExcludedNetworkIDs))
	d.Set("forward", resp.Forward)
	d.Set("full_duplex", resp.FullDuplex)
	d.Set("isolation", resp.Isolation)
	d.Set("lldpmed_enabled", resp.LldpmedEnabled)
	d.Set("lldpmed_notify_enabled", resp.LldpmedNotifyEnabled)
	d.Set("native_networkconf_id", resp.NATiveNetworkID)
	d.Set("name", resp.Name)
	d.Set("op_mode", resp.OpMode)
	d.Set("poe_mode", resp.PoeMode)
	d.Set("port_security_enabled", resp.PortSecurityEnabled)
	d.Set("port_security_mac_address", utils.StringSliceToSet(resp.PortSecurityMACAddress))
	d.Set("priority_queue1_level", resp.PriorityQueue1Level)
	d.Set("priority_queue2_level", resp.PriorityQueue2Level)
	d.Set("priority_queue3_level", resp.PriorityQueue3Level)
	d.Set("priority_queue4_level", resp.PriorityQueue4Level)
	d.Set("speed", resp.Speed)
	d.Set("stormctrl_bcast_enabled", resp.StormctrlBroadcastastEnabled)
	d.Set("stormctrl_bcast_level", resp.StormctrlBroadcastastLevel)
	d.Set("stormctrl_bcast_rate", resp.StormctrlBroadcastastRate)
	d.Set("stormctrl_mcast_enabled", resp.StormctrlMcastEnabled)
	d.Set("stormctrl_mcast_level", resp.StormctrlMcastLevel)
	d.Set("stormctrl_mcast_rate", resp.StormctrlMcastRate)
	d.Set("stormctrl_type", resp.StormctrlType)
	d.Set("stormctrl_ucast_enabled", resp.StormctrlUcastEnabled)
	d.Set("stormctrl_ucast_level", resp.StormctrlUcastLevel)
	d.Set("stormctrl_ucast_rate", resp.StormctrlUcastRate)
	d.Set("stp_port_mode", resp.StpPortMode)
	d.Set("tagged_vlan_mgmt", resp.TaggedVLANMgmt)
	d.Set("voice_networkconf_id", resp.VoiceNetworkID)

	return nil
}

func resourcePortProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	id := d.Id()

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	resp, err := c.GetPortProfile(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	return resourcePortProfileSetResourceData(resp, d, site)
}

func resourcePortProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	req, err := resourcePortProfileGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req.ID = d.Id()

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	req.SiteID = site

	resp, err := c.UpdatePortProfile(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourcePortProfileSetResourceData(resp, d, site)
}

func resourcePortProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	id := d.Id()

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	err := c.DeletePortProfile(ctx, site, id)
	return diag.FromErr(err)
}

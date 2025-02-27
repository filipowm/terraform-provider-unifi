package network

import (
	"context"
	"errors"
	"fmt"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/utils"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	wlanValidMinimumDataRate2g = []int{1000, 2000, 5500, 6000, 9000, 11000, 12000, 18000, 24000, 36000, 48000, 54000}
	wlanValidMinimumDataRate5g = []int{6000, 9000, 12000, 18000, 24000, 36000, 48000, 54000}
)

func ResourceWLAN() *schema.Resource {
	return &schema.Resource{
		Description: "The `unifi_wlan` resource manages wireless networks (SSIDs) on UniFi access points.\n\n" +
			"This resource allows you to create and manage WiFi networks with various security options including WPA2, WPA3, " +
			"and enterprise authentication. You can configure features such as guest policies, minimum data rates, band steering, " +
			"and scheduled availability.\n\n" +
			"Each WLAN can be customized with different security settings, VLAN assignments, and client options to meet specific " +
			"networking requirements.",

		CreateContext: resourceWLANCreate,
		ReadContext:   resourceWLANRead,
		UpdateContext: resourceWLANUpdate,
		DeleteContext: resourceWLANDelete,
		Importer: &schema.ResourceImporter{
			StateContext: base.ImportSiteAndID,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the wireless network in the UniFi controller.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"site": {
				Description: "The name of the UniFi site where the wireless network should be created. If not specified, the default site will be used.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},
			"name": {
				Description:  "The SSID (network name) that will be broadcast by the access points. Must be between 1 and 32 characters long.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 32),
			},
			"user_group_id": {
				Description: "The ID of the user group that defines the rate limiting and firewall rules for clients on this network.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"security": {
				Description: "The security protocol for the wireless network. Valid values are:\n" +
					"  * `wpapsk` - WPA Personal (PSK) with WPA2/WPA3 options\n" +
					"  * `wpaeap` - WPA Enterprise (802.1x)\n" +
					"  * `open` - Open network (no encryption)",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"wpapsk", "wpaeap", "open"}, false),
			},
			"wpa3_support": {
				Description: "Enable WPA3 security protocol. Requires security to be set to `wpapsk` and PMF mode to be enabled. WPA3 provides enhanced security features over WPA2.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"wpa3_transition": {
				Description: "Enable WPA3 transition mode, which allows both WPA2 and WPA3 clients to connect. This provides backward compatibility while gradually transitioning to WPA3." +
					" Requires security to be set to `wpapsk` and `wpa3_support` to be true.",
				Type:     schema.TypeBool,
				Optional: true,
			},
			"pmf_mode": {
				Description: "Protected Management Frames (PMF) mode. It cannot be disabled if using WPA3. Valid values are:\n" +
					"  * `required` - All clients must support PMF (required for WPA3)\n" +
					"  * `optional` - Clients can optionally use PMF (recommended when transitioning from WPA2 to WPA3)\n" +
					"  * `disabled` - PMF is disabled (not compatible with WPA3)",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"required", "optional", "disabled"}, false),
				Default:      "disabled",
			},
			"passphrase": {
				Description: "The WPA pre-shared key (password) for the network. Required when security is not set to `open`.",
				Type:        schema.TypeString,
				// only required if security != open
				Optional:  true,
				Sensitive: true,
			},
			"hide_ssid": {
				Description: "When enabled, the access points will not broadcast the network name (SSID). Clients will need to manually enter the SSID to connect.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"is_guest": {
				Description: "Mark this as a guest network. Guest networks are isolated from other networks and can have special restrictions like captive portals.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"multicast_enhance": {
				Description: "Enable multicast enhancement to convert multicast traffic to unicast for better reliability and performance, especially for applications like video streaming.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"mac_filter_enabled": {
				Description: "Enable MAC address filtering to control network access based on client MAC addresses. Works in conjunction with `mac_filter_list` and `mac_filter_policy`.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"mac_filter_list": {
				Description: "List of MAC addresses to filter in XX:XX:XX:XX:XX:XX format. Only applied when `mac_filter_enabled` is true. MAC addresses are case-insensitive.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateFunc:     validation.StringMatch(utils.MacAddressRegexp, "Mac address is invalid"),
					DiffSuppressFunc: utils.MacDiffSuppressFunc,
				},
			},
			"mac_filter_policy": {
				Description: "MAC address filter policy. Valid values are:\n" +
					"  * `allow` - Only allow listed MAC addresses\n" +
					"  * `deny` - Block listed MAC addresses",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "deny",
				ValidateFunc: validation.StringInSlice([]string{"allow", "deny"}, false),
			},
			"radius_profile_id": {
				Description: "ID of the RADIUS profile to use for WPA Enterprise authentication (when security is 'wpaeap'). Reference existing profiles using the `unifi_radius_profile` data source.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"schedule": {
				Description: "Time-based access control configuration for the wireless network. Allows automatic enabling/disabling of the network on specified schedules.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"day_of_week": {
							Description:  "Day of week. Valid values: `sun`, `mon`, `tue`, `wed`, `thu`, `fri`, `sat`.",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"}, false),
						},
						"start_hour": {
							Description:  "Start hour in 24-hour format (0-23).",
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 23),
						},
						"start_minute": {
							Description:  "Start minute (0-59).",
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntBetween(0, 59),
						},
						"duration": {
							Description:  "Duration in minutes that the network should remain active.",
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"name": {
							Description: "Friendly name for this schedule block (e.g., 'Business Hours', 'Weekend Access').",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
			"no2ghz_oui": {
				Description: "When enabled, devices from specific manufacturers (identified by their OUI - Organizationally Unique Identifier) " +
					"will be prevented from connecting on 2.4GHz and forced to use 5GHz. This improves overall network performance by " +
					"ensuring capable devices use the less congested 5GHz band. Common examples include newer smartphones and laptops.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"l2_isolation": {
				Description: "Isolates wireless clients from each other at layer 2 (ethernet) level. When enabled, devices on this WLAN " +
					"cannot communicate directly with each other, improving security especially for guest networks or IoT devices. " +
					"Each client can only communicate with the gateway/router.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"proxy_arp": {
				Description: "Enable ARP proxy on this WLAN. When enabled, the UniFi controller will respond to ARP requests on behalf " +
					"of clients, reducing broadcast traffic and potentially improving network performance. This is particularly useful " +
					"in high-density wireless environments.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"bss_transition": {
				Description: "Enable BSS Transition Management to help clients roam between APs more efficiently.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"uapsd": {
				Description: "Enable Unscheduled Automatic Power Save Delivery to improve battery life for mobile devices.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"fast_roaming_enabled": {
				Description: "Enable 802.11r Fast BSS Transition for seamless roaming between APs. Requires client device support.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"minimum_data_rate_2g_kbps": {
				Description: "Minimum data rate for 2.4GHz devices in Kbps. Use `0` to disable. Valid values: " +
					utils.MarkdownValueListInt(wlanValidMinimumDataRate2g),
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntInSlice(append([]int{0}, wlanValidMinimumDataRate2g...)),
			},
			"minimum_data_rate_5g_kbps": {
				Description: "Minimum data rate for 5GHz devices in Kbps. Use `0` to disable. Valid values: " +
					utils.MarkdownValueListInt(wlanValidMinimumDataRate5g),
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntInSlice(append([]int{0}, wlanValidMinimumDataRate5g...)),
			},
			"wlan_band": {
				Description: "Radio band selection. Valid values:\n" +
					"  * `both` - Both 2.4GHz and 5GHz (default)\n" +
					"  * `2g` - 2.4GHz only\n" +
					"  * `5g` - 5GHz only",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"2g", "5g", "both"}, false),
				Default:      "both",
			},
			"network_id": {
				Description: "ID of the network (VLAN) for this SSID. Used to assign the WLAN to a specific network segment.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"ap_group_ids": {
				Description: "IDs of the AP groups that should broadcast this SSID. Used to control which access points broadcast this network.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceWLANGetResourceData(d *schema.ResourceData, meta interface{}) (*unifi.WLAN, error) {
	c := meta.(*base.Client)

	security := d.Get("security").(string)
	passphrase := d.Get("passphrase").(string)
	switch security {
	case "open":
		passphrase = ""
	}

	pmf := d.Get("pmf_mode").(string)
	wpa3 := d.Get("wpa3_support").(bool)
	wpa3Transition := d.Get("wpa3_transition").(bool)
	switch security {
	case "wpapsk":
		// nothing
	default:
		if wpa3 || wpa3Transition {
			return nil, fmt.Errorf("wpa3_support and wpa3_transition are only valid for security type wpapsk")
		}
	}
	if !c.SupportsWPA3() {
		if wpa3 || wpa3Transition {
			return nil, fmt.Errorf("WPA 3 support is not available on controller version %q, you must be on %q or higher", c.Version, base.ControllerVersionWPA3)
		}
	}

	if wpa3Transition && pmf == "disabled" {
		return nil, fmt.Errorf("WPA 3 transition mode requires pmf_mode to be turned on.")
	} else if wpa3 && !wpa3Transition && pmf != "required" {
		return nil, fmt.Errorf("For WPA 3 you must set pmf_mode to required.")
	}

	macFilterEnabled := d.Get("mac_filter_enabled").(bool)
	macFilterList, err := utils.SetToStringSlice(d.Get("mac_filter_list").(*schema.Set))
	if err != nil {
		return nil, err
	}
	if !macFilterEnabled {
		macFilterList = nil
	}

	// version specific fields and validation
	networkID := d.Get("network_id").(string)
	apGroupIDs, err := utils.SetToStringSlice(d.Get("ap_group_ids").(*schema.Set))
	if err != nil {
		return nil, err
	}
	wlanBand := d.Get("wlan_band").(string)

	schedule, err := listToSchedules(d.Get("schedule").([]interface{}))
	if err != nil {
		return nil, fmt.Errorf("unable to process schedule block: %w", err)
	}

	minrateSettingPreference := "auto"
	if d.Get("minimum_data_rate_2g_kbps").(int) != 0 || d.Get("minimum_data_rate_5g_kbps").(int) != 0 {
		if d.Get("minimum_data_rate_2g_kbps").(int) == 0 || d.Get("minimum_data_rate_5g_kbps").(int) == 0 {
			// this is really only true I think in >= 7.2, but easier to just apply this in general
			return nil, fmt.Errorf("you must set minimum data rates on both 2g and 5g if setting either")
		}
		minrateSettingPreference = "manual"
	}

	return &unifi.WLAN{
		Name:                    d.Get("name").(string),
		XPassphrase:             passphrase,
		HideSSID:                d.Get("hide_ssid").(bool),
		IsGuest:                 d.Get("is_guest").(bool),
		NetworkID:               networkID,
		ApGroupIDs:              apGroupIDs,
		UserGroupID:             d.Get("user_group_id").(string),
		Security:                security,
		WPA3Support:             wpa3,
		WPA3Transition:          wpa3Transition,
		MulticastEnhanceEnabled: d.Get("multicast_enhance").(bool),
		MACFilterEnabled:        macFilterEnabled,
		MACFilterList:           macFilterList,
		MACFilterPolicy:         d.Get("mac_filter_policy").(string),
		RADIUSProfileID:         d.Get("radius_profile_id").(string),
		ScheduleWithDuration:    schedule,
		ScheduleEnabled:         len(schedule) > 0,
		WLANBand:                wlanBand,
		PMFMode:                 pmf,

		// TODO: add to schema
		WPAEnc:             "ccmp",
		WPAMode:            "wpa2",
		Enabled:            true,
		NameCombineEnabled: true,

		GroupRekey:         3600,
		DTIMMode:           "default",
		No2GhzOui:          d.Get("no2ghz_oui").(bool),
		L2Isolation:        d.Get("l2_isolation").(bool),
		ProxyArp:           d.Get("proxy_arp").(bool),
		BssTransition:      d.Get("bss_transition").(bool),
		UapsdEnabled:       d.Get("uapsd").(bool),
		FastRoamingEnabled: d.Get("fast_roaming_enabled").(bool),

		MinrateSettingPreference: minrateSettingPreference,

		MinrateNgEnabled:      d.Get("minimum_data_rate_2g_kbps").(int) != 0,
		MinrateNgDataRateKbps: d.Get("minimum_data_rate_2g_kbps").(int),

		MinrateNaEnabled:      d.Get("minimum_data_rate_5g_kbps").(int) != 0,
		MinrateNaDataRateKbps: d.Get("minimum_data_rate_5g_kbps").(int),
	}, nil
}

func resourceWLANCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	req, err := resourceWLANGetResourceData(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.CreateWLAN(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return resourceWLANSetResourceData(resp, d, meta, site)
}

func resourceWLANSetResourceData(resp *unifi.WLAN, d *schema.ResourceData, meta interface{}, site string) diag.Diagnostics {
	// c := meta.(*provider.Client)
	security := resp.Security
	passphrase := resp.XPassphrase
	wpa3 := false
	wpa3Transition := false
	switch security {
	case "open":
		passphrase = ""
	case "wpapsk":
		wpa3 = resp.WPA3Support
		wpa3Transition = resp.WPA3Transition
	}

	macFilterEnabled := resp.MACFilterEnabled
	var macFilterList *schema.Set
	macFilterPolicy := "deny"
	if macFilterEnabled {
		macFilterList = utils.StringSliceToSet(resp.MACFilterList)
		macFilterPolicy = resp.MACFilterPolicy
	}

	apGroupIDs := utils.StringSliceToSet(resp.ApGroupIDs)

	schedule := listFromSchedules(resp.ScheduleWithDuration)

	d.Set("site", site)
	d.Set("name", resp.Name)
	d.Set("user_group_id", resp.UserGroupID)
	d.Set("passphrase", passphrase)
	d.Set("hide_ssid", resp.HideSSID)
	d.Set("is_guest", resp.IsGuest)
	d.Set("security", security)
	d.Set("wpa3_support", wpa3)
	d.Set("wpa3_transition", wpa3Transition)
	d.Set("multicast_enhance", resp.MulticastEnhanceEnabled)
	d.Set("mac_filter_enabled", macFilterEnabled)
	d.Set("mac_filter_list", macFilterList)
	d.Set("mac_filter_policy", macFilterPolicy)
	d.Set("radius_profile_id", resp.RADIUSProfileID)
	d.Set("schedule", schedule)
	d.Set("wlan_band", resp.WLANBand)
	d.Set("no2ghz_oui", resp.No2GhzOui)
	d.Set("l2_isolation", resp.L2Isolation)
	d.Set("proxy_arp", resp.ProxyArp)
	d.Set("bss_transition", resp.BssTransition)
	d.Set("uapsd", resp.UapsdEnabled)
	d.Set("fast_roaming_enabled", resp.FastRoamingEnabled)
	d.Set("ap_group_ids", apGroupIDs)
	d.Set("network_id", resp.NetworkID)
	d.Set("pmf_mode", resp.PMFMode)
	if resp.MinrateSettingPreference != "auto" && resp.MinrateNgEnabled {
		d.Set("minimum_data_rate_2g_kbps", resp.MinrateNgDataRateKbps)
	} else {
		d.Set("minimum_data_rate_2g_kbps", 0)
	}
	if resp.MinrateSettingPreference != "auto" && resp.MinrateNaEnabled {
		d.Set("minimum_data_rate_5g_kbps", resp.MinrateNaDataRateKbps)
	} else {
		d.Set("minimum_data_rate_5g_kbps", 0)
	}

	return nil
}

func resourceWLANRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	id := d.Id()
	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.GetWLAN(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceWLANSetResourceData(resp, d, meta, site)
}

func resourceWLANUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	req, err := resourceWLANGetResourceData(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	req.ID = d.Id()
	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	req.SiteID = site

	resp, err := c.UpdateWLAN(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceWLANSetResourceData(resp, d, meta, site)
}

func resourceWLANDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	id := d.Id()
	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	err := c.DeleteWLAN(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		return nil
	}
	return diag.FromErr(err)
}

func listToSchedules(list []interface{}) ([]unifi.WLANScheduleWithDuration, error) {
	schedules := make([]unifi.WLANScheduleWithDuration, 0, len(list))
	for _, item := range list {
		data, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected data in block")
		}
		ss := toSchedule(data)
		schedules = append(schedules, ss)
	}
	return schedules, nil
}

func toSchedule(data map[string]interface{}) unifi.WLANScheduleWithDuration {
	// TODO: error check these?
	dow := data["day_of_week"].(string)
	startHour := data["start_hour"].(int)
	startMinute := data["start_minute"].(int)
	duration := data["duration"].(int)
	name := data["name"].(string)

	return unifi.WLANScheduleWithDuration{
		StartDaysOfWeek: []string{dow},
		StartHour:       startHour,
		StartMinute:     startMinute,
		DurationMinutes: duration,
		Name:            name,
	}
}

func fromSchedule(dow string, s unifi.WLANScheduleWithDuration) map[string]interface{} {
	return map[string]interface{}{
		"day_of_week":  dow,
		"start_hour":   s.StartHour,
		"start_minute": s.StartMinute,
		"duration":     s.DurationMinutes,
		"name":         s.Name,
	}
}

func listFromSchedules(ss []unifi.WLANScheduleWithDuration) []interface{} {
	// this explodes days of week lists in to individual schedules
	list := make([]interface{}, 0, len(ss))
	for _, s := range ss {
		for _, dow := range s.StartDaysOfWeek {
			v := fromSchedule(dow, s)
			list = append(list, v)
		}
	}
	return list
}

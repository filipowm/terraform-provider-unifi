package device

import (
	"context"
	"errors"
	"fmt"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/utils"
	"strconv"
	"strings"
	"time"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceDevice() *schema.Resource {
	return &schema.Resource{
		Description: "The `unifi_device` resource manages UniFi network devices such as access points, switches, gateways, etc.\n\n" +
			"Devices must first be adopted by the UniFi controller before they can be managed through Terraform. " +
			"This resource cannot create new devices, but instead allows you to manage existing devices that have already been adopted. " +
			"The recommended approach is to adopt devices through the UniFi controller UI first, then import them into Terraform using the device's MAC address.\n\n" +
			"This resource supports managing device names, port configurations, and other device-specific settings.",

		CreateContext: resourceDeviceCreate,
		ReadContext:   resourceDeviceRead,
		UpdateContext: resourceDeviceUpdate,
		DeleteContext: resourceDeviceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDeviceImport,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the device in the UniFi controller.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"site": {
				Description: "The name of the UniFi site where the device is located. If not specified, the default site will be used.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},
			"mac": {
				Description:      "The MAC address of the device in standard format (e.g., 'aa:bb:cc:dd:ee:ff'). This is used to identify and manage specific devices that have already been adopted by the controller.",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				DiffSuppressFunc: utils.MacDiffSuppressFunc,
				ValidateFunc:     validation.StringMatch(utils.MacAddressRegexp, "Mac address is invalid"),
			},
			"name": {
				Description: "A friendly name for the device that will be displayed in the UniFi controller UI. Examples:\n" +
					"* 'Office-AP-1' for an access point\n" +
					"* 'Core-Switch-01' for a switch\n" +
					"* 'Main-Gateway' for a gateway\n" +
					"Choose descriptive names that indicate location and purpose.",
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"disabled": {
				Description: "Whether the device is administratively disabled. When true, the device will not forward traffic or provide services.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"switch_vlan_enabled": {
				Description: "Whether per-port VLAN configuration is enabled on the device. Required for `port_override` blocks with VLAN-tagging profiles (e.g. an IoT-VLAN `port_profile_id`) to actually take effect on access points that expose passthrough Ethernet ports (UAP-UHDIW and similar in-wall units). " +
					"Switches honor port profile VLAN bindings unconditionally; APs ignore them unless this flag is true. " +
					"Note: the underlying field uses `omitempty` so setting this to `false` has no effect — once enabled on a device, it can only be disabled via the UI.",
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				// The controller ignores attempts to disable this (`omitempty` drops
				// a `false` from the payload), so a `true` -> `false` change would
				// otherwise read back as `true` and produce a perpetual diff.
				// Suppress that one transition to match the API's write-once behavior.
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old == "true" && new == "false"
				},
			},
			"port_override": {
				// TODO: this should really be a map or something when possible in the SDK
				// see https://github.com/hashicorp/terraform-plugin-sdk/issues/62
				Description: "A list of port-specific configuration overrides for UniFi switches. This allows you to customize individual port settings such as:\n" +
					"  * Port names and labels for easy identification\n" +
					"  * Port profiles for VLAN and security settings\n" +
					"  * Per-port native (untagged) and tagged VLAN behavior, inline, without authoring a `unifi_port_profile`\n" +
					"  * Operating modes for special functions\n\n" +
					"Common use cases include:\n" +
					"  * Setting up trunk ports for inter-switch connections\n" +
					"  * Configuring PoE settings for powered devices\n" +
					"  * Creating mirrored ports for network monitoring\n" +
					"  * Setting up link aggregation between switches or servers\n\n" +
					"**Warning:** the controller stores port overrides as a single array on the device and the provider replaces the " +
					"entire array on every apply. Any port whose override is set outside Terraform (e.g. via the UniFi UI or another " +
					"tool) and is NOT declared here will have its override reset to the controller default on the next apply. Declare " +
					"every port you want overridden.\n\n" +
					"**Tagged-VLAN model:** there is no positive \"allowed VLANs\" list. With `forward = \"customize\"`, tagged traffic is " +
					"*all* networks **minus** the ones listed in `excluded_network_ids`, so an empty `excluded_network_ids` means \"trunk " +
					"everything\", not \"trunk nothing\".",
				Type:     schema.TypeSet,
				Optional: true,
				// Key set identity by port `number` only (see portOverrideSetHash):
				// the controller auto-populates/echoes per-port VLAN fields
				// (e.g. setting_preference or a native VLAN) on override entries
				// the user never declared them on. Hashing the whole element would
				// let such an echo change an element's identity and churn the set
				// (perpetual add/remove diff). Combined with the Optional+Computed
				// VLAN attributes below, an undeclared field reads back the
				// controller value without producing a diff.
				Set: portOverrideSetHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"number": {
							Description: "The physical port number on the switch to configure.",
							Type:        schema.TypeInt,
							Required:    true,
						},
						"name": {
							Description: "A friendly name for the port that will be displayed in the UniFi controller UI. Examples:\n" +
								"  * 'Uplink to Core Switch'\n" +
								"  * 'Conference Room AP'\n" +
								"  * 'Server LACP Group 1'\n" +
								"  * 'VoIP Phone Port'",
							Type:     schema.TypeString,
							Optional: true,
						},
						"port_profile_id": {
							Description: "The ID of a pre-configured port profile to apply to this port. Port profiles define settings like VLANs, PoE, and other port-specific configurations.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"op_mode": {
							Description: "The operating mode of the port. Valid values are:\n" +
								"  * `switch` - Normal switching mode (default)\n" +
								"    - Standard port operation for connecting devices\n" +
								"    - Supports VLANs and all standard switching features\n" +
								"  * `mirror` - Port mirroring for traffic analysis\n" +
								"    - Copies traffic from other ports for monitoring\n" +
								"    - Useful for network troubleshooting and security\n" +
								"  * `aggregate` - Link aggregation/bonding mode\n" +
								"    - Combines multiple ports for increased bandwidth\n" +
								"    - Used for switch uplinks or high-bandwidth servers",
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "switch",
							ValidateFunc: validation.StringInSlice([]string{"switch", "mirror", "aggregate"}, false),
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if old == "" && new == "switch" {
									return true
								}
								return false
							},
						},
						"poe_mode": {
							Description: "The Power over Ethernet (PoE) mode for the port. Valid values are:\n" +
								"* `auto` - Automatically detect and power PoE devices (recommended)\n" +
								"  - Provides power based on device negotiation\n" +
								"  - Safest option for most PoE devices\n" +
								"* `pasv24` - Passive 24V PoE\n" +
								"  - For older UniFi devices requiring passive 24V\n" +
								"  - Use with caution to avoid damage\n" +
								"* `passthrough` - PoE passthrough mode\n" +
								"  - For daisy-chaining PoE devices\n" +
								"  - Available on select UniFi switches\n" +
								"* `off` - Disable PoE on the port\n" +
								"  - For non-PoE devices\n" +
								"  - To prevent unwanted power delivery",
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"auto", "pasv24", "passthrough", "off"}, false),
						},
						"aggregate_num_ports": {
							Description: "The number of ports to include in a link aggregation group (LAG). Valid range: 2-8 ports. Used when:\n" +
								"* Creating switch-to-switch uplinks for increased bandwidth\n" +
								"* Setting up high-availability connections\n" +
								"* Connecting to servers requiring more bandwidth\n" +
								"Note: All ports in the LAG must be sequential and have matching configurations.",
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(2, 8),
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if old == strconv.Itoa(0) && new == "" {
									return true
								}
								return false
							},
						},
						"native_networkconf_id": {
							Description: "The ID of the network to use as the native (untagged) network on this port. " +
								"This is typically used for:\n" +
								"* Access ports where devices need untagged access\n" +
								"* Trunk ports to specify the native VLAN\n" +
								"* Management networks for network devices\n\n" +
								"Computed when not set, so the controller's current value (which it may auto-populate on a port) " +
								"is preserved without producing a diff. Note: the underlying field uses `omitempty`, so once set it " +
								"cannot be cleared back to empty through Terraform — change it to another network ID instead.",
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"tagged_vlan_mgmt": {
							Description: "VLAN tagging behavior for the port. Valid values are:\n" +
								"* `auto` - Automatically handle VLAN tags (recommended)\n" +
								"* `block_all` - Block all VLAN tagged traffic\n" +
								"* `custom` - Custom VLAN configuration (use with `forward = \"customize\"` and `excluded_network_ids`)\n\n" +
								"Computed when not set, so the controller's current value is preserved without producing a diff. " +
								"Note: the underlying field uses `omitempty`, so once set it cannot be cleared back to empty " +
								"through Terraform — change it to another value instead.",
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"auto", "block_all", "custom"}, false),
						},
						"forward": {
							Description: "VLAN forwarding mode for the port. Valid values are:\n" +
								"  * `all` - Forward all VLANs (trunk port)\n" +
								"  * `native` - Only forward untagged traffic (access port)\n" +
								"  * `customize` - Forward selected VLANs (use with `excluded_network_ids`)\n" +
								"  * `disabled` - Disable VLAN forwarding\n\n" +
								"This attribute has NO default: leaving it unset keeps the port's existing forwarding behavior " +
								"(the value is computed from the controller). Note: the underlying field uses `omitempty`, so once " +
								"set it cannot be cleared back to empty through Terraform — change it to another value instead.",
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"all", "native", "customize", "disabled"}, false),
						},
						"excluded_network_ids": {
							Description: "Set of network IDs to exclude when `forward = \"customize\"`. Tagged traffic on the port is " +
								"*all* networks minus the ones listed here, so an empty set means \"trunk everything\". " +
								"Computed when not set, so the controller's current exclusions are preserved without producing a diff.",
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"voice_networkconf_id": {
							Description: "The ID of the network to use for Voice over IP (VoIP) traffic on this port, for automatic " +
								"voice-VLAN assignment in conjunction with LLDP-MED.\n\n" +
								"Computed when not set, so the controller's current value is preserved without producing a diff. " +
								"Note: the underlying field uses `omitempty`, so once set it cannot be cleared back to empty " +
								"through Terraform — change it to another network ID instead.",
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"setting_preference": {
							Description: "Whether the port's settings are taken from a profile (`auto`) or set per-port (`manual`). " +
								"Valid values are `auto` and `manual`. Per-port VLAN overrides (`native_networkconf_id`, " +
								"`tagged_vlan_mgmt`, `forward`, `excluded_network_ids`) generally require `setting_preference = \"manual\"` " +
								"to persist on the controller; with `auto` the controller may revert inline overrides to profile/auto " +
								"behavior. Setting this to `manual` also overrides any `port_profile_id` on the same port. " +
								"Computed when not set, so the value the controller attaches to the port is preserved without producing a diff.",
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"auto", "manual"}, false),
						},
					},
				},
			},

			"radio": {
				Description: "Per-band radio configuration for access points. Each block configures ONE band " +
					"(`ng` = 2.4GHz, `na` = 5GHz, `6e` = 6GHz). Only the bands you declare are managed — undeclared " +
					"bands are left untouched (the provider read-modify-writes the device's full radio table to preserve " +
					"them, so declaring just one band will not wipe the others). Common uses: disable a band " +
					"(`tx_power_mode = \"disabled\"`), pin a channel/width, or set a minimum-RSSI client kick. Applies to " +
					"access points; has no effect on switches.\n\n" +
					"Note: like other device fields, only non-zero values are written, so a field cannot be set back to its " +
					"zero value through Terraform — manage by overriding with explicit non-zero values.",
				Type:     schema.TypeSet,
				Optional: true,
				Set:      radioSetHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description:  "The radio band this block configures: `ng` (2.4GHz), `na` (5GHz), or `6e` (6GHz).",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"ng", "na", "6e"}, false),
						},
						"channel": {
							Description: "The channel for this radio (band-specific), or `auto` to let the controller choose.",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
						},
						"ht": {
							Description:  "Channel width in MHz for this radio (e.g. 20, 40, 80, 160, 320).",
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntInSlice([]int{20, 40, 80, 160, 240, 320, 1080, 2160, 4320}),
						},
						"tx_power_mode": {
							Description:  "Transmit-power mode: `auto`, `low`, `medium`, `high`, `custom`, or `disabled`. `disabled` turns the radio off (e.g. to suppress an unused 2.4GHz band on an in-wall AP).",
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"auto", "low", "medium", "high", "custom", "disabled"}, false),
						},
						"tx_power": {
							Description: "Custom transmit power in dBm, used when `tx_power_mode = \"custom\"`; otherwise leave unset.",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
						},
						"min_rssi_enabled": {
							Description: "Whether the minimum-RSSI client-disconnect threshold is enabled on this radio. Applied together with `min_rssi`.",
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
						},
						"min_rssi": {
							Description: "Minimum RSSI in dBm (negative) below which clients are disconnected, when `min_rssi_enabled` is true.",
							Type:        schema.TypeInt,
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},

			"ether_lighting": {
				Description: "Etherlighting configuration for switches with per-port LEDs (e.g. USW Pro Max). " +
					"`mode = \"network\"` colors each port's LED by the VLAN/network it serves (per-network colors come from " +
					"the site-level Etherlighting palette); `mode = \"speed\"` colors by link speed. Only the fields you set " +
					"are written — unset fields keep their controller-side values (read-modify-write overlay). Devices without " +
					"Etherlighting hardware ignore this object.",
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mode": {
							Description:  "Color scheme: `network` (color by VLAN/network) or `speed` (color by link speed).",
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"network", "speed"}, false),
						},
						"led_mode": {
							Description:  "`etherlighting` (colored per-port LEDs) or `standard` (plain status LEDs).",
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"etherlighting", "standard"}, false),
						},
						"behavior": {
							Description:  "LED animation: `steady` or `breath`.",
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"steady", "breath"}, false),
						},
						"brightness": {
							Description:  "LED brightness, 1-100.",
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(1, 100),
						},
					},
				},
			},

			"allow_adoption": {
				Description: "Whether to automatically adopt the device when creating this resource. When true:\n" +
					"* The controller will attempt to adopt the device\n" +
					"* Device must be in a pending adoption state\n" +
					"* Device must be accessible on the network\n" +
					"Set to false if you want to manage adoption manually.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"forget_on_destroy": {
				Description: "Whether to forget (un-adopt) the device when this resource is destroyed. When true:\n" +
					"* The device will be removed from the controller\n" +
					"* The device will need to be readopted to be managed again\n" +
					"* Device configuration will be reset\n" +
					"Set to false to keep the device adopted when removing from Terraform management.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceDeviceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	c := meta.(*base.Client)
	id := d.Id()
	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	if colons := strings.Count(id, ":"); colons == 1 || colons == 6 {
		importParts := strings.SplitN(id, ":", 2)
		site = importParts[0]
		id = importParts[1]
	}

	if utils.MacAddressRegexp.MatchString(id) {
		// look up id by mac
		mac := utils.CleanMAC(id)
		device, err := c.GetDeviceByMAC(ctx, site, mac)

		if err != nil {
			return nil, err
		}

		id = device.ID
	}

	if id != "" {
		d.SetId(id)
	}
	if site != "" {
		d.Set("site", site)
	}

	return []*schema.ResourceData{d}, nil
}

func resourceDeviceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	mac := d.Get("mac").(string)
	if mac == "" {
		return diag.Errorf("no MAC address specified, please import the device using terraform import")
	}

	mac = utils.CleanMAC(mac)
	device, err := c.GetDeviceByMAC(ctx, site, mac)

	if device == nil {
		return diag.Errorf("device not found using mac %q", mac)
	}
	if err != nil {
		return diag.FromErr(err)
	}

	if !device.Adopted {
		if !d.Get("allow_adoption").(bool) {
			return diag.Errorf("Device must be adopted before it can be managed")
		}

		err := c.AdoptDevice(ctx, site, mac)
		if err != nil {
			return diag.FromErr(err)
		}

		device, err = waitForDeviceState(ctx, d, meta, unifi.DeviceStateConnected, []unifi.DeviceState{unifi.DeviceStateAdopting, unifi.DeviceStatePending, unifi.DeviceStateProvisioning, unifi.DeviceStateUpgrading}, 2*time.Minute)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(device.ID)
	return resourceDeviceUpdate(ctx, d, meta)
}

func resourceDeviceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	req, err := resourceDeviceGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req.ID = d.Id()
	req.SiteID = site

	// Radio table and Etherlighting are controller-side structures managed
	// with patch semantics: fetch the device's current config once and overlay
	// only the declared fields, so undeclared bands/fields keep their
	// controller-side values. (Radio table additionally needs the full merged
	// array sent because UniFi replaces arrays wholesale on PUT.) When neither
	// block is declared, nothing extra is sent (prior behavior).
	radios := d.Get("radio").(*schema.Set)
	etherLighting := d.Get("ether_lighting").([]interface{})
	if radios.Len() > 0 || len(etherLighting) > 0 {
		current, err := c.GetDevice(ctx, site, d.Id())
		if err != nil {
			return diag.FromErr(fmt.Errorf("unable to read current device config for merge: %w", err))
		}
		if radios.Len() > 0 {
			req.RadioTable = mergeRadios(current.RadioTable, radios)
		}
		if len(etherLighting) > 0 {
			req.EtherLighting = mergeEtherLighting(current.EtherLighting, etherLighting[0].(map[string]interface{}))
		}
	}

	resp, err := c.UpdateDevice(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDeviceState(ctx, d, meta, unifi.DeviceStateConnected, []unifi.DeviceState{unifi.DeviceStateAdopting, unifi.DeviceStateProvisioning}, 1*time.Minute)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceDeviceSetResourceData(resp, d, site)
}

func resourceDeviceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	if !d.Get("forget_on_destroy").(bool) {
		return nil
	}

	site := d.Get("site").(string)
	mac := d.Get("mac").(string)

	if site == "" {
		site = c.Site
	}
	err := retry.RetryContext(ctx, 1*time.Minute, func() *retry.RetryError {
		internalErr := c.ForgetDevice(ctx, site, mac)
		if internalErr == nil {
			return nil
		}
		if utils.IsServerErrorContains(internalErr, "api.err.DeviceBusy") {
			return retry.RetryableError(internalErr)
		}
		return retry.NonRetryableError(internalErr)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDeviceState(ctx, d, meta, unifi.DeviceStatePending, []unifi.DeviceState{unifi.DeviceStateConnected, unifi.DeviceStateDeleting}, 1*time.Minute)
	if !errors.Is(err, unifi.ErrNotFound) {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDeviceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	id := d.Id()

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.GetDevice(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceDeviceSetResourceData(resp, d, site)
}

func resourceDeviceSetResourceData(resp *unifi.Device, d *schema.ResourceData, site string) diag.Diagnostics {
	portOverrides, err := setFromPortOverrides(resp.PortOverrides)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("site", site)
	d.Set("mac", resp.MAC)
	d.Set("name", resp.Name)
	d.Set("disabled", resp.Disabled)
	d.Set("switch_vlan_enabled", resp.SwitchVLANEnabled)
	d.Set("port_override", portOverrides)
	d.Set("radio", radiosFromDevice(resp, d))
	d.Set("ether_lighting", etherLightingFromDevice(resp, d))

	return nil
}

// etherLightingFromDevice returns ether_lighting state only when the user
// declares the block, so unmanaged devices never produce a diff.
func etherLightingFromDevice(resp *unifi.Device, d *schema.ResourceData) []map[string]interface{} {
	if len(d.Get("ether_lighting").([]interface{})) == 0 {
		return nil
	}
	return []map[string]interface{}{{
		"mode":       resp.EtherLighting.Mode,
		"led_mode":   resp.EtherLighting.LedMode,
		"behavior":   resp.EtherLighting.Behavior,
		"brightness": resp.EtherLighting.Brightness,
	}}
}

// mergeEtherLighting overlays the declared ether_lighting fields onto the
// device's current config, preserving any fields the user didn't set.
func mergeEtherLighting(current unifi.DeviceEtherLighting, m map[string]interface{}) unifi.DeviceEtherLighting {
	r := current
	if v, _ := m["mode"].(string); v != "" {
		r.Mode = v
	}
	if v, _ := m["led_mode"].(string); v != "" {
		r.LedMode = v
	}
	if v, _ := m["behavior"].(string); v != "" {
		r.Behavior = v
	}
	if v, _ := m["brightness"].(int); v != 0 {
		r.Brightness = v
	}
	return r
}

// radioSetHash keys the `radio` set by band only, so changes to Computed
// fields (channel, ht, …) don't churn set membership during plan/apply.
func radioSetHash(v interface{}) int {
	m := v.(map[string]interface{})
	return schema.HashString(m["name"].(string))
}

// portOverrideSetHash keys the `port_override` set by port `number` only, so the
// controller echoing/auto-populating per-port fields (e.g. setting_preference or
// a native VLAN) on an entry the user didn't declare them on does not change the
// element's set identity and churn the set. Together with the Optional+Computed
// VLAN attributes, an undeclared field reads back the controller value without a
// perpetual add/remove diff. `number` is Required and unique per port
// (setToPortOverrides already dedupes by PortIDX), so it is a sound stable key.
// Mirrors the radioSetHash precedent.
func portOverrideSetHash(v interface{}) int {
	m := v.(map[string]interface{})
	return schema.HashInt(m["number"].(int))
}

// radiosFromDevice returns radio state for only the bands the user manages
// (present in config/state), so undeclared bands on the device never produce
// a diff.
func radiosFromDevice(resp *unifi.Device, d *schema.ResourceData) []map[string]interface{} {
	managed := map[string]bool{}
	for _, item := range d.Get("radio").(*schema.Set).List() {
		managed[item.(map[string]interface{})["name"].(string)] = true
	}
	radios := make([]map[string]interface{}, 0, len(managed))
	for _, r := range resp.RadioTable {
		if managed[r.Radio] {
			radios = append(radios, fromRadio(r))
		}
	}
	return radios
}

func fromRadio(r unifi.DeviceRadioTable) map[string]interface{} {
	return map[string]interface{}{
		"name":             r.Radio,
		"channel":          r.Channel,
		"ht":               r.Ht,
		"tx_power_mode":    r.TxPowerMode,
		"tx_power":         r.TxPower,
		"min_rssi":         r.MinRssi,
		"min_rssi_enabled": r.MinRssiEnabled,
	}
}

// mergeRadios overlays the declared radio blocks onto the device's current
// radio_table, preserving every band's existing settings and changing only the
// non-zero fields the user specified. Bands not present in `current` are
// appended. The full merged list is returned so the wholesale-replace PUT keeps
// all bands intact.
func mergeRadios(current []unifi.DeviceRadioTable, set *schema.Set) []unifi.DeviceRadioTable {
	byBand := map[string]unifi.DeviceRadioTable{}
	order := make([]string, 0, len(current))
	for _, r := range current {
		byBand[r.Radio] = r
		order = append(order, r.Radio)
	}
	for _, item := range set.List() {
		m := item.(map[string]interface{})
		band := m["name"].(string)
		r, ok := byBand[band]
		if !ok {
			r = unifi.DeviceRadioTable{Radio: band}
			order = append(order, band)
		}
		if v, _ := m["channel"].(string); v != "" {
			r.Channel = v
		}
		if v, _ := m["ht"].(int); v != 0 {
			r.Ht = v
		}
		if v, _ := m["tx_power_mode"].(string); v != "" {
			r.TxPowerMode = v
		}
		if v, _ := m["tx_power"].(string); v != "" {
			r.TxPower = v
		}
		if v, _ := m["min_rssi"].(int); v != 0 {
			r.MinRssi = v
			r.MinRssiEnabled = m["min_rssi_enabled"].(bool)
		}
		byBand[band] = r
	}
	out := make([]unifi.DeviceRadioTable, 0, len(order))
	for _, b := range order {
		out = append(out, byBand[b])
	}
	return out
}

func resourceDeviceGetResourceData(d *schema.ResourceData) (*unifi.Device, error) {
	pos, err := setToPortOverrides(d.Get("port_override").(*schema.Set))
	if err != nil {
		return nil, fmt.Errorf("unable to process port_override block: %w", err)
	}

	//TODO: pass Disabled once we figure out how to enable the device afterwards

	return &unifi.Device{
		MAC:               d.Get("mac").(string),
		Name:              d.Get("name").(string),
		SwitchVLANEnabled: d.Get("switch_vlan_enabled").(bool),
		PortOverrides:     pos,
	}, nil
}

func setToPortOverrides(set *schema.Set) ([]unifi.DevicePortOverrides, error) {
	// use a map here to remove any duplication
	overrideMap := map[int]unifi.DevicePortOverrides{}
	for _, item := range set.List() {
		data, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected data in block")
		}
		po, err := toPortOverride(data)
		if err != nil {
			return nil, fmt.Errorf("unable to create port override: %w", err)
		}
		overrideMap[po.PortIDX] = po
	}

	pos := make([]unifi.DevicePortOverrides, 0, len(overrideMap))
	for _, item := range overrideMap {
		pos = append(pos, item)
	}
	return pos, nil
}

func setFromPortOverrides(pos []unifi.DevicePortOverrides) ([]map[string]interface{}, error) {
	list := make([]map[string]interface{}, 0, len(pos))
	for _, po := range pos {
		v, err := fromPortOverride(po)
		if err != nil {
			return nil, fmt.Errorf("unable to parse port override: %w", err)
		}
		list = append(list, v)
	}
	return list, nil
}

func toPortOverride(data map[string]interface{}) (unifi.DevicePortOverrides, error) {
	idx := data["number"].(int)
	name := data["name"].(string)
	profileID := data["port_profile_id"].(string)
	opMode := data["op_mode"].(string)
	poeMode := data["poe_mode"].(string)
	aggregateNumPorts := data["aggregate_num_ports"].(int)

	var excludedNetworkIDs []string
	if set, ok := data["excluded_network_ids"].(*schema.Set); ok {
		var err error
		excludedNetworkIDs, err = utils.SetToStringSlice(set)
		if err != nil {
			return unifi.DevicePortOverrides{}, fmt.Errorf("unable to process excluded_network_ids: %w", err)
		}
	}

	// Per-port VLAN overrides. All of these are `omitempty` on the controller
	// side, so an unset (empty) value is dropped from the PUT. The user declares
	// the whole set of ports, and the device PUT replaces the `port_overrides`
	// array wholesale (no read-modify-write merge). comma-ok reads tolerate a
	// partially-populated data map (e.g. in unit tests).
	nativeNetworkID, _ := data["native_networkconf_id"].(string)
	taggedVLANMgmt, _ := data["tagged_vlan_mgmt"].(string)
	forward, _ := data["forward"].(string)
	voiceNetworkID, _ := data["voice_networkconf_id"].(string)
	settingPreference, _ := data["setting_preference"].(string)

	po := unifi.DevicePortOverrides{
		PortIDX:            idx,
		Name:               name,
		PortProfileID:      profileID,
		OpMode:             opMode,
		PoeMode:            poeMode,
		NATiveNetworkID:    nativeNetworkID,
		TaggedVLANMgmt:     taggedVLANMgmt,
		Forward:            forward,
		ExcludedNetworkIDs: excludedNetworkIDs,
		VoiceNetworkID:     voiceNetworkID,
		SettingPreference:  settingPreference,
	}

	// go-unifi v1.9 tracks the current controller API, which expresses a LAG
	// as an explicit member list (`aggregate_members`) instead of the legacy
	// starting-port count (`aggregate_num_ports`). Translate the schema's
	// count into the equivalent contiguous member range — N sequential ports
	// starting at this port — which matches the documented schema semantics
	// ("All ports in the LAG must be sequential"), so existing practitioner
	// configs keep working unchanged. When unset (0), leave the slice nil so
	// the field is omitted from the payload entirely.
	if aggregateNumPorts > 0 {
		members := make([]int, aggregateNumPorts)
		for i := range members {
			members[i] = idx + i
		}
		po.AggregateMembers = members
	}

	return po, nil
}

func fromPortOverride(po unifi.DevicePortOverrides) (map[string]interface{}, error) {
	return map[string]interface{}{
		"number":          po.PortIDX,
		"name":            po.Name,
		"port_profile_id": po.PortProfileID,
		"op_mode":         po.OpMode,
		"poe_mode":        po.PoeMode,
		// Inverse of the translation in toPortOverride: the member-list
		// length is the LAG port count (0 / unset round-trips as an empty
		// list, preserving the previous zero-value behavior).
		"aggregate_num_ports": len(po.AggregateMembers),
		// Per-port VLAN overrides, round-tripped unconditionally to match the
		// existing fields above (keeps ImportStateVerify consistent). These
		// attributes are Optional+Computed and the set is keyed by port number
		// (portOverrideSetHash), so surfacing a value the controller populated on
		// a port the user didn't declare it on is absorbed as the computed value
		// instead of churning the set.
		"native_networkconf_id": po.NATiveNetworkID,
		"tagged_vlan_mgmt":      po.TaggedVLANMgmt,
		"forward":               po.Forward,
		"excluded_network_ids":  utils.StringSliceToSet(po.ExcludedNetworkIDs),
		"voice_networkconf_id":  po.VoiceNetworkID,
		"setting_preference":    po.SettingPreference,
	}, nil
}

func waitForDeviceState(ctx context.Context, d *schema.ResourceData, meta interface{}, targetState unifi.DeviceState, pendingStates []unifi.DeviceState, timeout time.Duration) (*unifi.Device, error) {
	c := meta.(*base.Client)

	site := d.Get("site").(string)
	mac := d.Get("mac").(string)

	if site == "" {
		site = c.Site
	}

	// Always consider unknown to be a pending state.
	pendingStates = append(pendingStates, unifi.DeviceStateUnknown)

	var pending []string
	for _, state := range pendingStates {
		pending = append(pending, state.String())
	}

	wait := retry.StateChangeConf{
		Pending: pending,
		Target:  []string{targetState.String()},
		Refresh: func() (interface{}, string, error) {
			device, err := c.GetDeviceByMAC(ctx, site, mac)

			if errors.Is(err, unifi.ErrNotFound) {
				err = nil
			}

			// When a device is forgotten, it will disappear from the UI for a few seconds before reappearing.
			// During this time, `device.GetDeviceByMAC` will return a 400.
			//
			// TODO: Improve handling of this situation in `go-unifi`.
			if err != nil && strings.Contains(err.Error(), "api.err.UnknownDevice") {
				err = nil
			}

			var state string
			if device != nil {
				state = device.State.String()
			}

			// TODO: Why is this needed???
			if device == nil {
				return nil, state, err
			}

			return device, state, err
		},
		Timeout:        timeout,
		NotFoundChecks: 30,
	}

	outputRaw, err := wait.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*unifi.Device); ok {
		return output, err
	}

	return nil, err
}

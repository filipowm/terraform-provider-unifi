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
			"port_override": {
				// TODO: this should really be a map or something when possible in the SDK
				// see https://github.com/hashicorp/terraform-plugin-sdk/issues/62
				Description: "A list of port-specific configuration overrides for UniFi switches. This allows you to customize individual port settings such as:\n" +
					"  * Port names and labels for easy identification\n" +
					"  * Port profiles for VLAN and security settings\n" +
					"  * Operating modes for special functions\n\n" +
					"Common use cases include:\n" +
					"  * Setting up trunk ports for inter-switch connections\n" +
					"  * Configuring PoE settings for powered devices\n" +
					"  * Creating mirrored ports for network monitoring\n" +
					"  * Setting up link aggregation between switches or servers",
				Type:     schema.TypeSet,
				Optional: true,
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
	d.Set("port_override", portOverrides)

	return nil
}

func resourceDeviceGetResourceData(d *schema.ResourceData) (*unifi.Device, error) {
	pos, err := setToPortOverrides(d.Get("port_override").(*schema.Set))
	if err != nil {
		return nil, fmt.Errorf("unable to process port_override block: %w", err)
	}

	//TODO: pass Disabled once we figure out how to enable the device afterwards

	return &unifi.Device{
		MAC:           d.Get("mac").(string),
		Name:          d.Get("name").(string),
		PortOverrides: pos,
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

	return unifi.DevicePortOverrides{
		PortIDX:           idx,
		Name:              name,
		PortProfileID:     profileID,
		OpMode:            opMode,
		PoeMode:           poeMode,
		AggregateNumPorts: aggregateNumPorts,
	}, nil
}

func fromPortOverride(po unifi.DevicePortOverrides) (map[string]interface{}, error) {
	return map[string]interface{}{
		"number":              po.PortIDX,
		"name":                po.Name,
		"port_profile_id":     po.PortProfileID,
		"op_mode":             po.OpMode,
		"poe_mode":            po.PoeMode,
		"aggregate_num_ports": po.AggregateNumPorts,
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

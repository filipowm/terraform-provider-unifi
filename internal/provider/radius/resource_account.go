package radius

import (
	"context"
	"errors"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/utils"
)

func ResourceAccount() *schema.Resource {
	return &schema.Resource{
		Description: "The `unifi_account` resource manages RADIUS user accounts in the UniFi controller's built-in RADIUS server.\n\n" +
			"This resource is used for:\n" +
			"  * WPA2/WPA3-Enterprise wireless authentication\n" +
			"  * 802.1X wired authentication\n" +
			"  * MAC-based device authentication\n" +
			"  * Dynamic VLAN assignment through RADIUS attributes (see the `vlan` attribute)\n\n" +
			"Important Notes:\n" +
			"1. For MAC-based authentication:\n" +
			"   * Use the device's MAC address as both username and password\n" +
			"   * Convert MAC address to uppercase with no separators (e.g., '00:11:22:33:44:55' becomes '001122334455')\n" +
			"2. VLAN Assignment:\n" +
			"   * Set the `vlan` attribute to the 802.1Q VLAN ID the controller should assign to authenticated clients\n" +
			"   * VLAN assignment is delivered using the standard RADIUS tunnel attributes (`tunnel_type`/`tunnel_medium_type`)\n" +
			"   * If no VLAN is specified, clients will use the network's untagged VLAN\n\n" +
			"Limitations:\n" +
			"  * MAC-based authentication works only for wireless and wired clients\n" +
			"  * L2TP remote access VPN is not supported with MAC authentication\n" +
			"  * Accounts must be unique within a site",

		CreateContext: resourceAccountCreate,
		ReadContext:   resourceAccountRead,
		UpdateContext: resourceAccountUpdate,
		DeleteContext: resourceAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: base.ImportSiteAndID,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the RADIUS account in the UniFi controller. This is automatically assigned.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"site": {
				Description: "The name of the UniFi site where this RADIUS account should be created. If not specified, the default site will be used.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The username for this RADIUS account. For regular users, this can be any unique identifier. For MAC-based " +
					"authentication, this must be the device's MAC address in uppercase with no separators (e.g., '001122334455').",
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Description: "The password for this RADIUS account. For MAC-based authentication, this must match the username (the MAC address). " +
					"For regular users, this should be a secure password following your organization's password policies.",
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"tunnel_type": {
				Description: "The RADIUS tunnel type attribute ([RFC 2868](https://tools.ietf.org/html/rfc2868), section 3.1). Common values:\n" +
					"  * `13` - VLAN (default)\n" +
					"  * `1` - Point-to-Point Protocol (PPTP)\n" +
					"  * `9` - Point-to-Point Protocol (L2TP)\n\n" +
					"Only change this if you need specific tunneling behavior.",
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      13,
				ValidateFunc: validation.IntBetween(1, 13),
			},
			"tunnel_medium_type": {
				Description: "The RADIUS tunnel medium type attribute ([RFC 2868](https://tools.ietf.org/html/rfc2868), section 3.2). Common values:\n" +
					"  * `6` - 802 (includes Ethernet, Token Ring, FDDI) (default)\n" +
					"  * `1` - IPv4\n" +
					"  * `2` - IPv6\n\n" +
					"Only change this if you need specific tunneling behavior.",
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      6,
				ValidateFunc: validation.IntBetween(1, 15),
			},
			"network_id": {
				Description: "The ID of a UniFi network configuration (the controller's `networkconf_id`) to associate with this " +
					"account. This is a reference to a network object and is distinct from the `vlan` attribute, which sets the " +
					"802.1Q VLAN ID delivered via RADIUS.",
				Type:     schema.TypeString,
				Optional: true,
			},
			"vlan": {
				Description: "The 802.1Q VLAN ID to assign to clients authenticating with this account, used for RADIUS dynamic " +
					"VLAN assignment. It is delivered together with the tunnel attributes (`tunnel_type`/`tunnel_medium_type`). " +
					"Omitting this attribute means no VLAN is assigned; if a VLAN was set out-of-band " +
					"(e.g. in the controller UI), omitting it here removes it on the next apply.",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(2),
			},
		},
	}
}

func resourceAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	req := resourceAccountGetResourceData(d)

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.CreateAccount(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return resourceAccountSetResourceData(resp, d, site)
}

func resourceAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	req := resourceAccountGetResourceData(d)

	req.ID = d.Id()
	req.SiteID = site

	// go-unifi v1.9.2's updateAccount converts a successful-but-empty PUT response
	// into unifi.ErrNotFound (see utils.ReReadOnUpdateNotFound / issue #98); re-read
	// to tell a spurious error from a genuine out-of-band deletion.
	resp, err := c.UpdateAccount(ctx, site, req)
	resp, found, err := utils.ReReadOnUpdateNotFound(resp, err, func() (*unifi.Account, error) {
		return c.GetAccount(ctx, site, req.ID)
	})
	if err != nil {
		return diag.FromErr(err)
	}
	if !found {
		d.SetId("")
		return nil
	}

	return resourceAccountSetResourceData(resp, d, site)
}

func resourceAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	// name := d.Get("name").(string)
	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	id := d.Id()
	err := c.DeleteAccount(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		return nil
	}
	return diag.FromErr(err)
}

func resourceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	id := d.Id()

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.GetAccount(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceAccountSetResourceData(resp, d, site)
}

func resourceAccountSetResourceData(resp *unifi.Account, d *schema.ResourceData, site string) diag.Diagnostics {
	if err := d.Set("site", site); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", resp.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("password", resp.XPassword); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tunnel_type", resp.TunnelType); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tunnel_medium_type", resp.TunnelMediumType); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("network_id", resp.NetworkID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("vlan", resp.VLAN); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceAccountGetResourceData(d *schema.ResourceData) *unifi.Account {
	name, _ := d.Get("name").(string)
	password, _ := d.Get("password").(string)
	tunnelType, _ := d.Get("tunnel_type").(int)
	tunnelMediumType, _ := d.Get("tunnel_medium_type").(int)
	networkID, _ := d.Get("network_id").(string)
	vlan, _ := d.Get("vlan").(int)
	return &unifi.Account{
		Name:             name,
		XPassword:        password,
		TunnelType:       tunnelType,
		TunnelMediumType: tunnelMediumType,
		NetworkID:        networkID,
		VLAN:             vlan,
	}
}

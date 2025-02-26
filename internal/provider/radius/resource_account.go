package radius

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

func ResourceAccount() *schema.Resource {
	return &schema.Resource{
		Description: "The `unifi_account` resource manages RADIUS user accounts in the UniFi controller's built-in RADIUS server.\n\n" +
			"This resource is used for:\n" +
			"  * WPA2/WPA3-Enterprise wireless authentication\n" +
			"  * 802.1X wired authentication\n" +
			"  * MAC-based device authentication\n" +
			"  * VLAN assignment through RADIUS attributes\n\n" +
			"Important Notes:\n" +
			"1. For MAC-based authentication:\n" +
			"   * Use the device's MAC address as both username and password\n" +
			"   * Convert MAC address to uppercase with no separators (e.g., '00:11:22:33:44:55' becomes '001122334455')\n" +
			"2. VLAN Assignment:\n" +
			"   * If no VLAN is specified in the profile, clients will use the network's untagged VLAN\n" +
			"   * VLAN assignment uses standard RADIUS tunnel attributes\n\n" +
			"Limitations:\n" +
			"  * MAC-based authentication works only for wireless and wired clients\n" +
			"  * L2TP remote access VPN is not supported with MAC authentication\n" +
			"  * Accounts must be unique within a site",

		CreateContext: resourceAccountCreate,
		ReadContext:   resourceAccountRead,
		UpdateContext: resourceAccountUpdate,
		DeleteContext: resourceAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: utils.ImportSiteAndID,
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
				Description: "The ID of the network (VLAN) to assign to clients authenticating with this account. This is used in " +
					"conjunction with the tunnel attributes to provide VLAN assignment via RADIUS.",
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	req, err := resourceAccountGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	site := d.Get("site").(string)
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
	c := meta.(*base.Client)

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	req, err := resourceAccountGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req.ID = d.Id()
	req.SiteID = site

	resp, err := c.UpdateAccount(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceAccountSetResourceData(resp, d, site)
}

func resourceAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	//name := d.Get("name").(string)
	site := d.Get("site").(string)
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
	c := meta.(*base.Client)

	id := d.Id()

	site := d.Get("site").(string)
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
	d.Set("site", site)
	d.Set("name", resp.Name)
	d.Set("password", resp.XPassword)
	d.Set("tunnel_type", resp.TunnelType)
	d.Set("tunnel_medium_type", resp.TunnelMediumType)
	d.Set("network_id", resp.NetworkID)
	return nil
}

func resourceAccountGetResourceData(d *schema.ResourceData) (*unifi.Account, error) {
	return &unifi.Account{
		Name:             d.Get("name").(string),
		XPassword:        d.Get("password").(string),
		TunnelType:       d.Get("tunnel_type").(int),
		TunnelMediumType: d.Get("tunnel_medium_type").(int),
		NetworkID:        d.Get("network_id").(string),
	}, nil
}

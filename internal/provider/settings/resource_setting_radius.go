package settings

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

func ResourceSettingRadius() *schema.Resource {
	return &schema.Resource{
		Description: "The `unifi_setting_radius` resource manages the built-in RADIUS server configuration in the UniFi controller.\n\n" +
			"This resource allows you to configure:\n" +
			"  * Authentication settings for network access control\n" +
			"  * Accounting settings for tracking user sessions\n" +
			"  * Security features like tunneled replies\n\n" +
			"The RADIUS server is commonly used for:\n" +
			"  * Enterprise WPA2/WPA3-Enterprise wireless networks\n" +
			"  * 802.1X port-based network access control\n" +
			"  * Centralized user authentication and accounting\n\n" +
			"When enabled, the RADIUS server can authenticate clients using the UniFi user database or external authentication sources.",

		CreateContext: resourceSettingRadiusCreate,
		ReadContext:   resourceSettingRadiusRead,
		UpdateContext: resourceSettingRadiusUpdate,
		DeleteContext: schema.NoopContext,
		Importer: &schema.ResourceImporter{
			StateContext: utils.ImportSiteAndID,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the RADIUS settings configuration in the UniFi controller.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"site": {
				Description: "The name of the UniFi site where these RADIUS settings should be applied. If not specified, the default site will be used.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},
			"accounting_enabled": {
				Description: "Enable RADIUS accounting to track user sessions, including connection time, data usage, and other metrics. " +
					"This information can be useful for billing, capacity planning, and security auditing.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"accounting_port": {
				Description: "The UDP port number for RADIUS accounting communications. The standard port is 1813. Only change this if you " +
					"need to avoid port conflicts or match specific network requirements.",
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1813,
				ValidateFunc: validation.IsPortNumber,
			},
			"auth_port": {
				Description: "The UDP port number for RADIUS authentication communications. The standard port is 1812. Only change this if you " +
					"need to avoid port conflicts or match specific network requirements.",
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1812,
				ValidateFunc: validation.IsPortNumber,
			},
			"interim_update_interval": {
				Description: "The interval (in seconds) at which the RADIUS server collects and updates statistics from connected clients. " +
					"Default is 3600 seconds (1 hour). Lower values provide more frequent updates but increase server load.",
				Type:     schema.TypeInt,
				Optional: true,
				Default:  3600,
			},
			"tunneled_reply": {
				Description: "Enable encrypted communication between the RADIUS server and clients using RADIUS tunneling. This adds an extra " +
					"layer of security by protecting RADIUS attributes in transit.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"secret": {
				Description: "The shared secret passphrase used to authenticate RADIUS clients (like wireless access points) with the RADIUS server. " +
					"This should be a strong, random string known only to the server and its clients.",
				Type:      schema.TypeString,
				Sensitive: true,
				Optional:  true,
				Default:   "",
			},
			"enabled": {
				Description: "Enable or disable the built-in RADIUS server. When disabled, no RADIUS authentication or accounting services " +
					"will be provided, affecting any network services that rely on RADIUS (like WPA2-Enterprise networks).",
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
		},
	}
}

func resourceSettingRadiusGetResourceData(d *schema.ResourceData, meta interface{}) (*unifi.SettingRadius, error) {
	return &unifi.SettingRadius{
		AccountingEnabled:     d.Get("accounting_enabled").(bool),
		Enabled:               d.Get("enabled").(bool),
		AcctPort:              d.Get("accounting_port").(int),
		AuthPort:              d.Get("auth_port").(int),
		ConfigureWholeNetwork: true,
		TunneledReply:         d.Get("tunneled_reply").(bool),
		XSecret:               d.Get("secret").(string),
		InterimUpdateInterval: d.Get("interim_update_interval").(int),
	}, nil
}

func resourceSettingRadiusCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	req, err := resourceSettingRadiusGetResourceData(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.UpdateSettingRadius(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return resourceSettingRadiusSetResourceData(resp, d, meta, site)
}

func resourceSettingRadiusSetResourceData(resp *unifi.SettingRadius, d *schema.ResourceData, meta interface{}, site string) diag.Diagnostics {
	d.Set("site", site)
	d.Set("enabled", resp.Enabled)
	d.Set("accounting_enabled", resp.AccountingEnabled)
	d.Set("accounting_port", resp.AcctPort)
	d.Set("auth_port", resp.AuthPort)
	d.Set("tunneled_reply", resp.TunneledReply)
	d.Set("secret", resp.XSecret)
	d.Set("interim_update_interval", resp.InterimUpdateInterval)
	return nil
}

func resourceSettingRadiusRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.GetSettingRadius(ctx, site)
	if errors.Is(err, unifi.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSettingRadiusSetResourceData(resp, d, meta, site)
}

func resourceSettingRadiusUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	req, err := resourceSettingRadiusGetResourceData(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	req.ID = d.Id()
	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.UpdateSettingRadius(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSettingRadiusSetResourceData(resp, d, meta, site)
}

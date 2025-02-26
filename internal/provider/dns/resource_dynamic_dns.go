package dns

import (
	"context"
	"errors"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/utils"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceDynamicDNS() *schema.Resource {
	return &schema.Resource{
		Description: "The `unifi_dynamic_dns` resource manages Dynamic DNS (DDNS).\n\n" +
			"Dynamic DNS allows you to access your network using a domain name even when your public IP address changes. This is useful for:\n" +
			"  * Remote access to your network\n" +
			"  * Hosting services from your home/office network\n" +
			"  * VPN connections to your network\n\n" +
			"The resource supports various DDNS providers including:\n" +
			"  * DynDNS\n" +
			"  * No-IP\n" +
			"  * Duck DNS\n" +
			"  * And many others\n\n" +
			"Each DDNS configuration can be associated with either the primary (WAN) or secondary (WAN2) interface.",

		CreateContext: resourceDynamicDNSCreate,
		ReadContext:   resourceDynamicDNSRead,
		UpdateContext: resourceDynamicDNSUpdate,
		DeleteContext: resourceDynamicDNSDelete,
		Importer: &schema.ResourceImporter{
			StateContext: utils.ImportSiteAndID,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the dynamic DNS configuration in the UniFi controller.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"site": {
				Description: "The name of the UniFi site where the dynamic DNS configuration should be created. If not specified, the default site will be used.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},
			"interface": {
				Description: "The WAN interface to use for the dynamic DNS updates. Valid values are:\n" +
					"  * `wan` - Primary WAN interface (default)\n" +
					"  * `wan2` - Secondary WAN interface",
				Type:     schema.TypeString,
				Optional: true,
				Default:  "wan",
				ForceNew: true,
			},
			"service": {
				Description: "The Dynamic DNS service provider. Common values include:\n" +
					"  * `dyndns` - DynDNS service\n" +
					"  * `noip` - No-IP service\n" +
					"  * `duckdns` - Duck DNS service\n" +
					"Check your UniFi controller for the complete list of supported providers.",
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"host_name": {
				Description: "The fully qualified domain name to update with your current public IP address (e.g., 'myhouse.dyndns.org' or 'myoffice.no-ip.com').",
				Type:        schema.TypeString,
				Required:    true,
			},
			"server": {
				Description: "The update server hostname for your DDNS provider. Usually not required as the UniFi controller knows the correct servers for common providers.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"login": {
				Description: "The username or login for your DDNS provider account.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"password": {
				Description: "The password or token for your DDNS provider account. This value will be stored securely and not displayed in logs.",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
			},

			//TODO: options support?
		},
	}
}

func resourceDynamicDNSCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	req, err := resourceDynamicDNSGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.CreateDynamicDNS(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return resourceDynamicDNSSetResourceData(resp, d, site)
}

func resourceDynamicDNSGetResourceData(d *schema.ResourceData) (*unifi.DynamicDNS, error) {
	r := &unifi.DynamicDNS{
		Interface: d.Get("interface").(string),
		Service:   d.Get("service").(string),

		HostName: d.Get("host_name").(string),

		Server:    d.Get("server").(string),
		Login:     d.Get("login").(string),
		XPassword: d.Get("password").(string),
	}

	return r, nil
}

func resourceDynamicDNSSetResourceData(resp *unifi.DynamicDNS, d *schema.ResourceData, site string) diag.Diagnostics {
	d.Set("interface", resp.Interface)
	d.Set("service", resp.Service)

	d.Set("host_name", resp.HostName)

	d.Set("server", resp.Server)
	d.Set("login", resp.Login)
	d.Set("password", resp.XPassword)

	return nil
}

func resourceDynamicDNSRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	id := d.Id()

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.GetDynamicDNS(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceDynamicDNSSetResourceData(resp, d, site)
}

func resourceDynamicDNSUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	req, err := resourceDynamicDNSGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req.ID = d.Id()

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	req.SiteID = site

	resp, err := c.UpdateDynamicDNS(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceDynamicDNSSetResourceData(resp, d, site)
}

func resourceDynamicDNSDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	id := d.Id()

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	err := c.DeleteDynamicDNS(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		return nil
	}
	return diag.FromErr(err)
}

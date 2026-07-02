package dns

import (
	"context"
	"errors"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/utils"
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
			StateContext: base.ImportSiteAndID,
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

			// TODO: options support?
		},
	}
}

func resourceDynamicDNSCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.FromErr(errors.New("meta must be of type *base.Client"))
	}

	req, err := resourceDynamicDNSGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	site, ok := d.Get("site").(string)
	if !ok {
		return diag.FromErr(errors.New("`site` must be a string"))
	}
	if site == "" {
		site = c.Site
	}

	resp, err := c.CreateDynamicDNS(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return resourceDynamicDNSSetResourceData(resp, d)
}

func resourceDynamicDNSGetResourceData(d *schema.ResourceData) (*unifi.DynamicDNS, error) {
	iface, ok := d.Get("interface").(string)
	if !ok {
		return nil, errors.New("interface must be a string")
	}
	service, ok := d.Get("service").(string)
	if !ok {
		return nil, errors.New("service must be a string")
	}
	hostName, ok := d.Get("host_name").(string)
	if !ok {
		return nil, errors.New("host_name must be a string")
	}
	server, ok := d.Get("server").(string)
	if !ok {
		return nil, errors.New("server must be a string")
	}
	login, ok := d.Get("login").(string)
	if !ok {
		return nil, errors.New("login must be a string")
	}
	password, ok := d.Get("password").(string)
	if !ok {
		return nil, errors.New("password must be a string")
	}

	r := &unifi.DynamicDNS{
		Interface: iface,
		Service:   service,

		HostName: hostName,

		Server:    server,
		Login:     login,
		XPassword: password,
	}

	return r, nil
}

func resourceDynamicDNSSetResourceData(resp *unifi.DynamicDNS, d *schema.ResourceData) diag.Diagnostics {
	if err := d.Set("interface", resp.Interface); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service", resp.Service); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("host_name", resp.HostName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("server", resp.Server); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("login", resp.Login); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("password", resp.XPassword); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDynamicDNSRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.FromErr(errors.New("meta must be of type *base.Client"))
	}

	id := d.Id()

	site, ok := d.Get("site").(string)
	if !ok {
		return diag.FromErr(errors.New("`site` must be a string"))
	}
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

	return resourceDynamicDNSSetResourceData(resp, d)
}

func resourceDynamicDNSUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.FromErr(errors.New("meta must be of type *base.Client"))
	}

	req, err := resourceDynamicDNSGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req.ID = d.Id()

	site, ok := d.Get("site").(string)
	if !ok {
		return diag.FromErr(errors.New("`site` must be a string"))
	}
	if site == "" {
		site = c.Site
	}
	req.SiteID = site

	// go-unifi v1.9.2's updateDynamicDNS converts a successful-but-empty PUT response
	// into unifi.ErrNotFound (see utils.ReReadOnUpdateNotFound / issue #98); re-read
	// to tell a spurious error from a genuine out-of-band deletion.
	resp, err := c.UpdateDynamicDNS(ctx, site, req)
	resp, found, err := utils.ReReadOnUpdateNotFound(resp, err, func() (*unifi.DynamicDNS, error) {
		return c.GetDynamicDNS(ctx, site, req.ID)
	})
	if err != nil {
		return diag.FromErr(err)
	}
	if !found {
		d.SetId("")
		return nil
	}

	return resourceDynamicDNSSetResourceData(resp, d)
}

func resourceDynamicDNSDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.FromErr(errors.New("meta must be of type *base.Client"))
	}

	id := d.Id()

	site, ok := d.Get("site").(string)
	if !ok {
		return diag.FromErr(errors.New("`site` must be a string"))
	}
	if site == "" {
		site = c.Site
	}
	err := c.DeleteDynamicDNS(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		return nil
	}
	return diag.FromErr(err)
}

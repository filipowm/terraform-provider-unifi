package routing

import (
	"context"
	"errors"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/utils"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourcePortForward() *schema.Resource {
	return &schema.Resource{
		Description: "The `unifi_port_forward` resource manages port forwarding rules on UniFi controllers.\n\n" +
			"Port forwarding allows external traffic to reach services hosted on your internal network by mapping external ports to internal IP addresses and ports. " +
			"This is commonly used for:\n" +
			"  * Hosting web servers, game servers, or other services\n" +
			"  * Remote access to internal services\n" +
			"  * Application-specific requirements\n\n" +
			"Each rule can be configured with source IP restrictions, protocol selection, and logging options for enhanced security and monitoring.",

		CreateContext: resourcePortForwardCreate,
		ReadContext:   resourcePortForwardRead,
		UpdateContext: resourcePortForwardUpdate,
		DeleteContext: resourcePortForwardDelete,
		Importer: &schema.ResourceImporter{
			StateContext: base.ImportSiteAndID,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the port forwarding rule in the UniFi controller.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"site": {
				Description: "The name of the UniFi site where the port forwarding rule should be created. If not specified, the default site will be used.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},
			"dst_port": {
				Description:  "The external port(s) that will be forwarded. Can be a single port (e.g., '80') or a port range (e.g., '8080:8090').",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validators.PortRangeV2,
			},
			// TODO: remove this, disabled rules should just be deleted.
			"enabled": {
				Description: "Specifies whether the port forwarding rule is enabled or not.",
				Type:        schema.TypeBool,
				Default:     true,
				Optional:    true,
				Deprecated: "This will attribute will be removed in a future release. Instead of disabling a " +
					"port forwarding rule you can remove it from your configuration.",
			},
			"fwd_ip": {
				Description:  "The internal IPv4 address of the device or service that will receive the forwarded traffic (e.g., '192.168.1.100').",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"fwd_port": {
				Description:  "The internal port(s) that will receive the forwarded traffic. Can be a single port (e.g., '8080') or a port range (e.g., '8080:8090').",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validators.PortRangeV2,
			},
			"log": {
				Description: "Enable logging of traffic matching this port forwarding rule. Useful for monitoring and troubleshooting.",
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
			},
			"name": {
				Description: "A friendly name for the port forwarding rule to help identify its purpose (e.g., 'Web Server' or 'Game Server').",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"port_forward_interface": {
				Description: "The WAN interface to apply the port forwarding rule to. Valid values are:\n" +
					"  * `wan` - Primary WAN interface\n" +
					"  * `wan2` - Secondary WAN interface\n" +
					"  * `both` - Both WAN interfaces",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"wan", "wan2", "both"}, false),
			},
			"protocol": {
				Description: "The network protocol(s) this rule applies to. Valid values are:\n" +
					"  * `tcp_udp` - Both TCP and UDP (default)\n" +
					"  * `tcp` - TCP only\n" +
					"  * `udp` - UDP only",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "tcp_udp",
				ValidateFunc: validation.StringInSlice([]string{"tcp_udp", "tcp", "udp"}, false),
			},
			"src_ip": {
				Description: "The source IP address or network in CIDR notation that is allowed to use this port forward. Use 'any' to allow all source IPs. " +
					"Examples: '203.0.113.1' for a single IP, '203.0.113.0/24' for a network, or 'any' for all IPs.",
				Type:     schema.TypeString,
				Optional: true,
				Default:  "any",
				ValidateFunc: validation.Any(
					validation.StringInSlice([]string{"any"}, false),
					validation.IsIPv4Address,
					utils.CidrValidate,
				),
			},
		},
	}
}

func resourcePortForwardCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	req, err := resourcePortForwardGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	resp, err := c.CreatePortForward(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return resourcePortForwardSetResourceData(resp, d, site)
}

func resourcePortForwardGetResourceData(d *schema.ResourceData) (*unifi.PortForward, error) {
	return &unifi.PortForward{
		DstPort:       d.Get("dst_port").(string),
		Enabled:       d.Get("enabled").(bool),
		Fwd:           d.Get("fwd_ip").(string),
		FwdPort:       d.Get("fwd_port").(string),
		Log:           d.Get("log").(bool),
		Name:          d.Get("name").(string),
		PfwdInterface: d.Get("port_forward_interface").(string),
		Proto:         d.Get("protocol").(string),
		Src:           d.Get("src_ip").(string),
	}, nil
}

func resourcePortForwardSetResourceData(resp *unifi.PortForward, d *schema.ResourceData, site string) diag.Diagnostics {
	d.Set("site", site)
	d.Set("dst_port", resp.DstPort)
	d.Set("enabled", resp.Enabled)
	d.Set("fwd_ip", resp.Fwd)
	d.Set("fwd_port", resp.FwdPort)
	d.Set("log", resp.Log)
	d.Set("name", resp.Name)
	d.Set("port_forward_interface", resp.PfwdInterface)
	d.Set("protocol", resp.Proto)
	d.Set("src_ip", resp.Src)

	return nil
}

func resourcePortForwardRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	id := d.Id()

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	resp, err := c.GetPortForward(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	return resourcePortForwardSetResourceData(resp, d, site)
}

func resourcePortForwardUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	req, err := resourcePortForwardGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req.ID = d.Id()

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	req.SiteID = site

	resp, err := c.UpdatePortForward(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourcePortForwardSetResourceData(resp, d, site)
}

func resourcePortForwardDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	id := d.Id()

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	err := c.DeletePortForward(ctx, site, id)
	return diag.FromErr(err)
}

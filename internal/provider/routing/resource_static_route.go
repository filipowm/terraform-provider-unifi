package routing

import (
	"context"
	"errors"
	"fmt"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/utils"
)

func ResourceStaticRoute() *schema.Resource {
	return &schema.Resource{
		Description: "The `unifi_static_route` resource manages static routes on UniFi Security Gateways (USG) and UniFi Dream Machines (UDM/UDM-Pro).\n\n" +
			"Static routes allow you to manually configure routing paths for specific networks. This is useful for:\n" +
			"  * Connecting to networks not directly connected to your UniFi gateway\n" +
			"  * Creating backup routes for redundancy\n" +
			"  * Implementing policy-based routing\n" +
			"  * Blocking traffic to specific networks using blackhole routes\n\n" +
			"Routes can be configured to use either a next-hop IP address, a specific interface, or as a blackhole route.",

		CreateContext: resourceStaticRouteCreate,
		ReadContext:   resourceStaticRouteRead,
		UpdateContext: resourceStaticRouteUpdate,
		DeleteContext: resourceStaticRouteDelete,
		Importer: &schema.ResourceImporter{
			StateContext: base.ImportSiteAndID,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the static route in the UniFi controller.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"site": {
				Description: "The name of the UniFi site where the static route should be created. If not specified, the default site will be used.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "A friendly name for the static route to help identify its purpose (e.g., 'Backup DC Link' or 'Cloud VPN Route').",
				Type:        schema.TypeString,
				Required:    true,
			},

			"network": {
				Description:      "The destination network in CIDR notation that this route will direct traffic to (e.g., '10.0.0.0/16' or '192.168.100.0/24').",
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     utils.CidrValidate,
				DiffSuppressFunc: utils.CidrDiffSuppress,
			},
			"type": {
				Description: "The type of static route. Valid values are:\n" +
					"  * `interface-route` - Route traffic through a specific interface\n" +
					"  * `nexthop-route` - Route traffic to a specific next-hop IP address\n" +
					"  * `blackhole` - Drop all traffic to this network",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"interface-route", "nexthop-route", "blackhole"}, false),
			},
			"distance": {
				Description: "The administrative distance for this route. Lower values are preferred. Use this to control route selection when multiple routes to the same destination exist.",
				Type:        schema.TypeInt,
				Required:    true,
			},

			"next_hop": {
				Description:  "The IP address of the next hop router for this route. Only used when type is set to 'nexthop-route'. This should be an IP address that is directly reachable from your UniFi gateway.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPAddress,
			},
			"interface": {
				Description: "The outbound interface to use for this route. Only used when type is set to 'interface-route'. Can be:\n" +
					"  * `WAN1` - Primary WAN interface\n" +
					"  * `WAN2` - Secondary WAN interface\n" +
					"  * A network ID for internal networks",
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceStaticRouteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	req, err := resourceStaticRouteGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.CreateRouting(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return resourceStaticRouteSetResourceData(resp, d, site)
}

func resourceStaticRouteGetResourceData(d *schema.ResourceData) (*unifi.Routing, error) {
	t, _ := d.Get("type").(string)

	name, _ := d.Get("name").(string)
	network, _ := d.Get("network").(string)
	distance, _ := d.Get("distance").(int)
	r := &unifi.Routing{
		Enabled: true,
		Type:    "static-route",

		Name:                name,
		StaticRouteNetwork:  utils.CidrZeroBased(network),
		StaticRouteDistance: distance,
		StaticRouteType:     t,
	}

	switch t {
	case "interface-route":
		r.StaticRouteInterface, _ = d.Get("interface").(string)
	case "nexthop-route":
		r.StaticRouteNexthop, _ = d.Get("next_hop").(string)
	case "blackhole":
	default:
		return nil, fmt.Errorf("unexpected route type: %q", t)
	}

	return r, nil
}

func resourceStaticRouteSetResourceData(resp *unifi.Routing, d *schema.ResourceData, site string) diag.Diagnostics {
	if err := d.Set("site", site); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", resp.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("network", utils.CidrZeroBased(resp.StaticRouteNetwork)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("distance", resp.StaticRouteDistance); err != nil {
		return diag.FromErr(err)
	}

	t := resp.StaticRouteType
	if err := d.Set("type", t); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("next_hop", ""); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("interface", ""); err != nil {
		return diag.FromErr(err)
	}

	switch t {
	case "interface-route":
		if err := d.Set("interface", resp.StaticRouteInterface); err != nil {
			return diag.FromErr(err)
		}
	case "nexthop-route":
		if err := d.Set("next_hop", resp.StaticRouteNexthop); err != nil {
			return diag.FromErr(err)
		}
	case "blackhole":
		// no additional attributes
	default:
		return diag.Errorf("unexpected static route type: %q", t)
	}

	return nil
}

func resourceStaticRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	id := d.Id()

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.GetRouting(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceStaticRouteSetResourceData(resp, d, site)
}

func resourceStaticRouteUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	req, err := resourceStaticRouteGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req.ID = d.Id()

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	req.SiteID = site

	// go-unifi v1.9.2's updateRouting converts a successful-but-empty PUT response
	// into unifi.ErrNotFound (see utils.ReReadOnUpdateNotFound / issue #98); re-read
	// to tell a spurious error from a genuine out-of-band deletion.
	resp, err := c.UpdateRouting(ctx, site, req)
	resp, found, err := utils.ReReadOnUpdateNotFound(resp, err, func() (*unifi.Routing, error) {
		return c.GetRouting(ctx, site, req.ID)
	})
	if err != nil {
		return diag.FromErr(err)
	}
	if !found {
		d.SetId("")
		return nil
	}

	return resourceStaticRouteSetResourceData(resp, d, site)
}

func resourceStaticRouteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	id := d.Id()

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	err := c.DeleteRouting(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		return nil
	}
	return diag.FromErr(err)
}

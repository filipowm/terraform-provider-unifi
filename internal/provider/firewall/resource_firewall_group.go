package firewall

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

func ResourceFirewallGroup() *schema.Resource {
	return &schema.Resource{
		Description: "The `unifi_firewall_group` resource manages reusable groups of addresses or ports that can be referenced in firewall rules (`unifi_firewall_rule`).\n\n" +
			"Firewall groups help organize and simplify firewall rule management by allowing you to:\n" +
			"  * Create collections of IP addresses or networks\n" +
			"  * Define sets of ports for specific services\n" +
			"  * Group IPv6 addresses for IPv6-specific rules\n\n" +
			"Common use cases include:\n" +
			"  * Creating groups of trusted IP addresses\n" +
			"  * Defining port groups for specific applications\n" +
			"  * Managing access control lists\n" +
			"  * Simplifying rule maintenance by using groups instead of individual IPs/ports",

		CreateContext: resourceFirewallGroupCreate,
		ReadContext:   resourceFirewallGroupRead,
		UpdateContext: resourceFirewallGroupUpdate,
		DeleteContext: resourceFirewallGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: utils.ImportSiteAndID,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the firewall group in the UniFi controller.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"site": {
				Description: "The name of the UniFi site where the firewall group should be created. If not specified, the default site will be used.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "A friendly name for the firewall group to help identify its purpose (e.g., 'Trusted IPs' or 'Web Server Ports'). " +
					"Must be unique within the site.",
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Description: "The type of firewall group. Valid values are:\n" +
					"  * `address-group` - Group of IPv4 addresses and/or networks (e.g., '192.168.1.10', '10.0.0.0/8')\n" +
					"  * `port-group` - Group of ports or port ranges (e.g., '80', '443', '8000-8080')\n" +
					"  * `ipv6-address-group` - Group of IPv6 addresses and/or networks",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"address-group", "port-group", "ipv6-address-group"}, false),
			},
			"members": {
				Description: "List of members in the group. The format depends on the group type:\n" +
					"  * For address-group: IPv4 addresses or CIDR notation (e.g., ['192.168.1.10', '10.0.0.0/8'])\n" +
					"  * For port-group: Port numbers or ranges (e.g., ['80', '443', '8000-8080'])\n" +
					"  * For ipv6-address-group: IPv6 addresses or CIDR notation",
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceFirewallGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	req, err := resourceFirewallGroupGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.CreateFirewallGroup(ctx, site, req)
	if err != nil {
		if base.IsServerErrorContains(err, "api.err.FirewallGroupExisted") {
			return diag.Errorf("firewall groups must have unique names: %s", err)
		}
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return resourceFirewallGroupSetResourceData(resp, d, site)
}

func resourceFirewallGroupGetResourceData(d *schema.ResourceData) (*unifi.FirewallGroup, error) {
	members, err := utils.SetToStringSlice(d.Get("members").(*schema.Set))
	if err != nil {
		return nil, err
	}

	return &unifi.FirewallGroup{
		Name:         d.Get("name").(string),
		GroupType:    d.Get("type").(string),
		GroupMembers: members,
	}, nil
}

func resourceFirewallGroupSetResourceData(resp *unifi.FirewallGroup, d *schema.ResourceData, site string) diag.Diagnostics {
	d.Set("site", site)
	d.Set("name", resp.Name)
	d.Set("type", resp.GroupType)
	d.Set("members", utils.StringSliceToSet(resp.GroupMembers))

	return nil
}

func resourceFirewallGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	id := d.Id()

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.GetFirewallGroup(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFirewallGroupSetResourceData(resp, d, site)
}

func resourceFirewallGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	req, err := resourceFirewallGroupGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req.ID = d.Id()

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	req.SiteID = site

	resp, err := c.UpdateFirewallGroup(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFirewallGroupSetResourceData(resp, d, site)
}

func resourceFirewallGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	id := d.Id()

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	err := c.DeleteFirewallGroup(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		return nil
	}
	return diag.FromErr(err)
}

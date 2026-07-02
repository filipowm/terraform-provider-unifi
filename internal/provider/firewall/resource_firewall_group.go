package firewall

import (
	"context"
	"errors"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/utils"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
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
			StateContext: base.ImportSiteAndID,
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
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	req, err := resourceFirewallGroupGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.CreateFirewallGroup(ctx, site, req)
	if err != nil {
		if utils.IsServerErrorContains(err, "api.err.FirewallGroupExisted") {
			return diag.Errorf("firewall groups must have unique names: %s", err)
		}
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return resourceFirewallGroupSetResourceData(resp, d, site)
}

func resourceFirewallGroupGetResourceData(d *schema.ResourceData) (*unifi.FirewallGroup, error) {
	membersSet, _ := d.Get("members").(*schema.Set)
	members, err := utils.SetToStringSlice(membersSet)
	if err != nil {
		return nil, err
	}

	name, _ := d.Get("name").(string)
	groupType, _ := d.Get("type").(string)

	return &unifi.FirewallGroup{
		Name:         name,
		GroupType:    groupType,
		GroupMembers: members,
	}, nil
}

func resourceFirewallGroupSetResourceData(resp *unifi.FirewallGroup, d *schema.ResourceData, site string) diag.Diagnostics {
	if err := d.Set("site", site); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", resp.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", resp.GroupType); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("members", utils.StringSliceToSet(resp.GroupMembers)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFirewallGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	id := d.Id()

	site, _ := d.Get("site").(string)
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
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	req, err := resourceFirewallGroupGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req.ID = d.Id()

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	req.SiteID = site

	// go-unifi v1.9.2's updateFirewallGroup converts a successful-but-empty PUT
	// response into unifi.ErrNotFound (see utils.ReReadOnUpdateNotFound / issue #98);
	// re-read to tell a spurious error from a genuine out-of-band deletion.
	resp, err := c.UpdateFirewallGroup(ctx, site, req)
	resp, found, err := utils.ReReadOnUpdateNotFound(resp, err, func() (*unifi.FirewallGroup, error) {
		return c.GetFirewallGroup(ctx, site, req.ID)
	})
	if err != nil {
		return diag.FromErr(err)
	}
	if !found {
		d.SetId("")
		return nil
	}

	return resourceFirewallGroupSetResourceData(resp, d, site)
}

func resourceFirewallGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	id := d.Id()

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	err := c.DeleteFirewallGroup(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		return nil
	}
	return diag.FromErr(err)
}

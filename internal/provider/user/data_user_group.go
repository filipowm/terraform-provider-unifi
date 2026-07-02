package user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
)

func DataUserGroup() *schema.Resource {
	return &schema.Resource{
		Description: "`unifi_user_group` data source can be used to retrieve the ID for a user group by name.",

		ReadContext: dataUserGroupRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of this AP group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"site": {
				Description: "The name of the site the user group is associated with.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
			},
			"name": {
				Description: "The name of the user group to look up.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Default",
			},

			"qos_rate_max_down": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"qos_rate_max_up": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataUserGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	name, _ := d.Get("name").(string)
	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	groups, err := c.ListUserGroup(ctx, site)
	if err != nil {
		return diag.FromErr(err)
	}
	for _, g := range groups {
		if g.Name == name {
			d.SetId(g.ID)

			if err := d.Set("site", site); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("qos_rate_max_down", g.QOSRateMaxDown); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("qos_rate_max_up", g.QOSRateMaxUp); err != nil {
				return diag.FromErr(err)
			}

			return nil
		}
	}

	return diag.Errorf("user group not found with name %s", name)
}

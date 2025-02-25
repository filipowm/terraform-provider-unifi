package network

import (
	"context"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataPortProfile() *schema.Resource {
	return &schema.Resource{
		Description: "`unifi_port_profile` data source can be used to retrieve the ID for a port profile by name.",

		ReadContext: dataPortProfileRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of this port profile.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"site": {
				Description: "The name of the site the port profile is associated with.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
			},
			"name": {
				Description: "The name of the port profile to look up.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "All",
			},
		},
	}
}

func dataPortProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	name := d.Get("name").(string)
	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	groups, err := c.ListPortProfile(ctx, site)
	if err != nil {
		return diag.FromErr(err)
	}
	for _, g := range groups {
		if g.Name == name {
			d.SetId(g.ID)

			d.Set("site", site)

			return nil
		}
	}

	return diag.Errorf("port profile not found with name %s", name)
}

package radius

import (
	"context"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataRADIUSProfile() *schema.Resource {
	return &schema.Resource{
		Description: "`unifi_radius_profile` data source can be used to retrieve the ID for a RADIUS profile by name.",

		ReadContext: dataRADIUSProfileRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of this AP group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"site": {
				Description: "The name of the site the RADIUS profile is associated with.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
			},
			"name": {
				Description: "The name of the RADIUS profile to look up.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Default",
			},
		},
	}
}

func dataRADIUSProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	name := d.Get("name").(string)
	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	profiles, err := c.ListRADIUSProfile(ctx, site)
	if err != nil {
		return diag.FromErr(err)
	}
	for _, g := range profiles {
		if g.Name == name {
			d.SetId(g.ID)
			d.Set("site", site)
			return nil
		}
	}

	return diag.Errorf("RADIUS profile not found with name %s", name)
}

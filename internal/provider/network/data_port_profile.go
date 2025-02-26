package network

import (
	"context"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataPortProfile() *schema.Resource {
	return &schema.Resource{
		Description: "`unifi_port_profile` data source can be used to retrieve port profile configurations from your UniFi network. " +
			"Port profiles define settings and behaviors for switch ports, including VLANs, PoE settings, and other port-specific configurations. " +
			"This data source is particularly useful when you need to reference existing port profiles in switch port configurations.",

		ReadContext: dataPortProfileRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the port profile. This is automatically assigned by UniFi and can be used " +
					"to reference this port profile in other resources.",
				Type:     schema.TypeString,
				Computed: true,
			},
			"site": {
				Description: "The name of the UniFi site where the port profile is configured. If not specified, the default site will be used.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
			},
			"name": {
				Description: "The name of the port profile to look up. This is the friendly name assigned to the profile in the UniFi controller. " +
					"Defaults to \"All\" if not specified, which is the default port profile in UniFi.",
				Type:     schema.TypeString,
				Optional: true,
				Default:  "All",
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

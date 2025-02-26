package site

import (
	"context"
	"errors"
	"fmt"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceSite() *schema.Resource {
	return &schema.Resource{
		Description: "The `unifi_site` resource manages UniFi sites, which are logical groupings of UniFi devices and their configurations.\n\n" +
			"Sites in UniFi are used to:\n" +
			"  * Organize network devices and settings for different physical locations\n" +
			"  * Isolate configurations between different networks or customers\n" +
			"  * Apply different policies and configurations to different groups of devices\n\n" +
			"Each site maintains its own:\n" +
			"  * Network configurations\n" +
			"  * Wireless networks (WLANs)\n" +
			"  * Security policies\n" +
			"  * Device configurations\n\n" +
			"A UniFi controller can manage multiple sites, making it ideal for multi-tenant or distributed network deployments.",

		CreateContext: resourceSiteCreate,
		ReadContext:   resourceSiteRead,
		UpdateContext: resourceSiteUpdate,
		DeleteContext: resourceSiteDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSiteImport,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the site in the UniFi controller. This is automatically generated when the site is created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "A human-readable description of the site (e.g., 'Main Office', 'Remote Branch', 'Client A Network'). " +
					"This is used as the display name in the UniFi controller interface.",
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Description: "The site's internal name in the UniFi system. This is automatically generated based on the description " +
					"and is used in API calls and configurations. It's typically a lowercase, hyphenated version of the description.",
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSiteImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	c := meta.(*base.Client)

	id := d.Id()
	_, err := c.GetSite(ctx, id)
	if err != nil {
		if !errors.Is(err, unifi.ErrNotFound) {
			return nil, err
		}
	} else {
		// id is a valid site
		return []*schema.ResourceData{d}, nil
	}

	// lookup site by name
	sites, err := c.ListSites(ctx)
	if err != nil {
		return nil, err
	}

	for _, s := range sites {
		if s.Name == id {
			d.SetId(s.ID)
			return []*schema.ResourceData{d}, nil
		}
	}

	return nil, fmt.Errorf("unable to find site %q on controller", id)
}

func resourceSiteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	description := d.Get("description").(string)

	resp, err := c.CreateSite(ctx, description)
	if err != nil {
		return diag.FromErr(err)
	}

	site := resp[0]
	d.SetId(site.ID)

	return resourceSiteSetResourceData(&site, d)
}

func resourceSiteSetResourceData(resp *unifi.Site, d *schema.ResourceData) diag.Diagnostics {
	d.Set("name", resp.Name)
	d.Set("description", resp.Description)
	return nil
}

func resourceSiteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	id := d.Id()

	site, err := c.GetSite(ctx, id)
	if errors.Is(err, unifi.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSiteSetResourceData(site, d)
}

func resourceSiteUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	site := &unifi.Site{
		ID:          d.Id(),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	resp, err := c.UpdateSite(ctx, site.Name, site.Description)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSiteSetResourceData(&resp[0], d)
}

func resourceSiteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)
	id := d.Id()
	_, err := c.DeleteSite(ctx, id)
	return diag.FromErr(err)
}

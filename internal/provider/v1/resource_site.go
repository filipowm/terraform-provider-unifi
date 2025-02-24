package v1

import (
	"context"
	"errors"
	"fmt"
	"github.com/filipowm/terraform-provider-unifi/internal/provider"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceSite() *schema.Resource {
	return &schema.Resource{
		Description: "`unifi_site` manages Unifi sites",

		CreateContext: resourceSiteCreate,
		ReadContext:   resourceSiteRead,
		UpdateContext: resourceSiteUpdate,
		DeleteContext: resourceSiteDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSiteImport,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the site.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "The description of the site.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "The name of the site.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceSiteImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	c := meta.(*provider.Client)

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
	c := meta.(*provider.Client)

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
	c := meta.(*provider.Client)

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
	c := meta.(*provider.Client)

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
	c := meta.(*provider.Client)
	id := d.Id()
	_, err := c.DeleteSite(ctx, id)
	return diag.FromErr(err)
}

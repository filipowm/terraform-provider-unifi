package user

import (
	"context"
	"errors"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceUserGroup() *schema.Resource {
	return &schema.Resource{
		Description: "The `unifi_user_group` resource manages client groups in the UniFi controller, which allow you to apply " +
			"common settings and restrictions to multiple network clients.\n\n" +
			"User groups are primarily used for:\n" +
			"  * Implementing Quality of Service (QoS) policies\n" +
			"  * Setting bandwidth limits for different types of users\n" +
			"  * Organizing clients into logical groups (e.g., Staff, Guests, IoT devices)\n\n" +
			"Key features include:\n" +
			"  * Download rate limiting\n" +
			"  * Upload rate limiting\n" +
			"  * Group-based policy application\n\n" +
			"User groups are particularly useful in:\n" +
			"  * Educational environments (different policies for staff and students)\n" +
			"  * Guest networks (limiting guest bandwidth)\n" +
			"  * Shared office spaces (managing different tenant groups)",

		CreateContext: resourceUserGroupCreate,
		ReadContext:   resourceUserGroupRead,
		UpdateContext: resourceUserGroupUpdate,
		DeleteContext: resourceUserGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: base.ImportSiteAndID,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the user group in the UniFi controller. This is automatically assigned.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"site": {
				Description: "The name of the UniFi site where this user group should be created. If not specified, the default site will be used.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "A descriptive name for the user group (e.g., 'Staff', 'Guests', 'IoT Devices'). This name will be " +
					"displayed in the UniFi controller interface and used when assigning clients to the group.",
				Type:     schema.TypeString,
				Required: true,
			},
			"qos_rate_max_down": {
				Description: "The maximum allowed download speed in Kbps (kilobits per second) for clients in this group. " +
					"Set to -1 for unlimited. Note: Values of 0 or 1 are not allowed.",
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
				// TODO: validate does not equal 0,1
			},
			"qos_rate_max_up": {
				Description: "The maximum allowed upload speed in Kbps (kilobits per second) for clients in this group. " +
					"Set to -1 for unlimited. Note: Values of 0 or 1 are not allowed.",
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
				// TODO: validate does not equal 0,1
			},
		},
	}
}

func resourceUserGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	req, err := resourceUserGroupGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.CreateUserGroup(context.TODO(), site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return resourceUserGroupSetResourceData(resp, d)
}

func resourceUserGroupGetResourceData(d *schema.ResourceData) (*unifi.UserGroup, error) {
	return &unifi.UserGroup{
		Name:           d.Get("name").(string),
		QOSRateMaxDown: d.Get("qos_rate_max_down").(int),
		QOSRateMaxUp:   d.Get("qos_rate_max_up").(int),
	}, nil
}

func resourceUserGroupSetResourceData(resp *unifi.UserGroup, d *schema.ResourceData) diag.Diagnostics {
	d.Set("name", resp.Name)
	d.Set("qos_rate_max_down", resp.QOSRateMaxDown)
	d.Set("qos_rate_max_up", resp.QOSRateMaxUp)

	return nil
}

func resourceUserGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	id := d.Id()

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.GetUserGroup(context.TODO(), site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceUserGroupSetResourceData(resp, d)
}

func resourceUserGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	req, err := resourceUserGroupGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req.ID = d.Id()

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	req.SiteID = site

	resp, err := c.UpdateUserGroup(context.TODO(), site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceUserGroupSetResourceData(resp, d)
}

func resourceUserGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	id := d.Id()

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	err := c.DeleteUserGroup(context.TODO(), site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		return nil
	}
	return diag.FromErr(err)
}

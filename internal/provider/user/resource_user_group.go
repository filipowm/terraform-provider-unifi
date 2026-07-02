package user

import (
	"context"
	"errors"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/utils"
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
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	req := resourceUserGroupGetResourceData(d)

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.CreateUserGroup(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return resourceUserGroupSetResourceData(resp, d)
}

func resourceUserGroupGetResourceData(d *schema.ResourceData) *unifi.UserGroup {
	name, _ := d.Get("name").(string)
	qosRateMaxDown, _ := d.Get("qos_rate_max_down").(int)
	qosRateMaxUp, _ := d.Get("qos_rate_max_up").(int)

	return &unifi.UserGroup{
		Name:           name,
		QOSRateMaxDown: qosRateMaxDown,
		QOSRateMaxUp:   qosRateMaxUp,
	}
}

func resourceUserGroupSetResourceData(resp *unifi.UserGroup, d *schema.ResourceData) diag.Diagnostics {
	if err := d.Set("name", resp.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("qos_rate_max_down", resp.QOSRateMaxDown); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("qos_rate_max_up", resp.QOSRateMaxUp); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceUserGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	id := d.Id()

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.GetUserGroup(ctx, site, id)
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
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	req := resourceUserGroupGetResourceData(d)

	req.ID = d.Id()

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	req.SiteID = site

	// go-unifi v1.9.2's updateUserGroup converts a successful-but-empty PUT response
	// into unifi.ErrNotFound (see utils.ReReadOnUpdateNotFound / issue #98); re-read
	// to tell a spurious error from a genuine out-of-band deletion.
	resp, err := c.UpdateUserGroup(ctx, site, req)
	resp, found, err := utils.ReReadOnUpdateNotFound(resp, err, func() (*unifi.UserGroup, error) {
		return c.GetUserGroup(ctx, site, req.ID)
	})
	if err != nil {
		return diag.FromErr(err)
	}
	if !found {
		d.SetId("")
		return nil
	}

	return resourceUserGroupSetResourceData(resp, d)
}

func resourceUserGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	id := d.Id()

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	err := c.DeleteUserGroup(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		return nil
	}
	return diag.FromErr(err)
}

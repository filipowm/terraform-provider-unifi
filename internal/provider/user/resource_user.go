package user

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

// TODO add validation: api.err.LocalDnsRecordRequiresFixedIp
// TODO require v7.3+ for local dns record.
func ResourceUser() *schema.Resource {
	return &schema.Resource{
		Description: "The `unifi_user` resource manages network clients in the UniFi controller, which are identified by their unique MAC addresses.\n\n" +
			"This resource allows you to manage:\n" +
			"  * Fixed IP assignments\n" +
			"  * User groups and network access\n" +
			"  * Network blocking and restrictions\n" +
			"  * Local DNS records\n\n" +
			"Important Notes:\n" +
			"  * Users are automatically created in the controller when devices connect to the network\n" +
			"  * By default, this resource can take over management of existing users (controlled by `allow_existing`)\n" +
			"  * Users can be 'forgotten' on destroy (controlled by `skip_forget_on_destroy`)\n\n" +
			"This resource is particularly useful for:\n" +
			"  * Managing static IP assignments\n" +
			"  * Implementing access control\n" +
			"  * Setting up local DNS records\n" +
			"  * Organizing devices into user groups",

		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: base.ImportSiteAndID,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the user in the UniFi controller. This is automatically assigned.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"site": {
				Description: "The name of the UniFi site where this user should be managed. If not specified, the default site will be used.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},
			"mac": {
				Description: "The MAC address of the device/client. This is used as the unique identifier and cannot be changed " +
					"after creation. Must be a valid MAC address format (e.g., '00:11:22:33:44:55'). MAC addresses are case-insensitive.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: utils.MacDiffSuppressFunc,
				ValidateFunc:     validation.StringMatch(utils.MacAddressRegexp, "Mac address is invalid"),
			},
			"name": {
				Description: "A friendly name for the device/client. This helps identify the device in the UniFi interface " +
					"(eg. 'Living Room TV', 'John's Laptop').",
				Type:     schema.TypeString,
				Required: true,
			},
			"user_group_id": {
				Description: "The ID of the user group this client belongs to. User groups can be used to apply common " +
					"settings and restrictions to multiple clients.",
				Type:     schema.TypeString,
				Optional: true,
			},
			"note": {
				Description: "Additional information about the client that you want to record (e.g., 'Company asset tag #12345', " +
					"'Guest device - expires 2024-03-01').",
				Type:     schema.TypeString,
				Optional: true,
			},
			// TODO: combine this with output IP for a single attribute ip_address?
			"fixed_ip": {
				Description: "A static IPv4 address to assign to this client. Ensure this IP is within the client's network range " +
					"and not already assigned to another device.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"network_id": {
				Description: "The ID of the network this client should be associated with. This is particularly important " +
					"when using VLANs or multiple networks.",
				Type:     schema.TypeString,
				Optional: true,
			},
			"blocked": {
				Description: "When true, this client will be blocked from accessing the network. Useful for temporarily " +
					"or permanently restricting network access for specific devices.",
				Type:     schema.TypeBool,
				Optional: true,
			},
			"dev_id_override": {
				Description: "Override the device fingerprint.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"local_dns_record": {
				Description: "A local DNS hostname for this client. When set, other devices on the network can resolve " +
					"this name to the client's IP address (e.g., 'printer.local', 'nas.home.arpa'). Such DNS record is automatically added to controller's DNS records.",
				Type:     schema.TypeString,
				Optional: true,
			},

			// these are "meta" attributes that control TF UX
			"allow_existing": {
				Description: "Allow this resource to take over management of an existing user in the UniFi controller. When true:\n" +
					"  * The resource can manage users that were automatically created when devices connected\n" +
					"  * Existing settings will be overwritten with the values specified in this resource\n" +
					"  * If false, attempting to manage an existing user will result in an error\n\n" +
					"Use with caution as it can modify settings for devices already connected to your network.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"skip_forget_on_destroy": {
				Description: "When false (default), the client will be 'forgotten' by the controller when this resource is destroyed. " +
					"Set to true to keep the client's history in the controller after the resource is removed from Terraform.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			// computed only attributes
			"hostname": {
				Description: "The hostname of the user.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ip": {
				Description: "The IP address of the user.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	req := resourceUserGetResourceData(d)

	allowExisting, _ := d.Get("allow_existing").(bool)

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.CreateUser(ctx, site, req)
	if err != nil {
		if !utils.IsServerErrorContains(err, "api.err.MacUsed") || !allowExisting {
			return diag.FromErr(err)
		}

		// mac in use, just absorb it
		mac, _ := d.Get("mac").(string)
		existing, err := c.GetUserByMAC(ctx, site, mac)
		if err != nil {
			return diag.FromErr(err)
		}

		req.ID = existing.ID
		req.SiteID = existing.SiteID

		// go-unifi v1.9.2's updateUser converts a successful-but-empty PUT response
		// into unifi.ErrNotFound (see utils.ReReadOnUpdateNotFound / issue #98); re-read
		// so reconciling an existing MAC does not spuriously fail.
		var found bool
		resp, err = c.UpdateUser(ctx, site, req)
		resp, found, err = utils.ReReadOnUpdateNotFound(resp, err, func() (*unifi.User, error) {
			return c.GetUser(ctx, site, req.ID)
		})
		if err != nil {
			return diag.FromErr(err)
		}
		if !found {
			return diag.Errorf("existing user %q (id %s) vanished while reconciling its configuration", req.MAC, req.ID)
		}
	}

	d.SetId(resp.ID)

	if blocked, _ := d.Get("blocked").(bool); blocked {
		mac, _ := d.Get("mac").(string)
		err := c.BlockUserByMAC(ctx, site, mac)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("dev_id_override") {
		mac, _ := d.Get("mac").(string)
		device, _ := d.Get("dev_id_override").(int)

		err := c.OverrideUserFingerprint(ctx, site, mac, device)
		if err != nil {
			return diag.FromErr(err)
		}

		resp.DevIdOverride = device
	}

	return resourceUserSetResourceData(resp, d, site)
}

func resourceUserGetResourceData(d *schema.ResourceData) *unifi.User {
	fixedIP, _ := d.Get("fixed_ip").(string)
	localDNSRecord, _ := d.Get("local_dns_record").(string)
	mac, _ := d.Get("mac").(string)
	name, _ := d.Get("name").(string)
	userGroupID, _ := d.Get("user_group_id").(string)
	note, _ := d.Get("note").(string)
	networkID, _ := d.Get("network_id").(string)
	blocked, _ := d.Get("blocked").(bool)
	devIDOverride, _ := d.Get("dev_id_override").(int)

	return &unifi.User{
		MAC:                   mac,
		Name:                  name,
		UserGroupID:           userGroupID,
		Note:                  note,
		FixedIP:               fixedIP,
		UseFixedIP:            fixedIP != "",
		LocalDNSRecord:        localDNSRecord,
		LocalDNSRecordEnabled: localDNSRecord != "",
		NetworkID:             networkID,
		// not sure if this matters/works
		Blocked:       blocked,
		DevIdOverride: devIDOverride,
	}
}

func resourceUserSetResourceData(resp *unifi.User, d *schema.ResourceData, site string) diag.Diagnostics {
	fixedIP := ""
	if resp.UseFixedIP {
		fixedIP = resp.FixedIP
	}

	localDNSRecord := ""
	if resp.LocalDNSRecordEnabled {
		localDNSRecord = resp.LocalDNSRecord
	}

	if err := d.Set("site", site); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("mac", resp.MAC); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", resp.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("user_group_id", resp.UserGroupID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("note", resp.Note); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("fixed_ip", fixedIP); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("local_dns_record", localDNSRecord); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("network_id", resp.NetworkID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("blocked", resp.Blocked); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("dev_id_override", resp.DevIdOverride); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("hostname", resp.Hostname); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ip", resp.IP); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	id := d.Id()

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.GetUser(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	// for some reason the IP address is only on this endpoint, so issue another request
	macResp, err := c.GetUserByMAC(ctx, site, resp.MAC)
	if errors.Is(err, unifi.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	// TODO: should this read the override fingerprint?

	resp.IP = macResp.IP

	return resourceUserSetResourceData(resp, d, site)
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	if d.HasChange("blocked") {
		mac, _ := d.Get("mac").(string)
		if blocked, _ := d.Get("blocked").(bool); blocked {
			err := c.BlockUserByMAC(ctx, site, mac)
			if err != nil {
				return diag.FromErr(err)
			}
		} else {
			err := c.UnblockUserByMAC(ctx, site, mac)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("dev_id_override") {
		mac, _ := d.Get("mac").(string)
		device, _ := d.Get("dev_id_override").(int)

		err := c.OverrideUserFingerprint(ctx, site, mac, device)
		if err != nil {
			return diag.FromErr(err)
		}

		if !d.HasChangesExcept("dev_id_override") {
			return nil
		}
	}

	req := resourceUserGetResourceData(d)

	req.ID = d.Id()
	req.SiteID = site

	// go-unifi v1.9.2's updateUser converts a successful-but-empty PUT response into
	// unifi.ErrNotFound (see utils.ReReadOnUpdateNotFound / issue #98); re-read to tell
	// a spurious error from a genuine out-of-band deletion.
	resp, err := c.UpdateUser(ctx, site, req)
	resp, found, err := utils.ReReadOnUpdateNotFound(resp, err, func() (*unifi.User, error) {
		return c.GetUser(ctx, site, req.ID)
	})
	if err != nil {
		return diag.FromErr(err)
	}
	if !found {
		d.SetId("")
		return nil
	}

	return resourceUserSetResourceData(resp, d, site)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	id := d.Id()

	if skip, _ := d.Get("skip_forget_on_destroy").(bool); skip {
		return nil
	}

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	// lookup MAC instead of trusting state
	u, err := c.GetUser(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	err = c.DeleteUserByMAC(ctx, site, u.MAC)
	return diag.FromErr(err)
}

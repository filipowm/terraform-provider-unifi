package user

import (
	"context"
	"errors"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

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
	c := meta.(*base.Client)

	req, err := resourceUserGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	allowExisting := d.Get("allow_existing").(bool)

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.CreateUser(ctx, site, req)
	if err != nil {
		if !base.IsServerErrorContains(err, "api.err.MacUsed") || !allowExisting {
			return diag.FromErr(err)
		}

		// mac in use, just absorb it
		mac := d.Get("mac").(string)
		existing, err := c.GetUserByMAC(ctx, site, mac)
		if err != nil {
			return diag.FromErr(err)
		}

		req.ID = existing.ID
		req.SiteID = existing.SiteID

		resp, err = c.UpdateUser(ctx, site, req)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(resp.ID)

	if d.Get("blocked").(bool) {
		err := c.BlockUserByMAC(ctx, site, d.Get("mac").(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("dev_id_override") {
		mac := d.Get("mac").(string)
		device := d.Get("dev_id_override").(int)

		err := c.OverrideUserFingerprint(context.TODO(), site, mac, device)
		if err != nil {
			return diag.FromErr(err)
		}

		resp.DevIdOverride = device
	}

	return resourceUserSetResourceData(resp, d, site)
}

func resourceUserGetResourceData(d *schema.ResourceData) (*unifi.User, error) {
	fixedIP := d.Get("fixed_ip").(string)
	localDnsRecord := d.Get("local_dns_record").(string)

	return &unifi.User{
		MAC:                   d.Get("mac").(string),
		Name:                  d.Get("name").(string),
		UserGroupID:           d.Get("user_group_id").(string),
		Note:                  d.Get("note").(string),
		FixedIP:               fixedIP,
		UseFixedIP:            fixedIP != "",
		LocalDNSRecord:        localDnsRecord,
		LocalDNSRecordEnabled: localDnsRecord != "",
		NetworkID:             d.Get("network_id").(string),
		// not sure if this matters/works
		Blocked:       d.Get("blocked").(bool),
		DevIdOverride: d.Get("dev_id_override").(int),
	}, nil
}

func resourceUserSetResourceData(resp *unifi.User, d *schema.ResourceData, site string) diag.Diagnostics {
	fixedIP := ""
	if resp.UseFixedIP {
		fixedIP = resp.FixedIP
	}

	localDnsRecord := ""
	if resp.LocalDNSRecordEnabled {
		localDnsRecord = resp.LocalDNSRecord
	}

	d.Set("site", site)
	d.Set("mac", resp.MAC)
	d.Set("name", resp.Name)
	d.Set("user_group_id", resp.UserGroupID)
	d.Set("note", resp.Note)
	d.Set("fixed_ip", fixedIP)
	d.Set("local_dns_record", localDnsRecord)
	d.Set("network_id", resp.NetworkID)
	d.Set("blocked", resp.Blocked)
	d.Set("dev_id_override", resp.DevIdOverride)
	d.Set("hostname", resp.Hostname)
	d.Set("ip", resp.IP)

	return nil
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	id := d.Id()

	site := d.Get("site").(string)
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
	c := meta.(*base.Client)

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	if d.HasChange("blocked") {
		mac := d.Get("mac").(string)
		if d.Get("blocked").(bool) {
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
		mac := d.Get("mac").(string)
		device := d.Get("dev_id_override").(int)

		err := c.OverrideUserFingerprint(context.TODO(), site, mac, device)
		if err != nil {
			return diag.FromErr(err)
		}

		if !d.HasChangesExcept("dev_id_override") {
			return nil
		}
	}

	req, err := resourceUserGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req.ID = d.Id()
	req.SiteID = site

	resp, err := c.UpdateUser(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceUserSetResourceData(resp, d, site)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	id := d.Id()

	if d.Get("skip_forget_on_destroy").(bool) {
		return nil
	}

	site := d.Get("site").(string)
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

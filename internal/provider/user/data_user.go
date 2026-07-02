package user

import (
	"context"
	"strings"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func DataUser() *schema.Resource {
	return &schema.Resource{
		Description: "`unifi_user` retrieves properties of a user (or \"client\" in the UI) of the network by MAC address.",

		ReadContext: dataUserRead,

		Schema: map[string]*schema.Schema{
			"site": {
				Description: "The name of the site the user is associated with.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
			},
			"mac": {
				Description:      "The MAC address of the user.",
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: utils.MacDiffSuppressFunc,
				ValidateFunc:     validation.StringMatch(utils.MacAddressRegexp, "Mac address is invalid"),
			},

			// read-only / computed
			"id": {
				Description: "The ID of the user.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "The name of the user.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"user_group_id": {
				Description: "The user group ID for the user.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"note": {
				Description: "A note with additional information for the user.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"fixed_ip": {
				Description: "Fixed IPv4 address set for this user.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"network_id": {
				Description: "The network ID for this user.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"blocked": {
				Description: "Specifies whether this user should be blocked from the network.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"dev_id_override": {
				Description: "Override the device fingerprint.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
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
			"local_dns_record": {
				Description: "The local DNS record for this user.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	mac, _ := d.Get("mac").(string)

	macResp, err := c.GetUserByMAC(ctx, site, strings.ToLower(mac))
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := c.GetUser(ctx, site, macResp.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	// for some reason the IP address is only on this endpoint, so issue another request

	resp.IP = macResp.IP
	fixedIP := ""
	if resp.UseFixedIP {
		fixedIP = resp.FixedIP
	}
	localDNSRecord := ""
	if resp.LocalDNSRecordEnabled {
		localDNSRecord = resp.LocalDNSRecord
	}
	d.SetId(resp.ID)
	if err := d.Set("blocked", resp.Blocked); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("dev_id_override", resp.DevIdOverride); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("fixed_ip", fixedIP); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("hostname", resp.Hostname); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ip", resp.IP); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("local_dns_record", localDNSRecord); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("mac", resp.MAC); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", resp.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("network_id", resp.NetworkID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("note", resp.Note); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("site", site); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("user_group_id", resp.UserGroupID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

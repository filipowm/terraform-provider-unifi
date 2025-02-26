package settings

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/utils"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceSettingUsgLock = sync.Mutex{}

func resourceSettingUsgLocker(f func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics) func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		resourceSettingUsgLock.Lock()
		defer resourceSettingUsgLock.Unlock()
		return f(ctx, d, meta)
	}
}

func ResourceSettingUsg() *schema.Resource {
	return &schema.Resource{
		Description: "The `unifi_setting_usg` resource manages advanced settings for UniFi Security Gateways (USG) and UniFi Dream Machines (UDM/UDM-Pro).\n\n" +
			"This resource allows you to configure gateway-specific features including:\n" +
			"  * Multicast DNS (mDNS) for service discovery\n" +
			"  * DHCP relay for forwarding DHCP requests to external servers\n\n" +
			"These settings are particularly useful for:\n" +
			"  * Enabling device discovery across VLANs (using mDNS)\n" +
			"  * Centralizing DHCP management in enterprise environments\n" +
			"  * Integration with existing network infrastructure\n\n" +
			"Note: Some settings may not be available on all controller versions. For example, multicast_dns_enabled is not supported on UniFi OS v7+.",

		CreateContext: resourceSettingUsgLocker(resourceSettingUsgUpsert),
		ReadContext:   resourceSettingUsgLocker(resourceSettingUsgRead),
		UpdateContext: resourceSettingUsgLocker(resourceSettingUsgUpsert),
		DeleteContext: schema.NoopContext,
		Importer: &schema.ResourceImporter{
			StateContext: utils.ImportSiteAndID,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the USG settings configuration in the UniFi controller.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"site": {
				Description: "The name of the UniFi site where these USG settings should be applied. If not specified, the default site will be used.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},
			"multicast_dns_enabled": {
				Description: "Enable multicast DNS (mDNS/Bonjour/Avahi) forwarding across VLANs. This allows devices to discover services " +
					"(like printers, Chromecasts, etc.) even when they are on different networks. Note: Not supported on UniFi OS v7+.",
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"dhcp_relay_servers": {
				Description: "List of up to 5 DHCP relay servers (specified by IP address) that will receive forwarded DHCP requests. " +
					"This is useful when you want to use external DHCP servers instead of the built-in DHCP server. " +
					"Example: ['192.168.1.5', '192.168.2.5']",
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 5,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.IsIPv4Address,
						// this doesn't let blank through
						validation.StringLenBetween(1, 50),
					),
				},
			},
		},
	}
}

func resourceSettingUsgUpdateResourceData(d *schema.ResourceData, meta interface{}, setting *unifi.SettingUsg) error {
	c := meta.(*base.Client)

	//nolint // GetOkExists is deprecated, but using here:
	if mdns, hasMdns := d.GetOkExists("multicast_dns_enabled"); hasMdns {
		if c.IsControllerV7() {
			return fmt.Errorf("multicast_dns_enabled is not supported on controller version %v", c.Version)
		}

		setting.MdnsEnabled = mdns.(bool)
	}

	dhcpRelay, err := utils.ListToStringSlice(d.Get("dhcp_relay_servers").([]interface{}))
	if err != nil {
		return fmt.Errorf("unable to convert dhcp_relay_servers to string slice: %w", err)
	}
	setting.DHCPRelayServer1 = append(dhcpRelay, "")[0]
	setting.DHCPRelayServer2 = append(dhcpRelay, "", "")[1]
	setting.DHCPRelayServer3 = append(dhcpRelay, "", "", "")[2]
	setting.DHCPRelayServer4 = append(dhcpRelay, "", "", "", "")[3]
	setting.DHCPRelayServer5 = append(dhcpRelay, "", "", "", "", "")[4]

	return nil
}

func resourceSettingUsgUpsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	req, err := c.GetSettingUsg(ctx, c.Site)
	if err != nil {
		return diag.FromErr(err)
	}

	err = resourceSettingUsgUpdateResourceData(d, meta, req)
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := c.UpdateSettingUsg(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)
	return resourceSettingUsgSetResourceData(resp, d, meta, site)
}

func resourceSettingUsgSetResourceData(resp *unifi.SettingUsg, d *schema.ResourceData, meta interface{}, site string) diag.Diagnostics {
	d.Set("site", site)
	d.Set("multicast_dns_enabled", resp.MdnsEnabled)

	dhcpRelay := []string{}
	for _, s := range []string{
		resp.DHCPRelayServer1,
		resp.DHCPRelayServer2,
		resp.DHCPRelayServer3,
		resp.DHCPRelayServer4,
		resp.DHCPRelayServer5,
	} {
		if s == "" {
			continue
		}
		dhcpRelay = append(dhcpRelay, s)
	}
	d.Set("dhcp_relay_servers", dhcpRelay)

	return nil
}

func resourceSettingUsgRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.GetSettingUsg(ctx, site)
	if errors.Is(err, unifi.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSettingUsgSetResourceData(resp, d, meta, site)
}

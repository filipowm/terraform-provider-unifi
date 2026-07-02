package radius

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/utils"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceRadiusProfile() *schema.Resource {
	return &schema.Resource{
		Description: "The `unifi_radius_profile` resource manages RADIUS authentication profiles for UniFi networks.\n\n" +
			"RADIUS (Remote Authentication Dial-In User Service) profiles enable enterprise-grade authentication and authorization for:\n" +
			"  * 802.1X network access control\n" +
			"  * WPA2/WPA3-Enterprise wireless networks\n" +
			"  * Dynamic VLAN assignment\n" +
			"  * User activity accounting\n\n" +
			"Each profile can be configured with:\n" +
			"  * Multiple authentication and accounting servers\n" +
			"  * VLAN assignment settings\n" +
			"  * Accounting update intervals",

		CreateContext: resourceRadiusProfileCreate,
		ReadContext:   resourceRadiusProfileRead,
		UpdateContext: resourceRadiusProfileUpdate,
		DeleteContext: resourceRadiusProfileDelete,
		Importer: &schema.ResourceImporter{
			StateContext: importRadiusProfile,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the RADIUS profile in the UniFi controller.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"site": {
				Description: "The name of the UniFi site where the RADIUS profile should be created. If not specified, the default site will be used.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "A friendly name for the RADIUS profile to help identify its purpose (e.g., 'Corporate Users' or 'Guest Access').",
				Type:        schema.TypeString,
				Required:    true,
			},
			"accounting_enabled": {
				Description: "Enable RADIUS accounting to track user sessions, including login/logout times and data usage. Useful for billing and audit purposes.",
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
			},
			"interim_update_enabled": {
				Description: "Enable periodic updates during active sessions. This allows tracking of ongoing session data like bandwidth usage.",
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
			},
			"interim_update_interval": {
				Description: "The interval (in seconds) between interim updates when `interim_update_enabled` is true. Default is 3600 seconds (1 hour).",
				Type:        schema.TypeInt,
				Default:     3600,
				Optional:    true,
			},
			"use_usg_acct_server": {
				Description: "Use the controller as a RADIUS accounting server. This allows local accounting without an external RADIUS server.",
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
			},
			"use_usg_auth_server": {
				Description: "Use the controller as a RADIUS authentication server. This allows local authentication without an external RADIUS server.",
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
			},
			"vlan_enabled": {
				Description: "Enable VLAN assignment for wired clients based on RADIUS attributes. This allows network segmentation based on user authentication.",
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
			},
			"vlan_wlan_mode": {
				Description: "VLAN assignment mode for wireless networks. Valid values are:\n" +
					"  * `disabled` - Do not use RADIUS-assigned VLANs\n" +
					"  * `optional` - Use RADIUS-assigned VLAN if provided\n" +
					"  * `required` - Require RADIUS-assigned VLAN for authentication to succeed",
				Type:         schema.TypeString,
				Default:      "",
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"disabled", "optional", "required"}, false),
			},
			"auth_server": {
				Description: "List of RADIUS authentication servers to use with this profile. Multiple servers provide failover - if the first " +
					"server is unreachable, the system will try the next server in the list. Each server requires:\n" +
					"  * IP address of the RADIUS server\n" +
					"  * Shared secret for secure communication",
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Description: "The IPv4 address of the RADIUS authentication server (e.g., '192.168.1.100'). Must be reachable from " +
								"your UniFi network.",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsIPAddress,
						},
						"port": {
							Description: "The UDP port number where the RADIUS authentication service is listening. The standard port is 1812, " +
								"but this can be changed if needed to match your server configuration.",
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1812,
							ValidateFunc: validation.IsPortNumber,
						},
						"xsecret": {
							Description: "The shared secret key used to secure communication between the UniFi controller and the RADIUS server. " +
								"This must match the secret configured on your RADIUS server.",
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
			"acct_server": {
				Description: "List of RADIUS accounting servers to use with this profile. Accounting servers track session data like " +
					"connection time and data usage. Each server requires:\n" +
					"  * IP address of the RADIUS server\n" +
					"  * Port number (default: 1813)\n" +
					"  * Shared secret for secure communication",
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Description: "The IPv4 address of the RADIUS accounting server (e.g., '192.168.1.100'). Must be reachable from " +
								"your UniFi network.",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsIPAddress,
						},
						"port": {
							Description: "The UDP port number where the RADIUS accounting service is listening. The standard port is 1813, " +
								"but this can be changed if needed to match your server configuration.",
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1813,
							ValidateFunc: validation.IsPortNumber,
						},
						"xsecret": {
							Description: "The shared secret key used to secure communication between the UniFi controller and the RADIUS server. " +
								"This must match the secret configured on your RADIUS server.",
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
		},
	}
}

func setToAuthServers(set []interface{}) ([]unifi.RADIUSProfileAuthServers, error) {
	var authServers []unifi.RADIUSProfileAuthServers
	for _, item := range set {
		data, ok := item.(map[string]interface{})
		if !ok {
			return nil, errors.New("unexpected data in block")
		}
		authServers = append(authServers, toAuthServer(data))
	}
	return authServers, nil
}

func setToAcctServers(set []interface{}) ([]unifi.RADIUSProfileAcctServers, error) {
	var acctServers []unifi.RADIUSProfileAcctServers
	for _, item := range set {
		data, ok := item.(map[string]interface{})
		if !ok {
			return nil, errors.New("unexpected data in block")
		}
		acctServers = append(acctServers, toAcctServer(data))
	}
	return acctServers, nil
}

func toAuthServer(data map[string]interface{}) unifi.RADIUSProfileAuthServers {
	ip, _ := data["ip"].(string)
	port, _ := data["port"].(int)
	xsecret, _ := data["xsecret"].(string)
	return unifi.RADIUSProfileAuthServers{
		IP:      ip,
		Port:    port,
		XSecret: xsecret,
	}
}

func toAcctServer(data map[string]interface{}) unifi.RADIUSProfileAcctServers {
	ip, _ := data["ip"].(string)
	port, _ := data["port"].(int)
	xsecret, _ := data["xsecret"].(string)
	return unifi.RADIUSProfileAcctServers{
		IP:      ip,
		Port:    port,
		XSecret: xsecret,
	}
}

func setFromAuthServers(authServers []unifi.RADIUSProfileAuthServers) []map[string]interface{} {
	list := make([]map[string]interface{}, 0, len(authServers))
	for _, authServer := range authServers {
		list = append(list, fromAuthServer(authServer))
	}
	return list
}

func setFromAcctServers(acctServers []unifi.RADIUSProfileAcctServers) []map[string]interface{} {
	list := make([]map[string]interface{}, 0, len(acctServers))
	for _, acctServer := range acctServers {
		list = append(list, fromAcctServer(acctServer))
	}
	return list
}

func fromAuthServer(sshKey unifi.RADIUSProfileAuthServers) map[string]interface{} {
	return map[string]interface{}{
		"ip":      sshKey.IP,
		"port":    sshKey.Port,
		"xsecret": sshKey.XSecret,
	}
}

func fromAcctServer(sshKey unifi.RADIUSProfileAcctServers) map[string]interface{} {
	return map[string]interface{}{
		"ip":      sshKey.IP,
		"port":    sshKey.Port,
		"xsecret": sshKey.XSecret,
	}
}

func resourceRadiusProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}
	req, err := resourceRadiusProfileGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	resp, err := c.CreateRADIUSProfile(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(resp.ID)

	return resourceRadiusProfileSetResourceData(resp, d, site)
}

func resourceRadiusProfileGetResourceData(d *schema.ResourceData) (*unifi.RADIUSProfile, error) {
	authServerData, _ := d.Get("auth_server").([]interface{})
	authServers, err := setToAuthServers(authServerData)
	if err != nil {
		return nil, fmt.Errorf("unable to auth_server ssh_key block: %w", err)
	}
	acctServerData, _ := d.Get("acct_server").([]interface{})
	acctServers, err := setToAcctServers(acctServerData)
	if err != nil {
		return nil, fmt.Errorf("unable to acct_server ssh_key block: %w", err)
	}
	name, _ := d.Get("name").(string)
	interimUpdateEnabled, _ := d.Get("interim_update_enabled").(bool)
	interimUpdateInterval, _ := d.Get("interim_update_interval").(int)
	accountingEnabled, _ := d.Get("accounting_enabled").(bool)
	useUsgAcctServer, _ := d.Get("use_usg_acct_server").(bool)
	useUsgAuthServer, _ := d.Get("use_usg_auth_server").(bool)
	vlanEnabled, _ := d.Get("vlan_enabled").(bool)
	vlanWLANMode, _ := d.Get("vlan_wlan_mode").(string)
	return &unifi.RADIUSProfile{
		Name:                  name,
		InterimUpdateEnabled:  interimUpdateEnabled,
		InterimUpdateInterval: interimUpdateInterval,
		AccountingEnabled:     accountingEnabled,
		UseUsgAcctServer:      useUsgAcctServer,
		UseUsgAuthServer:      useUsgAuthServer,
		VLANEnabled:           vlanEnabled,
		VLANWLANMode:          vlanWLANMode,
		AuthServers:           authServers,
		AcctServers:           acctServers,
	}, nil
}

func resourceRadiusProfileSetResourceData(resp *unifi.RADIUSProfile, d *schema.ResourceData, site string) diag.Diagnostics {
	authServers := setFromAuthServers(resp.AuthServers)
	acctServers := setFromAcctServers(resp.AcctServers)

	if err := d.Set("site", site); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", resp.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("interim_update_enabled", resp.InterimUpdateEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("interim_update_interval", resp.InterimUpdateInterval); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("accounting_enabled", resp.AccountingEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("use_usg_acct_server", resp.UseUsgAcctServer); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("use_usg_auth_server", resp.UseUsgAuthServer); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("vlan_enabled", resp.VLANEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("vlan_wlan_mode", resp.VLANWLANMode); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("auth_server", authServers); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("acct_server", acctServers); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceRadiusProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	id := d.Id()

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	resp, err := c.GetRADIUSProfile(ctx, site, id)
	if errors.Is(err, unifi.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceRadiusProfileSetResourceData(resp, d, site)
}

func resourceRadiusProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	req, err := resourceRadiusProfileGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req.ID = d.Id()

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	req.SiteID = site

	// go-unifi v1.9.2's updateRADIUSProfile converts a successful-but-empty PUT
	// response into unifi.ErrNotFound (see utils.ReReadOnUpdateNotFound / issue #98);
	// re-read to tell a spurious error from a genuine out-of-band deletion.
	resp, err := c.UpdateRADIUSProfile(ctx, site, req)
	resp, found, err := utils.ReReadOnUpdateNotFound(resp, err, func() (*unifi.RADIUSProfile, error) {
		return c.GetRADIUSProfile(ctx, site, req.ID)
	})
	if err != nil {
		return diag.FromErr(err)
	}
	if !found {
		d.SetId("")
		return nil
	}

	return resourceRadiusProfileSetResourceData(resp, d, site)
}

func resourceRadiusProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(*base.Client)
	if !ok {
		return diag.Errorf("unexpected meta type: %T", meta)
	}

	id := d.Id()

	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	err := c.DeleteRADIUSProfile(ctx, site, id)
	return diag.FromErr(err)
}

func importRadiusProfile(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	c, ok := meta.(*base.Client)
	if !ok {
		return nil, fmt.Errorf("unexpected meta type: %T", meta)
	}
	id := d.Id()
	site, _ := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	if strings.Contains(id, ":") {
		importParts := strings.SplitN(id, ":", 2)
		site = importParts[0]
		id = importParts[1]
	}

	if strings.HasPrefix(id, "name=") {
		targetName := strings.TrimPrefix(id, "name=")
		var err error
		if id, err = getRadiusProfileIDByName(ctx, c.Client, targetName, site); err != nil {
			return nil, err
		}
	}

	if id != "" {
		d.SetId(id)
	}
	if site != "" {
		if err := d.Set("site", site); err != nil {
			return nil, err
		}
	}

	return []*schema.ResourceData{d}, nil
}

func getRadiusProfileIDByName(ctx context.Context, client unifi.Client, profileName, site string) (string, error) {
	radiusProfiles, err := client.ListRADIUSProfile(ctx, site)
	if err != nil {
		return "", err
	}

	idMatchingName := ""
	allNames := []string{}
	for _, profile := range radiusProfiles {
		allNames = append(allNames, profile.Name)
		if profile.Name != profileName {
			continue
		}
		if idMatchingName != "" {
			return "", fmt.Errorf("found multiple RADIUS profiles with name '%s'", profileName)
		}
		idMatchingName = profile.ID
	}
	if idMatchingName == "" {
		return "", fmt.Errorf("found no RADIUS profile with name '%s', found: %s", profileName, strings.Join(allNames, ", "))
	}
	return idMatchingName, nil
}

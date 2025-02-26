package radius

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"

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
			return nil, fmt.Errorf("unexpected data in block")
		}
		authServer, err := toAuthServer(data)
		if err != nil {
			return nil, fmt.Errorf("unable to create port override: %w", err)
		}
		authServers = append(authServers, authServer)
	}
	return authServers, nil
}

func setToAcctServers(set []interface{}) ([]unifi.RADIUSProfileAcctServers, error) {
	var acctServers []unifi.RADIUSProfileAcctServers
	for _, item := range set {
		data, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected data in block")
		}
		accServer, err := toAcctServer(data)
		if err != nil {
			return nil, fmt.Errorf("unable to create port override: %w", err)
		}
		acctServers = append(acctServers, accServer)
	}
	return acctServers, nil
}

func toAuthServer(data map[string]interface{}) (unifi.RADIUSProfileAuthServers, error) {
	return unifi.RADIUSProfileAuthServers{
		IP:      data["ip"].(string),
		Port:    data["port"].(int),
		XSecret: data["xsecret"].(string),
	}, nil
}

func toAcctServer(data map[string]interface{}) (unifi.RADIUSProfileAcctServers, error) {
	return unifi.RADIUSProfileAcctServers{
		IP:      data["ip"].(string),
		Port:    data["port"].(int),
		XSecret: data["xsecret"].(string),
	}, nil
}

func setFromAuthServers(authServers []unifi.RADIUSProfileAuthServers) ([]map[string]interface{}, error) {
	list := make([]map[string]interface{}, 0, len(authServers))
	for _, authServer := range authServers {
		v, err := fromAuthServer(authServer)
		if err != nil {
			return nil, fmt.Errorf("unable to parse ssh key: %w", err)
		}
		list = append(list, v)
	}
	return list, nil
}

func setFromAcctServers(acctServers []unifi.RADIUSProfileAcctServers) ([]map[string]interface{}, error) {
	list := make([]map[string]interface{}, 0, len(acctServers))
	for _, acctServer := range acctServers {
		v, err := fromAcctServer(acctServer)
		if err != nil {
			return nil, fmt.Errorf("unable to parse ssh key: %w", err)
		}
		list = append(list, v)
	}
	return list, nil
}

func fromAuthServer(sshKey unifi.RADIUSProfileAuthServers) (map[string]interface{}, error) {
	return map[string]interface{}{
		"ip":      sshKey.IP,
		"port":    sshKey.Port,
		"xsecret": sshKey.XSecret,
	}, nil
}

func fromAcctServer(sshKey unifi.RADIUSProfileAcctServers) (map[string]interface{}, error) {
	return map[string]interface{}{
		"ip":      sshKey.IP,
		"port":    sshKey.Port,
		"xsecret": sshKey.XSecret,
	}, nil
}

func resourceRadiusProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)
	req, err := resourceRadiusProfileGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	site := d.Get("site").(string)
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
	authServers, err := setToAuthServers(d.Get("auth_server").([]interface{}))
	if err != nil {
		return nil, fmt.Errorf("unable to auth_server ssh_key block: %w", err)
	}
	acctServers, err := setToAcctServers(d.Get("acct_server").([]interface{}))
	if err != nil {
		return nil, fmt.Errorf("unable to acct_server ssh_key block: %w", err)
	}
	return &unifi.RADIUSProfile{
		Name:                  d.Get("name").(string),
		InterimUpdateEnabled:  d.Get("interim_update_enabled").(bool),
		InterimUpdateInterval: d.Get("interim_update_interval").(int),
		AccountingEnabled:     d.Get("accounting_enabled").(bool),
		UseUsgAcctServer:      d.Get("use_usg_acct_server").(bool),
		UseUsgAuthServer:      d.Get("use_usg_auth_server").(bool),
		VLANEnabled:           d.Get("vlan_enabled").(bool),
		VLANWLANMode:          d.Get("vlan_wlan_mode").(string),
		AuthServers:           authServers,
		AcctServers:           acctServers,
	}, nil
}

func resourceRadiusProfileSetResourceData(resp *unifi.RADIUSProfile, d *schema.ResourceData, site string) diag.Diagnostics {
	authServers, err := setFromAuthServers(resp.AuthServers)
	if err != nil {
		return diag.FromErr(err)
	}
	acctServers, err := setFromAcctServers(resp.AcctServers)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("site", site)
	d.Set("name", resp.Name)

	d.Set("interim_update_enabled", resp.InterimUpdateEnabled)
	d.Set("interim_update_interval", resp.InterimUpdateInterval)
	d.Set("accounting_enabled", resp.AccountingEnabled)
	d.Set("use_usg_acct_server", resp.UseUsgAcctServer)
	d.Set("use_usg_auth_server", resp.UseUsgAuthServer)
	d.Set("vlan_enabled", resp.VLANEnabled)
	d.Set("vlan_wlan_mode", resp.VLANWLANMode)
	d.Set("auth_server", authServers)
	d.Set("acct_server", acctServers)
	return nil
}

func resourceRadiusProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	id := d.Id()

	site := d.Get("site").(string)
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
	c := meta.(*base.Client)

	req, err := resourceRadiusProfileGetResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req.ID = d.Id()

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}
	req.SiteID = site

	resp, err := c.UpdateRADIUSProfile(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceRadiusProfileSetResourceData(resp, d, site)
}

func resourceRadiusProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	id := d.Id()

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	err := c.DeleteRADIUSProfile(ctx, site, id)
	return diag.FromErr(err)
}

func importRadiusProfile(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	c := meta.(*base.Client)
	id := d.Id()
	site := d.Get("site").(string)
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
		d.Set("site", site)
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
			return "", fmt.Errorf("Found multiple RADIUS profiles with name '%s'", profileName)
		}
		idMatchingName = profile.ID
	}
	if idMatchingName == "" {
		return "", fmt.Errorf("Found no RADIUS profile with name '%s', found: %s", profileName, strings.Join(allNames, ", "))
	}
	return idMatchingName, nil
}

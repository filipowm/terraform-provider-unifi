package v1

import (
	"context"
	"errors"
	"fmt"
	"github.com/filipowm/terraform-provider-unifi/internal/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
	"strings"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown

	schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
		desc := s.Description
		if s.Default != nil {
			desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
		}
		if s.Deprecated != "" {
			desc += " " + s.Deprecated
		}
		return strings.TrimSpace(desc)
	}
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"username": {
					Description: provider.ProviderUsernameDescription,
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("UNIFI_USERNAME", ""),
				},
				"password": {
					Description: provider.ProviderPasswordDescription,
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("UNIFI_PASSWORD", ""),
				},
				"api_key": {
					Description: provider.ProviderAPIKeyDescription,
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("UNIFI_API_KEY", ""),
				},
				"api_url": {
					Description: provider.ProviderAPIURLDescription,
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("UNIFI_API", ""),
				},
				"site": {
					Description: provider.ProviderSiteDescription,
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("UNIFI_SITE", "default"),
				},
				"allow_insecure": {
					Description: provider.ProviderAllowInsecureDescription,
					Type:        schema.TypeBool,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("UNIFI_INSECURE", false),
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"unifi_ap_group":       DataAPGroup(),
				"unifi_network":        DataNetwork(),
				"unifi_port_profile":   DataPortProfile(),
				"unifi_radius_profile": DataRADIUSProfile(),
				"unifi_user_group":     DataUserGroup(),
				"unifi_user":           DataUser(),
				"unifi_account":        DataAccount(),
			},
			ResourcesMap: map[string]*schema.Resource{
				// TODO: "unifi_ap_group"
				"unifi_device":         ResourceDevice(),
				"unifi_dynamic_dns":    ResourceDynamicDNS(),
				"unifi_firewall_group": ResourceFirewallGroup(),
				"unifi_firewall_rule":  ResourceFirewallRule(),
				"unifi_network":        ResourceNetwork(),
				"unifi_port_forward":   ResourcePortForward(),
				"unifi_static_route":   ResourceStaticRoute(),
				"unifi_wlan":           ResourceWLAN(),
				"unifi_port_profile":   ResourcePortProfile(),
				"unifi_site":           ResourceSite(),
				"unifi_account":        ResourceAccount(),
				"unifi_radius_profile": ResourceRadiusProfile(),

				"unifi_setting_mgmt":   ResourceSettingMgmt(),
				"unifi_setting_radius": ResourceSettingRadius(),
				"unifi_setting_usg":    ResourceSettingUsg(),
				"unifi_user_group":     ResourceUserGroup(),
				"unifi_user":           ResourceUser(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)
		return p
	}
}

func createHTTPTransport(insecure bool, subsystem string) http.RoundTripper {
	transport := provider.CreateHttpTransport(insecure)
	t := logging.NewSubsystemLoggingHTTPTransport(subsystem, transport)
	return t
}

func configure(v string, p *schema.Provider) schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		user := d.Get("username").(string)
		pass := d.Get("password").(string)
		apiKey := d.Get("api_key").(string)
		if apiKey != "" && (user != "" || pass != "") {
			return nil, diag.FromErr(errors.New("only one of `username`/`password` or `api_key` can be set"))
		} else if apiKey == "" && (user == "" || pass == "") {
			return nil, diag.FromErr(errors.New("either `username` and `password` or `api_key` must be set"))
		}
		baseURL := d.Get("api_url").(string)
		site := d.Get("site").(string)
		insecure := d.Get("allow_insecure").(bool)

		c, err := provider.NewClient(&provider.ClientConfig{
			Username: user,
			Password: pass,
			ApiKey:   apiKey,
			Url:      baseURL,
			Site:     site,
			HttpConfigurer: func() http.RoundTripper {
				return createHTTPTransport(insecure, "unifi")
			},
		})
		if err != nil {
			return nil, diag.FromErr(err)
		}
		return c, nil
	}
}

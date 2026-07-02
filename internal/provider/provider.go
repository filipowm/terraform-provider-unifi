package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/device"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/dns"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/firewall"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/network"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/radius"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/routing"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/settings"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/site"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/user"
)

const (
	ProviderUsernameDescription = "Local user name for the Unifi controller API. Can be specified with the `UNIFI_USERNAME` environment variable."
	ProviderPasswordDescription = "Password for the user accessing the API. Can be specified with the `UNIFI_PASSWORD` environment variable."
	ProviderAPIKeyDescription   = "API Key for the user accessing the API. Can be specified with the `UNIFI_API_KEY` environment variable. Controller version 9.0.108 or later is required." //nolint:gosec // G101 false positive: human-readable field description, not a credential
	ProviderAPIURLDescription   = "URL of the controller API. Can be specified with the `UNIFI_API` environment variable. " +
		"You should **NOT** supply the path (`/api`), the SDK will discover the appropriate paths. This is to support UDM Pro style API paths as well as more standard controller paths."
	ProviderSiteDescription          = "The site in the Unifi controller this provider will manage. Can be specified with the `UNIFI_SITE` environment variable. Default: `default`"
	ProviderAllowInsecureDescription = "Skip verification of TLS certificates of API requests. You may need to set this to `true` " +
		"if you are using your local API without setting up a signed certificate. Can be specified with the " +
		"`UNIFI_INSECURE` environment variable."
	ProviderMaxRetriesDescription = "Maximum number of additional attempts the provider makes when the controller returns a " +
		"transient response (network/connection errors, HTTP 5xx or 429 status codes, or an HTML body instead of JSON, " +
		"which can happen under parallel load). Only idempotent requests (`GET`, `HEAD`, `PUT`, `DELETE`, `OPTIONS`) are " +
		"retried. Defaults to `0`, which disables retries and preserves the default behavior. Can be specified with the " +
		"`UNIFI_MAX_RETRIES` environment variable."
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
					Description: ProviderUsernameDescription,
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("UNIFI_USERNAME", ""),
				},
				"password": {
					Description: ProviderPasswordDescription,
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("UNIFI_PASSWORD", ""),
				},
				"api_key": {
					Description: ProviderAPIKeyDescription,
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("UNIFI_API_KEY", ""),
				},
				"api_url": {
					Description: ProviderAPIURLDescription,
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("UNIFI_API", ""),
				},
				"site": {
					Description: ProviderSiteDescription,
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("UNIFI_SITE", "default"),
				},
				"allow_insecure": {
					Description: ProviderAllowInsecureDescription,
					Type:        schema.TypeBool,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("UNIFI_INSECURE", false),
				},
				"http_max_retries": {
					Description: ProviderMaxRetriesDescription,
					Type:        schema.TypeInt,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("UNIFI_MAX_RETRIES", 0),
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"unifi_network":        network.DataNetwork(),
				"unifi_port_profile":   device.DataPortProfile(),
				"unifi_radius_profile": radius.DataRADIUSProfile(),
				"unifi_user_group":     user.DataUserGroup(),
				"unifi_user":           user.DataUser(),
				"unifi_account":        radius.DataAccount(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"unifi_device":         device.ResourceDevice(),
				"unifi_dynamic_dns":    dns.ResourceDynamicDNS(),
				"unifi_firewall_group": firewall.ResourceFirewallGroup(),
				"unifi_firewall_rule":  firewall.ResourceFirewallRule(),
				"unifi_network":        network.ResourceNetwork(),
				"unifi_port_forward":   routing.ResourcePortForward(),
				"unifi_static_route":   routing.ResourceStaticRoute(),
				"unifi_wlan":           network.ResourceWLAN(),
				"unifi_port_profile":   device.ResourcePortProfile(),
				"unifi_site":           site.ResourceSite(),
				"unifi_account":        radius.ResourceAccount(),
				"unifi_radius_profile": radius.ResourceRadiusProfile(),
				"unifi_setting_radius": settings.ResourceSettingRadius(),
				"unifi_user_group":     user.ResourceUserGroup(),
				"unifi_user":           user.ResourceUser(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)
		return p
	}
}

func createHTTPTransport(insecure bool, subsystem string) http.RoundTripper {
	transport := base.CreateHTTPTransport(insecure)
	t := logging.NewSubsystemLoggingHTTPTransport(subsystem, transport)
	return t
}

func configure(v string, p *schema.Provider) schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		user, ok := d.Get("username").(string)
		if !ok {
			return nil, diag.FromErr(errors.New("`username` must be a string"))
		}
		pass, ok := d.Get("password").(string)
		if !ok {
			return nil, diag.FromErr(errors.New("`password` must be a string"))
		}
		apiKey, ok := d.Get("api_key").(string)
		if !ok {
			return nil, diag.FromErr(errors.New("`api_key` must be a string"))
		}
		if apiKey != "" && (user != "" || pass != "") {
			return nil, diag.FromErr(errors.New("only one of `username`/`password` or `api_key` can be set"))
		} else if apiKey == "" && (user == "" || pass == "") {
			return nil, diag.FromErr(errors.New("either `username` and `password` or `api_key` must be set"))
		}
		baseURL, ok := d.Get("api_url").(string)
		if !ok {
			return nil, diag.FromErr(errors.New("`api_url` must be a string"))
		}
		site, ok := d.Get("site").(string)
		if !ok {
			return nil, diag.FromErr(errors.New("`site` must be a string"))
		}
		insecure, ok := d.Get("allow_insecure").(bool)
		if !ok {
			return nil, diag.FromErr(errors.New("`allow_insecure` must be a boolean"))
		}
		maxRetries, ok := d.Get("http_max_retries").(int)
		if !ok {
			return nil, diag.FromErr(errors.New("`http_max_retries` must be an integer"))
		}

		c, err := base.NewClient(&base.ClientConfig{
			Username:   user,
			Password:   pass,
			APIKey:     apiKey,
			URL:        baseURL,
			Site:       site,
			Insecure:   insecure,
			MaxRetries: maxRetries,
			HTTPConfigurer: func() http.RoundTripper {
				return createHTTPTransport(insecure, "unifi")
			},
		})
		if err != nil {
			return nil, diag.FromErr(err)
		}
		return c, nil
	}
}

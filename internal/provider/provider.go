package provider

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
					Description: "Local user name for the Unifi controller API. Can be specified with the `UNIFI_USERNAME` " +
						"environment variable.",
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("UNIFI_USERNAME", ""),
				},
				"password": {
					Description: "Password for the user accessing the API. Can be specified with the `UNIFI_PASSWORD` " +
						"environment variable.",
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("UNIFI_PASSWORD", ""),
				},
				"api_url": {
					Description: "URL of the controller API. Can be specified with the `UNIFI_API` environment variable. " +
						"You should **NOT** supply the path (`/api`), the SDK will discover the appropriate paths. This is " +
						"to support UDM Pro style API paths as well as more standard controller paths.",

					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("UNIFI_API", ""),
				},
				"site": {
					Description: "The site in the Unifi controller this provider will manage. Can be specified with " +
						"the `UNIFI_SITE` environment variable. Default: `default`",
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("UNIFI_SITE", "default"),
				},
				"allow_insecure": {
					Description: "Skip verification of TLS certificates of API requests. You may need to set this to `true` " +
						"if you are using your local API without setting up a signed certificate. Can be specified with the " +
						"`UNIFI_INSECURE` environment variable.",
					Type:        schema.TypeBool,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("UNIFI_INSECURE", false),
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"unifi_ap_group":       dataAPGroup(),
				"unifi_network":        dataNetwork(),
				"unifi_port_profile":   dataPortProfile(),
				"unifi_radius_profile": dataRADIUSProfile(),
				"unifi_user_group":     dataUserGroup(),
				"unifi_user":           dataUser(),
				"unifi_account":        dataAccount(),
			},
			ResourcesMap: map[string]*schema.Resource{
				// TODO: "unifi_ap_group"
				"unifi_device":         resourceDevice(),
				"unifi_dynamic_dns":    resourceDynamicDNS(),
				"unifi_firewall_group": resourceFirewallGroup(),
				"unifi_firewall_rule":  resourceFirewallRule(),
				"unifi_network":        resourceNetwork(),
				"unifi_port_forward":   resourcePortForward(),
				"unifi_port_profile":   resourcePortProfile(),
				"unifi_radius_profile": resourceRadiusProfile(),
				"unifi_site":           resourceSite(),
				"unifi_static_route":   resourceStaticRoute(),
				"unifi_user_group":     resourceUserGroup(),
				"unifi_user":           resourceUser(),
				"unifi_wlan":           resourceWLAN(),
				"unifi_account":        resourceAccount(),

				"unifi_setting_mgmt":   resourceSettingMgmt(),
				"unifi_setting_radius": resourceSettingRadius(),
				"unifi_setting_usg":    resourceSettingUsg(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)
		return p
	}
}

func createHTTPTransport(insecure bool, subsystem string) http.RoundTripper {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,

		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}

	t := logging.NewSubsystemLoggingHTTPTransport(subsystem, transport)
	return t
}

func configure(version string, p *schema.Provider) schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		user := d.Get("username").(string)
		pass := d.Get("password").(string)
		baseURL := d.Get("api_url").(string)
		site := d.Get("site").(string)
		insecure := d.Get("allow_insecure").(bool)

		unifiClient, err := unifi.NewClient(&unifi.ClientConfig{
			URL:      baseURL,
			User:     user,
			Password: pass,
			HttpRoundTripperProvider: func() http.RoundTripper {
				return createHTTPTransport(insecure, "unifi")
			},
		})

		if err != nil {
			return nil, diag.FromErr(err)
		}
		err = checkMinimumControllerVersion(unifiClient.Version())
		log.Printf("[TRACE] Unifi controller version: %q", unifiClient.Version())
		if err != nil {
			return nil, diag.FromErr(err)
		}
		c := &client{
			c:    unifiClient,
			site: site,
		}

		return c, nil
	}
}

type client struct {
	c    unifi.Client
	site string
}

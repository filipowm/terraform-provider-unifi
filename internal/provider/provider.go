package provider

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkschema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &frameworkProvider{}
)

// frameworkProvider is the provider implementation for Plugin Framework.
type frameworkProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// frameworkProviderModel maps provider schema data to a Go type.
type frameworkProviderModel struct {
	Username      types.String `tfsdk:"username"`
	Password      types.String `tfsdk:"password"`
	APIUrl        types.String `tfsdk:"api_url"`
	Site          types.String `tfsdk:"site"`
	AllowInsecure types.Bool   `tfsdk:"allow_insecure"`
}

// New returns a new provider implementation with both SDK v2 and Plugin Framework support.
func New(version string) func() *sdkschema.Provider {
	return func() *sdkschema.Provider {
		p := &sdkschema.Provider{
			Schema: map[string]*sdkschema.Schema{
				"username": {
					Description: "Local user name for the Unifi controller API. Can be specified with the `UNIFI_USERNAME` " +
						"environment variable.",
					Type:        sdkschema.TypeString,
					Required:    true,
					DefaultFunc: sdkschema.EnvDefaultFunc("UNIFI_USERNAME", ""),
				},
				"password": {
					Description: "Password for the user accessing the API. Can be specified with the `UNIFI_PASSWORD` " +
						"environment variable.",
					Type:        sdkschema.TypeString,
					Required:    true,
					DefaultFunc: sdkschema.EnvDefaultFunc("UNIFI_PASSWORD", ""),
				},
				"api_url": {
					Description: "URL of the controller API. Can be specified with the `UNIFI_API` environment variable. " +
						"You should **NOT** supply the path (`/api`), the SDK will discover the appropriate paths. This is " +
						"to support UDM Pro style API paths as well as more standard controller paths.",
					Type:        sdkschema.TypeString,
					Required:    true,
					DefaultFunc: sdkschema.EnvDefaultFunc("UNIFI_API", ""),
				},
				"site": {
					Description: "The site in the Unifi controller this provider will manage. Can be specified with " +
						"the `UNIFI_SITE` environment variable. Default: `default`",
					Type:        sdkschema.TypeString,
					Required:    true,
					DefaultFunc: sdkschema.EnvDefaultFunc("UNIFI_SITE", "default"),
				},
				"allow_insecure": {
					Description: "Skip verification of TLS certificates of API requests. You may need to set this to `true` " +
						"if you are using your local API without setting up a signed certificate. Can be specified with the " +
						"`UNIFI_INSECURE` environment variable.",
					Type:        sdkschema.TypeBool,
					Optional:    true,
					DefaultFunc: sdkschema.EnvDefaultFunc("UNIFI_INSECURE", false),
				},
			},
			DataSourcesMap: map[string]*sdkschema.Resource{
				"unifi_ap_group":       dataAPGroup(),
				"unifi_network":        dataNetwork(),
				"unifi_port_profile":   dataPortProfile(),
				"unifi_radius_profile": dataRADIUSProfile(),
				"unifi_user_group":     dataUserGroup(),
				"unifi_user":           dataUser(),
				"unifi_account":        dataAccount(),
			},
			ResourcesMap: map[string]*sdkschema.Resource{
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

		p.ConfigureContextFunc = configure()
		return p
	}
}

// NewFrameworkProvider returns a new provider implementation using the Plugin Framework.
func NewFrameworkProvider(version string) func() provider.Provider {
	return func() provider.Provider {
		return &frameworkProvider{
			version: version,
		}
	}
}

// Metadata returns the provider type name.
func (p *frameworkProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "unifi"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *frameworkProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The UniFi provider provides resources to interact with a UniFi controller.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Description: "Local user name for the Unifi controller API. Can be specified with the `UNIFI_USERNAME` environment variable.",
				Required:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password for the user accessing the API. Can be specified with the `UNIFI_PASSWORD` environment variable.",
				Required:    true,
				Sensitive:   true,
			},
			"api_url": schema.StringAttribute{
				Description: "URL of the controller API. Can be specified with the `UNIFI_API` environment variable.",
				Required:    true,
			},
			"site": schema.StringAttribute{
				Description: "The site in the Unifi controller this provider will manage. Can be specified with the `UNIFI_SITE` environment variable.",
				Required:    true,
			},
			"allow_insecure": schema.BoolAttribute{
				Description: "Skip verification of TLS certificates of API requests. Can be specified with the `UNIFI_INSECURE` environment variable.",
				Optional:    true,
			},
		},
	}
}

// Configure prepares a UniFi API client for data sources and resources.
func (p *frameworkProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config frameworkProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Cannot use unknown value as username",
		)
		return
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Cannot use unknown value as password",
		)
		return
	}

	if config.APIUrl.IsUnknown() {
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Cannot use unknown value as api_url",
		)
		return
	}

	if config.Site.IsUnknown() {
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Cannot use unknown value as site",
		)
		return
	}

	c := &client{
		c: &lazyClient{
			user:     config.Username.ValueString(),
			pass:     config.Password.ValueString(),
			baseURL:  config.APIUrl.ValueString(),
			insecure: config.AllowInsecure.ValueBool(),
		},
		site: config.Site.ValueString(),
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

// DataSources defines the data sources implemented in the provider.
func (p *frameworkProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Add your framework data sources here
	}
}

// Resources defines the resources implemented in the provider.
func (p *frameworkProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Add your framework resources here
	}
}

func configure() schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		user := d.Get("username").(string)
		pass := d.Get("password").(string)
		baseURL := d.Get("api_url").(string)
		site := d.Get("site").(string)
		insecure := d.Get("allow_insecure").(bool)

		c := &client{
			c: &lazyClient{
				user:     user,
				pass:     pass,
				baseURL:  baseURL,
				insecure: insecure,
			},
			site: site,
		}

		return c, nil
	}
}

type unifiClient interface {
	Version() string

	ListUserGroup(ctx context.Context, site string) ([]unifi.UserGroup, error)
	DeleteUserGroup(ctx context.Context, site, id string) error
	CreateUserGroup(ctx context.Context, site string, d *unifi.UserGroup) (*unifi.UserGroup, error)
	GetUserGroup(ctx context.Context, site, id string) (*unifi.UserGroup, error)
	UpdateUserGroup(ctx context.Context, site string, d *unifi.UserGroup) (*unifi.UserGroup, error)

	ListFirewallGroup(ctx context.Context, site string) ([]unifi.FirewallGroup, error)
	DeleteFirewallGroup(ctx context.Context, site, id string) error
	CreateFirewallGroup(ctx context.Context, site string, d *unifi.FirewallGroup) (*unifi.FirewallGroup, error)
	GetFirewallGroup(ctx context.Context, site, id string) (*unifi.FirewallGroup, error)
	UpdateFirewallGroup(ctx context.Context, site string, d *unifi.FirewallGroup) (*unifi.FirewallGroup, error)

	ListFirewallRule(ctx context.Context, site string) ([]unifi.FirewallRule, error)
	DeleteFirewallRule(ctx context.Context, site, id string) error
	CreateFirewallRule(ctx context.Context, site string, d *unifi.FirewallRule) (*unifi.FirewallRule, error)
	GetFirewallRule(ctx context.Context, site, id string) (*unifi.FirewallRule, error)
	UpdateFirewallRule(ctx context.Context, site string, d *unifi.FirewallRule) (*unifi.FirewallRule, error)

	ListWLANGroup(ctx context.Context, site string) ([]unifi.WLANGroup, error)

	ListAPGroup(ctx context.Context, site string) ([]unifi.APGroup, error)

	DeleteNetwork(ctx context.Context, site, id, name string) error
	CreateNetwork(ctx context.Context, site string, d *unifi.Network) (*unifi.Network, error)
	GetNetwork(ctx context.Context, site, id string) (*unifi.Network, error)
	ListNetwork(ctx context.Context, site string) ([]unifi.Network, error)
	UpdateNetwork(ctx context.Context, site string, d *unifi.Network) (*unifi.Network, error)

	DeleteWLAN(ctx context.Context, site, id string) error
	CreateWLAN(ctx context.Context, site string, d *unifi.WLAN) (*unifi.WLAN, error)
	GetWLAN(ctx context.Context, site, id string) (*unifi.WLAN, error)
	UpdateWLAN(ctx context.Context, site string, d *unifi.WLAN) (*unifi.WLAN, error)

	GetDevice(ctx context.Context, site, id string) (*unifi.Device, error)
	GetDeviceByMAC(ctx context.Context, site, mac string) (*unifi.Device, error)
	CreateDevice(ctx context.Context, site string, d *unifi.Device) (*unifi.Device, error)
	UpdateDevice(ctx context.Context, site string, d *unifi.Device) (*unifi.Device, error)
	DeleteDevice(ctx context.Context, site, id string) error
	ListDevice(ctx context.Context, site string) ([]unifi.Device, error)
	AdoptDevice(ctx context.Context, site, mac string) error
	ForgetDevice(ctx context.Context, site, mac string) error

	GetUser(ctx context.Context, site, id string) (*unifi.User, error)
	GetUserByMAC(ctx context.Context, site, mac string) (*unifi.User, error)
	CreateUser(ctx context.Context, site string, d *unifi.User) (*unifi.User, error)
	BlockUserByMAC(ctx context.Context, site, mac string) error
	UnblockUserByMAC(ctx context.Context, site, mac string) error
	OverrideUserFingerprint(ctx context.Context, site, mac string, devIdOveride int) error
	UpdateUser(ctx context.Context, site string, d *unifi.User) (*unifi.User, error)
	DeleteUserByMAC(ctx context.Context, site, mac string) error

	GetPortForward(ctx context.Context, site, id string) (*unifi.PortForward, error)
	DeletePortForward(ctx context.Context, site, id string) error
	CreatePortForward(ctx context.Context, site string, d *unifi.PortForward) (*unifi.PortForward, error)
	UpdatePortForward(ctx context.Context, site string, d *unifi.PortForward) (*unifi.PortForward, error)

	ListRADIUSProfile(ctx context.Context, site string) ([]unifi.RADIUSProfile, error)
	GetRADIUSProfile(ctx context.Context, site, id string) (*unifi.RADIUSProfile, error)
	DeleteRADIUSProfile(ctx context.Context, site, id string) error
	CreateRADIUSProfile(ctx context.Context, site string, d *unifi.RADIUSProfile) (*unifi.RADIUSProfile, error)
	UpdateRADIUSProfile(ctx context.Context, site string, d *unifi.RADIUSProfile) (*unifi.RADIUSProfile, error)

	ListAccounts(ctx context.Context, site string) ([]unifi.Account, error)
	GetAccount(ctx context.Context, site, id string) (*unifi.Account, error)
	DeleteAccount(ctx context.Context, site, id string) error
	CreateAccount(ctx context.Context, site string, d *unifi.Account) (*unifi.Account, error)
	UpdateAccount(ctx context.Context, site string, d *unifi.Account) (*unifi.Account, error)

	GetSite(ctx context.Context, id string) (*unifi.Site, error)
	ListSites(ctx context.Context) ([]unifi.Site, error)
	CreateSite(ctx context.Context, Description string) ([]unifi.Site, error)
	UpdateSite(ctx context.Context, Name, Description string) ([]unifi.Site, error)
	DeleteSite(ctx context.Context, ID string) ([]unifi.Site, error)

	ListPortProfile(ctx context.Context, site string) ([]unifi.PortProfile, error)
	GetPortProfile(ctx context.Context, site, id string) (*unifi.PortProfile, error)
	DeletePortProfile(ctx context.Context, site, id string) error
	CreatePortProfile(ctx context.Context, site string, d *unifi.PortProfile) (*unifi.PortProfile, error)
	UpdatePortProfile(ctx context.Context, site string, d *unifi.PortProfile) (*unifi.PortProfile, error)

	ListRouting(ctx context.Context, site string) ([]unifi.Routing, error)
	GetRouting(ctx context.Context, site, id string) (*unifi.Routing, error)
	DeleteRouting(ctx context.Context, site, id string) error
	CreateRouting(ctx context.Context, site string, d *unifi.Routing) (*unifi.Routing, error)
	UpdateRouting(ctx context.Context, site string, d *unifi.Routing) (*unifi.Routing, error)

	ListDynamicDNS(ctx context.Context, site string) ([]unifi.DynamicDNS, error)
	GetDynamicDNS(ctx context.Context, site, id string) (*unifi.DynamicDNS, error)
	DeleteDynamicDNS(ctx context.Context, site, id string) error
	CreateDynamicDNS(ctx context.Context, site string, d *unifi.DynamicDNS) (*unifi.DynamicDNS, error)
	UpdateDynamicDNS(ctx context.Context, site string, d *unifi.DynamicDNS) (*unifi.DynamicDNS, error)

	GetSettingMgmt(ctx context.Context, id string) (*unifi.SettingMgmt, error)
	GetSettingUsg(ctx context.Context, id string) (*unifi.SettingUsg, error)
	UpdateSettingMgmt(ctx context.Context, site string, d *unifi.SettingMgmt) (*unifi.SettingMgmt, error)
	UpdateSettingUsg(ctx context.Context, site string, d *unifi.SettingUsg) (*unifi.SettingUsg, error)

	GetSettingRadius(ctx context.Context, id string) (*unifi.SettingRadius, error)
	UpdateSettingRadius(ctx context.Context, site string, d *unifi.SettingRadius) (*unifi.SettingRadius, error)
}

type client struct {
	c    unifiClient
	site string
}

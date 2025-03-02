package provider

import (
	"context"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/dns"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/settings"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
	"github.com/filipowm/terraform-provider-unifi/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewV2(version string) func() provider.Provider {
	return func() provider.Provider {
		return &unifiProvider{
			version: version,
		}
	}
}

var (
	_ provider.Provider = &unifiProvider{}
)

type unifiProvider struct {
	version string
}

type unifiProviderModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	APIKey   types.String `tfsdk:"api_key"`
	APIUrl   types.String `tfsdk:"api_url"`
	Site     types.String `tfsdk:"site"`
	Insecure types.Bool   `tfsdk:"allow_insecure"`
}

func (p *unifiProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "unifi"
	resp.Version = p.version
}

func (p *unifiProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				MarkdownDescription: ProviderUsernameDescription,
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: ProviderPasswordDescription,
				Optional:            true,
				Sensitive:           true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: ProviderAPIKeyDescription,
				Optional:            true,
				Sensitive:           true,
			},
			"api_url": schema.StringAttribute{
				MarkdownDescription: ProviderAPIURLDescription,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1), // workaround for `required: true`, because it fails on doc generation due to incorrectly detected difference between v1 and v2
					validators.HTTPSUrl(),
				},
				Optional: true,
			},
			"site": schema.StringAttribute{
				MarkdownDescription: ProviderSiteDescription,
				Optional:            true,
			},
			"allow_insecure": schema.BoolAttribute{
				MarkdownDescription: ProviderAllowInsecureDescription,
				Optional:            true,
			},
		},
	}
}

func (p *unifiProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Unifi provider...")
	// Retrieve provider data from the configuration
	var cfg unifiProviderModel
	diags := req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if cfg.APIUrl.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_url"),
			"Unknown UniFi Controller API URL",
			"The provider cannot create the UniFi Controller API client as there is an unknown configuration value "+
				"for the API endpoint. Either target apply the source of the value first, set the value statically in "+
				"the configuration, or use the UNIFI_API environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}
	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	// Check environment variables
	username := utils.GetAnyStringEnv("UNIFI_USERNAME")
	password := utils.GetAnyStringEnv("UNIFI_PASSWORD")
	apiKey := utils.GetAnyStringEnv("UNIFI_API_KEY")
	apiUrl := utils.GetAnyStringEnv("UNIFI_API")
	site := utils.GetAnyStringEnv("UNIFI_SITE")
	insecure := utils.GetAnyBoolEnv("UNIFI_INSECURE")

	if !cfg.Username.IsNull() {
		username = cfg.Username.ValueString()
	}
	if !cfg.Password.IsNull() {
		password = cfg.Password.ValueString()
	}
	if !cfg.APIKey.IsNull() {
		apiKey = cfg.APIKey.ValueString()
	}
	if !cfg.APIUrl.IsNull() {
		apiUrl = cfg.APIUrl.ValueString()
	}
	if !cfg.Site.IsNull() {
		site = cfg.Site.ValueString()
	}
	if !cfg.Insecure.IsNull() {
		insecure = cfg.Insecure.ValueBool()
	}
	if apiKey != "" && (username != "" || password != "") {
		resp.Diagnostics.AddAttributeError(path.Root("api_key"), "Two authentication methods configured", "Only one of `username`/`password` or `api_key` can be set")
	} else if apiKey == "" && (username == "" || password == "") {
		resp.Diagnostics.AddAttributeError(path.Root("api_key"), "Missing UniFi API credentials", "Either `username`/`password` or `api_key` must be set")
	}
	if apiUrl == "" {
		resp.Diagnostics.AddAttributeError(path.Root("api_url"), "Missing UniFi API URL", "The `api_url` attribute must be set")
	}
	if resp.Diagnostics.HasError() {
		return
	}
	if site == "" {
		site = "default" // set default site if not provided
	}
	c, err := base.NewClient(&base.ClientConfig{
		Username: username,
		Password: password,
		ApiKey:   apiKey,
		Url:      apiUrl,
		Site:     site,
		Insecure: insecure,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create UniFi client", err.Error())
		return
	}
	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *unifiProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		dns.NewDnsRecordResource,
		settings.NewAutoSpeedtestResource,
		settings.NewCountryResource,
		settings.NewLocaleResource,
		settings.NewMagicSiteToSiteVpnResource,
		settings.NewNetworkOptimizationResource,
		settings.NewNtpResource,
	}
}

func (p *unifiProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		dns.NewDnsRecordsDatasource,
		dns.NewDnsRecordDatasource,
	}
}

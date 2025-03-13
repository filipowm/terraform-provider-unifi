package settings

import (
	"context"
	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ntpModel represents the data model for NTP (Network Time Protocol) settings.
// It defines how NTP servers are configured for a UniFi site.
type ntpModel struct {
	base.Model
	NtpServer1 types.String `tfsdk:"ntp_server_1"`
	NtpServer2 types.String `tfsdk:"ntp_server_2"`
	NtpServer3 types.String `tfsdk:"ntp_server_3"`
	NtpServer4 types.String `tfsdk:"ntp_server_4"`
	Mode       types.String `tfsdk:"mode"`
}

func (d *ntpModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingNtp{
		ID:                d.ID.ValueString(),
		SettingPreference: d.Mode.ValueString(),
	}
	if d.Mode.ValueString() == "auto" {
		model.NtpServer1 = ""
		model.NtpServer2 = ""
		model.NtpServer3 = ""
		model.NtpServer4 = ""
	} else {
		if !base.IsEmptyString(d.NtpServer1) {
			model.NtpServer1 = d.NtpServer1.ValueString()
		}
		if !base.IsEmptyString(d.NtpServer2) {
			model.NtpServer2 = d.NtpServer2.ValueString()
		}
		if !base.IsEmptyString(d.NtpServer3) {
			model.NtpServer3 = d.NtpServer3.ValueString()
		}
		if !base.IsEmptyString(d.NtpServer4) {
			model.NtpServer4 = d.NtpServer4.ValueString()
		}
	}

	return model, diags
}

func (d *ntpModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingNtp)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingNtp")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.Mode = types.StringValue(model.SettingPreference)

	if model.NtpServer1 != "" {
		d.NtpServer1 = types.StringValue(model.NtpServer1)
	}
	if model.NtpServer2 != "" {
		d.NtpServer2 = types.StringValue(model.NtpServer2)
	}
	if model.NtpServer3 != "" {
		d.NtpServer3 = types.StringValue(model.NtpServer3)
	}
	if model.NtpServer4 != "" {
		d.NtpServer4 = types.StringValue(model.NtpServer4)
	}
	return diags
}

var (
	_ base.ResourceModel                    = &ntpModel{}
	_ resource.Resource                     = &ntpResource{}
	_ resource.ResourceWithConfigure        = &ntpResource{}
	_ resource.ResourceWithImportState      = &ntpResource{}
	_ resource.ResourceWithConfigValidators = &ntpResource{}
)

type ntpResource struct {
	*base.GenericResource[*ntpModel]
}

func (r *ntpResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		validators.RequiredNoneIf(path.MatchRoot("mode"), types.StringValue("auto"), path.MatchRoot("ntp_server_1"), path.MatchRoot("ntp_server_2"), path.MatchRoot("ntp_server_3"), path.MatchRoot("ntp_server_4")),
		validators.ResourceIf(path.MatchRoot("mode"),
			types.StringValue("manual"),
			resourcevalidator.AtLeastOneOf(path.MatchRoot("ntp_server_1"), path.MatchRoot("ntp_server_2"), path.MatchRoot("ntp_server_3"), path.MatchRoot("ntp_server_4")),
		),
	}
}

func (r *ntpResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	serverValidators := func() []validator.String {
		return []validator.String{
			stringvalidator.Any(validators.Hostname(), validators.IPv4()),
		}
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_ntp` resource allows you to configure Network Time Protocol (NTP) server settings for your UniFi network.\n\n" +
			"NTP servers provide time synchronization for your network devices. This resource supports both automatic and manual NTP configuration modes.",
		Attributes: map[string]schema.Attribute{
			"id":   base.ID(),
			"site": base.SiteAttribute(),
			"ntp_server_1": schema.StringAttribute{
				MarkdownDescription: "Primary NTP server hostname or IP address. Must be a valid hostname (e.g., `pool.ntp.org`) or IPv4 address. " +
					"Only applicable when `mode` is set to `manual`.",
				Optional:   true,
				Computed:   true,
				Validators: serverValidators(),
			},
			"ntp_server_2": schema.StringAttribute{
				MarkdownDescription: "Secondary NTP server hostname or IP address. Must be a valid hostname (e.g., `time.google.com`) or IPv4 address. " +
					"Only applicable when `mode` is set to `manual`.",
				Optional:   true,
				Computed:   true,
				Validators: serverValidators(),
			},
			"ntp_server_3": schema.StringAttribute{
				MarkdownDescription: "Tertiary NTP server hostname or IP address. Must be a valid hostname or IPv4 address. " +
					"Only applicable when `mode` is set to `manual`.",
				Optional:   true,
				Computed:   true,
				Validators: serverValidators(),
			},
			"ntp_server_4": schema.StringAttribute{
				MarkdownDescription: "Quaternary NTP server hostname or IP address. Must be a valid hostname or IPv4 address. " +
					"Only applicable when `mode` is set to `manual`.",
				Optional:   true,
				Computed:   true,
				Validators: serverValidators(),
			},
			"mode": schema.StringAttribute{
				MarkdownDescription: "NTP server configuration mode. Valid values are:\n" +
					"* `auto` - Use NTP servers configured on the controller\n" +
					"* `manual` - Use custom NTP servers specified in this resource\n\n" +
					"When set to `auto`, all NTP server fields will be cleared. " +
					"When set to `manual`, at least one NTP server must be specified.",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("auto", "manual"),
				},
			},
		},
	}
}

// NewNtpResource creates a new instance of the NTP resource.
func NewNtpResource() resource.Resource {
	r := &ntpResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_ntp",
		func() *ntpModel { return &ntpModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingNtp(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			return client.UpdateSettingNtp(ctx, site, body.(*unifi.SettingNtp))
		},
	)
	return r
}

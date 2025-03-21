package settings

import (
	"context"
	ut "github.com/filipowm/terraform-provider-unifi/internal/provider/types"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type rsyslogdModel struct {
	base.Model
	Enabled                     types.Bool   `tfsdk:"enabled"`
	Contents                    types.List   `tfsdk:"contents"`
	Debug                       types.Bool   `tfsdk:"debug"`
	IP                          types.String `tfsdk:"ip"`
	LogAllContents              types.Bool   `tfsdk:"log_all_contents"`
	NetconsoleEnabled           types.Bool   `tfsdk:"netconsole_enabled"`
	NetconsoleHost              types.String `tfsdk:"netconsole_host"`
	NetconsolePort              types.Int64  `tfsdk:"netconsole_port"`
	Port                        types.Int64  `tfsdk:"port"`
	ThisController              types.Bool   `tfsdk:"this_controller"`
	ThisControllerEncryptedOnly types.Bool   `tfsdk:"this_controller_encrypted_only"`
}

func (d *rsyslogdModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingRsyslogd{
		ID:       d.ID.ValueString(),
		Enabled:  d.Enabled.ValueBool(),
		Contents: []string{},
	}

	// Only set optional fields if rsyslogd is enabled
	if d.Enabled.ValueBool() {
		if !d.Debug.IsNull() {
			model.Debug = d.Debug.ValueBool()
		}

		if !d.IP.IsNull() {
			model.IP = d.IP.ValueString()
		}

		if !d.LogAllContents.IsNull() {
			model.LogAllContents = d.LogAllContents.ValueBool()
		}

		if !d.NetconsoleEnabled.IsNull() {
			model.NetconsoleEnabled = d.NetconsoleEnabled.ValueBool()
		}

		if !d.NetconsoleHost.IsNull() {
			model.NetconsoleHost = d.NetconsoleHost.ValueString()
		}

		if !d.NetconsolePort.IsNull() {
			model.NetconsolePort = int(d.NetconsolePort.ValueInt64())
		}

		if !d.Port.IsNull() {
			model.Port = int(d.Port.ValueInt64())
		}

		if !d.ThisController.IsNull() {
			model.ThisController = d.ThisController.ValueBool()
		}

		if !d.ThisControllerEncryptedOnly.IsNull() {
			model.ThisControllerEncryptedOnly = d.ThisControllerEncryptedOnly.ValueBool()
		}

		if !d.Contents.IsNull() {
			var contents []string
			diags.Append(ut.ListElementsAs(d.Contents, &contents)...)
			if diags.HasError() {
				return nil, diags
			}
			model.Contents = contents
		}
	}

	return model, diags
}

func (d *rsyslogdModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingRsyslogd)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingRsyslogd")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.Enabled = types.BoolValue(model.Enabled)

	// Only set optional fields if rsyslogd is enabled
	if model.Enabled {
		d.Debug = types.BoolValue(model.Debug)
		d.IP = types.StringValue(model.IP)
		d.LogAllContents = types.BoolValue(model.LogAllContents)
		d.NetconsoleEnabled = types.BoolValue(model.NetconsoleEnabled)
		d.NetconsoleHost = types.StringValue(model.NetconsoleHost)
		d.NetconsolePort = types.Int64Value(int64(model.NetconsolePort))
		d.Port = types.Int64Value(int64(model.Port))
		d.ThisController = types.BoolValue(model.ThisController)
		d.ThisControllerEncryptedOnly = types.BoolValue(model.ThisControllerEncryptedOnly)

		// Set the DHCP relay servers list
		contents, diags := types.ListValueFrom(ctx, types.StringType, model.Contents)
		if diags.HasError() {
			return diags
		}
		d.Contents = contents
	} else {
		d.Debug = types.BoolNull()
		d.IP = types.StringNull()
		d.LogAllContents = types.BoolNull()
		d.NetconsoleEnabled = types.BoolNull()
		d.NetconsoleHost = types.StringNull()
		d.NetconsolePort = types.Int64Null()
		d.Port = types.Int64Null()
		d.ThisController = types.BoolNull()
		d.ThisControllerEncryptedOnly = types.BoolNull()
		d.Contents = ut.EmptyList(types.StringType)
	}

	return diags
}

var (
	_ base.ResourceModel                    = &rsyslogdModel{}
	_ resource.Resource                     = &rsyslogdResource{}
	_ resource.ResourceWithConfigure        = &rsyslogdResource{}
	_ resource.ResourceWithImportState      = &rsyslogdResource{}
	_ resource.ResourceWithConfigValidators = &rsyslogdResource{}
	_ resource.ResourceWithModifyPlan       = &rsyslogdResource{}
)

type rsyslogdResource struct {
	*base.GenericResource[*rsyslogdModel]
}

func (r *rsyslogdResource) ModifyPlan(_ context.Context, _ resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	resp.Diagnostics.Append(r.RequireMinVersion("8.5")...)
}

func (r *rsyslogdResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		validators.RequiredNoneIf(path.MatchRoot("enabled"), types.BoolValue(false),
			path.MatchRoot("contents"),
			path.MatchRoot("debug"),
			path.MatchRoot("ip"),
			path.MatchRoot("log_all_contents"),
			path.MatchRoot("netconsole_enabled"),
			path.MatchRoot("netconsole_host"),
			path.MatchRoot("netconsole_port"),
			path.MatchRoot("port"),
			path.MatchRoot("this_controller"),
			path.MatchRoot("this_controller_encrypted_only"),
		),
		validators.RequiredTogetherIf(path.MatchRoot("enabled"), types.BoolValue(true), path.MatchRoot("contents"), path.MatchRoot("ip")),
	}
}

func (r *rsyslogdResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Remote Syslog (rsyslogd) settings for UniFi devices. Controller version 8.5 or later is required.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether remote syslog is enabled.",
				Required:            true,
			},
			"contents": schema.ListAttribute{
				MarkdownDescription: "List of log types to include in the remote syslog. Valid values: device, client, firewall_default_policy, triggers, updates, admin_activity, critical, security_detections, vpn.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(stringvalidator.OneOf("device", "client", "firewall_default_policy", "triggers", "updates", "admin_activity", "critical", "security_detections", "vpn")),
				},
			},
			"debug": schema.BoolAttribute{
				MarkdownDescription: "Whether debug logging is enabled.",
				Optional:            true,
				Computed:            true,
			},
			"ip": schema.StringAttribute{
				MarkdownDescription: "IP address of the remote syslog server.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validators.IPv4(),
				},
			},
			"log_all_contents": schema.BoolAttribute{
				MarkdownDescription: "Whether to log all content types.",
				Optional:            true,
				Computed:            true,
			},
			"netconsole_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether netconsole logging is enabled.",
				Optional:            true,
				Computed:            true,
			},
			"netconsole_host": schema.StringAttribute{
				MarkdownDescription: "Hostname or IP address of the netconsole server.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.Any(
						validators.Hostname(),
						validators.IPv4(),
					),
				},
			},
			"netconsole_port": schema.Int64Attribute{
				MarkdownDescription: "Port number for the netconsole server. Valid values: 1-65535.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Port number for the remote syslog server. Valid values: 1-65535.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"this_controller": schema.BoolAttribute{
				MarkdownDescription: "Whether to use this controller as the syslog server.",
				Optional:            true,
				Computed:            true,
			},
			"this_controller_encrypted_only": schema.BoolAttribute{
				MarkdownDescription: "Whether to only use encrypted connections to this controller for syslog.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func NewRsyslogdResource() resource.Resource {
	r := &rsyslogdResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_rsyslogd",
		func() *rsyslogdModel { return &rsyslogdModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingRsyslogd(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			return client.UpdateSettingRsyslogd(ctx, site, body.(*unifi.SettingRsyslogd))
		},
	)
	return r
}

package settings

import (
	"context"
	ut "github.com/filipowm/terraform-provider-unifi/internal/provider/types"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"

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

type lcmModel struct {
	base.Model
	Enabled     types.Bool  `tfsdk:"enabled"`
	Brightness  types.Int64 `tfsdk:"brightness"`
	IdleTimeout types.Int64 `tfsdk:"idle_timeout"`
	Sync        types.Bool  `tfsdk:"sync"`
	TouchEvent  types.Bool  `tfsdk:"touch_event"`
}

func (d *lcmModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingLcm{
		ID:      d.ID.ValueString(),
		Enabled: d.Enabled.ValueBool(),
	}

	// Only set optional fields if LCM is enabled
	if d.Enabled.ValueBool() {
		if !d.Brightness.IsNull() {
			model.Brightness = int(d.Brightness.ValueInt64())
		}
		if !d.IdleTimeout.IsNull() {
			model.IDleTimeout = int(d.IdleTimeout.ValueInt64())
		}
		if !d.Sync.IsNull() {
			model.Sync = d.Sync.ValueBool()
		}
		if !d.TouchEvent.IsNull() {
			model.TouchEvent = d.TouchEvent.ValueBool()
		}
	}

	return model, diags
}

func (d *lcmModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingLcm)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingLcm")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.Enabled = types.BoolValue(model.Enabled)

	// Only set optional fields if LCM is enabled
	if model.Enabled {
		d.Brightness = types.Int64Value(int64(model.Brightness))
		d.IdleTimeout = types.Int64Value(int64(model.IDleTimeout))
		d.Sync = types.BoolValue(model.Sync)
		d.TouchEvent = types.BoolValue(model.TouchEvent)
	} else {
		d.Brightness = types.Int64Null()
		d.IdleTimeout = types.Int64Null()
		d.Sync = types.BoolNull()
		d.TouchEvent = types.BoolNull()
	}

	return diags
}

var (
	_ base.ResourceModel                    = &lcmModel{}
	_ resource.Resource                     = &lcmResource{}
	_ resource.ResourceWithConfigure        = &lcmResource{}
	_ resource.ResourceWithImportState      = &lcmResource{}
	_ resource.ResourceWithConfigValidators = &lcmResource{}
)

type lcmResource struct {
	*base.GenericResource[*lcmModel]
}

func (r *lcmResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		validators.RequiredNoneIf(path.MatchRoot("enabled"), types.BoolValue(false), path.MatchRoot("brightness"), path.MatchRoot("idle_timeout"), path.MatchRoot("sync"), path.MatchRoot("touch_event")),
	}
}

func (r *lcmResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages LCD Monitor (LCM) settings for UniFi devices with built-in displays, such as the UniFi Dream Machine Pro (UDM Pro) and UniFi Network Video Recorder (UNVR).",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the LCD display is enabled.",
				Required:            true,
			},
			"brightness": schema.Int64Attribute{
				MarkdownDescription: "The brightness level of the LCD display. Valid values are 1-100.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.Between(1, 100),
				},
			},
			"idle_timeout": schema.Int64Attribute{
				MarkdownDescription: "The time in seconds after which the display turns off when idle. Valid values are 10-3600.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.Between(10, 3600),
				},
			},
			"sync": schema.BoolAttribute{
				MarkdownDescription: "Whether to synchronize display settings across multiple devices.",
				Optional:            true,
			},
			"touch_event": schema.BoolAttribute{
				MarkdownDescription: "Whether touch interactions with the display are enabled.",
				Optional:            true,
			},
		},
	}
}

func NewLcmResource() resource.Resource {
	r := &lcmResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_lcd_monitor",
		func() *lcmModel { return &lcmModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingLcm(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			return client.UpdateSettingLcm(ctx, site, body.(*unifi.SettingLcm))
		},
	)
	return r
}

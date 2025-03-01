package settings

import (
	"context"
	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &autoSpeedtestResource{}
	_ resource.ResourceWithConfigure   = &autoSpeedtestResource{}
	_ resource.ResourceWithImportState = &autoSpeedtestResource{}
	_ base.Resource                    = &autoSpeedtestResource{}
)

type autoSpeedtestModel struct {
	base.Model
	ID             types.String `tfsdk:"id"`
	CronExpression types.String `tfsdk:"cron"`
	Enabled        types.Bool   `tfsdk:"enabled"`
}

func (d *autoSpeedtestModel) AsUnifiModel() (interface{}, diag.Diagnostics) {
	return &unifi.SettingAutoSpeedtest{
		ID:       d.ID.ValueString(),
		CronExpr: d.CronExpression.ValueString(),
		Enabled:  d.Enabled.ValueBool(),
	}, diag.Diagnostics{}
}

func (d *autoSpeedtestModel) Merge(other interface{}) diag.Diagnostics {
	if typed, ok := other.(*unifi.SettingAutoSpeedtest); ok {
		d.ID = types.StringValue(typed.ID)
		d.CronExpression = types.StringValue(typed.CronExpr)
		d.Enabled = types.BoolValue(typed.Enabled)
	}
	return diag.Diagnostics{}
}

type autoSpeedtestResource struct {
	*BaseSettingResource[*autoSpeedtestModel]
}

func checkAutoSpeedtestUnsupportedError(err error, diag *diag.Diagnostics) bool {
	if base.IsServerErrorContains(err, "api.err.SpeedTestNotSupported") {
		if diag != nil {
			diag.AddError("Auto Speedtest is not supported", "Auto Speedtest is not supported on this controller")
		}
		return true
	}
	return false
}

func NewAutoSpeedtestResource() resource.Resource {
	r := &autoSpeedtestResource{}
	r.BaseSettingResource = NewBaseSettingResource(
		"unifi_setting_auto_speedtest",
		func() *autoSpeedtestModel { return &autoSpeedtestModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			res, err := client.GetSettingAutoSpeedtest(ctx, site)
			if err != nil && !checkAutoSpeedtestUnsupportedError(err, nil) {
				return nil, err
			}
			return res, nil
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			res, err := client.UpdateSettingAutoSpeedtest(ctx, site, body.(*unifi.SettingAutoSpeedtest))
			if err != nil && !checkAutoSpeedtestUnsupportedError(err, nil) {
				return nil, err
			}
			return res, nil
		},
	)
	return r
}

func (a *autoSpeedtestResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_auto_speedtest` resource manages the automatic speedtest settings in the UniFi controller." +
			"Automatic speedtests can be scheduled to run at regular intervals to monitor the network performance.\n\n" +
			"**NOTE:** Automatic speedtests where not verified and tested on all UniFi controller versions due to limitations of controller used in acceptance testing. ",
		Attributes: map[string]schema.Attribute{
			"id":   base.ID(),
			"site": base.SiteAttribute(),
			"cron": schema.StringAttribute{
				MarkdownDescription: "Cron expression defining the schedule for automatic speedtests.",
				Optional:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the automatic speedtest is enabled.",
				Required:            true,
			},
		},
	}
}

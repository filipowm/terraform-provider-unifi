package settings

import (
	"context"
	"errors"
	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type autoSpeedtestModel struct {
	base.Site
	ID             types.String `tfsdk:"id"`
	CronExpression types.String `tfsdk:"cron"`
	Enabled        types.Bool   `tfsdk:"enabled"`
}

func (d *autoSpeedtestModel) asUnifiModel() *unifi.SettingAutoSpeedtest {
	return &unifi.SettingAutoSpeedtest{
		ID:       d.ID.ValueString(),
		CronExpr: d.CronExpression.ValueString(),
		Enabled:  d.Enabled.ValueBool(),
	}
}

func (d *autoSpeedtestModel) merge(other *unifi.SettingAutoSpeedtest) {
	d.ID = types.StringValue(other.ID)
	d.CronExpression = types.StringValue(other.CronExpr)
	d.Enabled = types.BoolValue(other.Enabled)
}

var (
	_ resource.Resource                = &autoSpeedtestResource{}
	_ resource.ResourceWithConfigure   = &autoSpeedtestResource{}
	_ resource.ResourceWithImportState = &autoSpeedtestResource{}
	_ base.BaseData                    = &autoSpeedtestResource{}
)

type autoSpeedtestResource struct {
	client *base.Client
}

func NewAutoSpeedtestResource() resource.Resource {
	return &autoSpeedtestResource{}
}

func (a *autoSpeedtestResource) SetClient(client *base.Client) {
	a.client = client
}

func checkAutoSpeedtestUnsupportedError(err error, diag *diag.Diagnostics) {
	if base.IsServerErrorContains(err, "api.err.SpeedTestNotSupported") {
		diag.AddError("Auto Speedtest is not supported", "Auto Speedtest is not supported on this controller")

	}
}

func (a *autoSpeedtestResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, site := base.ImportIDWithSite(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	state := autoSpeedtestModel{
		ID:   types.StringValue(id),
		Site: base.NewSite(site),
	}
	a.read(ctx, site, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (a *autoSpeedtestResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	base.ConfigureResource(a, req, resp)
}

func (a *autoSpeedtestResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "unifi_setting_auto_speedtest"
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

func (a *autoSpeedtestResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan autoSpeedtestModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := plan.asUnifiModel()
	site := a.client.ResolveSite(&plan.Site)

	res, err := a.client.UpdateSettingAutoSpeedtest(ctx, site, body)
	if err != nil {
		checkAutoSpeedtestUnsupportedError(err, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.AddError("Error creating auto speedtest settings", err.Error())
		return
	}
	plan.merge(res)
	plan.Site.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (a *autoSpeedtestResource) read(ctx context.Context, site string, state *autoSpeedtestModel, diag *diag.Diagnostics) {
	res, err := a.client.GetSettingAutoSpeedtest(ctx, site)
	if err != nil {
		checkAutoSpeedtestUnsupportedError(err, diag)
		if diag.HasError() {
			return
		}
		if errors.Is(err, unifi.ErrNotFound) {
			diag.AddError("Auto speedtest settings not found", "The auto speedtest settings were not found in the UniFi controller")
		} else {
			diag.AddError("Error reading auto speedtest settings", err.Error())
		}
		return
	}
	state.merge(res)
}

func (a *autoSpeedtestResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state autoSpeedtestModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := a.client.ResolveSite(&state.Site)
	a.read(ctx, site, &state, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}
	(&state).Site.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (a *autoSpeedtestResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state autoSpeedtestModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := plan.asUnifiModel()
	site := a.client.ResolveSite(&plan.Site)

	res, err := a.client.UpdateSettingAutoSpeedtest(ctx, site, body)
	if err != nil {
		checkAutoSpeedtestUnsupportedError(err, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.AddError("Error updating auto speedtest settings", err.Error())
		return
	}
	state.merge(res)
	state.Site.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (a *autoSpeedtestResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Not supported
}

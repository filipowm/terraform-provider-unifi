package settings

import (
	"context"
	"errors"
	"github.com/biter777/countries"
	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type countryModel struct {
	base.Site
	ID          types.String `tfsdk:"id"`
	Code        types.String `tfsdk:"code"`
	CodeNumeric types.Int32  `tfsdk:"code_numeric"`
}

func (d *countryModel) asUnifiModel() *unifi.SettingCountry {
	code := countries.ByName(d.Code.ValueString())
	return &unifi.SettingCountry{
		ID:   d.ID.ValueString(),
		Code: int(code),
	}
}

func (d *countryModel) merge(other *unifi.SettingCountry) {
	d.ID = types.StringValue(other.ID)
	// UniFi uses numeric codes, so we need to convert the alpha-2 code to the numeric code, but we store both
	code := countries.ByNumeric(other.Code)
	d.Code = types.StringValue(code.Alpha2())
	d.CodeNumeric = types.Int32Value(int32(code))
}

var (
	_ resource.Resource                = &countryResource{}
	_ resource.ResourceWithConfigure   = &countryResource{}
	_ resource.ResourceWithImportState = &countryResource{}
	_ base.BaseData                    = &countryResource{}
)

type countryResource struct {
	client *base.Client
}

func (c *countryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, site := base.ImportIDWithSite(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	state := countryModel{
		ID:   types.StringValue(id),
		Site: base.NewSite(site),
	}
	c.read(ctx, site, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func NewCountryResource() resource.Resource {
	return &countryResource{}
}

func (c *countryResource) SetClient(client *base.Client) {
	c.client = client
}

func (c *countryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	base.ConfigureResource(c, req, resp)
}

func (c *countryResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "unifi_setting_country"
}

func (c *countryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_country` resource allows you to configure the country settings for your UniFi network. ",
		Attributes: map[string]schema.Attribute{
			"id":   base.ID(),
			"site": base.SiteAttribute(),
			"code": schema.StringAttribute{
				Description: "The country code to set for the UniFi site. The country code must be a valid ISO 3166-1 alpha-2 code.",
				Required:    true,
				Validators: []validator.String{
					validators.StringLengthExactly(2),
					validators.CountryCodeAlpha2(),
				},
			},
			"code_numeric": schema.Int32Attribute{
				Description: "The numeric representation in ISO 3166-1 of the country code.",
				Computed:    true,
			},
		},
	}
}

func (c *countryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan countryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := plan.asUnifiModel()
	site := c.client.ResolveSite(&plan.Site)

	res, err := c.client.UpdateSettingCountry(ctx, site, body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating country settings", err.Error())
		return
	}
	plan.merge(res)
	plan.Site.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (c *countryResource) read(ctx context.Context, site string, state *countryModel, diag *diag.Diagnostics) {
	res, err := c.client.GetSettingCountry(ctx, site)

	if err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			diag.AddError("Country settings not found", "The country settings were not found in the UniFi controller")
		} else {
			diag.AddError("Error reading country settings", err.Error())
		}
		return
	}
	state.merge(res)
}

func (c *countryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state countryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := c.client.ResolveSite(&state.Site)
	c.read(ctx, site, &state, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}
	(&state).Site.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (c *countryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state countryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := plan.asUnifiModel()
	site := c.client.ResolveSite(&plan.Site)

	res, err := c.client.UpdateSettingCountry(ctx, site, body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating country settings", err.Error())
		return
	}
	state.merge(res)
	state.Site.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (c *countryResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Not supported
}

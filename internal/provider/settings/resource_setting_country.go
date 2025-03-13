package settings

import (
	"context"
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

var (
	_ resource.Resource                = &countryResource{}
	_ resource.ResourceWithConfigure   = &countryResource{}
	_ resource.ResourceWithImportState = &countryResource{}
	_ base.Resource                    = &countryResource{}
)

type countryModel struct {
	base.Model
	Code        types.String `tfsdk:"code"`
	CodeNumeric types.Int32  `tfsdk:"code_numeric"`
}

func (d *countryModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	code := countries.ByName(d.Code.ValueString())
	return &unifi.SettingCountry{
		ID:   d.ID.ValueString(),
		Code: int(code),
	}, diag.Diagnostics{}
}

func (d *countryModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	if typed, ok := other.(*unifi.SettingCountry); ok {
		d.ID = types.StringValue(typed.ID)
		code := countries.ByNumeric(typed.Code)
		d.Code = types.StringValue(code.Alpha2())
		d.CodeNumeric = types.Int32Value(int32(code))
	}
	return diag.Diagnostics{}
}

type countryResource struct {
	*base.GenericResource[*countryModel]
}

func NewCountryResource() resource.Resource {
	r := &countryResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_country",
		func() *countryModel { return &countryModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingCountry(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			return client.UpdateSettingCountry(ctx, site, body.(*unifi.SettingCountry))
		},
	)
	return r
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

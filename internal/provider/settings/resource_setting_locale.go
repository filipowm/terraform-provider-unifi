package settings

import (
	"context"
	ut "github.com/filipowm/terraform-provider-unifi/internal/provider/types"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type localeModel struct {
	base.Model
	Timezone types.String `tfsdk:"timezone"`
}

func (d *localeModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingLocale{
		ID:       d.ID.ValueString(),
		Timezone: d.Timezone.ValueString(),
	}

	return model, diags
}

func (d *localeModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingLocale)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingLocale")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.Timezone = types.StringValue(model.Timezone)

	return diags
}

var (
	_ base.ResourceModel               = &localeModel{}
	_ resource.Resource                = &localeResource{}
	_ resource.ResourceWithConfigure   = &localeResource{}
	_ resource.ResourceWithImportState = &localeResource{}
)

type localeResource struct {
	*base.GenericResource[*localeModel]
}

func (r *localeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages locale settings for a UniFi site.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"timezone": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Timezone for the UniFi controller, e.g., `America/Los_Angeles`",
				Validators: []validator.String{
					validators.Timezone(),
				},
			},
		},
	}
}

func NewLocaleResource() resource.Resource {
	r := &localeResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_locale",
		func() *localeModel { return &localeModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingLocale(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			return client.UpdateSettingLocale(ctx, site, body.(*unifi.SettingLocale))
		},
	)
	return r
}

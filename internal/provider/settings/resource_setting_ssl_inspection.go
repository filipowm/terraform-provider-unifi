package settings

import (
	"context"
	ut "github.com/filipowm/terraform-provider-unifi/internal/provider/types"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type sslInspectionModel struct {
	base.Model
	State types.String `tfsdk:"state"`
}

func (d *sslInspectionModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingSslInspection{
		ID:    d.ID.ValueString(),
		State: d.State.ValueString(),
	}

	return model, diags
}

func (d *sslInspectionModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingSslInspection)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingSslInspection")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.State = types.StringValue(model.State)

	return diags
}

var (
	_ base.ResourceModel               = &sslInspectionModel{}
	_ resource.Resource                = &sslInspectionResource{}
	_ resource.ResourceWithConfigure   = &sslInspectionResource{}
	_ resource.ResourceWithImportState = &sslInspectionResource{}
)

type sslInspectionResource struct {
	*base.GenericResource[*sslInspectionModel]
}

func (r *sslInspectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages SSL Inspection settings for a UniFi site. SSL inspection is a security feature that allows the UniFi Security Gateway (USG) to inspect encrypted traffic for security threats.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"state": schema.StringAttribute{
				MarkdownDescription: "The mode of SSL inspection. Valid values are: `off`, `simple`, or `advanced`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("off", "simple", "advanced"),
				},
			},
		},
	}
}

func NewSslInspectionResource() resource.Resource {
	r := &sslInspectionResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_ssl_inspection",
		func() *sslInspectionModel { return &sslInspectionModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingSslInspection(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			return client.UpdateSettingSslInspection(ctx, site, body.(*unifi.SettingSslInspection))
		},
	)
	return r
}

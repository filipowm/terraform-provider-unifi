package apgroup

import (
	"context"
	"fmt"
	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	ut "github.com/filipowm/terraform-provider-unifi/internal/provider/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// APGroupDatasourceModel represents the data model for a UniFi AP Group data source
type APGroupDatasourceModel struct {
	base.Model
	Name types.String `tfsdk:"name"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model
func (m *APGroupDatasourceModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	return nil, diag.Diagnostics{}
}

// Merge updates the Terraform model with values from the UniFi API model
func (m *APGroupDatasourceModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	model, ok := other.(*unifi.APGroup)
	if !ok {
		diags.AddError("Invalid model type", "Expected *unifi.APGroup")
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.Name = types.StringValue(model.Name)

	return diags
}

var (
	_ datasource.DataSource              = &apGroupDatasource{}
	_ datasource.DataSourceWithConfigure = &apGroupDatasource{}
	_ base.Resource                      = &apGroupDatasource{}
)

type apGroupDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

func (d *apGroupDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func NewAPGroupDatasource() datasource.DataSource {
	return &apGroupDatasource{}
}

func (d *apGroupDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *apGroupDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *apGroupDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *apGroupDatasource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "unifi_ap_group"
}

func (d *apGroupDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_ap_group` data source can be used to retrieve the ID for an AP group by name.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the AP group to look up, leave blank to look up the default AP group.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		},
	}
}

func (d *apGroupDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state APGroupDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := d.client.ResolveSite(&state)

	groups, err := d.client.ListAPGroup(ctx, site)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list AP groups", err.Error())
		return
	}

	if len(groups) == 0 {
		resp.Diagnostics.AddError("AP group not found", "No AP groups found")
		return
	}

	name := state.Name.ValueString()
	var found *unifi.APGroup
	for _, g := range groups {
		if (name == "" && g.HiddenID == "default") || g.Name == name {
			found = &g
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError("AP group not found", fmt.Sprintf("No AP group with name %q found", name))
		return
	}

	(&state).Merge(ctx, found)
	state.SetID(found.ID)
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

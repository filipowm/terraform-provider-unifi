package apgroup

import (
	"context"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	ut "github.com/filipowm/terraform-provider-unifi/internal/provider/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// APGroupModel represents the data model for a UniFi AP Group
type APGroupModel struct {
	base.Model
	Name       types.String `tfsdk:"name"`
	DeviceMACs types.List   `tfsdk:"device_macs"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model
func (m *APGroupModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	var deviceMACs []string

	diags.Append(ut.ListElementsAs(m.DeviceMACs, &deviceMACs)...)
	if diags.HasError() {
		return nil, diags
	}

	return &unifi.APGroup{
		ID:         m.ID.ValueString(),
		Name:       m.Name.ValueString(),
		DeviceMACs: deviceMACs,
	}, diags
}

// Merge updates the Terraform model with values from the UniFi API model
func (m *APGroupModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	model, ok := other.(*unifi.APGroup)
	if !ok {
		diags.AddError("Invalid model type", "Expected *unifi.APGroup")
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.Name = types.StringValue(model.Name)

	deviceMACs, d := types.ListValueFrom(ctx, types.StringType, model.DeviceMACs)
	diags = append(diags, d...)
	m.DeviceMACs = deviceMACs

	return diags
}

var (
	_ resource.Resource                = &apGroupResource{}
	_ resource.ResourceWithConfigure   = &apGroupResource{}
	_ resource.ResourceWithImportState = &apGroupResource{}
	_ base.Resource                    = &apGroupResource{}
)

type apGroupResource struct {
	*base.GenericResource[*APGroupModel]
}

// NewAPGroupResource creates a new instance of the AP group resource
func NewAPGroupResource() resource.Resource {
	return &apGroupResource{
		GenericResource: base.NewGenericResource(
			"unifi_ap_group",
			func() *APGroupModel { return &APGroupModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetAPGroup(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					return client.CreateAPGroup(ctx, site, model.(*unifi.APGroup))
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					return client.UpdateAPGroup(ctx, site, model.(*unifi.APGroup))
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteAPGroup(ctx, site, id)
				},
			},
		),
	}
}

// Schema defines the schema for the resource
func (r *apGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_ap_group` resource manages Access Point groups in the UniFi controller.\n\n" +
				"AP groups allow you to organize and manage multiple access points together. " +
				"This resource allows you to create, update, and delete AP groups.",

		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the AP group.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"device_macs": schema.ListAttribute{
				MarkdownDescription: "List of AP devices MAC addresses to include in this AP group.",
				Required:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(validators.Mac),
				},
			},
		},
	}
}

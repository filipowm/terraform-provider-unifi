package apgroup

import (
	"context"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/utils"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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
	DeviceMACs types.Set    `tfsdk:"device_macs"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model
func (m *APGroupModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	var deviceMACs []string

	diags.Append(m.DeviceMACs.ElementsAs(ctx, &deviceMACs, false)...)
	if diags.HasError() {
		return nil, diags
	}

	// Normalize to the controller's canonical MAC form (lowercase, colon)
	// defensively, in case a value reaches here without passing through the
	// plan modifier.
	for i, mac := range deviceMACs {
		deviceMACs[i] = utils.CleanMAC(mac)
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

	deviceMACs, d := types.SetValueFrom(ctx, types.StringType, model.DeviceMACs)
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
			"device_macs": schema.SetAttribute{
				MarkdownDescription: "Set of AP device MAC addresses to include in this AP group. " +
					"MAC addresses are case-insensitive and may use `:` or `-` separators; " +
					"they are normalized to lowercase, colon-separated form.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(validators.Mac),
				},
				PlanModifiers: []planmodifier.Set{
					ut.NormalizeMAC(),
				},
			},
		},
	}
}

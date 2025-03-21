package firewall

import (
	"context"
	"github.com/filipowm/go-unifi/unifi/features"
	ut "github.com/filipowm/terraform-provider-unifi/internal/provider/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &firewallZoneResource{}
	_ resource.ResourceWithConfigure   = &firewallZoneResource{}
	_ resource.ResourceWithImportState = &firewallZoneResource{}
	_ resource.ResourceWithModifyPlan  = &firewallZoneResource{}
	_ base.Resource                    = &firewallZoneResource{}
)

// firewallZoneModel represents the data model for a UniFi Firewall Zone
type firewallZoneModel struct {
	base.Model
	Name     types.String `tfsdk:"name"`
	Networks types.List   `tfsdk:"networks"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model
func (m *firewallZoneModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	var networkIDs []string

	diags.Append(ut.ListElementsAs(m.Networks, &networkIDs)...)
	if diags.HasError() {
		return nil, diags
	}

	return &unifi.FirewallZone{
		ID:         m.ID.ValueString(),
		Name:       m.Name.ValueString(),
		NetworkIDs: networkIDs,
	}, diags
}

// Merge updates the Terraform model with values from the UniFi API model
func (m *firewallZoneModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	model, ok := other.(*unifi.FirewallZone)
	if !ok {
		diags.AddError("Invalid model type", "Expected *unifi.FirewallZone")
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.Name = types.StringValue(model.Name)

	networkIDs, d := types.ListValueFrom(ctx, types.StringType, model.NetworkIDs)
	diags = append(diags, d...)
	m.Networks = networkIDs

	return diags
}

type firewallZoneResource struct {
	*base.GenericResource[*firewallZoneModel]
}

func (r *firewallZoneResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	resp.Diagnostics.Append(r.RequireMinVersion("9.0.0")...)
	site, diags := r.GetClient().ResolveSiteFromConfig(ctx, req.Config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	resp.Diagnostics.Append(r.RequireFeaturesEnabled(ctx, site, features.ZoneBasedFirewall, features.ZoneBasedFirewallMigration)...)
}

// NewFirewallZoneResource creates a new instance of the firewall zone resource
func NewFirewallZoneResource() resource.Resource {
	return &firewallZoneResource{
		GenericResource: base.NewGenericResource(
			"unifi_firewall_zone",
			func() *firewallZoneModel { return &firewallZoneModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetFirewallZone(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					return client.CreateFirewallZone(ctx, site, model.(*unifi.FirewallZone))
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					return client.UpdateFirewallZone(ctx, site, model.(*unifi.FirewallZone))
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteFirewallZone(ctx, site, id)
				},
			},
		),
	}
}

// Schema defines the schema for the resource
func (r *firewallZoneResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_firewall_zone` resource manages firewall zones in the UniFi controller.\n\n" +
			"Firewall zones allow you to group networks together for firewall rule application. " +
			"This resource allows you to create, update, and delete firewall zones.",

		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the firewall zone.",
				Required:            true,
			},
			"networks": schema.ListAttribute{
				MarkdownDescription: "List of network IDs to include in this firewall zone.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             ut.DefaultEmptyList(types.StringType),
			},
		},
	}
}

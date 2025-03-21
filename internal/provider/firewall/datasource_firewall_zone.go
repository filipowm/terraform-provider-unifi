package firewall

import (
	"context"
	"fmt"
	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &firewallZoneDatasource{}
	_ datasource.DataSourceWithConfigure = &firewallZoneDatasource{}
	_ base.Resource                      = &firewallZoneDatasource{}
)

type firewallZoneDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

func (d *firewallZoneDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func NewFirewallZoneDatasource() datasource.DataSource {
	return &firewallZoneDatasource{}
}

func (d *firewallZoneDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *firewallZoneDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *firewallZoneDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *firewallZoneDatasource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "unifi_firewall_zone"
}

func (d *firewallZoneDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_firewall_zone` datsources allows retrieving existing firewall zone details from the UniFi controller by the zone name.",
		Attributes: map[string]schema.Attribute{
			"id":   base.ID(),
			"site": base.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the firewall zone.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"networks": schema.ListAttribute{
				MarkdownDescription: "List of network IDs that this firewall zone contains.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *firewallZoneDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	resp.Diagnostics.Append(d.RequireMinVersion("9.0.0")...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state firewallZoneModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := d.client.ResolveSite(&state)

	list, err := d.client.ListFirewallZone(ctx, site)

	if err != nil {
		resp.Diagnostics.AddError("Failed to list Firewall zones", err.Error())
		return
	}
	if len(list) == 0 {
		resp.Diagnostics.AddError("Firewall zone not found", "No firewall zone found")
		return
	}

	expectedName := state.Name.ValueString()
	var found *unifi.FirewallZone
	for _, zone := range list {
		if zone.Name == expectedName {
			found = &zone
			break
		}
	}
	if found == nil {
		resp.Diagnostics.AddError("Firewall zone not found", fmt.Sprintf("No firewall zone with name %q found", expectedName))
		return
	}

	(&state).Merge(ctx, found)
	state.SetID(found.ID)
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

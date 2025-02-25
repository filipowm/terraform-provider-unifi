package dns

import (
	"context"
	"fmt"
	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ datasource.DataSource                     = &dnsRecordDatasource{}
	_ datasource.DataSourceWithConfigure        = &dnsRecordDatasource{}
	_ base.BaseData                             = &dnsRecordDatasource{}
	_ datasource.DataSourceWithConfigValidators = &dnsRecordDatasource{}
)

type dnsRecordDatasource struct {
	client *base.Client
}

func NewDnsRecordDatasource() datasource.DataSource {
	return &dnsRecordDatasource{}
}

func (d dnsRecordDatasource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("filter").AtName("name"),
			path.MatchRoot("filter").AtName("record"),
		),
	}
}

func (d dnsRecordDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d dnsRecordDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d dnsRecordDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, resourceName)
}

func (d dnsRecordDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about a specific DNS record.",
		Attributes:  dnsRecordDatasourceAttributes,
		Blocks: map[string]schema.Block{
			"filter": schema.SingleNestedBlock{
				Description: "Filter to apply to the DNS record.",
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Description: "DNS record name.",
						Optional:    true,
					},
					"record": schema.StringAttribute{
						Description: "DNS record content.",
						Optional:    true,
					},
				},
			},
		},
	}
}

func (d dnsRecordDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if !d.client.SupportsDnsRecords() {
		resp.Diagnostics.AddError("DNS Records are not supported", fmt.Sprintf("The Unifi controller in version %q does not support DNS records. Required controller version: %q", d.client.Version, base.ControllerVersionDnsRecords))
	}
	var state dnsRecordDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
	if state.Filter == nil {
		// TODO remove after testing validation
		resp.Diagnostics.AddError("Filter is required", "Filter is required. Validation should prevent this from happening.")
		return
	}
	list, err := d.client.ListDNSRecord(ctx, d.client.Site)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list DNS records", err.Error())
		return
	}
	if len(list) == 0 {
		resp.Diagnostics.AddError("DNS record not found", "No DNS record found")
		return
	}
	var nameFilter, recordFilter string
	if utils.IsStringValueNotEmpty(state.Filter.Name) {
		nameFilter = state.Filter.Name.ValueString()
	}
	if utils.IsStringValueNotEmpty(state.Filter.Record) {
		recordFilter = state.Filter.Record.ValueString()
	}
	if nameFilter != "" && recordFilter != "" {
		// TODO remove after testing validation
		resp.Diagnostics.AddError("Filter is invalid", "Only one of 'name' or 'record' can be specified. Validation should prevent this from happening.")
		return
	}
	var found *unifi.DNSRecord
	for _, record := range list {
		if nameFilter != "" && record.Key == nameFilter {
			found = &record
			break
		}
		if recordFilter != "" && record.Value == recordFilter {
			found = &record
			break
		}
	}
	if found == nil {
		resp.Diagnostics.AddError("DNS record not found", "No DNS record found")
		return
	}
	state.dnsRecordModel.merge(found)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

package dns

import (
	"context"
	"fmt"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var (
	_ datasource.DataSource                     = &dnsRecordDatasource{}
	_ datasource.DataSourceWithConfigure        = &dnsRecordDatasource{}
	_ base.Resource                             = &dnsRecordDatasource{}
	_ datasource.DataSourceWithConfigValidators = &dnsRecordDatasource{}
)

type dnsRecordDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

func (d *dnsRecordDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func NewDnsRecordDatasource() datasource.DataSource {
	return &dnsRecordDatasource{}
}

func (d *dnsRecordDatasource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("name"),
			path.MatchRoot("record"),
		),
	}
}

func (d *dnsRecordDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *dnsRecordDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *dnsRecordDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *dnsRecordDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, resourceName)
}

func (d *dnsRecordDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about a specific DNS record configured in your UniFi network. " +
			"This data source allows you to look up DNS records by either their name or record content. " +
			"It's particularly useful for validating existing DNS configurations or referencing DNS records in other resources.",
		Attributes: dnsRecordDatasourceAttributes,
	}
}

func (d *dnsRecordDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if !d.client.SupportsDnsRecords() {
		resp.Diagnostics.AddError("DNS Records are not supported", fmt.Sprintf("The Unifi controller in version %q does not support DNS records. Required controller version: %q", d.client.Version, base.ControllerVersionDnsRecords))
	}
	var state dnsRecordModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := d.client.ResolveSite(&state)

	list, err := d.client.ListDNSRecord(ctx, site)

	if err != nil {
		resp.Diagnostics.AddError("Failed to list DNS records", err.Error())
		return
	}
	if len(list) == 0 {
		resp.Diagnostics.AddError("DNS record not found", "No DNS record found")
		return
	}

	var nameFilter, recordFilter string
	if utils.IsStringValueNotEmpty(state.Name) {
		nameFilter = state.Name.ValueString()
	}
	if utils.IsStringValueNotEmpty(state.Record) {
		recordFilter = state.Record.ValueString()
	}
	if nameFilter != "" && recordFilter != "" {
		// TODO remove after testing validation
		resp.Diagnostics.AddError("Filter is invalid", "Only one of 'name' or 'record' can be specified. Validation should prevent this from happening.")
		return
	}
	var found []*unifi.DNSRecord
	for _, record := range list {
		if nameFilter != "" && record.Key == nameFilter {
			found = append(found, &record)
			break
		}
		if recordFilter != "" && record.Value == recordFilter {
			found = append(found, &record)
			break
		}
	}

	if len(found) == 0 {
		resp.Diagnostics.AddError("DNS record not found", "No DNS record found")
		return
	} else if len(found) > 1 {
		resp.Diagnostics.AddError("Multiple DNS records found", "More than one DNS record found")
		return
	}
	(&state).Merge(ctx, found[0])
	state.SetID(found[0].ID)
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

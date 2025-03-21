package dns

import (
	"context"
	"fmt"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	ut "github.com/filipowm/terraform-provider-unifi/internal/provider/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &dnsRecordsDatasource{}
	_ datasource.DataSourceWithConfigure = &dnsRecordsDatasource{}
	_ base.Resource                      = &dnsRecordsDatasource{}
)

type dnsRecordsDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

func NewDnsRecordsDatasource() datasource.DataSource {
	return &dnsRecordsDatasource{}
}

func (d *dnsRecordsDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *dnsRecordsDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *dnsRecordsDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *dnsRecordsDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *dnsRecordsDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%ss", req.ProviderTypeName, resourceName)
}

func (d *dnsRecordsDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about a all DNS records.",
		Attributes: map[string]schema.Attribute{
			"site": ut.SiteAttribute(),
			"result": schema.ListNestedAttribute{
				Description: "The list of DNS records.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: dnsRecordDatasourceAttributes,
				},
			},
		},
	}
}

func (d *dnsRecordsDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state dnsRecordsDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	site := d.client.ResolveSite(&state)
	records, err := d.client.ListDNSRecord(ctx, site)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list DNS records", err.Error())
		return
	}
	for _, record := range records {
		state.Records = append(state.Records, &dnsRecordModel{
			Model: base.Model{
				ID:   types.StringValue(record.ID),
				Site: types.StringValue(site),
			},
			Name:     types.StringValue(record.Key),
			Record:   types.StringValue(record.Value),
			Enabled:  types.BoolValue(record.Enabled),
			Port:     types.Int32Value(int32(record.Port)),
			Priority: types.Int32Value(int32(record.Priority)),
			Type:     types.StringValue(record.RecordType),
			TTL:      types.Int32Value(int32(record.Ttl)),
			Weight:   types.Int32Value(int32(record.Weight)),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

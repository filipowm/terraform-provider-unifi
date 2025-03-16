package dns

import (
	"context"
	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const resourceName = "dns_record"

var _ base.ResourceModel = &dnsRecordModel{}

type dnsRecordModel struct {
	base.Model
	Name     types.String `tfsdk:"name"`
	Record   types.String `tfsdk:"record"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	Port     types.Int32  `tfsdk:"port"`
	Priority types.Int32  `tfsdk:"priority"`
	Type     types.String `tfsdk:"type"`
	TTL      types.Int32  `tfsdk:"ttl"`
	Weight   types.Int32  `tfsdk:"weight"`
}

type dnsRecordsDatasourceModel struct {
	Site    types.String      `tfsdk:"site"`
	Records []*dnsRecordModel `tfsdk:"result"`
}

func (b *dnsRecordsDatasourceModel) GetSite() string {
	return b.Site.ValueString()
}

func (b *dnsRecordsDatasourceModel) GetRawSite() types.String {
	return b.Site
}

func (b *dnsRecordsDatasourceModel) SetSite(site string) {
	b.Site = types.StringValue(site)
}

var dnsRecordDatasourceAttributes = map[string]schema.Attribute{
	"id":   base.ID(),
	"site": base.SiteAttribute(),
	"name": schema.StringAttribute{
		Description: "DNS record name.",
		Optional:    true,
		Computed:    true,
		Validators: []validator.String{
			stringvalidator.ConflictsWith(path.MatchRoot("record")),
		},
	},
	"record": schema.StringAttribute{
		Description: "DNS record content.",
		Optional:    true,
		Computed:    true,
		Validators: []validator.String{
			stringvalidator.ConflictsWith(path.MatchRoot("name")),
		},
	},
	"enabled": schema.BoolAttribute{
		Description: "Whether the DNS record is enabled.",
		Computed:    true,
	},
	"port": schema.Int32Attribute{
		Description: "The port of the DNS record.",
		Computed:    true,
	},
	"priority": schema.Int32Attribute{
		Description: "Priority of the DNS records. Present only for MX and SRV records; unused by other record types.",
		Computed:    true,
	},
	"type": schema.StringAttribute{
		Description: "The type of the DNS record.",
		Computed:    true,
	},
	"ttl": schema.Int32Attribute{
		Description: "Time To Live (TTL) of the DNS record in seconds. Setting to 0 means 'automatic'.",
		Computed:    true,
	},
	"weight": schema.Int32Attribute{
		Description: "A numeric value indicating the relative weight of the record.",
		Computed:    true,
	},
}

func (d *dnsRecordModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	return &unifi.DNSRecord{
		ID:         d.ID.ValueString(),
		Key:        d.Name.ValueString(),
		Value:      d.Record.ValueString(),
		Enabled:    d.Enabled.ValueBool(),
		Port:       int(d.Port.ValueInt32()),
		Priority:   int(d.Priority.ValueInt32()),
		RecordType: d.Type.ValueString(),
		Ttl:        int(d.TTL.ValueInt32()),
		Weight:     int(d.Weight.ValueInt32()),
	}, diag.Diagnostics{}
}

func (d *dnsRecordModel) Merge(_ context.Context, i interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	other, ok := i.(*unifi.DNSRecord)
	if !ok {
		diags.AddError("Invalid model type", "Expected *unifi.DNSRecord")
		return diags
	}
	d.ID = types.StringValue(other.ID)
	d.Name = types.StringValue(other.Key)
	d.Record = types.StringValue(other.Value)
	d.Enabled = types.BoolValue(other.Enabled)
	d.Port = types.Int32Value(int32(other.Port))
	d.Priority = types.Int32Value(int32(other.Priority))
	d.Type = types.StringValue(other.RecordType)
	d.TTL = types.Int32Value(int32(other.Ttl))
	d.Weight = types.Int32Value(int32(other.Weight))
	return diags
}

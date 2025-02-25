package dns

import (
	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const resourceName = "dns_record"

type dnsRecordModel struct {
	ID       types.String `tfsdk:"id"`
	SiteID   types.String `tfsdk:"site_id"`
	Name     types.String `tfsdk:"name"`
	Record   types.String `tfsdk:"record"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	Port     types.Int32  `tfsdk:"port"`
	Priority types.Int32  `tfsdk:"priority"`
	Type     types.String `tfsdk:"type"`
	TTL      types.Int32  `tfsdk:"ttl"`
	Weight   types.Int32  `tfsdk:"weight"`
}

type dnsRecordDatasourceModel struct {
	*dnsRecordModel
	Filter *dnsRecordFilterModel `tfsdk:"filter"`
}

type dnsRecordsDatasourceModel struct {
	Records []*dnsRecordModel `tfsdk:"result"`
}

var dnsRecordDatasourceAttributes = map[string]schema.Attribute{
	"id": utils.ID(),
	"name": schema.StringAttribute{
		Description: "DNS record name.",
		Computed:    true,
	},
	"record": schema.StringAttribute{
		Description: "DNS record content.",
		Computed:    true,
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

type dnsRecordFilterModel struct {
	Name   types.String `tfsdk:"name"`
	Record types.String `tfsdk:"record"`
}

func (d *dnsRecordModel) asUnifiModel() *unifi.DNSRecord {
	return &unifi.DNSRecord{
		ID:         d.ID.ValueString(),
		SiteID:     d.SiteID.ValueString(),
		Key:        d.Name.ValueString(),
		Value:      d.Record.ValueString(),
		Enabled:    d.Enabled.ValueBool(),
		Port:       int(d.Port.ValueInt32()),
		Priority:   int(d.Priority.ValueInt32()),
		RecordType: d.Type.ValueString(),
		Ttl:        int(d.TTL.ValueInt32()),
		Weight:     int(d.Weight.ValueInt32()),
	}
}

func (d *dnsRecordModel) merge(other *unifi.DNSRecord) {
	d.ID = types.StringValue(other.ID)
	d.SiteID = types.StringValue(other.SiteID)
	d.Name = types.StringValue(other.Key)
	d.Record = types.StringValue(other.Value)
	d.Enabled = types.BoolValue(other.Enabled)
	d.Port = types.Int32Value(int32(other.Port))
	d.Priority = types.Int32Value(int32(other.Priority))
	d.Type = types.StringValue(other.RecordType)
	d.TTL = types.Int32Value(int32(other.Ttl))
	d.Weight = types.Int32Value(int32(other.Weight))
}

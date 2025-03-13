package dns

import (
	"context"
	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                     = &dnsRecordResource{}
	_ resource.ResourceWithConfigure        = &dnsRecordResource{}
	_ resource.ResourceWithImportState      = &dnsRecordResource{}
	_ resource.ResourceWithModifyPlan       = &dnsRecordResource{}
	_ resource.ResourceWithConfigValidators = &dnsRecordResource{}
	_ base.Resource                         = &dnsRecordResource{}
)

type dnsRecordResource struct {
	*base.GenericResource[*dnsRecordModel]
}

func (d *dnsRecordResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		validators.RequiredNoneIf(path.MatchRoot("type"), types.StringValue("A"), path.MatchRoot("priority"), path.MatchRoot("weight"), path.MatchRoot("port")),
		validators.RequiredNoneIf(path.MatchRoot("type"), types.StringValue("AAAA"), path.MatchRoot("priority"), path.MatchRoot("weight"), path.MatchRoot("port")),
		validators.RequiredNoneIf(path.MatchRoot("type"), types.StringValue("CNAME"), path.MatchRoot("priority"), path.MatchRoot("weight"), path.MatchRoot("port")),
		validators.RequiredNoneIf(path.MatchRoot("type"), types.StringValue("MX"), path.MatchRoot("weight"), path.MatchRoot("port")),
		validators.RequiredNoneIf(path.MatchRoot("type"), types.StringValue("NS"), path.MatchRoot("priority"), path.MatchRoot("weight"), path.MatchRoot("port")),
		validators.RequiredNoneIf(path.MatchRoot("type"), types.StringValue("PTR"), path.MatchRoot("priority"), path.MatchRoot("weight"), path.MatchRoot("port")),
		validators.RequiredNoneIf(path.MatchRoot("type"), types.StringValue("SOA"), path.MatchRoot("priority"), path.MatchRoot("weight"), path.MatchRoot("port")),
		validators.RequiredNoneIf(path.MatchRoot("type"), types.StringValue("TXT"), path.MatchRoot("priority"), path.MatchRoot("weight"), path.MatchRoot("port")),
	}
}

func (d *dnsRecordResource) ModifyPlan(_ context.Context, _ resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	resp.Diagnostics.Append(d.RequireMinVersion("8.2")...)
}

func NewDnsRecordResource() resource.Resource {
	return &dnsRecordResource{
		GenericResource: base.NewGenericResource(
			"unifi_dns_record",
			func() *dnsRecordModel { return &dnsRecordModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetDNSRecord(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					return client.CreateDNSRecord(ctx, site, model.(*unifi.DNSRecord))
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					return client.UpdateDNSRecord(ctx, site, model.(*unifi.DNSRecord))
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteDNSRecord(ctx, site, id)
				},
			},
		),
	}
}

func (d *dnsRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_dns_record` resource manages DNS records in the UniFi controller's DNS server.\n\n" +
			"This resource allows you to configure various types of DNS records for local name resolution. Common use cases include:\n" +
			"  * Creating A records for local servers and devices\n" +
			"  * Setting up CNAME aliases for internal services\n" +
			"  * Configuring MX records for local mail servers\n" +
			"  * Adding TXT records for service verification\n\n",

		Attributes: map[string]schema.Attribute{
			"id":   base.ID(),
			"site": base.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "DNS record name.",
				Required:            true,
			},
			"record": schema.StringAttribute{
				MarkdownDescription: "The content of the DNS record. The expected value depends on the record type:\n" +
					"  * For A records: IPv4 address (e.g., '192.168.1.10')\n" +
					"  * For AAAA records: IPv6 address\n" +
					"  * For CNAME records: Canonical name (e.g., 'server1.example.com')\n" +
					"  * For MX records: Mail server hostname\n" +
					"  * For TXT records: Text content (e.g., 'v=spf1 include:_spf.example.com ~all')",
				Required: true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the DNS record is active. Defaults to true. Set to false to temporarily disable resolution without removing the record.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"port": schema.Int32Attribute{
				MarkdownDescription: "The port number for SRV records. Valid values are between 1 and 65535. Only used with SRV records.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int32{
					int32validator.Between(1, 65535),
				},
			},
			"priority": schema.Int32Attribute{
				MarkdownDescription: "Priority value for MX and SRV records. Lower values indicate higher priority. Required for MX and SRV records, ignored for other types.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int32{
					int32validator.AtLeast(1),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of DNS record. Valid values are:\n" +
					"  * `A` - Maps a hostname to IPv4 address\n" +
					"  * `AAAA` - Maps a hostname to IPv6 address\n" +
					"  * `CNAME` - Creates an alias for another domain name\n" +
					"  * `MX` - Specifies mail servers for the domain\n" +
					"  * `NS` - Delegates a subdomain to a set of name servers\n" +
					"  * `PTR` - Creates a pointer to a canonical name (reverse DNS)\n" +
					"  * `SOA` - Specifies authoritative information about the domain\n" +
					"  * `SRV` - Specifies location of services (hostname and port)\n" +
					"  * `TXT` - Holds descriptive text",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("A", "AAAA", "CNAME", "MX", "NS", "PTR", "SOA", "SRV", "TXT"),
				},
			},
			"ttl": schema.Int32Attribute{
				MarkdownDescription: "Time To Live (TTL) in seconds, determines how long DNS resolvers should cache this record. Set to 0 for automatic TTL. " +
					"Common values: 300 (5 minutes), 3600 (1 hour), 86400 (1 day).",
				Optional: true,
				Computed: true,
			},
			"weight": schema.Int32Attribute{
				MarkdownDescription: "A relative weight for SRV records with the same priority. Higher values get proportionally more traffic. Only used with SRV records.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

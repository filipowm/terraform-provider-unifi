package dns

import (
	"context"
	"fmt"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &dnsRecordResource{}
	_ resource.ResourceWithConfigure   = &dnsRecordResource{}
	_ resource.ResourceWithImportState = &dnsRecordResource{}
	_ base.BaseData                    = &dnsRecordResource{}
)

type dnsRecordResource struct {
	client *base.Client
}

func (d *dnsRecordResource) SetClient(client *base.Client) {
	d.client = client
}

func NewDnsRecordResource() resource.Resource {
	return &dnsRecordResource{}
}

func (d *dnsRecordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	base.ConfigureResource(d, req, resp)
}

func (d *dnsRecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, resourceName)
}

func (d *dnsRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a DNS record in the Unifi controller.",

		Attributes: map[string]schema.Attribute{
			"id":      utils.ID(),
			"site_id": utils.ID("The site ID where the DNS record is located."),
			"name": schema.StringAttribute{
				MarkdownDescription: "DNS record name.",
				Required:            true,
			},
			"record": schema.StringAttribute{
				MarkdownDescription: "DNS record content.",
				Required:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the DNS record is enabled.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"port": schema.Int32Attribute{
				MarkdownDescription: "The port of the DNS record.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int32{
					int32validator.Between(1, 65535),
				},
			},
			"priority": schema.Int32Attribute{
				MarkdownDescription: "Required for MX and SRV records; unused by other record types. Records with lower priorities are preferred",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int32{
					int32validator.AtLeast(1),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the DNS record.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("A", "AAAA", "CNAME", "MX", "NS", "PTR", "SOA", "SRV", "TXT"),
				},
			},
			"ttl": schema.Int32Attribute{
				MarkdownDescription: "Time To Live (TTL) of the DNS record in seconds. Setting to 0 means 'automatic'.",
				Optional:            true,
				Computed:            true,
			},
			"weight": schema.Int32Attribute{
				MarkdownDescription: "A numeric value indicating the relative weight of the record.",
				Optional:            true,
				Computed:            true,
			},
		},
	}

}

func (d *dnsRecordResource) checkSupportsDnsRecords(diag *diag.Diagnostics) {
	if !d.client.SupportsDnsRecords() {
		diag.AddError("DNS Records are not supported", fmt.Sprintf("The Unifi controller in version %q does not support DNS records. Required controller version: %q", d.client.Version, base.ControllerVersionDnsRecords))
	}
}

func (d *dnsRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	d.checkSupportsDnsRecords(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan dnsRecordModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	body := plan.asUnifiModel()

	res, err := d.client.CreateDNSRecord(ctx, d.client.Site, body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating DNS record", err.Error())
		return
	}
	plan.merge(res)

	resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (d *dnsRecordResource) read(ctx context.Context, state *dnsRecordModel, diag *diag.Diagnostics) {
	res, err := d.client.GetDNSRecord(ctx, d.client.Site, state.ID.ValueString())
	if err != nil {
		diag.AddError("Error reading DNS record", err.Error())
		return
	}
	state.merge(res)
}

func (d *dnsRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	d.checkSupportsDnsRecords(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	var state dnsRecordModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	d.read(ctx, &state, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (d *dnsRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	d.checkSupportsDnsRecords(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan, state dnsRecordModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	body := plan.asUnifiModel()

	res, err := d.client.UpdateDNSRecord(ctx, d.client.Site, body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating DNS record", err.Error())
		return
	}
	state.merge(res)
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (d *dnsRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	d.checkSupportsDnsRecords(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	var state dnsRecordModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := d.client.DeleteDNSRecord(ctx, d.client.Site, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting DNS record", err.Error())
		return
	}
}

func (d *dnsRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	d.checkSupportsDnsRecords(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	id := req.ID
	if id == "" {
		resp.Diagnostics.AddError("Invalid import ID", "The ID must be set")
		return
	}

	state := dnsRecordModel{
		ID: types.StringValue(id),
	}
	d.read(ctx, &state, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}
	d.read(ctx, &state, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

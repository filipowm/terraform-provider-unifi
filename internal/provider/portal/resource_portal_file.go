package portal

import (
	"context"
	"fmt"
	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	ut "github.com/filipowm/terraform-provider-unifi/internal/provider/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
)

var (
	_ resource.Resource                = &portalFileResource{}
	_ resource.ResourceWithConfigure   = &portalFileResource{}
	_ resource.ResourceWithImportState = &portalFileResource{}
	_ base.Resource                    = &portalFileResource{}
)

type portalFileResource struct {
	*base.GenericResource[*portalFileModel]
}

type portalFileModel struct {
	base.Model
	Filename     types.String `tfsdk:"filename"`
	FilePath     types.String `tfsdk:"file_path"`
	ContentType  types.String `tfsdk:"content_type"`
	FileSize     types.Int64  `tfsdk:"file_size"`
	MD5          types.String `tfsdk:"md5"`
	URL          types.String `tfsdk:"url"`
	LastModified types.Int64  `tfsdk:"last_modified"`
}

func (m *portalFileModel) Merge(_ context.Context, data interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	portalFile, ok := data.(*unifi.PortalFile)
	if !ok {
		diags.AddError("Invalid data type", fmt.Sprintf("Expected *unifi.PortalFile, got: %T", data))
		return diags
	}

	m.ID = types.StringValue(portalFile.ID)
	m.Filename = types.StringValue(portalFile.Filename)
	m.ContentType = types.StringValue(portalFile.ContentType)
	m.FileSize = types.Int64Value(int64(portalFile.FileSize))
	m.MD5 = types.StringValue(portalFile.MD5)
	m.URL = types.StringValue(portalFile.URL)
	m.LastModified = types.Int64Value(int64(portalFile.LastModified))

	return diags
}

func (m *portalFileModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	// Not used for upload - we don't convert the model to a UniFi model
	// The file path is used directly for upload
	return nil, diag.Diagnostics{}
}

func NewPortalFileResource() resource.Resource {
	return &portalFileResource{
		GenericResource: base.NewGenericResource(
			"unifi_portal_file",
			func() *portalFileModel { return &portalFileModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetPortalFile(ctx, site, id)
				},
				Create: nil, // Custom implementation in CreateWithContext
				Update: nil, // Portal files cannot be updated, only replaced
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeletePortalFile(ctx, site, id)
				},
			},
		),
	}
}

func (r *portalFileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_portal_file` resource manages files uploaded to the UniFi guest portal. " +
			"This resource allows you to upload images that can be used in customizing " +
			"the UniFi guest portal interface.\n\n" +
			"**Note:** This resource uploads files to the UniFi controller. The file must exist on the local filesystem " +
			"where Terraform is executed.",

		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"file_path": schema.StringAttribute{
				MarkdownDescription: "Path to the file on the local filesystem to upload to the UniFi controller. " +
					"The file must exist and be readable.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"filename": schema.StringAttribute{
				MarkdownDescription: "Name of the file as stored in the UniFi controller.",
				Computed:            true,
			},
			"content_type": schema.StringAttribute{
				MarkdownDescription: "MIME type of the file.",
				Computed:            true,
			},
			"file_size": schema.Int64Attribute{
				MarkdownDescription: "Size of the file in bytes.",
				Computed:            true,
			},
			"md5": schema.StringAttribute{
				MarkdownDescription: "MD5 hash of the file content.",
				Computed:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "URL where the file can be accessed on the UniFi controller.",
				Computed:            true,
			},
			"last_modified": schema.Int64Attribute{
				MarkdownDescription: "Timestamp when the file was last modified.",
				Computed:            true,
			},
		},
	}
}

func (r *portalFileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data portalFileModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get file path
	filePath := data.FilePath.ValueString()
	if filePath == "" {
		resp.Diagnostics.AddError("File path is required", "A valid file path must be provided")
		return
	}

	// Check if file exists
	_, err := os.Stat(filePath)
	if err != nil {
		resp.Diagnostics.AddError("Invalid file path", fmt.Sprintf("Error accessing file: %s", err))
		return
	}
	site := r.GetClient().ResolveSite(&data)

	portalFile, err := r.GetClient().UploadPortalFile(ctx, site, filePath)
	if err != nil {
		resp.Diagnostics.AddError("Error uploading file", fmt.Sprintf("Could not upload file: %s", err))
		return
	}

	// Map response back to model
	resp.Diagnostics.Append(data.Merge(ctx, portalFile)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Site = types.StringValue(site)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *portalFileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddError("Import is not supported", "The `unifi_portal_file` resource does not support import")
}

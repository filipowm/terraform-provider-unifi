package base

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Resource interface {
	SetClient(client *Client)
	SetVersionValidator(validator ControllerVersionValidator)
}

// ResourceModel defines the interface that all setting models must implement
type ResourceModel interface {
	GetSite() string
	GetRawSite() types.String
	SetSite(string)
	GetID() string
	GetRawID() types.String
	SetID(string)
	AsUnifiModel(context.Context) (interface{}, diag.Diagnostics)
	Merge(context.Context, interface{}) diag.Diagnostics
}

type Model struct {
	ID   types.String `tfsdk:"id"`
	Site types.String `tfsdk:"site"`
}

func (b *Model) GetID() string {
	return b.ID.ValueString()
}

func (b *Model) GetRawID() types.String {
	return b.ID
}

func (b *Model) SetID(id string) {
	b.ID = types.StringValue(id)
}

func (b *Model) GetSite() string {
	return b.Site.ValueString()
}

func (b *Model) GetRawSite() types.String {
	return b.Site
}

func (b *Model) SetSite(site string) {
	b.Site = types.StringValue(site)
}

func ConfigureDatasource(base Resource, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Datasource Configure Type", fmt.Sprintf("Expected provider.Client, got: %T", req.ProviderData))
		return
	}
	if cfg == nil {
		resp.Diagnostics.AddError("Empty configuration", "provider.Client is nil")
		return
	}
	base.SetClient(cfg)
	base.SetVersionValidator(NewControllerVersionValidator(cfg))
}

func ConfigureResource(base Resource, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected provider.Client, got: %T", req.ProviderData))
		return
	}
	if cfg == nil {
		resp.Diagnostics.AddError("Empty configuration", "provider.Client is nil")
		return
	}
	base.SetClient(cfg)
	base.SetVersionValidator(NewControllerVersionValidator(cfg))
}

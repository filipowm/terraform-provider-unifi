package base

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type BaseData interface {
	SetClient(client *Client)
}

type Site struct {
	Site types.String `tfsdk:"site"`
}

func NewSite(str string) Site {
	return Site{
		Site: types.StringValue(str),
	}
}

func (s *Site) SetSite(site string) {
	s.Site = types.StringValue(site)
}

func (s *Site) AsString() string {
	return s.Site.ValueString()
}

func ConfigureDatasource(base BaseData, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
}

func ConfigureResource(base BaseData, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
}

package v2

import (
	"fmt"
	"github.com/filipowm/terraform-provider-unifi/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type BaseData interface {
	SetClient(client *provider.Client)
}

func ConfigureDatasource(base BaseData, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(*provider.Client)
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

	cfg, ok := req.ProviderData.(*provider.Client)
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

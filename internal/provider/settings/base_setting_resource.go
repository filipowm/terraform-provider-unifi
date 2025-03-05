package settings

import (
	"context"
	"errors"
	"fmt"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// BaseSettingResource provides common functionality for all setting resources
type BaseSettingResource[T base.ResourceModel] struct {
	base.ControllerVersionValidator
	client       *base.Client
	typeName     string
	modelFactory func() T
	getter       func(context.Context, *base.Client, string) (interface{}, error)
	updater      func(context.Context, *base.Client, string, interface{}) (interface{}, error)
}

// NewBaseSettingResource creates a new base setting resource
func NewBaseSettingResource[T base.ResourceModel](
	typeName string,
	modelFactory func() T,
	getter func(context.Context, *base.Client, string) (interface{}, error),
	updater func(context.Context, *base.Client, string, interface{}) (interface{}, error),
) *BaseSettingResource[T] {
	return &BaseSettingResource[T]{
		typeName:     typeName,
		modelFactory: modelFactory,
		getter:       getter,
		updater:      updater,
	}
}

// GetClient returns the UniFi client
func (b *BaseSettingResource[T]) GetClient() *base.Client {
	return b.client
}

// SetClient sets the UniFi client
func (b *BaseSettingResource[T]) SetClient(client *base.Client) {
	b.client = client
}

func (b *BaseSettingResource[T]) SetVersionValidator(validator base.ControllerVersionValidator) {
	b.ControllerVersionValidator = validator
}

func (b *BaseSettingResource[T]) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	base.ConfigureResource(b, req, resp)
}

func (b *BaseSettingResource[T]) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = b.typeName
}

func (b *BaseSettingResource[T]) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, site := base.ImportIDWithSite(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	state := b.modelFactory()
	state.SetID(id)
	state.SetSite(site)
	b.read(ctx, site, state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (b *BaseSettingResource[T]) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if b.client == nil {
		resp.Diagnostics.AddError(
			"Client Not Configured",
			"Expected configured client. Please report this issue to the provider developers.",
		)
		return
	}

	var plan T
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := b.client.ResolveSite(plan)

	body, diags := plan.AsUnifiModel(ctx)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	res, err := b.updater(ctx, b.client, site, body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating settings", err.Error())
		return
	}
	if res == nil {
		resp.Diagnostics.AddError("Error creating settings", fmt.Sprintf("No %[1]s settings returned from the UniFi controller. %[1]s might not be supported on this controller", b.typeName))
		return
	}
	plan.Merge(ctx, res)
	plan.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (b *BaseSettingResource[T]) read(ctx context.Context, site string, state T, diag *diag.Diagnostics) {
	if b.client == nil {
		diag.AddError(
			"Client Not Configured",
			"Expected configured client. Please report this issue to the provider developers.",
		)
		return
	}

	res, err := b.getter(ctx, b.client, site)
	if err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			diag.AddError("Settings not found", "The settings were not found in the UniFi controller")
		} else {
			diag.AddError("Error reading settings", err.Error())
		}
		return
	}
	if res != nil {
		state.Merge(ctx, res)
	}
}

func (b *BaseSettingResource[T]) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if b.client == nil {
		resp.Diagnostics.AddError(
			"Client Not Configured",
			"Expected configured client. Please report this issue to the provider developers.",
		)
		return
	}

	var state T
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	site := b.client.ResolveSite(state)
	b.read(ctx, site, state, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (b *BaseSettingResource[T]) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if b.client == nil {
		resp.Diagnostics.AddError(
			"Client Not Configured",
			"Expected configured client. Please report this issue to the provider developers.",
		)
		return
	}

	var plan, state T
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, diags := plan.AsUnifiModel(ctx)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	site := b.client.ResolveSite(plan)

	res, err := b.updater(ctx, b.client, site, body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating settings", err.Error())
		return
	}
	state.Merge(ctx, res)
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (b *BaseSettingResource[T]) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Not supported
}

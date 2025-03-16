package base

import (
	"context"
	"errors"
	"fmt"
	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type ResourceFunctions struct {
	Read   func(ctx context.Context, client *Client, site string, id string) (interface{}, error)
	Create func(ctx context.Context, client *Client, site string, body interface{}) (interface{}, error)
	Update func(ctx context.Context, client *Client, site string, body interface{}) (interface{}, error)
	Delete func(ctx context.Context, client *Client, site string, id string) error
}

// GenericResource provides common functionality for all resources
type GenericResource[T ResourceModel] struct {
	ControllerVersionValidator
	FeatureValidator
	client       *Client
	typeName     string
	modelFactory func() T
	Handlers     ResourceFunctions
}

// NewGenericResource creates a new base resource
func NewGenericResource[T ResourceModel](
	typeName string,
	modelFactory func() T,
	handlers ResourceFunctions,
) *GenericResource[T] {
	return &GenericResource[T]{
		typeName:     typeName,
		modelFactory: modelFactory,
		Handlers:     handlers,
	}
}

// GetClient returns the UniFi client
func (b *GenericResource[T]) GetClient() *Client {
	return b.client
}

// SetClient sets the UniFi client
func (b *GenericResource[T]) SetClient(client *Client) {
	b.client = client
}

func (b *GenericResource[T]) SetVersionValidator(validator ControllerVersionValidator) {
	b.ControllerVersionValidator = validator
}

func (b *GenericResource[T]) SetFeatureValidator(validator FeatureValidator) {
	b.FeatureValidator = validator
}

func (b *GenericResource[T]) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	ConfigureResource(b, req, resp)
}

func (b *GenericResource[T]) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = b.typeName
}

func (b *GenericResource[T]) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(checkClientConfigured(b.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, site := ImportIDWithSite(req, resp)
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

func (b *GenericResource[T]) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if b.Handlers.Create == nil {
		// Create is not supported
		return
	}
	resp.Diagnostics.Append(checkClientConfigured(b.client)...)
	if resp.Diagnostics.HasError() {
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

	res, err := b.Handlers.Create(ctx, b.client, site, body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating resource", err.Error())
		return
	}
	if res == nil {
		resp.Diagnostics.AddError("Error creating resource", fmt.Sprintf("No %[1]s resource returned from the UniFi controller. %[1]s might not be supported on this controller", b.typeName))
		return
	}
	plan.Merge(ctx, res)
	plan.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (b *GenericResource[T]) read(ctx context.Context, site string, state T, diag *diag.Diagnostics) {
	res, err := b.Handlers.Read(ctx, b.client, site, state.GetID())
	if err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			diag.AddError("Resource not found", "The resource was not found in the UniFi controller")
		} else {
			diag.AddError("Error reading resource", err.Error())
		}
		return
	}
	if res != nil {
		state.Merge(ctx, res)
	}
}

func (b *GenericResource[T]) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if b.Handlers.Read == nil {
		resp.Diagnostics.AddError("Read Not Supported", "Read operation is not supported for this resource. Please report this issue to the provider developers cause this is unexpected issue.")
		return
	}
	resp.Diagnostics.Append(checkClientConfigured(b.client)...)
	if resp.Diagnostics.HasError() {
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

func (b *GenericResource[T]) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if b.Handlers.Update == nil {
		// Update is not supported
		return
	}
	resp.Diagnostics.Append(checkClientConfigured(b.client)...)
	if resp.Diagnostics.HasError() {
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

	res, err := b.Handlers.Update(ctx, b.client, site, body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating resource", err.Error())
		return
	}
	state.Merge(ctx, res)
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (b *GenericResource[T]) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if b.Handlers.Delete == nil {
		// Delete is not supported
		return
	}
	resp.Diagnostics.Append(checkClientConfigured(b.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state T
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := b.client.ResolveSite(state)
	err := b.Handlers.Delete(ctx, b.client, site, state.GetID())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting resource", err.Error())
		return
	}
}

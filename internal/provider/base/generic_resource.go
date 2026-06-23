package base

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// readAfterWriteMaxAttempts and readAfterWriteBackoff bound the
// poll-until-consistent loop used by the opt-in ReadAfterWrite path. The
// controller's GET is eventually consistent for some collections, so a single
// read can briefly return them empty; we re-read a few times before giving up
// and surfacing whatever the controller actually returned.
const (
	readAfterWriteMaxAttempts = 5
	readAfterWriteBackoff     = 500 * time.Millisecond
)

// WriteConsistencyChecker is an OPTIONAL interface a model may implement so the
// ReadAfterWrite path can poll past the controller's eventual-consistency
// window. The poll is gated by BOTH ResourceFunctions.ReadAfterWrite==true and
// the model implementing this interface, so non-settings resources and settings
// models that don't implement it keep the single-read behavior unchanged.
type WriteConsistencyChecker interface {
	// ConsistentAfterWrite reports whether THIS model (just read back via GET)
	// already reflects the values that were written. 'planned' is the same
	// concrete model type carrying the planned/just-written framework values.
	// Return true when the read-back is acceptable (stop polling), false to
	// retry. Only fields known to be eventually-consistent need to be checked;
	// everything else should be treated as consistent (return true) so
	// unrelated diffs never cause a poll.
	ConsistentAfterWrite(planned any) bool
}

type ResourceFunctions struct {
	Read   func(ctx context.Context, client *Client, site string, id string) (interface{}, error)
	Create func(ctx context.Context, client *Client, site string, body interface{}) (interface{}, error)
	Update func(ctx context.Context, client *Client, site string, body interface{}) (interface{}, error)
	Delete func(ctx context.Context, client *Client, site string, id string) error

	// ReadAfterWrite, when true, makes Create and Update populate the final
	// state by performing a Read (Handlers.Read + Merge) against the persisted
	// datastore instead of merging the write-operation response echo. This is
	// opt-in because some resources are eventually consistent on create (an
	// immediate post-create GET may 404). Defaults to false, which preserves
	// the historical behavior of merging the write echo for all resources.
	ReadAfterWrite bool
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

	// Opt-in read-after-write: re-read from the persisted datastore so the
	// final state reflects the GET response rather than the (possibly
	// eventually-consistent) write echo. Merge above already populated the ID
	// so the Read handler can resolve the resource.
	if b.Handlers.ReadAfterWrite && b.Handlers.Read != nil {
		// Use the configured values (req.Plan) as the consistency reference, NOT
		// the write echo: the PUT echo can itself be eventually-consistent and
		// come back empty (the very reason read-after-write exists), so deriving
		// the reference from it would skip the poll in exactly the failure case.
		// req.Plan reliably carries what the user configured. 'plan' (now holding
		// the echo) is the read target, preserving the single-read-into-plan flow.
		var planned T
		resp.Diagnostics.Append(req.Plan.Get(ctx, &planned)...)
		if resp.Diagnostics.HasError() {
			return
		}

		b.readUntilConsistent(ctx, site, plan, planned, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.SetSite(site)
	}

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

// readUntilConsistent performs the ReadAfterWrite read (Handlers.Read + Merge
// into target) and, when target implements WriteConsistencyChecker, polls until
// the read-back is consistent with the planned/just-written values or the cap is
// reached. When target does NOT implement the interface it performs exactly one
// read, preserving the historical single-read behavior. On cap exhaustion it
// leaves target holding the LAST REAL read result (no fabricated/stale values),
// so a value that genuinely never persists still surfaces as an inconsistency.
func (b *GenericResource[T]) readUntilConsistent(ctx context.Context, site string, target T, planned T, diags *diag.Diagnostics) {
	for attempt := 0; ; attempt++ {
		b.read(ctx, site, target, diags)
		if diags.HasError() {
			return
		}

		checker, ok := any(target).(WriteConsistencyChecker)
		if !ok {
			// Model opts out of consistency polling: single read, identical to
			// the historical behavior.
			return
		}
		if checker.ConsistentAfterWrite(planned) {
			return
		}
		if attempt+1 >= readAfterWriteMaxAttempts {
			// Cap reached: fall through with the last real read result rather
			// than masking the inconsistency.
			return
		}

		// Short backoff before re-reading, respecting cancellation.
		select {
		case <-ctx.Done():
			return
		case <-time.After(readAfterWriteBackoff):
		}
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

	// Opt-in read-after-write: re-read from the persisted datastore so the
	// final state reflects the GET response rather than the (possibly
	// eventually-consistent) write echo. State already carries the ID. 'plan'
	// (from req.Plan) is untouched by the merge above and carries the
	// planned/just-written values used for the consistency check; 'state' is the
	// read target (preserving the single-read-into-state behavior).
	if b.Handlers.ReadAfterWrite && b.Handlers.Read != nil {
		b.readUntilConsistent(ctx, site, state, plan, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		state.SetSite(site)
	}

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

package firewall

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/go-unifi/unifi/features"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	ut "github.com/filipowm/terraform-provider-unifi/internal/provider/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                     = &firewallZonePolicyOrderResource{}
	_ resource.ResourceWithConfigure        = &firewallZonePolicyOrderResource{}
	_ resource.ResourceWithConfigValidators = &firewallZonePolicyOrderResource{}
	_ resource.ResourceWithImportState      = &firewallZonePolicyOrderResource{}
	_ resource.ResourceWithModifyPlan       = &firewallZonePolicyOrderResource{}
	_ base.Resource                         = &firewallZonePolicyOrderResource{}
)

// FirewallZonePolicyOrderModel represents the ordering of custom firewall zone
// policies within a single source -> destination zone pair.
type FirewallZonePolicyOrderModel struct {
	base.Model
	SourceZoneID        types.String `tfsdk:"source_zone_id"`
	DestinationZoneID   types.String `tfsdk:"destination_zone_id"`
	BeforePredefinedIDs types.List   `tfsdk:"before_predefined_ids"` // ordered []string
	AfterPredefinedIDs  types.List   `tfsdk:"after_predefined_ids"`  // ordered []string
}

// AsUnifiModel builds the FirewallPolicyOrderUpdate payload from the model.
//
// The go-unifi `before_predefined_ids` / `after_predefined_ids` JSON tags carry
// no `omitempty`, so a nil slice marshals as JSON `null`. We therefore always
// send non-nil slices (`[]string{}` when a list is unset) so the controller
// receives `[]` rather than `null`.
func (m *FirewallZonePolicyOrderModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	before := []string{}
	after := []string{}
	diags.Append(ut.ListElementsAs(m.BeforePredefinedIDs, &before)...)
	diags.Append(ut.ListElementsAs(m.AfterPredefinedIDs, &after)...)
	if diags.HasError() {
		return nil, diags
	}

	return &unifi.FirewallPolicyOrderUpdate{
		SourceZoneId:        m.SourceZoneID.ValueString(),
		DestinationZoneId:   m.DestinationZoneID.ValueString(),
		BeforePredefinedIds: before,
		AfterPredefinedIds:  after,
	}, diags
}

// partitionZonePairOrder reconstructs the before/after-predefined custom-policy
// ordering for a single source -> destination zone pair out of the full policy
// list returned by the controller.
//
// INFERENCE (validated by the acceptance test, NOT yet confirmed against a live
// controller): the controller expresses policy order via the integer `Index`.
// We treat the smallest `Index` among the pair's PREDEFINED (built-in) policies
// as the boundary — custom policies whose index is below it run BEFORE the
// predefined policies, the rest run AFTER. Each partition is sorted by ascending
// `Index` and the policy IDs are collected in that order. Policies belonging to
// any other zone pair are ignored.
//
// Kept as a small, well-named pure function so the partition heuristic is easy
// to adjust once the live-controller semantics are confirmed.
func partitionZonePairOrder(policies []unifi.FirewallZonePolicy, sourceZoneID, destZoneID string) (before, after []string) {
	type indexedPolicy struct {
		id    string
		index int
	}

	var customs []indexedPolicy
	hasPredefined := false
	minPredefinedIndex := 0

	for _, p := range policies {
		if p.Source.ZoneID != sourceZoneID || p.Destination.ZoneID != destZoneID {
			continue
		}
		if p.Predefined {
			if !hasPredefined || p.Index < minPredefinedIndex {
				minPredefinedIndex = p.Index
				hasPredefined = true
			}
			continue
		}
		customs = append(customs, indexedPolicy{id: p.ID, index: p.Index})
	}

	sort.SliceStable(customs, func(i, j int) bool { return customs[i].index < customs[j].index })

	before = []string{}
	after = []string{}
	for _, c := range customs {
		// Inferred fallback: when the pair has NO predefined policies we cannot
		// locate the boundary, so every custom policy is treated as "after"
		// (the controller's default placement for new custom policies). This is
		// an inference about controller semantics that the acceptance test
		// validates.
		if hasPredefined && c.index < minPredefinedIndex {
			before = append(before, c.id)
		} else {
			after = append(after, c.id)
		}
	}
	return before, after
}

// Merge implements base.ResourceModel. It reconstructs the ordering for this
// zone pair from a full controller policy list, managing ALL custom policies in
// the pair (no subset filtering). This is the behavior used at import time,
// where there is no prior state to scope ownership; Read uses applyOrder with a
// managed set instead.
func (m *FirewallZonePolicyOrderModel) Merge(ctx context.Context, data interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	policies, ok := data.([]unifi.FirewallZonePolicy)
	if !ok {
		diags.AddError("Invalid data type", fmt.Sprintf("Expected []unifi.FirewallZonePolicy, got: %T", data))
		return diags
	}

	return m.applyOrder(ctx, policies, nil)
}

// applyOrder reconstructs the before/after-predefined ordering for this zone
// pair from the controller policy list and writes it into the model.
//
// The managed argument selects which custom policies this resource owns:
//   - nil  -> manage EVERY custom policy in the pair (used at import, where there
//     is no prior state to scope ownership; see Merge).
//   - set  -> subset ownership: keep only the listed IDs, preserving the
//     reconstructed order, and ignore any other custom policy in the pair (used
//     by Read so unlisted policies don't cause a perpetual diff).
//
// To avoid a spurious null-vs-`[]` plan diff, an empty reconstructed list
// inherits the null-ness of the corresponding prior-state list: an attribute the
// practitioner omitted (null) round-trips as null, while an explicit empty list
// stays an empty list. The prior-state lists are read from the receiver before
// they are overwritten.
func (m *FirewallZonePolicyOrderModel) applyOrder(ctx context.Context, policies []unifi.FirewallZonePolicy, managed map[string]struct{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	// Capture prior null-ness before overwriting the lists (FIX 3).
	priorBeforeNull := m.BeforePredefinedIDs.IsNull()
	priorAfterNull := m.AfterPredefinedIDs.IsNull()

	source := m.SourceZoneID.ValueString()
	dest := m.DestinationZoneID.ValueString()

	before, after := partitionZonePairOrder(policies, source, dest)
	if managed != nil {
		before = filterManaged(before, managed)
		after = filterManaged(after, managed)
	}

	beforeList, d := listPreservingNull(ctx, before, priorBeforeNull)
	diags.Append(d...)
	afterList, d := listPreservingNull(ctx, after, priorAfterNull)
	diags.Append(d...)

	m.BeforePredefinedIDs = beforeList
	m.AfterPredefinedIDs = afterList
	m.ID = types.StringValue(source + ":" + dest)

	return diags
}

// filterManaged keeps only the IDs present in the managed set, preserving the
// input (reconstructed) order. It implements Read's subset ownership: custom
// policies the resource does not list are dropped from state.
func filterManaged(ids []string, managed map[string]struct{}) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, ok := managed[id]; ok {
			out = append(out, id)
		}
	}
	return out
}

// managedIDSet returns the set of custom-policy IDs this resource currently
// manages: the union of the IDs in the prior-state before/after lists. Read uses
// it to scope ownership to the listed policies only (a non-nil, possibly empty
// set means subset ownership).
func managedIDSet(m *FirewallZonePolicyOrderModel) map[string]struct{} {
	set := make(map[string]struct{})
	var before, after []string
	// Prior-state lists are already-validated string lists; null/unknown lists
	// contribute nothing, so the returned diagnostics are not actionable here.
	ut.ListElementsAs(m.BeforePredefinedIDs, &before)
	ut.ListElementsAs(m.AfterPredefinedIDs, &after)
	for _, id := range before {
		set[id] = struct{}{}
	}
	for _, id := range after {
		set[id] = struct{}{}
	}
	return set
}

// listPreservingNull converts a slice to a list value. A non-empty slice becomes
// a known list. An empty slice maps to null when priorNull is true (the
// attribute was omitted) or to an explicit empty list otherwise, so an Optional
// attribute round-trips without a null-vs-`[]` plan diff.
func listPreservingNull(ctx context.Context, vals []string, priorNull bool) (types.List, diag.Diagnostics) {
	if len(vals) == 0 {
		if priorNull {
			return types.ListNull(types.StringType), diag.Diagnostics{}
		}
		return types.ListValueFrom(ctx, types.StringType, []string{})
	}
	return types.ListValueFrom(ctx, types.StringType, vals)
}

type firewallZonePolicyOrderResource struct {
	*base.GenericResource[*FirewallZonePolicyOrderModel]
}

// NewFirewallZonePolicyOrderResource creates a new instance of the firewall zone
// policy order resource. It embeds GenericResource purely to inherit the shared
// Configure/Metadata/client/version+feature-gating infrastructure; all CRUD is
// overridden below, so an empty ResourceFunctions{} is passed.
func NewFirewallZonePolicyOrderResource() resource.Resource {
	return &firewallZonePolicyOrderResource{
		GenericResource: base.NewGenericResource(
			"unifi_firewall_zone_policy_order",
			func() *FirewallZonePolicyOrderModel { return &FirewallZonePolicyOrderModel{} },
			base.ResourceFunctions{},
		),
	}
}

// Schema defines the schema for the resource.
func (r *firewallZonePolicyOrderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_firewall_zone_policy_order` resource controls the ordering of the custom " +
			"`unifi_firewall_zone_policy` policies within a single source -> destination zone pair.\n\n" +
			"This resource manages the relative order of ONLY the custom policies listed in `before_predefined_ids` " +
			"and `after_predefined_ids` within the zone pair. Any other (unlisted) custom policies in the same zone " +
			"pair are left untouched and ignored — they are neither reordered by this resource nor surfaced in its " +
			"state.\n\n" +
			"Policies are evaluated top-to-bottom; the order within each list is significant. Policies listed in " +
			"`before_predefined_ids` run BEFORE the controller's predefined (built-in) policies for the zone pair, " +
			"and policies listed in `after_predefined_ids` run AFTER them. At least one of the two lists must be set; " +
			"prefer OMITTING an unused list over setting it to an empty list (`[]`).\n\n" +
			"Because `index` on `unifi_firewall_zone_policy` is controller-assigned and read-only, this resource is " +
			"the supported way to make per-zone-pair policy order deterministic. Use `depends_on` to ensure the " +
			"referenced policies exist before this resource is applied.\n\n" +
			"~> Deleting this resource does NOT change the order on the controller; it only stops Terraform from " +
			"managing the ordering for the zone pair.\n\n" +
			"!> This is experimental feature, that requires UniFi OS 9.0.0 or later and Zone Based Firewall feature enabled. " +
			"Check [official documentation](https://help.ui.com/hc/en-us/articles/28223082254743-Migrating-to-Zone-Based-Firewalls-in-UniFi) how to migrate to Zone-Based firewalls.",

		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"source_zone_id": schema.StringAttribute{
				MarkdownDescription: "ID of the source firewall zone of the pair whose policy order is managed. " +
					"Changing the zone pair forces a new resource.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destination_zone_id": schema.StringAttribute{
				MarkdownDescription: "ID of the destination firewall zone of the pair whose policy order is managed. " +
					"Changing the zone pair forces a new resource.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"before_predefined_ids": schema.ListAttribute{
				MarkdownDescription: "Ordered IDs of custom `unifi_firewall_zone_policy` policies that run BEFORE the " +
					"predefined (built-in) policies for this zone pair. Order within the list is significant. " +
					"Omit this attribute when it is unused rather than setting it to an empty list (`[]`).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"after_predefined_ids": schema.ListAttribute{
				MarkdownDescription: "Ordered IDs of custom `unifi_firewall_zone_policy` policies that run AFTER the " +
					"predefined (built-in) policies for this zone pair. Order within the list is significant. " +
					"Omit this attribute when it is unused rather than setting it to an empty list (`[]`).",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// ConfigValidators requires at least one of the two ordering lists to be set.
func (r *firewallZonePolicyOrderResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("before_predefined_ids"),
			path.MatchRoot("after_predefined_ids"),
		),
	}
}

// ModifyPlan gates the resource on controller version and the Zone-Based
// Firewall feature flags, mirroring unifi_firewall_zone(_policy).
func (r *firewallZonePolicyOrderResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Destroy plan: nothing to configure, so skip gating (the controller need
	// not still support the feature for Terraform to drop the resource).
	if req.Plan.Raw.IsNull() {
		return
	}
	resp.Diagnostics.Append(r.RequireMinVersion("9.0.0")...)
	site, diags := r.GetClient().ResolveSiteFromConfig(ctx, req.Config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	resp.Diagnostics.Append(r.RequireFeaturesEnabled(ctx, site, features.ZoneBasedFirewall, features.ZoneBasedFirewallMigration)...)
}

// applyReorder builds the reorder payload from the model, calls the controller,
// validates the response, and sets ONLY the derived fields (id, site) on the
// model.
//
// It deliberately does NOT reconstruct before_predefined_ids /
// after_predefined_ids from the controller response. Those attributes are
// Optional (not Computed), so post-apply state must equal the configured plan
// value; rebuilding them here would risk a "Provider produced inconsistent
// result after apply" error (issue #122). Reconstruction happens only in Read /
// ImportState, where it reflects observed controller state rather than the plan.
func (r *firewallZonePolicyOrderResource) applyReorder(ctx context.Context, model *FirewallZonePolicyOrderModel, diags *diag.Diagnostics) {
	site := r.GetClient().ResolveSite(model)

	body, d := model.AsUnifiModel(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	update, ok := body.(*unifi.FirewallPolicyOrderUpdate)
	if !ok {
		diags.AddError("Unexpected payload type", fmt.Sprintf("Expected *unifi.FirewallPolicyOrderUpdate, got: %T", body))
		return
	}

	requested := make([]string, 0, len(update.BeforePredefinedIds)+len(update.AfterPredefinedIds))
	requested = append(requested, update.BeforePredefinedIds...)
	requested = append(requested, update.AfterPredefinedIds...)

	result, err := r.GetClient().ReorderFirewallPolicies(ctx, site, update)
	if err != nil {
		diags.AddError("Error reordering firewall zone policies", err.Error())
		return
	}

	// go-unifi does not assert that the reorder response covers the requested
	// IDs (see its TODO at unifi/firewall_zone_policy.go), so validate here.
	if err := validateReorderResponse(requested, result); err != nil {
		diags.AddError("Unexpected reorder response", err.Error())
		return
	}

	// Derived fields only — the Optional ordering lists are left exactly as
	// configured so Create/Update echo the plan verbatim.
	model.ID = types.StringValue(model.SourceZoneID.ValueString() + ":" + model.DestinationZoneID.ValueString())
	model.SetSite(site)
}

// validateReorderResponse ensures the controller returned a non-empty policy set
// that includes every ID we asked it to reorder.
func validateReorderResponse(requestedIDs []string, result []unifi.FirewallZonePolicy) error {
	if len(result) == 0 {
		return fmt.Errorf("the controller returned no policies from the reorder operation")
	}
	present := make(map[string]struct{}, len(result))
	for _, p := range result {
		present[p.ID] = struct{}{}
	}
	var missing []string
	for _, id := range requestedIDs {
		if _, ok := present[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("the reorder response did not include the following requested policy IDs "+
			"(they may have been deleted outside Terraform): %s", strings.Join(missing, ", "))
	}
	return nil
}

func (r *firewallZonePolicyOrderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Client Not Configured", "Expected configured client. Please report this issue to the provider developers.")
		return
	}

	var model FirewallZonePolicyOrderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.applyReorder(ctx, &model, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// applyReorder set only id+site; the Optional ordering lists keep their
	// configured plan values so post-apply state matches the plan (FIX 1).
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *firewallZonePolicyOrderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Client Not Configured", "Expected configured client. Please report this issue to the provider developers.")
		return
	}

	var model FirewallZonePolicyOrderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.applyReorder(ctx, &model, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// As in Create, the configured ordering lists are echoed verbatim (FIX 1).
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *firewallZonePolicyOrderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Client Not Configured", "Expected configured client. Please report this issue to the provider developers.")
		return
	}

	var model FirewallZonePolicyOrderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	site := r.GetClient().ResolveSite(&model)

	// Subset ownership (FIX 2): this resource manages only the custom policies it
	// already lists. Capture that set from prior state BEFORE reconstructing, so
	// unlisted custom policies in the same zone pair are ignored rather than
	// pulled into state (which would cause a perpetual diff).
	managed := managedIDSet(&model)

	policies, err := r.GetClient().ListFirewallZonePolicy(ctx, site)
	if err != nil {
		resp.Diagnostics.AddError("Error reading firewall zone policies", err.Error())
		return
	}

	// Reconstruct order for the zone pair, keeping only the managed policies. If
	// none of them remain the lists come back null/empty; we intentionally keep
	// the resource in state rather than removing it, since there is no controller
	// object to 404.
	resp.Diagnostics.Append(model.applyOrder(ctx, policies, managed)...)
	model.SetSite(site)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Delete is a state-only no-op: the controller exposes no "un-order" operation,
// so removing this resource simply stops Terraform from managing the ordering
// for the zone pair and leaves the controller-side order untouched.
func (r *firewallZonePolicyOrderResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

// ImportState imports an order resource via `<site>:<source_zone_id>:<destination_zone_id>`.
func (r *firewallZonePolicyOrderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Client Not Configured", "Expected configured client. Please report this issue to the provider developers.")
		return
	}

	// ImportIDWithSite splits on the FIRST colon, so `site:source:dest` yields
	// site=`site`, id=`source:dest`.
	id, site := base.ImportIDWithSite(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected import ID in the format `<site>:<source_zone_id>:<destination_zone_id>`",
		)
		return
	}

	model := FirewallZonePolicyOrderModel{}
	model.SourceZoneID = types.StringValue(parts[0])
	model.DestinationZoneID = types.StringValue(parts[1])
	model.SetID(id)
	model.SetSite(site)

	policies, err := r.GetClient().ListFirewallZonePolicy(ctx, site)
	if err != nil {
		resp.Diagnostics.AddError("Error reading firewall zone policies", err.Error())
		return
	}

	resp.Diagnostics.Append(model.Merge(ctx, policies)...)
	model.SetSite(site)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

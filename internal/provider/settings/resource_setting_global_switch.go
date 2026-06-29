package settings

import (
	"context"
	"errors"
	"fmt"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	ut "github.com/filipowm/terraform-provider-unifi/internal/provider/types"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/utils"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// aclL3IsolationModel models a single layer-3 isolation entry: a source network
// and the set of destination networks it is isolated from. Values are UniFi
// network IDs (the `_id` of a unifi_network), not names or CIDRs.
type aclL3IsolationModel struct {
	SourceNetwork       types.String `tfsdk:"source_network"`
	DestinationNetworks types.Set    `tfsdk:"destination_networks"`
}

func (m *aclL3IsolationModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"source_network": types.StringType,
		"destination_networks": types.SetType{
			ElemType: types.StringType,
		},
	}
}

// globalSwitchModel is the Terraform model for unifi_setting_global_switch. It
// is intentionally NARROW: it models only the three switch-isolation fields of
// the controller's `global_switch` setting object. The remaining fields of that
// object (dhcp_snoop, dot1x_*, stp_version, jumboframe_enabled, etc.) are NOT
// modeled and are preserved verbatim by the read-modify-write write path.
type globalSwitchModel struct {
	base.Model
	AclDeviceIsolation types.Set `tfsdk:"acl_device_isolation"`
	AclL3Isolation     types.Set `tfsdk:"acl_l3_isolation"`
	SwitchExclusions   types.Set `tfsdk:"switch_exclusions"`
}

// overlay applies only the configured (known and non-null) isolation fields of
// the model onto cur, leaving every other field of cur untouched. This is the
// core of the read-modify-write write path: cur is the current controller
// object, so unmanaged fields survive the subsequent full-object PUT.
//
// It is a pure function (no client access) so it can be unit-tested.
func (m *globalSwitchModel) overlay(ctx context.Context, cur *unifi.SettingGlobalSwitch) diag.Diagnostics {
	diags := diag.Diagnostics{}

	if ut.IsDefined(m.AclDeviceIsolation) {
		var v []string
		diags.Append(m.AclDeviceIsolation.ElementsAs(ctx, &v, false)...)
		if diags.HasError() {
			return diags
		}
		cur.AclDeviceIsolation = v
	}

	if ut.IsDefined(m.SwitchExclusions) {
		var v []string
		diags.Append(m.SwitchExclusions.ElementsAs(ctx, &v, false)...)
		if diags.HasError() {
			return diags
		}
		// Defensive normalization in case a value reaches here without passing
		// through the plan-time NormalizeMAC modifier.
		for i := range v {
			v[i] = utils.CleanMAC(v[i])
		}
		cur.SwitchExclusions = v
	}

	if ut.IsDefined(m.AclL3Isolation) {
		var entries []aclL3IsolationModel
		diags.Append(m.AclL3Isolation.ElementsAs(ctx, &entries, false)...)
		if diags.HasError() {
			return diags
		}
		seen := make(map[string]struct{}, len(entries))
		result := make([]unifi.SettingGlobalSwitchAclL3Isolation, 0, len(entries))
		for _, e := range entries {
			sn := e.SourceNetwork.ValueString()
			// Defensive dedup; the plan-time uniqueSourceNetworkValidator should
			// already have rejected duplicate source networks.
			if _, ok := seen[sn]; ok {
				continue
			}
			seen[sn] = struct{}{}
			var dest []string
			diags.Append(e.DestinationNetworks.ElementsAs(ctx, &dest, false)...)
			if diags.HasError() {
				return diags
			}
			result = append(result, unifi.SettingGlobalSwitchAclL3Isolation{
				SourceNetwork:       sn,
				DestinationNetworks: dest,
			})
		}
		cur.AclL3Isolation = result
	}

	return diags
}

// AsUnifiModel builds a UniFi model from the plan by overlaying the configured
// isolation fields onto a fresh struct. It is required to satisfy
// base.ResourceModel, but the dedicated Create/Update overrides on
// globalSwitchResource do NOT use it (they overlay onto the current controller
// object instead, to preserve unmanaged fields). It is kept correct so the
// model is a well-behaved base.ResourceModel.
func (m *globalSwitchModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	cur := &unifi.SettingGlobalSwitch{}
	diags := m.overlay(ctx, cur)
	return cur, diags
}

// Merge populates the model from a UniFi model. Empty collections round-trip as
// empty Sets (not null) via types.SetValueFrom so the Optional+Computed
// attributes stay consistent post-apply.
func (m *globalSwitchModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingGlobalSwitch)
	if !ok {
		diags.AddError("Invalid model type", "Expected *unifi.SettingGlobalSwitch")
		return diags
	}

	m.ID = types.StringValue(model.ID)

	// types.SetValueFrom maps a nil slice to a null Set; normalize nil to an
	// empty slice so empty collections round-trip as empty (known) Sets, keeping
	// the Optional+Computed attributes consistent after apply.
	deviceIso, d := types.SetValueFrom(ctx, types.StringType, nonNilStrings(model.AclDeviceIsolation))
	diags.Append(d...)
	m.AclDeviceIsolation = deviceIso

	switchExcl, d := types.SetValueFrom(ctx, types.StringType, nonNilStrings(model.SwitchExclusions))
	diags.Append(d...)
	m.SwitchExclusions = switchExcl

	entries := make([]aclL3IsolationModel, 0, len(model.AclL3Isolation))
	for _, e := range model.AclL3Isolation {
		dest, dd := types.SetValueFrom(ctx, types.StringType, nonNilStrings(e.DestinationNetworks))
		diags.Append(dd...)
		entries = append(entries, aclL3IsolationModel{
			SourceNetwork:       types.StringValue(e.SourceNetwork),
			DestinationNetworks: dest,
		})
	}
	l3Set, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: (&aclL3IsolationModel{}).AttributeTypes()}, entries)
	diags.Append(d...)
	m.AclL3Isolation = l3Set

	return diags
}

// nonNilStrings returns s, or an empty (non-nil) slice when s is nil, so that
// types.SetValueFrom produces an empty Set rather than a null Set.
func nonNilStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// decideBaseGlobalSwitch implements the read-modify-write GET-failure decision:
//   - on unifi.ErrNotFound the setting is absent, so we create it from a fresh
//     struct (overlaying only configured fields);
//   - on any other error we abort (returning a non-nil error) so the caller does
//     NOT issue a destructive PUT;
//   - otherwise the current controller object is used as the overlay base.
//
// It is pure so it can be unit-tested.
func decideBaseGlobalSwitch(cur *unifi.SettingGlobalSwitch, err error) (*unifi.SettingGlobalSwitch, error) {
	if err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			return &unifi.SettingGlobalSwitch{}, nil
		}
		return nil, err
	}
	if cur == nil {
		return &unifi.SettingGlobalSwitch{}, nil
	}
	return cur, nil
}

type globalSwitchResource struct {
	*base.GenericResource[*globalSwitchModel]
}

func NewGlobalSwitchResource() resource.Resource {
	r := &globalSwitchResource{}
	r.GenericResource = base.NewGenericResource(
		"unifi_setting_global_switch",
		func() *globalSwitchModel { return &globalSwitchModel{} },
		base.ResourceFunctions{
			// Read is wired so the promoted ImportState path (and any base
			// helper) can resolve the setting. Create/Update/Delete are left nil
			// on purpose: Create/Update are overridden below with a
			// read-modify-write path, and Delete is a no-op for this singleton.
			Read: func(ctx context.Context, client *base.Client, site, _ string) (interface{}, error) {
				return client.GetSettingGlobalSwitch(ctx, site)
			},
		},
	)
	return r
}

func (r *globalSwitchResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_global_switch` resource manages the switch isolation settings " +
			"(device isolation and ACL-based layer-3 isolation) for a UniFi site, exposed in the controller UI " +
			"under **Settings → Network → Switch Isolation Settings**.\n\n" +
			"This resource is intentionally narrow: it manages only the isolation-related fields of the " +
			"controller's `global_switch` setting object. All other fields of that object (such as DHCP snooping, " +
			"802.1X, STP, jumbo frames, and flow control) are preserved untouched using a read-modify-write write " +
			"path, so this resource can be adopted without clobbering settings managed elsewhere " +
			"(for example, DHCP snooping via `unifi_setting_usw`).\n\n" +
			"~> **Clearing collections is not supported.** Because the underlying controller fields are " +
			"`omitempty`, setting any of `acl_device_isolation`, `acl_l3_isolation`, or `switch_exclusions` to an " +
			"empty value cannot reliably clear it via the API. Configure at least one element, or remove the " +
			"attribute to stop managing it (the last applied value is retained). Empty values are rejected at plan time.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"acl_device_isolation": schema.SetAttribute{
				MarkdownDescription: "Set of device identifiers to isolate (the controller's **Device Isolation** control). " +
					"Each element is sent to the controller verbatim, with no validation or normalization: the UniFi " +
					"`global_switch` API does not constrain this field's format, so supply the identifiers exactly as the " +
					"controller expects them (refer to the controller UI). Reordering has no effect (this is an unordered " +
					"set). At least one element is required when set; remove the attribute to stop managing it.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"acl_l3_isolation": schema.SetNestedAttribute{
				MarkdownDescription: "Set of layer-3 (network-to-network) isolation rules. Each entry isolates a source " +
					"network from a set of destination networks. All values are UniFi network IDs (the `id` of a " +
					"`unifi_network` resource), not network names or CIDRs. Reordering has no effect (unordered set).",
				Optional: true,
				Computed: true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					uniqueSourceNetworkValidator{},
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"source_network": schema.StringAttribute{
							MarkdownDescription: "The UniFi network ID (the `id` of a `unifi_network`) that this rule " +
								"applies to. Must be unique across all entries.",
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"destination_networks": schema.SetAttribute{
							MarkdownDescription: "Set of UniFi network IDs that the source network is isolated from. " +
								"At least one destination network is required.",
							ElementType: types.StringType,
							Required:    true,
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
						},
					},
				},
			},
			"switch_exclusions": schema.SetAttribute{
				MarkdownDescription: "Set of switch MAC addresses excluded from isolation enforcement. MAC addresses " +
					"are case-insensitive and may use `:` or `-` separators; they are normalized to lowercase, " +
					"colon-separated form. At least one element is required when set; remove the attribute to stop " +
					"managing it.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(validators.Mac),
				},
				PlanModifiers: []planmodifier.Set{
					ut.NormalizeMAC(),
					setplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// write performs the read-modify-write used by both Create and Update:
//
//	GET current -> overlay only configured fields -> PUT -> re-read -> set state.
//
// On a GET failure that is not ErrNotFound it aborts WITHOUT issuing a PUT, so a
// transient/auth/decode error can never clobber the controller object.
func (r *globalSwitchResource) write(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State, diags *diag.Diagnostics) {
	client := r.GetClient()
	if client == nil {
		diags.AddError("Provider not configured", "The UniFi client is not configured. Please report this issue to the provider developers.")
		return
	}

	var model globalSwitchModel
	diags.Append(plan.Get(ctx, &model)...)
	if diags.HasError() {
		return
	}
	site := client.ResolveSite(&model)

	cur, err := client.GetSettingGlobalSwitch(ctx, site)
	cur, abort := decideBaseGlobalSwitch(cur, err)
	if abort != nil {
		diags.AddError("Unable to read current Global Switch settings", abort.Error())
		return
	}

	diags.Append(model.overlay(ctx, cur)...)
	if diags.HasError() {
		return
	}

	// Full-object PUT. cur carries all unmanaged toggles intact. Tolerate an
	// ErrNotFound write echo (eventual consistency) and rely on the re-read.
	if _, err := client.UpdateSettingGlobalSwitch(ctx, site, cur); err != nil && !errors.Is(err, unifi.ErrNotFound) {
		diags.AddError("Unable to update Global Switch settings", err.Error())
		return
	}

	// Read-after-write: re-read from the persisted datastore so post-apply state
	// is reliable (settings PUT echoes are eventually consistent).
	reread, err := client.GetSettingGlobalSwitch(ctx, site)
	if err != nil {
		diags.AddError("Unable to read Global Switch settings after write", err.Error())
		return
	}

	var result globalSwitchModel
	diags.Append(result.Merge(ctx, reread)...)
	if diags.HasError() {
		return
	}
	result.SetSite(site)
	diags.Append(state.Set(ctx, &result)...)
}

func (r *globalSwitchResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.write(ctx, req.Plan, &resp.State, &resp.Diagnostics)
}

func (r *globalSwitchResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.write(ctx, req.Plan, &resp.State, &resp.Diagnostics)
}

// Read overrides the promoted base Read so that an absent setting (ErrNotFound)
// removes the resource from state (drift handling) rather than raising an error.
func (r *globalSwitchResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	client := r.GetClient()
	if client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The UniFi client is not configured. Please report this issue to the provider developers.")
		return
	}

	var state globalSwitchModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := client.ResolveSite(&state)

	cur, err := client.GetSettingGlobalSwitch(ctx, site)
	if err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read Global Switch settings", err.Error())
		return
	}

	resp.Diagnostics.Append(state.Merge(ctx, cur)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ImportState overrides the promoted base ImportState so that importing when the
// setting is absent surfaces a clear, actionable error.
func (r *globalSwitchResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	client := r.GetClient()
	if client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The UniFi client is not configured. Please report this issue to the provider developers.")
		return
	}

	_, site := base.ImportIDWithSite(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	cur, err := client.GetSettingGlobalSwitch(ctx, site)
	if err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			resp.Diagnostics.AddError(
				"Global Switch setting not present",
				fmt.Sprintf("The Global Switch (switch isolation) setting was not found on site %q. "+
					"Import using the form '<site>:global_switch' for a site that has this setting configured.", site),
			)
			return
		}
		resp.Diagnostics.AddError("Unable to import Global Switch settings", err.Error())
		return
	}

	state := &globalSwitchModel{}
	resp.Diagnostics.Append(state.Merge(ctx, cur)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// uniqueSourceNetworkValidator rejects an acl_l3_isolation set that contains two
// entries with the same source_network at plan time.
type uniqueSourceNetworkValidator struct{}

func (v uniqueSourceNetworkValidator) Description(_ context.Context) string {
	return "Each acl_l3_isolation entry must have a unique source_network."
}

func (v uniqueSourceNetworkValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v uniqueSourceNetworkValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	var entries []aclL3IsolationModel
	resp.Diagnostics.Append(req.ConfigValue.ElementsAs(ctx, &entries, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	seen := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		if e.SourceNetwork.IsNull() || e.SourceNetwork.IsUnknown() {
			continue
		}
		sn := e.SourceNetwork.ValueString()
		if _, ok := seen[sn]; ok {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Duplicate source_network",
				fmt.Sprintf("source_network %q appears in more than one acl_l3_isolation entry; each source network must be unique.", sn),
			)
			return
		}
		seen[sn] = struct{}{}
	}
}

var (
	_ base.ResourceModel               = &globalSwitchModel{}
	_ resource.Resource                = &globalSwitchResource{}
	_ resource.ResourceWithConfigure   = &globalSwitchResource{}
	_ resource.ResourceWithImportState = &globalSwitchResource{}
	_ validator.Set                    = uniqueSourceNetworkValidator{}
)

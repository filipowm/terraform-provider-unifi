package types

import (
	"context"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// normalizeMACModifier rewrites every element of a set of MAC addresses to the
// controller's canonical form (lowercase, colon-separated) at plan time via
// utils.CleanMAC. The UniFi controller always stores and returns MACs in that
// form, so without this the planned value for a config such as
// ["AA:BB-CC:DD:EE:FF"] would differ from the post-apply state, producing a
// "Provider produced inconsistent result after apply" error (or a perpetual
// diff on read). This mirrors the legacy utils.MacDiffSuppressFunc behavior.
type normalizeMACModifier struct{}

// NormalizeMAC returns a plan modifier that canonicalizes a set of MAC
// addresses to lowercase, colon-separated form. It is shared across resources
// (e.g. unifi_ap_group, unifi_setting_global_switch) that accept MAC sets.
func NormalizeMAC() planmodifier.Set {
	return normalizeMACModifier{}
}

func (m normalizeMACModifier) Description(_ context.Context) string {
	return "Normalizes MAC addresses to lowercase, colon-separated canonical form."
}

func (m normalizeMACModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m normalizeMACModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	// Read elements as typed strings so unknown/null elements (e.g. MACs
	// interpolated from another resource) are preserved rather than erroring.
	var macs []types.String
	resp.Diagnostics.Append(req.PlanValue.ElementsAs(ctx, &macs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, mac := range macs {
		// Leave the value untouched until every element is known; the
		// modifier runs again on a later plan once they resolve.
		if mac.IsUnknown() || mac.IsNull() {
			return
		}
	}

	normalized := make([]string, len(macs))
	for i, mac := range macs {
		normalized[i] = utils.CleanMAC(mac.ValueString())
	}

	setVal, diags := types.SetValueFrom(ctx, types.StringType, normalized)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.PlanValue = setVal
}

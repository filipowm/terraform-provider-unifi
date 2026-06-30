package validators

import (
	"context"
	"fmt"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/utils"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// UniqueMACs returns a Set validator that rejects a set whose elements denote the
// same MAC address once normalized (utils.CleanMAC: lowercase, colon-separated).
//
// Set membership is by the raw element string, so "AA-BB-CC-DD-EE-FF" and
// "aa:bb:cc:dd:ee:ff" are two distinct members even though they are the same
// hardware address. The MACType element type only reconciles such values during
// state comparison, not set membership, so without this check both would be sent
// to the controller, which collapses them to one entry — yielding a "Provider
// produced inconsistent result after apply" error. This surfaces the problem at
// plan time with a clear message instead.
func UniqueMACs() validator.Set {
	return uniqueMACsValidator{}
}

type uniqueMACsValidator struct{}

func (v uniqueMACsValidator) Description(_ context.Context) string {
	return "MAC addresses must be unique after normalization (case- and separator-insensitive)."
}

func (v uniqueMACsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v uniqueMACsValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	seen := make(map[string]string, len(req.ConfigValue.Elements()))
	for _, el := range req.ConfigValue.Elements() {
		sv, ok := el.(basetypes.StringValuable)
		if !ok {
			continue
		}
		s, diags := sv.ToStringValue(ctx)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Unknown elements (e.g. a MAC interpolated from another resource) cannot
		// be normalized yet; the validator re-runs once they are known.
		if s.IsUnknown() || s.IsNull() {
			continue
		}

		norm := utils.CleanMAC(s.ValueString())
		if first, dup := seen[norm]; dup {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Duplicate MAC address",
				fmt.Sprintf("%q and %q denote the same MAC address; each address may appear only once.", first, s.ValueString()),
			)
			return
		}
		seen[norm] = s.ValueString()
	}
}

var _ validator.Set = uniqueMACsValidator{}

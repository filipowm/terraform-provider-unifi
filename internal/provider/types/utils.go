package types

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ShouldBeRemoved evaluates if an attribute should be removed from the plan during update.
func ShouldBeRemoved(plan attr.Value, state attr.Value, isClone bool) bool {
	return !IsDefined(plan) && IsDefined(state) && !isClone
}

// IsDefined returns true if attribute is known and not null.
func IsDefined(v attr.Value) bool {
	return !v.IsNull() && !v.IsUnknown()
}

func IsEmptyString(s types.String) bool {
	return s.IsNull() || s.IsUnknown() || s.ValueString() == ""
}

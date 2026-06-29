package utils

import "github.com/hashicorp/go-cty/cty"

// IsRawConfigSet reports whether attribute name is set in the raw config: non-null,
// and for a known value non-empty (string), non-zero (number), or non-empty
// (collection). An unknown/interpolated value (e.g. var.x) counts as set, so it
// is not mistaken for unset at plan time.
func IsRawConfigSet(raw cty.Value, name string) bool {
	// A null raw config (e.g. on destroy, or in a unit test that builds
	// ResourceData without a config block) is known-but-null: HasAttribute still
	// reports true, but GetAttr panics on it because go-cty only guards the
	// unknown case, not the null one. Treat it as "not set", like a missing attr.
	if raw.IsNull() {
		return false
	}
	if !raw.Type().HasAttribute(name) {
		return false
	}
	v := raw.GetAttr(name)
	if v.IsNull() {
		return false
	}
	if !v.IsKnown() {
		return true
	}
	switch {
	case v.Type() == cty.String:
		return v.AsString() != ""
	case v.Type() == cty.Number:
		return !v.RawEquals(cty.Zero)
	default: // list / set / tuple / map
		return v.LengthInt() > 0
	}
}

package utils

import "github.com/hashicorp/go-cty/cty"

// IsRawConfigSet reports whether attribute name is set in the raw config: non-null,
// and for a known value non-empty (string), non-zero (number), non-empty
// (collection), or any bool — including an explicit false, which unlike "" or 0 is
// distinguishable from absence here. An unknown/interpolated value (e.g. var.x)
// counts as set, so it is not mistaken for unset at plan time.
//
// Note this answers "did the user write this?", not "does this ask for anything?".
// A caller rejecting a field as inapplicable to the resource's purpose usually
// wants the latter, and should treat an explicit false as harmless.
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
	case v.Type() == cty.Bool:
		// Reached only for a non-null, known bool, i.e. one the user actually
		// wrote. Unlike "" or 0, `false` is not indistinguishable from absent
		// here -- the raw config represents absent as null -- so an explicit
		// `false` counts as set, and a caller asking "is this attribute valid
		// on this resource?" gets the right answer for it.
		//
		// Without this case a bool fell through to LengthInt(), which panics on
		// a non-collection.
		return true
	default: // list / set / tuple / map
		return v.LengthInt() > 0
	}
}

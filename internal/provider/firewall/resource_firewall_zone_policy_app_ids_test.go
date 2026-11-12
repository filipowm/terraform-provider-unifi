package firewall

import (
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func listStrings(t *testing.T, l types.List) []string {
	t.Helper()
	out := make([]string, 0, len(l.Elements()))
	for _, e := range l.Elements() {
		s, ok := e.(types.String)
		require.True(t, ok, "expected string element, got %T", e)
		out = append(out, s.ValueString())
	}
	return out
}

func readDestination(t *testing.T, dest unifi.FirewallZonePolicyDestination) FirewallZonePolicyDestinationModel {
	t.Helper()
	m := &FirewallZonePolicyModel{}
	diags := m.mergeDestination(t.Context(), &unifi.FirewallZonePolicy{Destination: dest})
	require.False(t, diags.HasError(), "mergeDestination returned errors: %v", diags)
	require.False(t, m.Destination.IsUnknown(), "destination must be known after read")

	var out FirewallZonePolicyDestinationModel
	diags = m.Destination.As(t.Context(), &out, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError(), "destination object conversion failed: %v", diags)
	return out
}

// The controller (and go-unifi) model app and app category ids as []int, while
// the schema exposes them as lists of strings. Reading them used to fail the
// int -> string list conversion, which left the destination unknown after
// import ("Failed to marshal plan to json: value is not known").
func TestFirewallZonePolicyDestinationAppIDsRead(t *testing.T) {
	t.Run("App", func(t *testing.T) {
		out := readDestination(t, unifi.FirewallZonePolicyDestination{
			MatchingTarget: "APP",
			AppIDs:         []int{1245268},
			ZoneID:         "zone1",
		})
		assert.Equal(t, []string{"1245268"}, listStrings(t, out.AppIDs))
		assert.True(t, out.AppCategoryIDs.IsNull())
	})

	t.Run("AppCategory", func(t *testing.T) {
		out := readDestination(t, unifi.FirewallZonePolicyDestination{
			MatchingTarget: "APP_CATEGORY",
			AppCategoryIDs: []int{5, 0, 24},
			ZoneID:         "zone1",
		})
		assert.Equal(t, []string{"5", "0", "24"}, listStrings(t, out.AppCategoryIDs))
		assert.True(t, out.AppIDs.IsNull())
	})
}

func TestStringListToInts(t *testing.T) {
	strList := func(vals ...string) types.List {
		elems := make([]attr.Value, 0, len(vals))
		for _, v := range vals {
			elems = append(elems, types.StringValue(v))
		}
		l, d := types.ListValue(types.StringType, elems)
		require.False(t, d.HasError())
		return l
	}

	t.Run("Valid", func(t *testing.T) {
		var got []int
		diags := stringListToInts(t.Context(), strList("5", "0", "24"), &got)
		require.False(t, diags.HasError(), "unexpected errors: %v", diags)
		assert.Equal(t, []int{5, 0, 24}, got)
	})

	t.Run("NonNumericIsError", func(t *testing.T) {
		var got []int
		diags := stringListToInts(t.Context(), strList("5", "abc"), &got)
		assert.True(t, diags.HasError(), "expected an error for a non-numeric id")
	})
}

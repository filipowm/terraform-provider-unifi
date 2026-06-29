package firewall

import (
	"context"
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFirewallZonePolicyOrderSchema is a controller-free schema-introspection
// test: it builds the resource schema and asserts the ordering attributes have
// the expected shape (Required zone IDs, Optional list-of-string ordering lists).
func TestFirewallZonePolicyOrderSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewFirewallZonePolicyOrderResource().(*firewallZonePolicyOrderResource).Schema(context.Background(), resource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "schema build returned diagnostics: %v", resp.Diagnostics)

	t.Run("source_zone_id is required string", func(t *testing.T) {
		a, ok := resp.Schema.Attributes["source_zone_id"]
		require.True(t, ok, "expected a `source_zone_id` attribute")
		sa, ok := a.(schema.StringAttribute)
		require.True(t, ok, "expected `source_zone_id` to be a StringAttribute, got %T", a)
		assert.True(t, sa.Required, "`source_zone_id` must be Required")
		assert.False(t, sa.Optional, "`source_zone_id` must not be Optional")
	})

	t.Run("destination_zone_id is required string", func(t *testing.T) {
		a, ok := resp.Schema.Attributes["destination_zone_id"]
		require.True(t, ok, "expected a `destination_zone_id` attribute")
		sa, ok := a.(schema.StringAttribute)
		require.True(t, ok, "expected `destination_zone_id` to be a StringAttribute, got %T", a)
		assert.True(t, sa.Required, "`destination_zone_id` must be Required")
		assert.False(t, sa.Optional, "`destination_zone_id` must not be Optional")
	})

	for _, name := range []string{"before_predefined_ids", "after_predefined_ids"} {
		name := name
		t.Run(name+" is optional list of string", func(t *testing.T) {
			a, ok := resp.Schema.Attributes[name]
			require.True(t, ok, "expected a %q attribute", name)
			la, ok := a.(schema.ListAttribute)
			require.True(t, ok, "expected %q to be a ListAttribute, got %T", name, a)
			assert.True(t, la.Optional, "%q must be Optional", name)
			assert.False(t, la.Required, "%q must not be Required", name)
			assert.Equal(t, types.StringType, la.ElementType, "%q must be a list of strings", name)
		})
	}
}

func customPolicy(id, source, dest string, index int) unifi.FirewallZonePolicy {
	return unifi.FirewallZonePolicy{
		ID:          id,
		Index:       index,
		Predefined:  false,
		Source:      unifi.FirewallZonePolicySource{ZoneID: source},
		Destination: unifi.FirewallZonePolicyDestination{ZoneID: dest},
	}
}

func predefinedPolicy(id, source, dest string, index int) unifi.FirewallZonePolicy {
	p := customPolicy(id, source, dest, index)
	p.Predefined = true
	return p
}

// TestPartitionZonePairOrder pins the before/after-predefined reconstruction
// heuristic: custom policies below the predefined boundary go to `before`,
// the rest to `after`, each sorted by ascending Index, and policies from any
// other zone pair are ignored.
func TestPartitionZonePairOrder(t *testing.T) {
	t.Run("partitions and ignores other zone pairs", func(t *testing.T) {
		policies := []unifi.FirewallZonePolicy{
			// Out of order on purpose to exercise the sort.
			customPolicy("c4", "src", "dst", 12000),
			customPolicy("c1", "src", "dst", 9000),
			predefinedPolicy("pre", "src", "dst", 10000), // boundary
			customPolicy("c3", "src", "dst", 11000),
			customPolicy("c2", "src", "dst", 9500),
			// Different zone pair: must be ignored even though its index is low.
			customPolicy("other", "src2", "dst", 1),
		}

		before, after := partitionZonePairOrder(policies, "src", "dst")
		assert.Equal(t, []string{"c1", "c2"}, before)
		assert.Equal(t, []string{"c3", "c4"}, after)
	})

	t.Run("no predefined policies falls back to after", func(t *testing.T) {
		policies := []unifi.FirewallZonePolicy{
			customPolicy("c2", "src", "dst", 9500),
			customPolicy("c1", "src", "dst", 9000),
		}

		before, after := partitionZonePairOrder(policies, "src", "dst")
		assert.Empty(t, before)
		assert.Equal(t, []string{"c1", "c2"}, after)
	})
}

// TestFirewallZonePolicyOrderMerge verifies the Merge reconstruction populates
// the model lists in the right order and maps empty partitions to null.
func TestFirewallZonePolicyOrderMerge(t *testing.T) {
	ctx := context.Background()

	m := &FirewallZonePolicyOrderModel{
		SourceZoneID:      types.StringValue("src"),
		DestinationZoneID: types.StringValue("dst"),
	}

	policies := []unifi.FirewallZonePolicy{
		customPolicy("c1", "src", "dst", 9000),
		predefinedPolicy("pre", "src", "dst", 10000),
		customPolicy("c2", "src", "dst", 11000),
		customPolicy("c3", "src", "dst", 12000),
		customPolicy("ignored", "other", "dst", 5),
	}

	diags := m.Merge(ctx, policies)
	require.False(t, diags.HasError(), "Merge returned diagnostics: %v", diags)

	assert.Equal(t, "src:dst", m.ID.ValueString())

	var before, after []string
	require.False(t, m.BeforePredefinedIDs.ElementsAs(ctx, &before, false).HasError())
	require.False(t, m.AfterPredefinedIDs.ElementsAs(ctx, &after, false).HasError())
	assert.Equal(t, []string{"c1"}, before)
	assert.Equal(t, []string{"c2", "c3"}, after)

	t.Run("empty partition maps to null", func(t *testing.T) {
		m2 := &FirewallZonePolicyOrderModel{
			SourceZoneID:      types.StringValue("src"),
			DestinationZoneID: types.StringValue("dst"),
		}
		// Only `after` policies (no predefined boundary), so `before` is empty.
		onlyAfter := []unifi.FirewallZonePolicy{
			customPolicy("c1", "src", "dst", 9000),
		}
		d := m2.Merge(ctx, onlyAfter)
		require.False(t, d.HasError(), "Merge returned diagnostics: %v", d)
		assert.True(t, m2.BeforePredefinedIDs.IsNull(), "empty `before` partition should map to a null list")
		assert.False(t, m2.AfterPredefinedIDs.IsNull())
	})

	t.Run("wrong data type returns error", func(t *testing.T) {
		m3 := &FirewallZonePolicyOrderModel{
			SourceZoneID:      types.StringValue("src"),
			DestinationZoneID: types.StringValue("dst"),
		}
		d := m3.Merge(ctx, "not a policy slice")
		assert.True(t, d.HasError(), "expected an error for an unexpected data type")
	})
}

// mustStringList builds a known (non-null) list-of-string value for tests.
func mustStringList(t *testing.T, vals ...string) types.List {
	t.Helper()
	if vals == nil {
		vals = []string{}
	}
	l, diags := types.ListValueFrom(context.Background(), types.StringType, vals)
	require.False(t, diags.HasError(), "building list value: %v", diags)
	return l
}

// TestFilterManaged pins the subset-ownership filter: only IDs in the managed
// set survive, and the input (reconstructed) order is preserved.
func TestFilterManaged(t *testing.T) {
	got := filterManaged([]string{"a", "b", "c", "d"}, map[string]struct{}{"d": {}, "b": {}})
	assert.Equal(t, []string{"b", "d"}, got, "must keep only managed IDs, in reconstructed order")

	assert.Empty(t, filterManaged([]string{"a", "b"}, map[string]struct{}{}), "empty managed set drops everything")
	assert.Empty(t, filterManaged(nil, map[string]struct{}{"a": {}}), "nil input yields empty result")
}

// TestFirewallZonePolicyOrderApplyOrderSubsetFiltering exercises the Read path
// (FIX 2): the zone pair has four custom policies, but the resource manages only
// a subset (c1 before, c3 after). applyOrder must drop the unlisted policies (c2,
// c4) while preserving the reconstructed order of the managed ones.
func TestFirewallZonePolicyOrderApplyOrderSubsetFiltering(t *testing.T) {
	ctx := context.Background()

	policies := []unifi.FirewallZonePolicy{
		customPolicy("c1", "src", "dst", 9000),
		customPolicy("c2", "src", "dst", 9500),
		predefinedPolicy("pre", "src", "dst", 10000), // boundary
		customPolicy("c3", "src", "dst", 11000),
		customPolicy("c4", "src", "dst", 12000),
	}

	// Prior state: this resource owns c1 (before) and c3 (after); c2 and c4 are
	// unlisted custom policies in the same pair.
	m := &FirewallZonePolicyOrderModel{
		SourceZoneID:        types.StringValue("src"),
		DestinationZoneID:   types.StringValue("dst"),
		BeforePredefinedIDs: mustStringList(t, "c1"),
		AfterPredefinedIDs:  mustStringList(t, "c3"),
	}
	managed := managedIDSet(m)

	diags := m.applyOrder(ctx, policies, managed)
	require.False(t, diags.HasError(), "applyOrder returned diagnostics: %v", diags)

	assert.Equal(t, "src:dst", m.ID.ValueString())

	var before, after []string
	require.False(t, m.BeforePredefinedIDs.ElementsAs(ctx, &before, false).HasError())
	require.False(t, m.AfterPredefinedIDs.ElementsAs(ctx, &after, false).HasError())
	assert.Equal(t, []string{"c1"}, before, "unlisted c2 must be dropped")
	assert.Equal(t, []string{"c3"}, after, "unlisted c4 must be dropped")

	t.Run("filter preserves reconstructed order, not managed-set order", func(t *testing.T) {
		// The resource lists c4 then c3, but reconstruction order (by ascending
		// index) is c3, c4; filtering must preserve the reconstructed order.
		m2 := &FirewallZonePolicyOrderModel{
			SourceZoneID:       types.StringValue("src"),
			DestinationZoneID:  types.StringValue("dst"),
			AfterPredefinedIDs: mustStringList(t, "c4", "c3"),
		}
		d := m2.applyOrder(ctx, policies, managedIDSet(m2))
		require.False(t, d.HasError(), "applyOrder returned diagnostics: %v", d)
		var after2 []string
		require.False(t, m2.AfterPredefinedIDs.ElementsAs(ctx, &after2, false).HasError())
		assert.Equal(t, []string{"c3", "c4"}, after2)
		assert.True(t, m2.BeforePredefinedIDs.IsNull(), "omitted before list stays null (FIX 3)")
	})

	t.Run("explicit empty list round-trips as empty, not null", func(t *testing.T) {
		// before is an explicit empty list in prior state; with nothing managed in
		// the before partition it must stay an empty list rather than flip to null.
		m3 := &FirewallZonePolicyOrderModel{
			SourceZoneID:        types.StringValue("src"),
			DestinationZoneID:   types.StringValue("dst"),
			BeforePredefinedIDs: mustStringList(t), // explicit []
			AfterPredefinedIDs:  mustStringList(t, "c3"),
		}
		d := m3.applyOrder(ctx, policies, managedIDSet(m3))
		require.False(t, d.HasError(), "applyOrder returned diagnostics: %v", d)
		assert.False(t, m3.BeforePredefinedIDs.IsNull(), "explicit empty before list must stay non-null (FIX 3)")
		assert.Empty(t, m3.BeforePredefinedIDs.Elements(), "explicit empty before list must stay empty")
	})
}

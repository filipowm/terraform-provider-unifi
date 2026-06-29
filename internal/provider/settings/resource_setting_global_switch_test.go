package settings

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDecideBaseGlobalSwitch covers the read-modify-write GET-failure decision
// (the ErrNotFound-fallback / abort-on-other-error / use-current branches).
// These branches cannot be exercised against the always-present singleton on
// the Docker controller, so they are only covered here.
func TestDecideBaseGlobalSwitch(t *testing.T) {
	t.Parallel()

	transient := errors.New("503 service unavailable")
	existing := &unifi.SettingGlobalSwitch{JumboframeEnabled: true}

	tests := map[string]struct {
		cur       *unifi.SettingGlobalSwitch
		err       error
		wantBase  *unifi.SettingGlobalSwitch
		wantAbort error
	}{
		"not found falls back to a fresh struct": {
			cur:      nil,
			err:      unifi.ErrNotFound,
			wantBase: &unifi.SettingGlobalSwitch{},
		},
		"wrapped not found falls back to a fresh struct": {
			cur:      nil,
			err:      fmt.Errorf("get setting: %w", unifi.ErrNotFound),
			wantBase: &unifi.SettingGlobalSwitch{},
		},
		"transient error aborts without a base": {
			cur:       nil,
			err:       transient,
			wantAbort: transient,
		},
		"success uses the current object": {
			cur:      existing,
			err:      nil,
			wantBase: existing,
		},
		"nil current without error falls back to a fresh struct": {
			cur:      nil,
			err:      nil,
			wantBase: &unifi.SettingGlobalSwitch{},
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			base, abort := decideBaseGlobalSwitch(test.cur, test.err)
			if test.wantAbort != nil {
				require.Error(t, abort)
				assert.ErrorIs(t, abort, test.wantAbort)
				assert.Nil(t, base)
				return
			}
			require.NoError(t, abort)
			assert.Equal(t, test.wantBase, base)
		})
	}
}

// TestOverlayPreservesUnmanagedFields proves the central correctness mandate:
// overlay never touches the non-modeled fields of the controller object, and it
// only writes the modeled fields that are configured (known and non-null).
func TestOverlayPreservesUnmanagedFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cur := &unifi.SettingGlobalSwitch{
		JumboframeEnabled: true,
		DHCPSnoop:         true,
		StpVersion:        "rstp",
		RADIUSProfileID:   "rp-1",
		// Pre-existing isolation values that must survive when not configured.
		SwitchExclusions: []string{"aa:bb:cc:dd:ee:ff"},
	}

	// Only acl_device_isolation configured; the other two are null.
	m := &globalSwitchModel{
		AclDeviceIsolation: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("dev-1")}),
		AclL3Isolation:     types.SetNull(types.ObjectType{AttrTypes: (&aclL3IsolationModel{}).AttributeTypes()}),
		SwitchExclusions:   types.SetNull(types.StringType),
	}

	diags := m.overlay(ctx, cur)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	// Unmanaged fields untouched.
	assert.True(t, cur.JumboframeEnabled)
	assert.True(t, cur.DHCPSnoop)
	assert.Equal(t, "rstp", cur.StpVersion)
	assert.Equal(t, "rp-1", cur.RADIUSProfileID)
	// Configured field applied.
	assert.Equal(t, []string{"dev-1"}, cur.AclDeviceIsolation)
	// Non-configured isolation collection left as-is.
	assert.Equal(t, []string{"aa:bb:cc:dd:ee:ff"}, cur.SwitchExclusions)
}

// TestOverlayNormalizesAndDedups checks MAC normalization on switch_exclusions
// and source_network dedup on acl_l3_isolation.
func TestOverlayNormalizesAndDedups(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cur := &unifi.SettingGlobalSwitch{}

	l3Type := types.ObjectType{AttrTypes: (&aclL3IsolationModel{}).AttributeTypes()}
	entry := func(src string, dests ...string) attr.Value {
		vals := make([]attr.Value, len(dests))
		for i, d := range dests {
			vals[i] = types.StringValue(d)
		}
		return types.ObjectValueMust((&aclL3IsolationModel{}).AttributeTypes(), map[string]attr.Value{
			"source_network":       types.StringValue(src),
			"destination_networks": types.SetValueMust(types.StringType, vals),
		})
	}

	m := &globalSwitchModel{
		AclDeviceIsolation: types.SetNull(types.StringType),
		SwitchExclusions: types.SetValueMust(types.StringType, []attr.Value{
			types.StringValue("AA-BB-CC-DD-EE-FF"),
		}),
		AclL3Isolation: types.SetValueMust(l3Type, []attr.Value{
			entry("net-a", "net-b"),
			entry("net-a", "net-c"), // duplicate source, must be dropped
		}),
	}

	diags := m.overlay(ctx, cur)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	assert.Equal(t, []string{"aa:bb:cc:dd:ee:ff"}, cur.SwitchExclusions)
	require.Len(t, cur.AclL3Isolation, 1)
	assert.Equal(t, "net-a", cur.AclL3Isolation[0].SourceNetwork)
}

// TestMergeRoundTrip checks API->model mapping, including nil collections
// mapping to empty (not null) sets via types.SetValueFrom.
func TestMergeRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	model := &unifi.SettingGlobalSwitch{
		ID:                 "gs-id",
		AclDeviceIsolation: []string{"dev-1", "dev-2"},
		SwitchExclusions:   nil, // should become an empty set, not null
		AclL3Isolation: []unifi.SettingGlobalSwitchAclL3Isolation{
			{SourceNetwork: "net-a", DestinationNetworks: []string{"net-b"}},
		},
	}

	m := &globalSwitchModel{}
	diags := m.Merge(ctx, model)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	assert.Equal(t, "gs-id", m.ID.ValueString())

	assert.False(t, m.AclDeviceIsolation.IsNull())
	assert.Len(t, m.AclDeviceIsolation.Elements(), 2)

	// nil slice -> empty (known) set, not null
	assert.False(t, m.SwitchExclusions.IsNull())
	assert.Len(t, m.SwitchExclusions.Elements(), 0)

	assert.False(t, m.AclL3Isolation.IsNull())
	require.Len(t, m.AclL3Isolation.Elements(), 1)

	// Round-trip back via overlay to confirm the nested structure is intact.
	cur := &unifi.SettingGlobalSwitch{}
	require.False(t, m.overlay(ctx, cur).HasError())
	require.Len(t, cur.AclL3Isolation, 1)
	assert.Equal(t, "net-a", cur.AclL3Isolation[0].SourceNetwork)
	assert.Equal(t, []string{"net-b"}, cur.AclL3Isolation[0].DestinationNetworks)
}

// TestUniqueSourceNetworkValidator covers the plan-time uniqueness check.
func TestUniqueSourceNetworkValidator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	l3Type := types.ObjectType{AttrTypes: (&aclL3IsolationModel{}).AttributeTypes()}
	entry := func(src string) attr.Value {
		return types.ObjectValueMust((&aclL3IsolationModel{}).AttributeTypes(), map[string]attr.Value{
			"source_network":       types.StringValue(src),
			"destination_networks": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("net-x")}),
		})
	}

	tests := map[string]struct {
		set     types.Set
		wantErr bool
	}{
		"unique source networks are accepted": {
			set: types.SetValueMust(l3Type, []attr.Value{entry("net-a"), entry("net-b")}),
		},
		"duplicate source networks are rejected": {
			set:     types.SetValueMust(l3Type, []attr.Value{entry("net-a"), entry("net-a")}),
			wantErr: true,
		},
		"null set is accepted": {
			set: types.SetNull(l3Type),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			req := validator.SetRequest{
				Path:        path.Root("acl_l3_isolation"),
				ConfigValue: test.set,
			}
			resp := &validator.SetResponse{}
			uniqueSourceNetworkValidator{}.ValidateSet(ctx, req, resp)
			assert.Equal(t, test.wantErr, resp.Diagnostics.HasError())
		})
	}
}

package device

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
)

// rawEntries decodes controller-shaped JSON, so numbers arrive as float64 the
// way they do from a real GET.
func rawEntries(t *testing.T, s string) []map[string]any {
	t.Helper()
	var out []map[string]any
	require.NoError(t, json.Unmarshal([]byte(s), &out))
	return out
}

func entryByIdx(t *testing.T, entries []map[string]any, idx int) map[string]any {
	t.Helper()
	for _, e := range entries {
		if i, ok := portIdxOf(e); ok && i == idx {
			return e
		}
	}
	t.Fatalf("no entry for port %d in %v", idx, entries)
	return nil
}

// The bug this fixes: settings the resource does not model must survive an
// update to a modeled field on the same port. Includes `false` values, which
// go-unifi's omitempty bools cannot express.
func TestMergePortOverrides_PreservesUnmodeledKeys(t *testing.T) {
	current := rawEntries(t, `[
		{"port_idx": 1, "name": "Uplink", "forward": "all", "native_networkconf_id": "net-default",
		 "autoneg": false, "speed": 10000, "full_duplex": true,
		 "stp_port_mode": true, "stp_edge_state": "auto", "stp_bpdu_guard_enabled": false},
		{"port_idx": 4, "name": "workstation", "forward": "native", "native_networkconf_id": "net-trusted",
		 "tagged_vlan_mgmt": "block_all", "setting_preference": "manual",
		 "stp_edge_state": "enabled", "stp_bpdu_guard_enabled": true, "port_security_enabled": false}
	]`)
	declared := []unifi.DevicePortOverrides{
		{PortIDX: 1, Name: "Uplink to core", Forward: "all", NATiveNetworkID: "net-default"},
		{PortIDX: 4, Name: "workstation (renamed)", Forward: "native", NATiveNetworkID: "net-trusted", TaggedVLANMgmt: "block_all", SettingPreference: "manual"},
	}

	got, err := mergePortOverrides(current, declared)
	require.NoError(t, err)
	require.Len(t, got, 2)

	p1 := entryByIdx(t, got, 1)
	assert.Equal(t, "Uplink to core", p1["name"])
	assert.Equal(t, false, p1["autoneg"], "autoneg=false must survive (omitempty would drop it)")
	assert.Equal(t, float64(10000), p1["speed"])
	assert.Equal(t, true, p1["full_duplex"])
	assert.Equal(t, "auto", p1["stp_edge_state"])

	p4 := entryByIdx(t, got, 4)
	assert.Equal(t, "workstation (renamed)", p4["name"])
	assert.Equal(t, "enabled", p4["stp_edge_state"])
	assert.Equal(t, true, p4["stp_bpdu_guard_enabled"])
	assert.Equal(t, false, p4["port_security_enabled"])
}

// A modeled key removed from the config is removed from the entry, not kept
// from the current state.
func TestMergePortOverrides_ClearedModeledKeyIsRemoved(t *testing.T) {
	current := rawEntries(t, `[{"port_idx": 2, "name": "old name", "poe_mode": "auto", "stp_edge_state": "enabled"}]`)
	got, err := mergePortOverrides(current, []unifi.DevicePortOverrides{{PortIDX: 2}})
	require.NoError(t, err)

	p2 := entryByIdx(t, got, 2)
	assert.NotContains(t, p2, "name")
	assert.NotContains(t, p2, "poe_mode")
	assert.Equal(t, "enabled", p2["stp_edge_state"], "unmodeled key still preserved")
}

// Assigning a port profile starts the entry clean (what the UI stores), so
// stale per-port settings don't sit underneath the profile.
func TestMergePortOverrides_ProfilePortStartsClean(t *testing.T) {
	current := rawEntries(t, `[{"port_idx": 3, "name": "was manual", "stp_edge_state": "enabled",
		"stp_bpdu_guard_enabled": true, "native_networkconf_id": "net-x", "setting_preference": "manual"}]`)
	got, err := mergePortOverrides(current, []unifi.DevicePortOverrides{
		{PortIDX: 3, Name: "server-1", PortProfileID: "profile-server", SettingPreference: "auto"},
	})
	require.NoError(t, err)

	assert.Equal(t, map[string]any{
		"port_idx":           3,
		"name":               "server-1",
		"portconf_id":        "profile-server",
		"setting_preference": "auto",
	}, entryByIdx(t, got, 3))
}

// Declared-set semantics are unchanged: an undeclared port gets no entry, and a
// newly declared port with no current entry is just the overlay.
func TestMergePortOverrides_DeclaredSetSemantics(t *testing.T) {
	current := rawEntries(t, `[
		{"port_idx": 5, "name": "five", "stp_edge_state": "enabled"},
		{"port_idx": 6, "name": "six"}
	]`)
	got, err := mergePortOverrides(current, []unifi.DevicePortOverrides{
		{PortIDX: 5, Name: "five"},
		{PortIDX: 7, Name: "seven"},
	})
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, "enabled", entryByIdx(t, got, 5)["stp_edge_state"])
	assert.Equal(t, map[string]any{"port_idx": 7, "name": "seven"}, entryByIdx(t, got, 7))
	for _, e := range got {
		idx, _ := portIdxOf(e)
		assert.NotEqual(t, 6, idx, "undeclared port 6 must not be carried over")
	}
}

// portOverrideModeledKeys is hand-written, so pin it to what toPortOverride
// actually emits. A field added there but not to the list could never be
// cleared; a key in the list that toPortOverride doesn't set would be wiped
// from the controller on every update.
func TestPortOverrideModeledKeys_MatchToPortOverride(t *testing.T) {
	po, err := toPortOverride(map[string]interface{}{
		"number":                1,
		"name":                  "name",
		"port_profile_id":       "profile",
		"op_mode":               "switch",
		"poe_mode":              "auto",
		"aggregate_num_ports":   2,
		"excluded_network_ids":  schema.NewSet(schema.HashString, []interface{}{"net-excluded"}),
		"native_networkconf_id": "net-native",
		"tagged_vlan_mgmt":      "custom",
		"forward":               "customize",
		"voice_networkconf_id":  "net-voice",
		"setting_preference":    "manual",
	})
	require.NoError(t, err)

	emitted, err := structToMap(po)
	require.NoError(t, err)

	var keys []string
	for k := range emitted {
		keys = append(keys, k)
	}
	assert.ElementsMatch(t, portOverrideModeledKeys, keys)
}

func TestMergePortOverrides_SortedByPortIdx(t *testing.T) {
	got, err := mergePortOverrides(nil, []unifi.DevicePortOverrides{{PortIDX: 8}, {PortIDX: 2}, {PortIDX: 5}})
	require.NoError(t, err)
	var order []int
	for _, e := range got {
		idx, _ := portIdxOf(e)
		order = append(order, idx)
	}
	assert.Equal(t, []int{2, 5, 8}, order)
}

func TestPortIdxOf(t *testing.T) {
	for name, tc := range map[string]struct {
		in   any
		want int
		ok   bool
	}{
		"float64": {float64(3), 3, true},
		"int":     {4, 4, true},
		"string":  {"5", 5, true},
		"number":  {json.Number("6"), 6, true},
		"missing": {nil, 0, false},
		"garbage": {"x", 0, false},
	} {
		t.Run(name, func(t *testing.T) {
			got, ok := portIdxOf(map[string]any{"port_idx": tc.in})
			assert.Equal(t, tc.ok, ok)
			if tc.ok {
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

// fakeDeviceAPI implements only the raw Get/Put the update path uses.
type fakeDeviceAPI struct {
	unifi.Client
	device  map[string]any
	putPath string
	putBody map[string]any
}

func (f *fakeDeviceAPI) Get(_ context.Context, _ string, _ any, respBody any) error {
	return roundTrip(map[string]any{"data": []any{f.device}}, respBody)
}

func (f *fakeDeviceAPI) Put(_ context.Context, apiPath string, reqBody any, respBody any) error {
	f.putPath = apiPath
	f.putBody = map[string]any{}
	if err := roundTrip(reqBody, &f.putBody); err != nil {
		return err
	}
	return roundTrip(map[string]any{"data": []any{f.putBody}}, respBody)
}

func roundTrip(in, out any) error {
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

// End to end through the raw client calls: the PUT body carries the merged
// entries (unmodeled keys intact) plus the device's modeled top-level fields.
func TestUpdateDeviceWithRawPortOverrides_SendsMergedEntries(t *testing.T) {
	api := &fakeDeviceAPI{device: map[string]any{
		"_id": "dev-1",
		"port_overrides": []any{
			map[string]any{"port_idx": 4, "name": "old", "stp_edge_state": "enabled", "stp_bpdu_guard_enabled": true},
		},
	}}
	c := &base.Client{Client: api, Site: "default"}

	current, err := getRawPortOverrides(context.Background(), c, "default", "dev-1")
	require.NoError(t, err)

	req := &unifi.Device{ID: "dev-1", Name: "switch", PortOverrides: []unifi.DevicePortOverrides{{PortIDX: 4, Name: "new"}}}
	merged, err := mergePortOverrides(current, req.PortOverrides)
	require.NoError(t, err)

	resp, err := updateDeviceWithRawPortOverrides(context.Background(), c, "default", req, merged)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "s/default/rest/device/dev-1", api.putPath)
	assert.Equal(t, "switch", api.putBody["name"])
	sent, ok := api.putBody["port_overrides"].([]any)
	require.True(t, ok)
	require.Len(t, sent, 1)
	p4, _ := sent[0].(map[string]any)
	assert.Equal(t, "new", p4["name"])
	assert.Equal(t, "enabled", p4["stp_edge_state"])
	assert.Equal(t, true, p4["stp_bpdu_guard_enabled"])
}

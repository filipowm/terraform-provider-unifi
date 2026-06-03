package device

import (
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func radioSet(items ...map[string]interface{}) *schema.Set {
	raw := make([]interface{}, len(items))
	for i, m := range items {
		raw[i] = m
	}
	return schema.NewSet(radioSetHash, raw)
}

func radioByBand(rs []unifi.DeviceRadioTable, band string) (unifi.DeviceRadioTable, bool) {
	for _, r := range rs {
		if r.Radio == band {
			return r, true
		}
	}
	return unifi.DeviceRadioTable{}, false
}

// The core safety property: declaring one band must not wipe the others.
func TestMergeRadios_PreservesUndeclaredBands(t *testing.T) {
	current := []unifi.DeviceRadioTable{
		{Radio: "ng", Channel: "1", Ht: 20, TxPowerMode: "auto"},
		{Radio: "na", Channel: "36", Ht: 80, TxPowerMode: "high"},
		{Radio: "6e", Channel: "37", Ht: 160, TxPowerMode: "auto"},
	}
	got := mergeRadios(current, radioSet(map[string]interface{}{
		"name": "ng", "tx_power_mode": "disabled",
	}))
	if len(got) != 3 {
		t.Fatalf("expected all 3 bands preserved, got %d: %+v", len(got), got)
	}
	ng, _ := radioByBand(got, "ng")
	if ng.TxPowerMode != "disabled" {
		t.Errorf("ng tx_power_mode = %q, want disabled", ng.TxPowerMode)
	}
	if ng.Channel != "1" || ng.Ht != 20 {
		t.Errorf("ng channel/ht clobbered: got channel=%q ht=%d, want 1/20", ng.Channel, ng.Ht)
	}
	if na, _ := radioByBand(got, "na"); na.Channel != "36" || na.Ht != 80 || na.TxPowerMode != "high" {
		t.Errorf("na band modified: %+v", na)
	}
	if sixE, _ := radioByBand(got, "6e"); sixE.Channel != "37" || sixE.Ht != 160 {
		t.Errorf("6e band modified: %+v", sixE)
	}
}

// Only non-zero declared fields overlay; unset fields keep the controller value.
func TestMergeRadios_OverlaysOnlyNonZero(t *testing.T) {
	current := []unifi.DeviceRadioTable{{Radio: "ng", Channel: "6", Ht: 40, TxPowerMode: "medium"}}
	got := mergeRadios(current, radioSet(map[string]interface{}{
		"name": "ng", "channel": "11",
	}))
	ng, _ := radioByBand(got, "ng")
	if ng.Channel != "11" {
		t.Errorf("channel = %q, want 11", ng.Channel)
	}
	if ng.Ht != 40 || ng.TxPowerMode != "medium" {
		t.Errorf("unset fields clobbered: ht=%d tx_power_mode=%q, want 40/medium", ng.Ht, ng.TxPowerMode)
	}
}

// A declared band missing from the device is appended.
func TestMergeRadios_AppendsMissingBand(t *testing.T) {
	current := []unifi.DeviceRadioTable{{Radio: "na", Channel: "36"}}
	got := mergeRadios(current, radioSet(map[string]interface{}{
		"name": "ng", "tx_power_mode": "disabled",
	}))
	if len(got) != 2 {
		t.Fatalf("expected 2 bands, got %d", len(got))
	}
	if ng, ok := radioByBand(got, "ng"); !ok || ng.TxPowerMode != "disabled" {
		t.Errorf("ng not appended correctly: %+v ok=%v", ng, ok)
	}
}

// min_rssi pairs with min_rssi_enabled only when a non-zero threshold is set.
func TestMergeRadios_MinRssiPairing(t *testing.T) {
	current := []unifi.DeviceRadioTable{{Radio: "na"}}
	got := mergeRadios(current, radioSet(map[string]interface{}{
		"name": "na", "min_rssi": -75, "min_rssi_enabled": true,
	}))
	na, _ := radioByBand(got, "na")
	if na.MinRssi != -75 || !na.MinRssiEnabled {
		t.Errorf("min_rssi pairing failed: %+v", na)
	}
}

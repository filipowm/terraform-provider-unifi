package network

import "testing"

// TestValidateDefaultGatewayCombo covers the cross-field rule enforced by the
// dhcpd_gateway / dhcpd_gateway_enabled CustomizeDiff: the override IP and its
// enable toggle must be configured together.
func TestValidateDefaultGatewayCombo(t *testing.T) {
	tests := map[string]struct {
		gatewaySet bool
		enabled    bool
		wantErr    bool
	}{
		"neither set (auto default gateway)": {gatewaySet: false, enabled: false, wantErr: false},
		"both set (manual override)":         {gatewaySet: true, enabled: true, wantErr: false},
		"gateway set but disabled":           {gatewaySet: true, enabled: false, wantErr: true},
		"enabled but no gateway":             {gatewaySet: false, enabled: true, wantErr: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := validateDefaultGatewayCombo(tc.gatewaySet, tc.enabled)
			if tc.wantErr && err == nil {
				t.Fatalf("validateDefaultGatewayCombo(%t, %t) = nil, want error", tc.gatewaySet, tc.enabled)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("validateDefaultGatewayCombo(%t, %t) = %v, want nil", tc.gatewaySet, tc.enabled, err)
			}
		})
	}
}

package utils

import "testing"

func TestWireguardKeyValidate(t *testing.T) {
	cases := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{"valid 32-byte key", "0WvUlUyZZ0yTUibNCAdBrQ6XJd+8V37zmk/j8y/V9g4=", false},
		{"not base64", "not valid base64!!!", true},
		{"valid base64 but wrong length", "c2hvcnQ=", true}, // "short" -> 5 bytes
		{"empty string", "", true},
		{"not a string", 123, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, errs := WireguardKeyValidate(tc.input, "k"); (len(errs) > 0) != tc.wantErr {
				t.Errorf("WireguardKeyValidate(%v): got errs=%v, wantErr=%v", tc.input, errs, tc.wantErr)
			}
		})
	}
}

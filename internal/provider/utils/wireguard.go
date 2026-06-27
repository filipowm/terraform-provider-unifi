package utils

import (
	"encoding/base64"
	"fmt"
)

// WireguardKeyValidate ensures a value is a base64-encoded 32-byte WireGuard key
// (a private, public, or preshared key), catching typos at plan time instead of
// letting them slip through to the controller.
func WireguardKeyValidate(raw interface{}, key string) ([]string, []error) {
	v, ok := raw.(string)
	if !ok {
		return nil, []error{fmt.Errorf("%s: expected string, got %T", key, raw)}
	}
	b, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return nil, []error{fmt.Errorf("%s must be a base64-encoded WireGuard key: %w", key, err)}
	}
	if len(b) != 32 {
		return nil, []error{fmt.Errorf("%s must decode to 32 bytes, got %d", key, len(b))}
	}
	return nil, nil
}

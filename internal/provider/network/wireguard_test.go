package network

import (
	"encoding/base64"
	"testing"
)

// TestWireguardPublicKey verifies the Curve25519 derivation against a reference
// (private, public) pair computed independently with Python's `cryptography`
// X25519 implementation (the same value `wg pubkey` would produce).
func TestWireguardPublicKey(t *testing.T) {
	const priv = "80dbJvq14felsz3JzyP4u/hELE3FOX+wHhm2KxsCr8E="
	const wantPub = "0WvUlUyZZ0yTUibNCAdBrQ6XJd+8V37zmk/j8y/V9g4="

	got, err := wireguardPublicKey(priv)
	if err != nil {
		t.Fatalf("wireguardPublicKey() unexpected error: %v", err)
	}
	if got != wantPub {
		t.Errorf("wireguardPublicKey() = %q, want %q", got, wantPub)
	}
}

func TestWireguardPublicKey_errors(t *testing.T) {
	if _, err := wireguardPublicKey("not valid base64!!!"); err == nil {
		t.Error("expected an error for invalid base64 input")
	}
	if _, err := wireguardPublicKey(base64.StdEncoding.EncodeToString([]byte("short"))); err == nil {
		t.Error("expected an error for a key that is not 32 bytes")
	}
}

// TestGenerateWireguardPrivateKey checks the generated key is a 32-byte,
// Curve25519-clamped key from which a public key can be derived.
func TestGenerateWireguardPrivateKey(t *testing.T) {
	k, err := generateWireguardPrivateKey()
	if err != nil {
		t.Fatalf("generateWireguardPrivateKey() unexpected error: %v", err)
	}
	raw, err := base64.StdEncoding.DecodeString(k)
	if err != nil {
		t.Fatalf("generated key is not valid base64: %v", err)
	}
	if len(raw) != 32 {
		t.Fatalf("generated key = %d bytes, want 32", len(raw))
	}
	// Clamping per RFC 7748.
	if raw[0]&0b111 != 0 || raw[31]&0b1000_0000 != 0 || raw[31]&0b0100_0000 == 0 {
		t.Error("generated key is not Curve25519-clamped")
	}
	if _, err := wireguardPublicKey(k); err != nil {
		t.Errorf("could not derive a public key from the generated private key: %v", err)
	}
}

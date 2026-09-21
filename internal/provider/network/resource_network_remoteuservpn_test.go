package network

import (
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
)

// serverRaw builds a raw config for a remote-user-vpn network. Attributes are
// passed as a map so a case can omit one entirely -- which is the distinction
// these rules turn on, and is not the same as passing a null.
func serverRaw(attrs map[string]cty.Value) cty.Value {
	base := map[string]cty.Value{
		"purpose":            cty.StringVal("remote-user-vpn"),
		"vpn_type":           cty.StringVal("wireguard-server"),
		"subnet":             cty.StringVal("192.168.3.1/24"),
		"remote_vpn_subnets": cty.ListVal([]cty.Value{cty.StringVal("192.168.3.0/24")}),
	}
	for k, v := range attrs {
		if v == cty.NilVal {
			delete(base, k)
			continue
		}
		base[k] = v
	}
	return cty.ObjectVal(base)
}

func wantErr(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected an error containing %q, got nil", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("expected an error containing %q, got: %v", want, err)
	}
}

func wantOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected the config to validate, got: %v", err)
	}
}

func TestValidateRemoteUserVPNRawConfig_valid(t *testing.T) {
	wantOK(t, validateRemoteUserVPNRawConfig(serverRaw(nil)))
}

// A null raw config is what a destroy plan supplies; it must not error.
func TestValidateRemoteUserVPNRawConfig_nullConfig(t *testing.T) {
	wantOK(t, validateRemoteUserVPNRawConfig(cty.NullVal(cty.EmptyObject)))
}

func TestValidateRemoteUserVPNRawConfig_requiresVPNType(t *testing.T) {
	wantErr(t, validateRemoteUserVPNRawConfig(serverRaw(map[string]cty.Value{"vpn_type": cty.NilVal})),
		`"vpn_type" is required when purpose = "remote-user-vpn"`)
}

func TestValidateRemoteUserVPNRawConfig_requiresSubnet(t *testing.T) {
	wantErr(t, validateRemoteUserVPNRawConfig(serverRaw(map[string]cty.Value{"subnet": cty.NilVal})),
		`"subnet" (the VPN Server's own tunnel address) is required`)
}

// Either an explicit pool or controller allocation, never neither: with both
// absent the server has no addresses to hand out.
func TestValidateRemoteUserVPNRawConfig_poolOrDynamic(t *testing.T) {
	wantErr(t, validateRemoteUserVPNRawConfig(serverRaw(map[string]cty.Value{"remote_vpn_subnets": cty.NilVal})),
		`"remote_vpn_subnets" is required`)

	wantOK(t, validateRemoteUserVPNRawConfig(serverRaw(map[string]cty.Value{
		"remote_vpn_subnets":                 cty.NilVal,
		"remote_vpn_dynamic_subnets_enabled": cty.True,
	})))

	// Explicitly false is the same as absent for this rule.
	wantErr(t, validateRemoteUserVPNRawConfig(serverRaw(map[string]cty.Value{
		"remote_vpn_subnets":                 cty.NilVal,
		"remote_vpn_dynamic_subnets_enabled": cty.False,
	})), `"remote_vpn_subnets" is required`)
}

// An interpolated value is unknown at plan time but is still "set", so it must
// satisfy the requirement rather than trip a false "required" error.
func TestValidateRemoteUserVPNRawConfig_unknownCountsAsSet(t *testing.T) {
	wantOK(t, validateRemoteUserVPNRawConfig(serverRaw(map[string]cty.Value{
		"remote_vpn_subnets": cty.UnknownVal(cty.List(cty.String)),
	})))
}

// The server-only fields must not leak onto other purposes.
func TestValidateRemoteUserVPNRawConfig_rejectsServerFieldsElsewhere(t *testing.T) {
	for _, field := range []string{
		"remote_vpn_subnets", "remote_vpn_dynamic_subnets_enabled",
		"uid_vpn_type", "uid_vpn_masquerade_enabled",
		"uid_vpn_max_connection_time_seconds", "uid_vpn_default_dns_suffix",
		"uid_vpn_sync_public_ip",
	} {
		attrs := map[string]cty.Value{
			"purpose": cty.StringVal("corporate"),
			"subnet":  cty.StringVal("10.10.1.1/24"),
		}
		switch field {
		case "remote_vpn_subnets":
			attrs[field] = cty.ListVal([]cty.Value{cty.StringVal("10.0.0.0/24")})
		case "uid_vpn_type", "uid_vpn_default_dns_suffix":
			attrs[field] = cty.StringVal("x")
		case "uid_vpn_max_connection_time_seconds":
			attrs[field] = cty.NumberIntVal(3600)
		default:
			attrs[field] = cty.True
		}
		err := validateRemoteUserVPNRawConfig(cty.ObjectVal(attrs))
		wantErr(t, err, `"`+field+`" is only valid when purpose = "remote-user-vpn"`)
	}
}

// vpn_type is shared between the two VPN purposes, so each value stays pinned to
// its own purpose rather than being accepted on both.
func TestValidateRemoteUserVPNRawConfig_vpnTypePinnedToPurpose(t *testing.T) {
	wantErr(t, validateRemoteUserVPNRawConfig(serverRaw(map[string]cty.Value{
		"vpn_type": cty.StringVal("wireguard-client"),
	})), `vpn_type "wireguard-client" is only valid when purpose = "vpn-client"`)

	wantErr(t, validateRemoteUserVPNRawConfig(cty.ObjectVal(map[string]cty.Value{
		"purpose":  cty.StringVal("corporate"),
		"vpn_type": cty.StringVal("wireguard-server"),
	})), `vpn_type "wireguard-server" is only valid when purpose = "remote-user-vpn"`)

	// The vpn-client purpose keeps its own type.
	wantOK(t, validateRemoteUserVPNRawConfig(cty.ObjectVal(map[string]cty.Value{
		"purpose":  cty.StringVal("vpn-client"),
		"vpn_type": cty.StringVal("wireguard-client"),
	})))
}

// Enum coverage: the new values must be accepted and the still-unmodelled ones
// must not be.
func TestNetworkEnums_pass2(t *testing.T) {
	purpose := ResourceNetwork().Schema["purpose"]
	if _, errs := purpose.ValidateFunc("remote-user-vpn", "purpose"); len(errs) != 0 {
		t.Errorf("purpose \"remote-user-vpn\" should be accepted, got: %v", errs)
	}
	if _, errs := purpose.ValidateFunc("site-vpn", "purpose"); len(errs) == 0 {
		t.Error("purpose \"site-vpn\" is not modelled and should still be rejected")
	}

	vpnType := ResourceNetwork().Schema["vpn_type"]
	for _, ok := range []string{"wireguard-server", "wireguard-client"} {
		if _, errs := vpnType.ValidateFunc(ok, "vpn_type"); len(errs) != 0 {
			t.Errorf("vpn_type %q should be accepted, got: %v", ok, errs)
		}
	}
	if _, errs := vpnType.ValidateFunc("openvpn-server", "vpn_type"); len(errs) == 0 {
		t.Error("vpn_type \"openvpn-server\" is not modelled and should be rejected")
	}
}

// Regression: removing vpn_type from customizeNetworkVPNClient's exclusivity list
// left "wireguard-client" accepted on every purpose that is not remote-user-vpn.
// Neither value belongs on a non-VPN network.
func TestValidateRemoteUserVPNRawConfig_rejectsVPNTypeOnNonVPNPurpose(t *testing.T) {
	for _, vpnType := range []string{"wireguard-client", "wireguard-server"} {
		err := validateRemoteUserVPNRawConfig(cty.ObjectVal(map[string]cty.Value{
			"purpose":  cty.StringVal("corporate"),
			"subnet":   cty.StringVal("10.10.1.1/24"),
			"vpn_type": cty.StringVal(vpnType),
		}))
		wantErr(t, err, `is only valid when purpose`)
	}
}

// An interpolated vpn_type is set, just not knowable at plan time. The schema enum
// still constrains it at apply, so the rule must defer rather than read it as "".
func TestValidateRemoteUserVPNRawConfig_unknownVPNType(t *testing.T) {
	wantOK(t, validateRemoteUserVPNRawConfig(serverRaw(map[string]cty.Value{
		"vpn_type": cty.UnknownVal(cty.String),
	})))
}

// Same for the dynamic-pool toggle: unknown is not false, and reading it as false
// fails a config whose pool is legitimately controller-allocated.
func TestValidateRemoteUserVPNRawConfig_unknownDynamicSubnets(t *testing.T) {
	wantOK(t, validateRemoteUserVPNRawConfig(serverRaw(map[string]cty.Value{
		"remote_vpn_subnets":                 cty.NilVal,
		"remote_vpn_dynamic_subnets_enabled": cty.UnknownVal(cty.Bool),
	})))
}

// A module that passes every optional argument from a variable with a `false`
// default must not be rejected: an explicit `false` asks for nothing the purpose
// cannot give, so it does not count as "configured for another purpose".
func TestValidateRemoteUserVPNRawConfig_explicitFalseIsNotAServerField(t *testing.T) {
	for _, field := range []string{
		"remote_vpn_dynamic_subnets_enabled", "uid_vpn_masquerade_enabled", "uid_vpn_sync_public_ip",
	} {
		wantOK(t, validateRemoteUserVPNRawConfig(cty.ObjectVal(map[string]cty.Value{
			"purpose": cty.StringVal("corporate"),
			"subnet":  cty.StringVal("10.10.1.1/24"),
			field:     cty.False,
		})))
	}
}

// Only the WireGuard server is modelled, so the two "which server" fields must
// agree; uid_vpn_type = "openvpn" alongside vpn_type = "wireguard-server"
// describes an object this resource cannot produce.
func TestValidateRemoteUserVPNRawConfig_uidVPNTypeMustMatch(t *testing.T) {
	wantErr(t, validateRemoteUserVPNRawConfig(serverRaw(map[string]cty.Value{
		"uid_vpn_type": cty.StringVal("openvpn"),
	})), `"uid_vpn_type" must be "wireguard"`)

	wantOK(t, validateRemoteUserVPNRawConfig(serverRaw(map[string]cty.Value{
		"uid_vpn_type": cty.StringVal("wireguard"),
	})))
}

// A VPN Server needs its own keypair, and the fields that carry one must not be
// Optional-without-Computed dead ends: uid_vpn_max_connection_time_seconds and
// uid_vpn_default_dns_suffix both carry `omitempty` in go-unifi, so without
// Computed a removed attribute never reaches the controller and the diff never
// converges.
func TestNetworkSchema_uidVPNOptionalsAreComputed(t *testing.T) {
	s := ResourceNetwork().Schema
	for _, name := range []string{"uid_vpn_max_connection_time_seconds", "uid_vpn_default_dns_suffix"} {
		if !s[name].Computed {
			t.Errorf("%q must be Computed: its go-unifi field is omitempty, so an unset value is dropped from the payload and the controller's old value is read back forever", name)
		}
	}
}

// A wireguard-server network had no way to obtain a private key: the mint was
// keyed on "wireguard-client" only, while x_wireguard_private_key was rejected
// outright on any purpose but vpn-client. Either way the create carried no key.
func TestNeedsGeneratedWireguardKey(t *testing.T) {
	for _, tc := range []struct {
		vpnType, key string
		want         bool
	}{
		{"wireguard-server", "", true},
		{"wireguard-client", "", true},
		{"wireguard-server", "existing", false},
		{"wireguard-client", "existing", false},
		{"", "", false},
	} {
		if got := needsGeneratedWireguardKey(tc.vpnType, tc.key); got != tc.want {
			t.Errorf("needsGeneratedWireguardKey(%q, %q) = %v, want %v", tc.vpnType, tc.key, got, tc.want)
		}
	}
}

// The gateway's own keypair and egress WAN belong to both WireGuard purposes, and
// to neither of the others.
func TestValidateWireguardFieldPurposes(t *testing.T) {
	raw := func(field string) cty.Value {
		return cty.ObjectVal(map[string]cty.Value{field: cty.StringVal("wan")})
	}
	for _, field := range []string{"wireguard_interface", "x_wireguard_private_key"} {
		wantOK(t, validateWireguardFieldPurposes(raw(field), "vpn-client"))
		wantOK(t, validateWireguardFieldPurposes(raw(field), "remote-user-vpn"))
		wantErr(t, validateWireguardFieldPurposes(raw(field), "corporate"),
			`"`+field+`" is only valid when purpose = "vpn-client" or "remote-user-vpn"`)
	}
}

// Every rule in the validator is keyed on purpose, so an interpolated purpose
// makes none of them decidable — including the widened vpn_type exclusivity.
func TestValidateRemoteUserVPNRawConfig_unknownPurpose(t *testing.T) {
	wantOK(t, validateRemoteUserVPNRawConfig(cty.ObjectVal(map[string]cty.Value{
		"purpose":  cty.UnknownVal(cty.String),
		"vpn_type": cty.StringVal("wireguard-server"),
		"subnet":   cty.StringVal("192.168.3.1/24"),
	})))
}

// clientRaw builds a raw config for a vpn-client network, complete enough that
// every "required" rule is satisfied, so a case can isolate one attribute.
func clientRaw(attrs map[string]cty.Value) cty.Value {
	base := map[string]cty.Value{
		"purpose":                          cty.StringVal("vpn-client"),
		"vpn_type":                         cty.StringVal("wireguard-client"),
		"subnet":                           cty.StringVal("10.0.0.2/32"),
		"dhcp_dns":                         cty.ListVal([]cty.Value{cty.StringVal("1.1.1.1")}),
		"wireguard_client_peer_ip":         cty.StringVal("198.51.100.1"),
		"wireguard_client_peer_public_key": cty.StringVal("Zm9vYmFyZm9vYmFyZm9vYmFyZm9vYmFyZm9vYmFyZm8="),
		"wireguard_client_peer_port":       cty.NumberIntVal(51820),
	}
	for k, v := range attrs {
		if v == cty.NilVal {
			delete(base, k)
			continue
		}
		base[k] = v
	}
	return cty.ObjectVal(base)
}

func TestValidateVPNClientRawConfig_valid(t *testing.T) {
	wantOK(t, validateVPNClientRawConfig(clientRaw(nil)))
}

func TestValidateVPNClientRawConfig_nullConfig(t *testing.T) {
	wantOK(t, validateVPNClientRawConfig(cty.NullVal(cty.EmptyObject)))
}

// The pre-existing vpn-client rules must keep firing.
func TestValidateVPNClientRawConfig_requiredFields(t *testing.T) {
	wantErr(t, validateVPNClientRawConfig(clientRaw(map[string]cty.Value{"subnet": cty.NilVal})),
		`"subnet" (the tunnel interface address`)
	wantErr(t, validateVPNClientRawConfig(clientRaw(map[string]cty.Value{"dhcp_dns": cty.NilVal})),
		`"dhcp_dns" (interface DNS) is required`)
	wantErr(t, validateVPNClientRawConfig(clientRaw(map[string]cty.Value{"wireguard_client_peer_ip": cty.NilVal})),
		`"wireguard_client_peer_ip" is required`)
	wantErr(t, validateVPNClientRawConfig(clientRaw(map[string]cty.Value{"subnet": cty.StringVal("10.0.0.0/24")})),
		`must be a /32 tunnel interface address`)
	wantErr(t, validateVPNClientRawConfig(cty.ObjectVal(map[string]cty.Value{
		"purpose":                  cty.StringVal("corporate"),
		"wireguard_client_peer_ip": cty.StringVal("198.51.100.1"),
	})), `"wireguard_client_peer_ip" is only valid when purpose = "vpn-client"`)
}

// Every rule is keyed on purpose, so an interpolated purpose makes none of them
// decidable. d.Get read an unknown purpose as "", which reads as "not
// vpn-client" and rejected the very fields the config legitimately carries.
func TestValidateVPNClientRawConfig_unknownPurpose(t *testing.T) {
	wantOK(t, validateVPNClientRawConfig(cty.ObjectVal(map[string]cty.Value{
		"purpose":                 cty.UnknownVal(cty.String),
		"wireguard_interface":     cty.StringVal("wan"),
		"x_wireguard_private_key": cty.StringVal("Zm9vYmFyZm9vYmFyZm9vYmFyZm9vYmFyZm9vYmFyZm8="),
	})))
	// ...including the vpn-client-only fields, which behaved the same way before.
	wantOK(t, validateVPNClientRawConfig(cty.ObjectVal(map[string]cty.Value{
		"purpose":                  cty.UnknownVal(cty.String),
		"wireguard_client_peer_ip": cty.StringVal("198.51.100.1"),
	})))
}

// An interpolated vpn_type is set but not comparable at plan time.
func TestValidateVPNClientRawConfig_unknownVPNType(t *testing.T) {
	wantOK(t, validateVPNClientRawConfig(clientRaw(map[string]cty.Value{
		"vpn_type": cty.UnknownVal(cty.String),
	})))
	// An interpolated subnet cannot be parsed for the /32 rule either.
	wantOK(t, validateVPNClientRawConfig(clientRaw(map[string]cty.Value{
		"subnet": cty.UnknownVal(cty.String),
	})))
}

func TestRawKnownString(t *testing.T) {
	raw := cty.ObjectVal(map[string]cty.Value{
		"set":     cty.StringVal("v"),
		"empty":   cty.StringVal(""),
		"null":    cty.NullVal(cty.String),
		"unknown": cty.UnknownVal(cty.String),
		"notStr":  cty.True,
	})
	for _, tc := range []struct {
		name, want string
		wantKnown  bool
	}{
		{"set", "v", true},
		{"empty", "", true},
		{"null", "", false},
		{"unknown", "", false},
		{"notStr", "", false},
		{"missing", "", false},
	} {
		got, known := rawKnownString(raw, tc.name)
		if got != tc.want || known != tc.wantKnown {
			t.Errorf("rawKnownString(%q) = (%q, %v), want (%q, %v)", tc.name, got, known, tc.want, tc.wantKnown)
		}
	}
	if _, known := rawKnownString(cty.NullVal(cty.EmptyObject), "x"); known {
		t.Error("a null raw config has no readable attributes")
	}
}

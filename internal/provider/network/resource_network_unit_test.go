package network

import (
	"context"
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
)

// fakeNetworkClient is a minimal unifi.Client used to drive resourceNetworkUpdate through the
// re-read-on-ErrNotFound path introduced for issue #98. It embeds the unifi.Client interface so
// every method we do not override panics if unexpectedly called.
type fakeNetworkClient struct {
	unifi.Client

	updateResp *unifi.Network
	updateErr  error

	getResp *unifi.Network
	getErr  error

	getCalled bool
	getID     string
}

func (f *fakeNetworkClient) UpdateNetwork(_ context.Context, _ string, _ *unifi.Network) (*unifi.Network, error) {
	return f.updateResp, f.updateErr
}

func (f *fakeNetworkClient) GetNetwork(_ context.Context, _ string, id string) (*unifi.Network, error) {
	f.getCalled = true
	f.getID = id
	return f.getResp, f.getErr
}

// TestResourceNetworkUpdate_ReReadOnNotFound covers the issue #98 regression: go-unifi v1.9.3
// turns a successful-but-empty PUT into unifi.ErrNotFound, so the Update handler must re-read
// instead of surfacing the spurious error.
func TestResourceNetworkUpdate_ReReadOnNotFound(t *testing.T) {
	t.Run("update returns ErrNotFound but object still exists - re-read succeeds", func(t *testing.T) {
		fake := &fakeNetworkClient{
			updateResp: nil,
			updateErr:  unifi.ErrNotFound,
			getResp:    &unifi.Network{ID: "net1", Name: "lan-renamed", Purpose: "corporate"},
			getErr:     nil,
		}
		client := &base.Client{Client: fake, Site: "default"}

		d := ResourceNetwork().TestResourceData()
		d.SetId("net1")
		require.NoError(t, d.Set("site", "default"))
		require.NoError(t, d.Set("name", "lan"))
		require.NoError(t, d.Set("purpose", "corporate"))

		diags := resourceNetworkUpdate(context.Background(), d, client)

		assert.False(t, diags.HasError(), "spurious ErrNotFound from Update must not surface: %v", diags)
		assert.True(t, fake.getCalled, "Update must re-read via GetNetwork on ErrNotFound")
		assert.Equal(t, "net1", fake.getID, "re-read must use the resource ID")
		assert.Equal(t, "net1", d.Id(), "ID must be retained when the object still exists")
		assert.Equal(t, "lan-renamed", d.Get("name"), "state must be repopulated from the re-read object, not the preloaded state")
	})

	t.Run("update and re-read both return ErrNotFound - genuine deletion clears state", func(t *testing.T) {
		fake := &fakeNetworkClient{
			updateResp: nil,
			updateErr:  unifi.ErrNotFound,
			getResp:    nil,
			getErr:     unifi.ErrNotFound,
		}
		client := &base.Client{Client: fake, Site: "default"}

		d := ResourceNetwork().TestResourceData()
		d.SetId("net1")
		require.NoError(t, d.Set("site", "default"))
		require.NoError(t, d.Set("name", "lan"))
		require.NoError(t, d.Set("purpose", "corporate"))

		diags := resourceNetworkUpdate(context.Background(), d, client)

		assert.False(t, diags.HasError(), "genuine deletion must clear state, not error: %v", diags)
		assert.True(t, fake.getCalled, "Update must re-read via GetNetwork on ErrNotFound")
		assert.Equal(t, "", d.Id(), "genuinely deleted network must clear the ID so it is recreated")
	})
}

// TestResourceNetwork_ipv6FieldsOptionalComputed is the offline guard for issue #96.
// The IPv6/DHCPv6 value fields whose go-unifi struct tags carry `,omitempty` must be
// Optional+Computed so a sparse post-import config inherits the controller's value
// instead of planning it to null (a diff the controller can never apply, producing a
// perpetual post-import diff). This Docker-free test catches accidental removal of the
// `Computed:` flag even when the acceptance coverage is skipped.
func TestResourceNetwork_ipv6FieldsOptionalComputed(t *testing.T) {
	s := ResourceNetwork().Schema

	// These five reported fields plus the two same-class siblings carry `,omitempty`
	// in the go-unifi Network struct, so they must round-trip via Optional+Computed.
	computed := []string{
		"dhcp_v6_start",
		"dhcp_v6_stop",
		"ipv6_pd_start",
		"ipv6_pd_stop",
		"ipv6_ra_priority",
		"ipv6_static_subnet",
		"ipv6_pd_interface",
	}
	for _, name := range computed {
		attr, ok := s[name]
		assert.True(t, ok, "schema must define %q", name)
		if !ok {
			continue
		}
		assert.True(t, attr.Optional, "%q must remain Optional", name)
		assert.True(t, attr.Computed, "%q must be Computed so a sparse post-import config inherits the controller value (issue #96)", name)
		// Computed + Default is rejected by SDKv2 at init; these must stay Default-free.
		assert.Nil(t, attr.Default, "%q must not declare a Default (illegal with Computed)", name)
	}

	// ipv6_pd_prefixid is deliberately EXCLUDED: its go-unifi tag has no `,omitempty`,
	// so an empty config value IS sent and clears it. Making it Computed would silently
	// remove that working clear-by-omit behavior — a real BC break. Guard against it.
	if attr, ok := s["ipv6_pd_prefixid"]; ok {
		assert.False(t, attr.Computed, "ipv6_pd_prefixid must stay Optional-only: it has no go-unifi omitempty tag, so making it Computed would break clear-by-omit (issue #96)")
	}
}

// TestValidateIpV6InterfaceType is the issue #99 regression guard: `single_network` must be
// accepted by the plan-time validator (previously the allow-list was only none|pd|static, so the
// value was rejected at terraform plan/validate before any controller call). The invalid set is
// coupled to the anchored regexp form (`^(none|pd|static|single_network)$`): "xstaticy" only fails
// because of anchoring; if the validator is ever loosened to an unanchored alternation, drop it.
func TestValidateIpV6InterfaceType(t *testing.T) {
	for _, v := range []string{"none", "static", "pd", "single_network"} { // single_network = #99 guard
		_, errs := validateIPV6InterfaceType(v, "ipv6_interface_type")
		assert.Emptyf(t, errs, "%q must be accepted", v)
	}
	for _, v := range []string{"", "bogus", "xstaticy", "PD", "single"} { // anchored form
		_, errs := validateIPV6InterfaceType(v, "ipv6_interface_type")
		assert.NotEmptyf(t, errs, "%q must be rejected", v)
	}
}

// TestResourceNetworkIpV6InterfaceTypeValidatorWired guards against a future detachment of the
// validator from the schema attribute (the #99 fix is only effective while it stays wired).
func TestResourceNetworkIpV6InterfaceTypeValidatorWired(t *testing.T) {
	attr := ResourceNetwork().Schema["ipv6_interface_type"]
	assert.NotNil(t, attr, "ipv6_interface_type attribute must exist")
	assert.NotNil(t, attr.ValidateFunc, "ipv6_interface_type must keep a ValidateFunc wired")
}

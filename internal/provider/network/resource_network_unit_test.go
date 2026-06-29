package network

import (
	"context"
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/stretchr/testify/assert"
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

// TestResourceNetworkUpdate_ReReadOnNotFound covers the issue #98 regression: go-unifi v1.9.2
// turns a successful-but-empty PUT into unifi.ErrNotFound, so the Update handler must re-read
// instead of surfacing the spurious error.
func TestResourceNetworkUpdate_ReReadOnNotFound(t *testing.T) {
	t.Run("update returns ErrNotFound but object still exists - re-read succeeds", func(t *testing.T) {
		fake := &fakeNetworkClient{
			updateResp: nil,
			updateErr:  unifi.ErrNotFound,
			getResp:    &unifi.Network{ID: "net1", Name: "lan", Purpose: "corporate"},
			getErr:     nil,
		}
		client := &base.Client{Client: fake, Site: "default"}

		d := ResourceNetwork().TestResourceData()
		d.SetId("net1")
		d.Set("site", "default")
		d.Set("name", "lan")
		d.Set("purpose", "corporate")

		diags := resourceNetworkUpdate(context.Background(), d, client)

		assert.False(t, diags.HasError(), "spurious ErrNotFound from Update must not surface: %v", diags)
		assert.True(t, fake.getCalled, "Update must re-read via GetNetwork on ErrNotFound")
		assert.Equal(t, "net1", fake.getID, "re-read must use the resource ID")
		assert.Equal(t, "net1", d.Id(), "ID must be retained when the object still exists")
		assert.Equal(t, "lan", d.Get("name"), "state must be populated from the re-read object")
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
		d.Set("site", "default")
		d.Set("name", "lan")
		d.Set("purpose", "corporate")

		diags := resourceNetworkUpdate(context.Background(), d, client)

		assert.False(t, diags.HasError(), "genuine deletion must clear state, not error: %v", diags)
		assert.True(t, fake.getCalled, "Update must re-read via GetNetwork on ErrNotFound")
		assert.Equal(t, "", d.Id(), "genuinely deleted network must clear the ID so it is recreated")
	})
}

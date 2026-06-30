package device

import (
	"context"
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/stretchr/testify/assert"
)

// fakePortProfileClient is a minimal unifi.Client used to drive resourcePortProfileUpdate
// through the re-read-on-ErrNotFound path introduced for issue #98. It embeds the unifi.Client
// interface so every method we do not override panics if unexpectedly called.
type fakePortProfileClient struct {
	unifi.Client

	updateResp *unifi.PortProfile
	updateErr  error

	getResp *unifi.PortProfile
	getErr  error

	getCalled bool
	getID     string
}

func (f *fakePortProfileClient) UpdatePortProfile(_ context.Context, _ string, _ *unifi.PortProfile) (*unifi.PortProfile, error) {
	return f.updateResp, f.updateErr
}

func (f *fakePortProfileClient) GetPortProfile(_ context.Context, _ string, id string) (*unifi.PortProfile, error) {
	f.getCalled = true
	f.getID = id
	return f.getResp, f.getErr
}

// TestResourcePortProfileUpdate_ReReadOnNotFound covers the issue #98 regression: go-unifi
// v1.9.2 turns a successful-but-empty PUT into unifi.ErrNotFound, so the Update handler must
// re-read instead of surfacing the spurious error.
func TestResourcePortProfileUpdate_ReReadOnNotFound(t *testing.T) {
	t.Run("update returns ErrNotFound but object still exists - re-read succeeds", func(t *testing.T) {
		fake := &fakePortProfileClient{
			updateResp: nil,
			updateErr:  unifi.ErrNotFound,
			getResp:    &unifi.PortProfile{ID: "pp1", Name: "office", Forward: "customize"},
			getErr:     nil,
		}
		client := &base.Client{Client: fake, Site: "default"}

		d := ResourcePortProfile().TestResourceData()
		d.SetId("pp1")
		d.Set("site", "default")
		d.Set("name", "office")

		diags := resourcePortProfileUpdate(context.Background(), d, client)

		assert.False(t, diags.HasError(), "spurious ErrNotFound from Update must not surface: %v", diags)
		assert.True(t, fake.getCalled, "Update must re-read via GetPortProfile on ErrNotFound")
		assert.Equal(t, "pp1", fake.getID, "re-read must use the resource ID")
		assert.Equal(t, "pp1", d.Id(), "ID must be retained when the object still exists")
		assert.Equal(t, "office", d.Get("name"), "state must be populated from the re-read object")
		assert.Equal(t, "customize", d.Get("forward"), "controller normalization must be reflected in state")
	})

	t.Run("update and re-read both return ErrNotFound - genuine deletion clears state", func(t *testing.T) {
		fake := &fakePortProfileClient{
			updateResp: nil,
			updateErr:  unifi.ErrNotFound,
			getResp:    nil,
			getErr:     unifi.ErrNotFound,
		}
		client := &base.Client{Client: fake, Site: "default"}

		d := ResourcePortProfile().TestResourceData()
		d.SetId("pp1")
		d.Set("site", "default")
		d.Set("name", "office")

		diags := resourcePortProfileUpdate(context.Background(), d, client)

		assert.False(t, diags.HasError(), "genuine deletion must clear state, not error: %v", diags)
		assert.True(t, fake.getCalled, "Update must re-read via GetPortProfile on ErrNotFound")
		assert.Equal(t, "", d.Id(), "genuinely deleted profile must clear the ID so it is recreated")
	})
}

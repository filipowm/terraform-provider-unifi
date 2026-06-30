package utils

import (
	"errors"
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/stretchr/testify/assert"
)

// TestReReadOnUpdateNotFound covers the issue #98 workaround shared by every SDKv2
// Update handler: go-unifi v1.9.2 turns a successful-but-empty PUT into
// unifi.ErrNotFound, so the helper must re-read instead of surfacing the spurious
// error, while still distinguishing a genuine out-of-band deletion and propagating
// real errors.
func TestReReadOnUpdateNotFound(t *testing.T) {
	type obj struct{ name string }

	t.Run("update succeeds - returns update result, no re-read", func(t *testing.T) {
		updated := &obj{name: "from-update"}
		reReadCalled := false
		got, found, err := ReReadOnUpdateNotFound(updated, nil, func() (*obj, error) {
			reReadCalled = true
			return nil, nil
		})
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Same(t, updated, got)
		assert.False(t, reReadCalled, "re-read must not run when update succeeds")
	})

	t.Run("update returns non-NotFound error - propagated, no re-read", func(t *testing.T) {
		sentinel := errors.New("boom")
		reReadCalled := false
		got, found, err := ReReadOnUpdateNotFound((*obj)(nil), sentinel, func() (*obj, error) {
			reReadCalled = true
			return nil, nil
		})
		assert.ErrorIs(t, err, sentinel)
		assert.False(t, found)
		assert.Nil(t, got)
		assert.False(t, reReadCalled, "re-read must not run for a real error")
	})

	t.Run("spurious ErrNotFound but object exists - returns re-read result", func(t *testing.T) {
		reRead := &obj{name: "from-reread"}
		got, found, err := ReReadOnUpdateNotFound((*obj)(nil), unifi.ErrNotFound, func() (*obj, error) {
			return reRead, nil
		})
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Same(t, reRead, got, "re-read object (with controller normalization) must be returned")
	})

	t.Run("update and re-read both ErrNotFound - genuine deletion", func(t *testing.T) {
		got, found, err := ReReadOnUpdateNotFound((*obj)(nil), unifi.ErrNotFound, func() (*obj, error) {
			return nil, unifi.ErrNotFound
		})
		assert.NoError(t, err, "genuine deletion must not error so the caller can clear state")
		assert.False(t, found)
		assert.Nil(t, got)
	})

	t.Run("spurious ErrNotFound but re-read fails - propagates re-read error", func(t *testing.T) {
		sentinel := errors.New("read failed")
		got, found, err := ReReadOnUpdateNotFound((*obj)(nil), unifi.ErrNotFound, func() (*obj, error) {
			return nil, sentinel
		})
		assert.ErrorIs(t, err, sentinel)
		assert.False(t, found)
		assert.Nil(t, got)
	})
}

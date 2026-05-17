// Code generated from ace.jar fields *.json files
// DO NOT EDIT.

package unifi

import (
	"context"
	"encoding/json"
	"fmt"
)

// just to fix compile issues with the import
var (
	_ context.Context
	_ fmt.Formatter
	_ json.Marshaler
)

const SettingSuperFwupdateKey = "super_fwupdate"

type SettingSuperFwupdate struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	Hidden   bool   `json:"attr_hidden,omitempty"`
	HiddenID string `json:"attr_hidden_id,omitempty"`
	NoDelete bool   `json:"attr_no_delete,omitempty"`
	NoEdit   bool   `json:"attr_no_edit,omitempty"`

	Key string `json:"key"`

	ControllerChannel string `json:"controller_channel,omitempty" validate:"omitempty,oneof=internal alpha beta release-candidate release"` // internal|alpha|beta|release-candidate|release
	FirmwareChannel   string `json:"firmware_channel,omitempty" validate:"omitempty,oneof=internal alpha beta release-candidate release"`   // internal|alpha|beta|release-candidate|release
	SsoEnabled        bool   `json:"sso_enabled"`
}

func (dst *SettingSuperFwupdate) UnmarshalJSON(b []byte) error {
	type Alias SettingSuperFwupdate
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(dst),
	}

	err := json.Unmarshal(b, &aux)
	if err != nil {
		return fmt.Errorf("unable to unmarshal alias: %w", err)
	}

	return nil
}

// GetSettingSuperFwupdate Experimental! This function is not yet stable and may change in the future.
func (c *client) GetSettingSuperFwupdate(ctx context.Context, site string) (*SettingSuperFwupdate, error) {
	s, f, err := c.GetSetting(ctx, site, SettingSuperFwupdateKey)
	if err != nil {
		return nil, err
	}
	if s.Key != SettingSuperFwupdateKey {
		return nil, fmt.Errorf("unexpected setting key received. Requested: %q, received: %q", SettingSuperFwupdateKey, s.Key)
	}
	return f.(*SettingSuperFwupdate), nil
}

// UpdateSettingSuperFwupdate Experimental! This function is not yet stable and may change in the future.
func (c *client) UpdateSettingSuperFwupdate(ctx context.Context, site string, s *SettingSuperFwupdate) (*SettingSuperFwupdate, error) {
	s.Key = SettingSuperFwupdateKey
	result, err := c.SetSetting(ctx, site, SettingSuperFwupdateKey, s)
	if err != nil {
		return nil, err
	}
	return result.(*SettingSuperFwupdate), nil
}

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

const SettingSuperSdnKey = "super_sdn"

type SettingSuperSdn struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	Hidden   bool   `json:"attr_hidden,omitempty"`
	HiddenID string `json:"attr_hidden_id,omitempty"`
	NoDelete bool   `json:"attr_no_delete,omitempty"`
	NoEdit   bool   `json:"attr_no_edit,omitempty"`

	Key string `json:"key"`

	AuthToken       string `json:"auth_token,omitempty"`
	DeviceID        string `json:"device_id"`
	Enabled         bool   `json:"enabled"`
	Migrated        bool   `json:"migrated"`
	SsoLoginEnabled string `json:"sso_login_enabled,omitempty"`
	UbicUuid        string `json:"ubic_uuid,omitempty"`
}

func (dst *SettingSuperSdn) UnmarshalJSON(b []byte) error {
	type Alias SettingSuperSdn
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

// GetSettingSuperSdn Experimental! This function is not yet stable and may change in the future.
func (c *client) GetSettingSuperSdn(ctx context.Context, site string) (*SettingSuperSdn, error) {
	s, f, err := c.GetSetting(ctx, site, SettingSuperSdnKey)
	if err != nil {
		return nil, err
	}
	if s.Key != SettingSuperSdnKey {
		return nil, fmt.Errorf("unexpected setting key received. Requested: %q, received: %q", SettingSuperSdnKey, s.Key)
	}
	return f.(*SettingSuperSdn), nil
}

// UpdateSettingSuperSdn Experimental! This function is not yet stable and may change in the future.
func (c *client) UpdateSettingSuperSdn(ctx context.Context, site string, s *SettingSuperSdn) (*SettingSuperSdn, error) {
	s.Key = SettingSuperSdnKey
	result, err := c.SetSetting(ctx, site, SettingSuperSdnKey, s)
	if err != nil {
		return nil, err
	}
	return result.(*SettingSuperSdn), nil
}

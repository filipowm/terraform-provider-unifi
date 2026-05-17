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

const SettingDpiKey = "dpi"

type SettingDpi struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	Hidden   bool   `json:"attr_hidden,omitempty"`
	HiddenID string `json:"attr_hidden_id,omitempty"`
	NoDelete bool   `json:"attr_no_delete,omitempty"`
	NoEdit   bool   `json:"attr_no_edit,omitempty"`

	Key string `json:"key"`

	Enabled               bool `json:"enabled"`
	FingerprintingEnabled bool `json:"fingerprintingEnabled"`
}

func (dst *SettingDpi) UnmarshalJSON(b []byte) error {
	type Alias SettingDpi
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

// GetSettingDpi Experimental! This function is not yet stable and may change in the future.
func (c *client) GetSettingDpi(ctx context.Context, site string) (*SettingDpi, error) {
	s, f, err := c.GetSetting(ctx, site, SettingDpiKey)
	if err != nil {
		return nil, err
	}
	if s.Key != SettingDpiKey {
		return nil, fmt.Errorf("unexpected setting key received. Requested: %q, received: %q", SettingDpiKey, s.Key)
	}
	return f.(*SettingDpi), nil
}

// UpdateSettingDpi Experimental! This function is not yet stable and may change in the future.
func (c *client) UpdateSettingDpi(ctx context.Context, site string, s *SettingDpi) (*SettingDpi, error) {
	s.Key = SettingDpiKey
	result, err := c.SetSetting(ctx, site, SettingDpiKey, s)
	if err != nil {
		return nil, err
	}
	return result.(*SettingDpi), nil
}

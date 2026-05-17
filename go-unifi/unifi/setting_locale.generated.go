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

const SettingLocaleKey = "locale"

type SettingLocale struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	Hidden   bool   `json:"attr_hidden,omitempty"`
	HiddenID string `json:"attr_hidden_id,omitempty"`
	NoDelete bool   `json:"attr_no_delete,omitempty"`
	NoEdit   bool   `json:"attr_no_edit,omitempty"`

	Key string `json:"key"`

	Timezone string `json:"timezone,omitempty"`
}

func (dst *SettingLocale) UnmarshalJSON(b []byte) error {
	type Alias SettingLocale
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

// GetSettingLocale Experimental! This function is not yet stable and may change in the future.
func (c *client) GetSettingLocale(ctx context.Context, site string) (*SettingLocale, error) {
	s, f, err := c.GetSetting(ctx, site, SettingLocaleKey)
	if err != nil {
		return nil, err
	}
	if s.Key != SettingLocaleKey {
		return nil, fmt.Errorf("unexpected setting key received. Requested: %q, received: %q", SettingLocaleKey, s.Key)
	}
	return f.(*SettingLocale), nil
}

// UpdateSettingLocale Experimental! This function is not yet stable and may change in the future.
func (c *client) UpdateSettingLocale(ctx context.Context, site string, s *SettingLocale) (*SettingLocale, error) {
	s.Key = SettingLocaleKey
	result, err := c.SetSetting(ctx, site, SettingLocaleKey, s)
	if err != nil {
		return nil, err
	}
	return result.(*SettingLocale), nil
}

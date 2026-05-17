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

const SettingCountryKey = "country"

type SettingCountry struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	Hidden   bool   `json:"attr_hidden,omitempty"`
	HiddenID string `json:"attr_hidden_id,omitempty"`
	NoDelete bool   `json:"attr_no_delete,omitempty"`
	NoEdit   bool   `json:"attr_no_edit,omitempty"`

	Key string `json:"key"`

	Code int `json:"code,omitempty"`
}

func (dst *SettingCountry) UnmarshalJSON(b []byte) error {
	type Alias SettingCountry
	aux := &struct {
		Code emptyStringInt `json:"code"`

		*Alias
	}{
		Alias: (*Alias)(dst),
	}

	err := json.Unmarshal(b, &aux)
	if err != nil {
		return fmt.Errorf("unable to unmarshal alias: %w", err)
	}
	dst.Code = int(aux.Code)

	return nil
}

// GetSettingCountry Experimental! This function is not yet stable and may change in the future.
func (c *client) GetSettingCountry(ctx context.Context, site string) (*SettingCountry, error) {
	s, f, err := c.GetSetting(ctx, site, SettingCountryKey)
	if err != nil {
		return nil, err
	}
	if s.Key != SettingCountryKey {
		return nil, fmt.Errorf("unexpected setting key received. Requested: %q, received: %q", SettingCountryKey, s.Key)
	}
	return f.(*SettingCountry), nil
}

// UpdateSettingCountry Experimental! This function is not yet stable and may change in the future.
func (c *client) UpdateSettingCountry(ctx context.Context, site string, s *SettingCountry) (*SettingCountry, error) {
	s.Key = SettingCountryKey
	result, err := c.SetSetting(ctx, site, SettingCountryKey, s)
	if err != nil {
		return nil, err
	}
	return result.(*SettingCountry), nil
}

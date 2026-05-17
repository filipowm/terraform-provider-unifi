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

const SettingBaresipKey = "baresip"

type SettingBaresip struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	Hidden   bool   `json:"attr_hidden,omitempty"`
	HiddenID string `json:"attr_hidden_id,omitempty"`
	NoDelete bool   `json:"attr_no_delete,omitempty"`
	NoEdit   bool   `json:"attr_no_edit,omitempty"`

	Key string `json:"key"`

	Enabled       bool   `json:"enabled"`
	OutboundProxy string `json:"outbound_proxy,omitempty"`
	PackageUrl    string `json:"package_url,omitempty"`
	Server        string `json:"server,omitempty"`
}

func (dst *SettingBaresip) UnmarshalJSON(b []byte) error {
	type Alias SettingBaresip
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

// GetSettingBaresip Experimental! This function is not yet stable and may change in the future.
func (c *client) GetSettingBaresip(ctx context.Context, site string) (*SettingBaresip, error) {
	s, f, err := c.GetSetting(ctx, site, SettingBaresipKey)
	if err != nil {
		return nil, err
	}
	if s.Key != SettingBaresipKey {
		return nil, fmt.Errorf("unexpected setting key received. Requested: %q, received: %q", SettingBaresipKey, s.Key)
	}
	return f.(*SettingBaresip), nil
}

// UpdateSettingBaresip Experimental! This function is not yet stable and may change in the future.
func (c *client) UpdateSettingBaresip(ctx context.Context, site string, s *SettingBaresip) (*SettingBaresip, error) {
	s.Key = SettingBaresipKey
	result, err := c.SetSetting(ctx, site, SettingBaresipKey, s)
	if err != nil {
		return nil, err
	}
	return result.(*SettingBaresip), nil
}

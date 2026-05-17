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

const SettingDohKey = "doh"

type SettingDoh struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	Hidden   bool   `json:"attr_hidden,omitempty"`
	HiddenID string `json:"attr_hidden_id,omitempty"`
	NoDelete bool   `json:"attr_no_delete,omitempty"`
	NoEdit   bool   `json:"attr_no_edit,omitempty"`

	Key string `json:"key"`

	CustomServers []SettingDohCustomServers `json:"custom_servers,omitempty"`
	ServerNames   []string                  `json:"server_names,omitempty"`
	State         string                    `json:"state,omitempty" validate:"omitempty,oneof=off auto manual custom"` // off|auto|manual|custom
}

func (dst *SettingDoh) UnmarshalJSON(b []byte) error {
	type Alias SettingDoh
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

type SettingDohCustomServers struct {
	Enabled    bool   `json:"enabled"`
	SdnsStamp  string `json:"sdns_stamp,omitempty"`
	ServerName string `json:"server_name,omitempty"`
}

func (dst *SettingDohCustomServers) UnmarshalJSON(b []byte) error {
	type Alias SettingDohCustomServers
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

// GetSettingDoh Experimental! This function is not yet stable and may change in the future.
func (c *client) GetSettingDoh(ctx context.Context, site string) (*SettingDoh, error) {
	s, f, err := c.GetSetting(ctx, site, SettingDohKey)
	if err != nil {
		return nil, err
	}
	if s.Key != SettingDohKey {
		return nil, fmt.Errorf("unexpected setting key received. Requested: %q, received: %q", SettingDohKey, s.Key)
	}
	return f.(*SettingDoh), nil
}

// UpdateSettingDoh Experimental! This function is not yet stable and may change in the future.
func (c *client) UpdateSettingDoh(ctx context.Context, site string, s *SettingDoh) (*SettingDoh, error) {
	s.Key = SettingDohKey
	result, err := c.SetSetting(ctx, site, SettingDohKey, s)
	if err != nil {
		return nil, err
	}
	return result.(*SettingDoh), nil
}

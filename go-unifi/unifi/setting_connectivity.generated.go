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

const SettingConnectivityKey = "connectivity"

type SettingConnectivity struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	Hidden   bool   `json:"attr_hidden,omitempty"`
	HiddenID string `json:"attr_hidden_id,omitempty"`
	NoDelete bool   `json:"attr_no_delete,omitempty"`
	NoEdit   bool   `json:"attr_no_edit,omitempty"`

	Key string `json:"key"`

	EnableIsolatedWLAN bool   `json:"enable_isolated_wlan"`
	Enabled            bool   `json:"enabled"`
	UplinkHost         string `json:"uplink_host,omitempty"`
	UplinkType         string `json:"uplink_type,omitempty"`
	XMeshEssid         string `json:"x_mesh_essid,omitempty"`
	XMeshPsk           string `json:"x_mesh_psk,omitempty"`
}

func (dst *SettingConnectivity) UnmarshalJSON(b []byte) error {
	type Alias SettingConnectivity
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

// GetSettingConnectivity Experimental! This function is not yet stable and may change in the future.
func (c *client) GetSettingConnectivity(ctx context.Context, site string) (*SettingConnectivity, error) {
	s, f, err := c.GetSetting(ctx, site, SettingConnectivityKey)
	if err != nil {
		return nil, err
	}
	if s.Key != SettingConnectivityKey {
		return nil, fmt.Errorf("unexpected setting key received. Requested: %q, received: %q", SettingConnectivityKey, s.Key)
	}
	return f.(*SettingConnectivity), nil
}

// UpdateSettingConnectivity Experimental! This function is not yet stable and may change in the future.
func (c *client) UpdateSettingConnectivity(ctx context.Context, site string, s *SettingConnectivity) (*SettingConnectivity, error) {
	s.Key = SettingConnectivityKey
	result, err := c.SetSetting(ctx, site, SettingConnectivityKey, s)
	if err != nil {
		return nil, err
	}
	return result.(*SettingConnectivity), nil
}

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

const SettingNetworkOptimizationKey = "network_optimization"

type SettingNetworkOptimization struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	Hidden   bool   `json:"attr_hidden,omitempty"`
	HiddenID string `json:"attr_hidden_id,omitempty"`
	NoDelete bool   `json:"attr_no_delete,omitempty"`
	NoEdit   bool   `json:"attr_no_edit,omitempty"`

	Key string `json:"key"`

	Enabled bool `json:"enabled"`
}

func (dst *SettingNetworkOptimization) UnmarshalJSON(b []byte) error {
	type Alias SettingNetworkOptimization
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

// GetSettingNetworkOptimization Experimental! This function is not yet stable and may change in the future.
func (c *client) GetSettingNetworkOptimization(ctx context.Context, site string) (*SettingNetworkOptimization, error) {
	s, f, err := c.GetSetting(ctx, site, SettingNetworkOptimizationKey)
	if err != nil {
		return nil, err
	}
	if s.Key != SettingNetworkOptimizationKey {
		return nil, fmt.Errorf("unexpected setting key received. Requested: %q, received: %q", SettingNetworkOptimizationKey, s.Key)
	}
	return f.(*SettingNetworkOptimization), nil
}

// UpdateSettingNetworkOptimization Experimental! This function is not yet stable and may change in the future.
func (c *client) UpdateSettingNetworkOptimization(ctx context.Context, site string, s *SettingNetworkOptimization) (*SettingNetworkOptimization, error) {
	s.Key = SettingNetworkOptimizationKey
	result, err := c.SetSetting(ctx, site, SettingNetworkOptimizationKey, s)
	if err != nil {
		return nil, err
	}
	return result.(*SettingNetworkOptimization), nil
}

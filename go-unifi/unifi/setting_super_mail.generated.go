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

const SettingSuperMailKey = "super_mail"

type SettingSuperMail struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	Hidden   bool   `json:"attr_hidden,omitempty"`
	HiddenID string `json:"attr_hidden_id,omitempty"`
	NoDelete bool   `json:"attr_no_delete,omitempty"`
	NoEdit   bool   `json:"attr_no_edit,omitempty"`

	Key string `json:"key"`

	Provider string `json:"provider,omitempty" validate:"omitempty,oneof=smtp cloud disabled"` // smtp|cloud|disabled
}

func (dst *SettingSuperMail) UnmarshalJSON(b []byte) error {
	type Alias SettingSuperMail
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

// GetSettingSuperMail Experimental! This function is not yet stable and may change in the future.
func (c *client) GetSettingSuperMail(ctx context.Context, site string) (*SettingSuperMail, error) {
	s, f, err := c.GetSetting(ctx, site, SettingSuperMailKey)
	if err != nil {
		return nil, err
	}
	if s.Key != SettingSuperMailKey {
		return nil, fmt.Errorf("unexpected setting key received. Requested: %q, received: %q", SettingSuperMailKey, s.Key)
	}
	return f.(*SettingSuperMail), nil
}

// UpdateSettingSuperMail Experimental! This function is not yet stable and may change in the future.
func (c *client) UpdateSettingSuperMail(ctx context.Context, site string, s *SettingSuperMail) (*SettingSuperMail, error) {
	s.Key = SettingSuperMailKey
	result, err := c.SetSetting(ctx, site, SettingSuperMailKey, s)
	if err != nil {
		return nil, err
	}
	return result.(*SettingSuperMail), nil
}

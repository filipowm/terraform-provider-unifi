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

const SettingEvaluationScoreKey = "evaluation_score"

type SettingEvaluationScore struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	Hidden   bool   `json:"attr_hidden,omitempty"`
	HiddenID string `json:"attr_hidden_id,omitempty"`
	NoDelete bool   `json:"attr_no_delete,omitempty"`
	NoEdit   bool   `json:"attr_no_edit,omitempty"`

	Key string `json:"key"`

	DismissedIDs []string `json:"dismissed_ids,omitempty"` // ^[a-zA-Z]{2}[0-9]{2,3}$|^$
}

func (dst *SettingEvaluationScore) UnmarshalJSON(b []byte) error {
	type Alias SettingEvaluationScore
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

// GetSettingEvaluationScore Experimental! This function is not yet stable and may change in the future.
func (c *client) GetSettingEvaluationScore(ctx context.Context, site string) (*SettingEvaluationScore, error) {
	s, f, err := c.GetSetting(ctx, site, SettingEvaluationScoreKey)
	if err != nil {
		return nil, err
	}
	if s.Key != SettingEvaluationScoreKey {
		return nil, fmt.Errorf("unexpected setting key received. Requested: %q, received: %q", SettingEvaluationScoreKey, s.Key)
	}
	return f.(*SettingEvaluationScore), nil
}

// UpdateSettingEvaluationScore Experimental! This function is not yet stable and may change in the future.
func (c *client) UpdateSettingEvaluationScore(ctx context.Context, site string, s *SettingEvaluationScore) (*SettingEvaluationScore, error) {
	s.Key = SettingEvaluationScoreKey
	result, err := c.SetSetting(ctx, site, SettingEvaluationScoreKey, s)
	if err != nil {
		return nil, err
	}
	return result.(*SettingEvaluationScore), nil
}

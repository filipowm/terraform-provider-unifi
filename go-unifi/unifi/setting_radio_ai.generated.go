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

const SettingRadioAiKey = "radio_ai"

type SettingRadioAi struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	Hidden   bool   `json:"attr_hidden,omitempty"`
	HiddenID string `json:"attr_hidden_id,omitempty"`
	NoDelete bool   `json:"attr_no_delete,omitempty"`
	NoEdit   bool   `json:"attr_no_edit,omitempty"`

	Key string `json:"key"`

	AutoAdjustChannelsToCountry bool                              `json:"auto_adjust_channels_to_country"`
	Channels6E                  []int                             `json:"channels_6e,omitempty"` // [1-9]|[1-2][0-9]|3[3-9]|[4-5][0-9]|6[0-1]|6[5-9]|[7-8][0-9]|9[0-3]|9[7-9]|1[0-1][0-9]|12[0-5]|129|1[3-4][0-9]|15[0-7]|16[1-9]|1[7-8][0-9]|19[3-9]|2[0-1][0-9]|22[0-1]|22[5-9]|233
	ChannelsBlacklist           []SettingRadioAiChannelsBlacklist `json:"channels_blacklist,omitempty"`
	ChannelsNa                  []int                             `json:"channels_na,omitempty" validate:"omitempty,oneof=34 36 38 40 42 44 46 48 52 56 60 64 100 104 108 112 116 120 124 128 132 136 140 144 149 153 157 161 165 169"` // 34|36|38|40|42|44|46|48|52|56|60|64|100|104|108|112|116|120|124|128|132|136|140|144|149|153|157|161|165|169
	ChannelsNg                  []int                             `json:"channels_ng,omitempty" validate:"omitempty,oneof=1 2 3 4 5 6 7 8 9 10 11 12 13 14"`                                                                            // 1|2|3|4|5|6|7|8|9|10|11|12|13|14
	CronExpr                    string                            `json:"cron_expr,omitempty"`
	Default                     bool                              `json:"default"`
	Enabled                     bool                              `json:"enabled"`
	ExcludeDevices              []string                          `json:"exclude_devices,omitempty"`                                           // ([0-9a-z]{2}:){5}[0-9a-z]{2}
	HtModesNa                   []int                             `json:"ht_modes_na,omitempty" validate:"omitempty,oneof=20 40 80 160"`       // ^(20|40|80|160)$
	HtModesNg                   []int                             `json:"ht_modes_ng,omitempty" validate:"omitempty,oneof=20 40"`              // ^(20|40)$
	Optimize                    []string                          `json:"optimize,omitempty" validate:"omitempty,oneof=channel power"`         // channel|power
	Radios                      []string                          `json:"radios,omitempty" validate:"omitempty,oneof=na ng"`                   // na|ng
	SettingPreference           string                            `json:"setting_preference,omitempty" validate:"omitempty,oneof=auto manual"` // auto|manual
	UseXy                       bool                              `json:"useXY"`
}

func (dst *SettingRadioAi) UnmarshalJSON(b []byte) error {
	type Alias SettingRadioAi
	aux := &struct {
		Channels6E []emptyStringInt `json:"channels_6e"`
		ChannelsNa []emptyStringInt `json:"channels_na"`
		ChannelsNg []emptyStringInt `json:"channels_ng"`
		HtModesNa  []emptyStringInt `json:"ht_modes_na"`
		HtModesNg  []emptyStringInt `json:"ht_modes_ng"`

		*Alias
	}{
		Alias: (*Alias)(dst),
	}

	err := json.Unmarshal(b, &aux)
	if err != nil {
		return fmt.Errorf("unable to unmarshal alias: %w", err)
	}
	dst.Channels6E = make([]int, len(aux.Channels6E))
	for i, v := range aux.Channels6E {
		dst.Channels6E[i] = int(v)
	}
	dst.ChannelsNa = make([]int, len(aux.ChannelsNa))
	for i, v := range aux.ChannelsNa {
		dst.ChannelsNa[i] = int(v)
	}
	dst.ChannelsNg = make([]int, len(aux.ChannelsNg))
	for i, v := range aux.ChannelsNg {
		dst.ChannelsNg[i] = int(v)
	}
	dst.HtModesNa = make([]int, len(aux.HtModesNa))
	for i, v := range aux.HtModesNa {
		dst.HtModesNa[i] = int(v)
	}
	dst.HtModesNg = make([]int, len(aux.HtModesNg))
	for i, v := range aux.HtModesNg {
		dst.HtModesNg[i] = int(v)
	}

	return nil
}

type SettingRadioAiChannelsBlacklist struct {
	Channel      int    `json:"channel,omitempty"`                                                       // [1-9]|[1-9][0-9]|1[0-9][0-9]|2[0-9]|2[0-1][0-9]|22[0-1]|22[5-9]|233
	ChannelWidth int    `json:"channel_width,omitempty" validate:"omitempty,oneof=20 40 80 160 240 320"` // 20|40|80|160|240|320
	Radio        string `json:"radio,omitempty" validate:"omitempty,oneof=na ng 6e"`                     // na|ng|6e
}

func (dst *SettingRadioAiChannelsBlacklist) UnmarshalJSON(b []byte) error {
	type Alias SettingRadioAiChannelsBlacklist
	aux := &struct {
		Channel      emptyStringInt `json:"channel"`
		ChannelWidth emptyStringInt `json:"channel_width"`

		*Alias
	}{
		Alias: (*Alias)(dst),
	}

	err := json.Unmarshal(b, &aux)
	if err != nil {
		return fmt.Errorf("unable to unmarshal alias: %w", err)
	}
	dst.Channel = int(aux.Channel)
	dst.ChannelWidth = int(aux.ChannelWidth)

	return nil
}

// GetSettingRadioAi Experimental! This function is not yet stable and may change in the future.
func (c *client) GetSettingRadioAi(ctx context.Context, site string) (*SettingRadioAi, error) {
	s, f, err := c.GetSetting(ctx, site, SettingRadioAiKey)
	if err != nil {
		return nil, err
	}
	if s.Key != SettingRadioAiKey {
		return nil, fmt.Errorf("unexpected setting key received. Requested: %q, received: %q", SettingRadioAiKey, s.Key)
	}
	return f.(*SettingRadioAi), nil
}

// UpdateSettingRadioAi Experimental! This function is not yet stable and may change in the future.
func (c *client) UpdateSettingRadioAi(ctx context.Context, site string, s *SettingRadioAi) (*SettingRadioAi, error) {
	s.Key = SettingRadioAiKey
	result, err := c.SetSetting(ctx, site, SettingRadioAiKey, s)
	if err != nil {
		return nil, err
	}
	return result.(*SettingRadioAi), nil
}

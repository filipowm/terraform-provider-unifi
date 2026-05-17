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

type FirewallZoneMatrix struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	Hidden   bool   `json:"attr_hidden,omitempty"`
	HiddenID string `json:"attr_hidden_id,omitempty"`
	NoDelete bool   `json:"attr_no_delete,omitempty"`
	NoEdit   bool   `json:"attr_no_edit,omitempty"`

	Data    []FirewallZoneMatrixData `json:"data,omitempty"`
	Name    string                   `json:"name,omitempty"`
	ZoneKey string                   `json:"zone_key,omitempty"`
}

func (dst *FirewallZoneMatrix) UnmarshalJSON(b []byte) error {
	type Alias FirewallZoneMatrix
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

type FirewallZoneMatrixData struct {
	Action      string `json:"action,omitempty"`
	PolicyCount int    `json:"policy_count,omitempty"`
}

func (dst *FirewallZoneMatrixData) UnmarshalJSON(b []byte) error {
	type Alias FirewallZoneMatrixData
	aux := &struct {
		PolicyCount emptyStringInt `json:"policy_count"`

		*Alias
	}{
		Alias: (*Alias)(dst),
	}

	err := json.Unmarshal(b, &aux)
	if err != nil {
		return fmt.Errorf("unable to unmarshal alias: %w", err)
	}
	dst.PolicyCount = int(aux.PolicyCount)

	return nil
}

func (c *client) listFirewallZoneMatrix(ctx context.Context, site string) ([]FirewallZoneMatrix, error) {
	var respBody []FirewallZoneMatrix

	err := c.Get(ctx, fmt.Sprintf("%s/site/%s/firewall/zone-matrix", c.apiPaths.ApiV2Path, site), nil, &respBody)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

func (c *client) getFirewallZoneMatrix(ctx context.Context, site, id string) (*FirewallZoneMatrix, error) {
	var respBody FirewallZoneMatrix

	err := c.Get(ctx, fmt.Sprintf("%s/site/%s/firewall/zone-matrix/%s", c.apiPaths.ApiV2Path, site, id), nil, &respBody)

	if err != nil {
		return nil, err
	}
	if respBody.ID == "" {
		return nil, ErrNotFound
	}
	return &respBody, nil
}

func (c *client) deleteFirewallZoneMatrix(ctx context.Context, site, id string) error {
	err := c.Delete(ctx, fmt.Sprintf("%s/site/%s/firewall/zone-matrix/%s", c.apiPaths.ApiV2Path, site, id), struct{}{}, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *client) createFirewallZoneMatrix(ctx context.Context, site string, d *FirewallZoneMatrix) (*FirewallZoneMatrix, error) {
	var respBody FirewallZoneMatrix

	err := c.Post(ctx, fmt.Sprintf("%s/site/%s/firewall/zone-matrix", c.apiPaths.ApiV2Path, site), d, &respBody)
	if err != nil {
		return nil, err
	}

	return &respBody, nil
}

func (c *client) updateFirewallZoneMatrix(ctx context.Context, site string, d *FirewallZoneMatrix) (*FirewallZoneMatrix, error) {
	var respBody FirewallZoneMatrix

	err := c.Put(ctx, fmt.Sprintf("%s/site/%s/firewall/zone-matrix/%s", c.apiPaths.ApiV2Path, site, d.ID), d, &respBody)
	if err != nil {
		return nil, err
	}
	return &respBody, nil
}

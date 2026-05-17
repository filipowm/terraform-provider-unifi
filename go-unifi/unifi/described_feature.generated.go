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

type DescribedFeature struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	Hidden   bool   `json:"attr_hidden,omitempty"`
	HiddenID string `json:"attr_hidden_id,omitempty"`
	NoDelete bool   `json:"attr_no_delete,omitempty"`
	NoEdit   bool   `json:"attr_no_edit,omitempty"`

	FeatureExists bool   `json:"feature_exists"`
	Name          string `json:"name,omitempty"`
}

func (dst *DescribedFeature) UnmarshalJSON(b []byte) error {
	type Alias DescribedFeature
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

func (c *client) listDescribedFeature(ctx context.Context, site string) ([]DescribedFeature, error) {
	var respBody []DescribedFeature

	err := c.Get(ctx, fmt.Sprintf("%s/site/%s/described-features?includeSystemFeatures=true", c.apiPaths.ApiV2Path, site), nil, &respBody)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

func (c *client) getDescribedFeature(ctx context.Context, site, id string) (*DescribedFeature, error) {
	var respBody DescribedFeature

	err := c.Get(ctx, fmt.Sprintf("%s/site/%s/described-features?includeSystemFeatures=true/%s", c.apiPaths.ApiV2Path, site, id), nil, &respBody)

	if err != nil {
		return nil, err
	}
	if respBody.ID == "" {
		return nil, ErrNotFound
	}
	return &respBody, nil
}

func (c *client) deleteDescribedFeature(ctx context.Context, site, id string) error {
	err := c.Delete(ctx, fmt.Sprintf("%s/site/%s/described-features?includeSystemFeatures=true/%s", c.apiPaths.ApiV2Path, site, id), struct{}{}, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *client) createDescribedFeature(ctx context.Context, site string, d *DescribedFeature) (*DescribedFeature, error) {
	var respBody DescribedFeature

	err := c.Post(ctx, fmt.Sprintf("%s/site/%s/described-features?includeSystemFeatures=true", c.apiPaths.ApiV2Path, site), d, &respBody)
	if err != nil {
		return nil, err
	}

	return &respBody, nil
}

func (c *client) updateDescribedFeature(ctx context.Context, site string, d *DescribedFeature) (*DescribedFeature, error) {
	var respBody DescribedFeature

	err := c.Put(ctx, fmt.Sprintf("%s/site/%s/described-features?includeSystemFeatures=true/%s", c.apiPaths.ApiV2Path, site, d.ID), d, &respBody)
	if err != nil {
		return nil, err
	}
	return &respBody, nil
}

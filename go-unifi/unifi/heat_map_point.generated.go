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

type HeatMapPoint struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	Hidden   bool   `json:"attr_hidden,omitempty"`
	HiddenID string `json:"attr_hidden_id,omitempty"`
	NoDelete bool   `json:"attr_no_delete,omitempty"`
	NoEdit   bool   `json:"attr_no_edit,omitempty"`

	DownloadSpeed float64 `json:"download_speed,omitempty"`
	HeatmapID     string  `json:"heatmap_id"`
	UploadSpeed   float64 `json:"upload_speed,omitempty"`
	X             float64 `json:"x,omitempty"`
	Y             float64 `json:"y,omitempty"`
}

func (dst *HeatMapPoint) UnmarshalJSON(b []byte) error {
	type Alias HeatMapPoint
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

func (c *client) listHeatMapPoint(ctx context.Context, site string) ([]HeatMapPoint, error) {
	var respBody struct {
		Meta Meta           `json:"meta"`
		Data []HeatMapPoint `json:"data"`
	}

	err := c.Get(ctx, fmt.Sprintf("s/%s/rest/heatmappoint", site), nil, &respBody)
	if err != nil {
		return nil, err
	}

	return respBody.Data, nil
}

func (c *client) getHeatMapPoint(ctx context.Context, site, id string) (*HeatMapPoint, error) {
	var respBody struct {
		Meta Meta           `json:"meta"`
		Data []HeatMapPoint `json:"data"`
	}

	err := c.Get(ctx, fmt.Sprintf("s/%s/rest/heatmappoint/%s", site, id), nil, &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) != 1 {
		return nil, ErrNotFound
	}

	d := respBody.Data[0]
	return &d, nil
}

func (c *client) deleteHeatMapPoint(ctx context.Context, site, id string) error {
	err := c.Delete(ctx, fmt.Sprintf("s/%s/rest/heatmappoint/%s", site, id), struct{}{}, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *client) createHeatMapPoint(ctx context.Context, site string, d *HeatMapPoint) (*HeatMapPoint, error) {
	var respBody struct {
		Meta Meta           `json:"meta"`
		Data []HeatMapPoint `json:"data"`
	}

	err := c.Post(ctx, fmt.Sprintf("s/%s/rest/heatmappoint", site), d, &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) != 1 {
		return nil, ErrNotFound
	}

	new := respBody.Data[0]

	return &new, nil
}

func (c *client) updateHeatMapPoint(ctx context.Context, site string, d *HeatMapPoint) (*HeatMapPoint, error) {
	var respBody struct {
		Meta Meta           `json:"meta"`
		Data []HeatMapPoint `json:"data"`
	}

	err := c.Put(ctx, fmt.Sprintf("s/%s/rest/heatmappoint/%s", site, d.ID), d, &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) != 1 {
		return nil, ErrNotFound
	}

	new := respBody.Data[0]

	return &new, nil
}

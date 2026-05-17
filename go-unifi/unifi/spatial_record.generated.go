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

type SpatialRecord struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	Hidden   bool   `json:"attr_hidden,omitempty"`
	HiddenID string `json:"attr_hidden_id,omitempty"`
	NoDelete bool   `json:"attr_no_delete,omitempty"`
	NoEdit   bool   `json:"attr_no_edit,omitempty"`

	Devices []SpatialRecordDevices `json:"devices,omitempty"`
	Name    string                 `json:"name,omitempty" validate:"omitempty,gte=1,lte=128"` // .{1,128}
}

func (dst *SpatialRecord) UnmarshalJSON(b []byte) error {
	type Alias SpatialRecord
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

type SpatialRecordDevices struct {
	MAC      string                `json:"mac,omitempty" validate:"omitempty,mac"` // ^([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})$
	Position SpatialRecordPosition `json:"position,omitempty"`
}

func (dst *SpatialRecordDevices) UnmarshalJSON(b []byte) error {
	type Alias SpatialRecordDevices
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

type SpatialRecordPosition struct {
	X float64 `json:"x,omitempty"` // (^([-]?[\d]+)$)|(^([-]?[\d]+[.]?[\d]+)$)
	Y float64 `json:"y,omitempty"` // (^([-]?[\d]+)$)|(^([-]?[\d]+[.]?[\d]+)$)
	Z float64 `json:"z,omitempty"` // (^([-]?[\d]+)$)|(^([-]?[\d]+[.]?[\d]+)$)
}

func (dst *SpatialRecordPosition) UnmarshalJSON(b []byte) error {
	type Alias SpatialRecordPosition
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

func (c *client) listSpatialRecord(ctx context.Context, site string) ([]SpatialRecord, error) {
	var respBody struct {
		Meta Meta            `json:"meta"`
		Data []SpatialRecord `json:"data"`
	}

	err := c.Get(ctx, fmt.Sprintf("s/%s/rest/spatialrecord", site), nil, &respBody)
	if err != nil {
		return nil, err
	}

	return respBody.Data, nil
}

func (c *client) getSpatialRecord(ctx context.Context, site, id string) (*SpatialRecord, error) {
	var respBody struct {
		Meta Meta            `json:"meta"`
		Data []SpatialRecord `json:"data"`
	}

	err := c.Get(ctx, fmt.Sprintf("s/%s/rest/spatialrecord/%s", site, id), nil, &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) != 1 {
		return nil, ErrNotFound
	}

	d := respBody.Data[0]
	return &d, nil
}

func (c *client) deleteSpatialRecord(ctx context.Context, site, id string) error {
	err := c.Delete(ctx, fmt.Sprintf("s/%s/rest/spatialrecord/%s", site, id), struct{}{}, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *client) createSpatialRecord(ctx context.Context, site string, d *SpatialRecord) (*SpatialRecord, error) {
	var respBody struct {
		Meta Meta            `json:"meta"`
		Data []SpatialRecord `json:"data"`
	}

	err := c.Post(ctx, fmt.Sprintf("s/%s/rest/spatialrecord", site), d, &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) != 1 {
		return nil, ErrNotFound
	}

	new := respBody.Data[0]

	return &new, nil
}

func (c *client) updateSpatialRecord(ctx context.Context, site string, d *SpatialRecord) (*SpatialRecord, error) {
	var respBody struct {
		Meta Meta            `json:"meta"`
		Data []SpatialRecord `json:"data"`
	}

	err := c.Put(ctx, fmt.Sprintf("s/%s/rest/spatialrecord/%s", site, d.ID), d, &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) != 1 {
		return nil, ErrNotFound
	}

	new := respBody.Data[0]

	return &new, nil
}

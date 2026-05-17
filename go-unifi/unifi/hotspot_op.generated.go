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

type HotspotOp struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	Hidden   bool   `json:"attr_hidden,omitempty"`
	HiddenID string `json:"attr_hidden_id,omitempty"`
	NoDelete bool   `json:"attr_no_delete,omitempty"`
	NoEdit   bool   `json:"attr_no_edit,omitempty"`

	Name      string `json:"name,omitempty" validate:"omitempty,gte=1,lte=256"` // .{1,256}
	Note      string `json:"note,omitempty"`
	XPassword string `json:"x_password,omitempty" validate:"omitempty,gte=1,lte=256"` // .{1,256}
}

func (dst *HotspotOp) UnmarshalJSON(b []byte) error {
	type Alias HotspotOp
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

func (c *client) listHotspotOp(ctx context.Context, site string) ([]HotspotOp, error) {
	var respBody struct {
		Meta Meta        `json:"meta"`
		Data []HotspotOp `json:"data"`
	}

	err := c.Get(ctx, fmt.Sprintf("s/%s/rest/hotspotop", site), nil, &respBody)
	if err != nil {
		return nil, err
	}

	return respBody.Data, nil
}

func (c *client) getHotspotOp(ctx context.Context, site, id string) (*HotspotOp, error) {
	var respBody struct {
		Meta Meta        `json:"meta"`
		Data []HotspotOp `json:"data"`
	}

	err := c.Get(ctx, fmt.Sprintf("s/%s/rest/hotspotop/%s", site, id), nil, &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) != 1 {
		return nil, ErrNotFound
	}

	d := respBody.Data[0]
	return &d, nil
}

func (c *client) deleteHotspotOp(ctx context.Context, site, id string) error {
	err := c.Delete(ctx, fmt.Sprintf("s/%s/rest/hotspotop/%s", site, id), struct{}{}, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *client) createHotspotOp(ctx context.Context, site string, d *HotspotOp) (*HotspotOp, error) {
	var respBody struct {
		Meta Meta        `json:"meta"`
		Data []HotspotOp `json:"data"`
	}

	err := c.Post(ctx, fmt.Sprintf("s/%s/rest/hotspotop", site), d, &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) != 1 {
		return nil, ErrNotFound
	}

	new := respBody.Data[0]

	return &new, nil
}

func (c *client) updateHotspotOp(ctx context.Context, site string, d *HotspotOp) (*HotspotOp, error) {
	var respBody struct {
		Meta Meta        `json:"meta"`
		Data []HotspotOp `json:"data"`
	}

	err := c.Put(ctx, fmt.Sprintf("s/%s/rest/hotspotop/%s", site, d.ID), d, &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) != 1 {
		return nil, ErrNotFound
	}

	new := respBody.Data[0]

	return &new, nil
}

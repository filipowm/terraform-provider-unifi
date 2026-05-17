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

type DpiGroup struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	Hidden   bool   `json:"attr_hidden,omitempty"`
	HiddenID string `json:"attr_hidden_id,omitempty"`
	NoDelete bool   `json:"attr_no_delete,omitempty"`
	NoEdit   bool   `json:"attr_no_edit,omitempty"`

	DPIappIDs []string `json:"dpiapp_ids,omitempty" validate:"omitempty,w_regex"` // [\d\w]+
	Enabled   bool     `json:"enabled"`
	Name      string   `json:"name,omitempty" validate:"omitempty,gte=1,lte=128"` // .{1,128}
}

func (dst *DpiGroup) UnmarshalJSON(b []byte) error {
	type Alias DpiGroup
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

func (c *client) listDpiGroup(ctx context.Context, site string) ([]DpiGroup, error) {
	var respBody struct {
		Meta Meta       `json:"meta"`
		Data []DpiGroup `json:"data"`
	}

	err := c.Get(ctx, fmt.Sprintf("s/%s/rest/dpigroup", site), nil, &respBody)
	if err != nil {
		return nil, err
	}

	return respBody.Data, nil
}

func (c *client) getDpiGroup(ctx context.Context, site, id string) (*DpiGroup, error) {
	var respBody struct {
		Meta Meta       `json:"meta"`
		Data []DpiGroup `json:"data"`
	}

	err := c.Get(ctx, fmt.Sprintf("s/%s/rest/dpigroup/%s", site, id), nil, &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) != 1 {
		return nil, ErrNotFound
	}

	d := respBody.Data[0]
	return &d, nil
}

func (c *client) deleteDpiGroup(ctx context.Context, site, id string) error {
	err := c.Delete(ctx, fmt.Sprintf("s/%s/rest/dpigroup/%s", site, id), struct{}{}, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *client) createDpiGroup(ctx context.Context, site string, d *DpiGroup) (*DpiGroup, error) {
	var respBody struct {
		Meta Meta       `json:"meta"`
		Data []DpiGroup `json:"data"`
	}

	err := c.Post(ctx, fmt.Sprintf("s/%s/rest/dpigroup", site), d, &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) != 1 {
		return nil, ErrNotFound
	}

	new := respBody.Data[0]

	return &new, nil
}

func (c *client) updateDpiGroup(ctx context.Context, site string, d *DpiGroup) (*DpiGroup, error) {
	var respBody struct {
		Meta Meta       `json:"meta"`
		Data []DpiGroup `json:"data"`
	}

	err := c.Put(ctx, fmt.Sprintf("s/%s/rest/dpigroup/%s", site, d.ID), d, &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) != 1 {
		return nil, ErrNotFound
	}

	new := respBody.Data[0]

	return &new, nil
}

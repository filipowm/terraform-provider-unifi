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

type MediaFile struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	Hidden   bool   `json:"attr_hidden,omitempty"`
	HiddenID string `json:"attr_hidden_id,omitempty"`
	NoDelete bool   `json:"attr_no_delete,omitempty"`
	NoEdit   bool   `json:"attr_no_edit,omitempty"`

	Name string `json:"name,omitempty"`
}

func (dst *MediaFile) UnmarshalJSON(b []byte) error {
	type Alias MediaFile
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

func (c *client) listMediaFile(ctx context.Context, site string) ([]MediaFile, error) {
	var respBody struct {
		Meta Meta        `json:"meta"`
		Data []MediaFile `json:"data"`
	}

	err := c.Get(ctx, fmt.Sprintf("s/%s/rest/mediafile", site), nil, &respBody)
	if err != nil {
		return nil, err
	}

	return respBody.Data, nil
}

func (c *client) getMediaFile(ctx context.Context, site, id string) (*MediaFile, error) {
	var respBody struct {
		Meta Meta        `json:"meta"`
		Data []MediaFile `json:"data"`
	}

	err := c.Get(ctx, fmt.Sprintf("s/%s/rest/mediafile/%s", site, id), nil, &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) != 1 {
		return nil, ErrNotFound
	}

	d := respBody.Data[0]
	return &d, nil
}

func (c *client) deleteMediaFile(ctx context.Context, site, id string) error {
	err := c.Delete(ctx, fmt.Sprintf("s/%s/rest/mediafile/%s", site, id), struct{}{}, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *client) createMediaFile(ctx context.Context, site string, d *MediaFile) (*MediaFile, error) {
	var respBody struct {
		Meta Meta        `json:"meta"`
		Data []MediaFile `json:"data"`
	}

	err := c.Post(ctx, fmt.Sprintf("s/%s/rest/mediafile", site), d, &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) != 1 {
		return nil, ErrNotFound
	}

	new := respBody.Data[0]

	return &new, nil
}

func (c *client) updateMediaFile(ctx context.Context, site string, d *MediaFile) (*MediaFile, error) {
	var respBody struct {
		Meta Meta        `json:"meta"`
		Data []MediaFile `json:"data"`
	}

	err := c.Put(ctx, fmt.Sprintf("s/%s/rest/mediafile/%s", site, d.ID), d, &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) != 1 {
		return nil, ErrNotFound
	}

	new := respBody.Data[0]

	return &new, nil
}

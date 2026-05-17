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

type UserGroup struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	Hidden   bool   `json:"attr_hidden,omitempty"`
	HiddenID string `json:"attr_hidden_id,omitempty"`
	NoDelete bool   `json:"attr_no_delete,omitempty"`
	NoEdit   bool   `json:"attr_no_edit,omitempty"`

	Name           string `json:"name,omitempty" validate:"omitempty,gte=1,lte=128"` // .{1,128}
	QOSRateMaxDown int    `json:"qos_rate_max_down,omitempty"`                       // -1|[2-9]|[1-9][0-9]{1,4}|100000
	QOSRateMaxUp   int    `json:"qos_rate_max_up,omitempty"`                         // -1|[2-9]|[1-9][0-9]{1,4}|100000
}

func (dst *UserGroup) UnmarshalJSON(b []byte) error {
	type Alias UserGroup
	aux := &struct {
		QOSRateMaxDown emptyStringInt `json:"qos_rate_max_down"`
		QOSRateMaxUp   emptyStringInt `json:"qos_rate_max_up"`

		*Alias
	}{
		Alias: (*Alias)(dst),
	}

	err := json.Unmarshal(b, &aux)
	if err != nil {
		return fmt.Errorf("unable to unmarshal alias: %w", err)
	}
	dst.QOSRateMaxDown = int(aux.QOSRateMaxDown)
	dst.QOSRateMaxUp = int(aux.QOSRateMaxUp)

	return nil
}

func (c *client) listUserGroup(ctx context.Context, site string) ([]UserGroup, error) {
	var respBody struct {
		Meta Meta        `json:"meta"`
		Data []UserGroup `json:"data"`
	}

	err := c.Get(ctx, fmt.Sprintf("s/%s/rest/usergroup", site), nil, &respBody)
	if err != nil {
		return nil, err
	}

	return respBody.Data, nil
}

func (c *client) getUserGroup(ctx context.Context, site, id string) (*UserGroup, error) {
	var respBody struct {
		Meta Meta        `json:"meta"`
		Data []UserGroup `json:"data"`
	}

	err := c.Get(ctx, fmt.Sprintf("s/%s/rest/usergroup/%s", site, id), nil, &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) != 1 {
		return nil, ErrNotFound
	}

	d := respBody.Data[0]
	return &d, nil
}

func (c *client) deleteUserGroup(ctx context.Context, site, id string) error {
	err := c.Delete(ctx, fmt.Sprintf("s/%s/rest/usergroup/%s", site, id), struct{}{}, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *client) createUserGroup(ctx context.Context, site string, d *UserGroup) (*UserGroup, error) {
	var respBody struct {
		Meta Meta        `json:"meta"`
		Data []UserGroup `json:"data"`
	}

	err := c.Post(ctx, fmt.Sprintf("s/%s/rest/usergroup", site), d, &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) != 1 {
		return nil, ErrNotFound
	}

	new := respBody.Data[0]

	return &new, nil
}

func (c *client) updateUserGroup(ctx context.Context, site string, d *UserGroup) (*UserGroup, error) {
	var respBody struct {
		Meta Meta        `json:"meta"`
		Data []UserGroup `json:"data"`
	}

	err := c.Put(ctx, fmt.Sprintf("s/%s/rest/usergroup/%s", site, d.ID), d, &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) != 1 {
		return nil, ErrNotFound
	}

	new := respBody.Data[0]

	return &new, nil
}

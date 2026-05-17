package unifi

import (
	"context"
	"errors"
	"fmt"
)

// GetUserByMAC returns slightly different information than GetUser, as they
// use separate endpoints for their lookups. Specifically IP is only returned
// by this method.
func (c *client) GetUserByMAC(ctx context.Context, site, mac string) (*User, error) {
	var respBody struct {
		Meta Meta   `json:"Meta"`
		Data []User `json:"data"`
	}

	err := c.Get(ctx, fmt.Sprintf("s/%s/stat/user/%s", site, mac), nil, &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) != 1 {
		return nil, ErrNotFound
	}

	d := respBody.Data[0]
	return &d, nil
}

func (c *client) CreateUser(ctx context.Context, site string, d *User) (*User, error) {
	reqBody := struct {
		Objects []struct {
			Data *User `json:"data"`
		} `json:"objects"`
	}{
		Objects: []struct {
			Data *User `json:"data"`
		}{
			{Data: d},
		},
	}

	var respBody struct {
		Meta Meta `json:"Meta"`
		Data []struct {
			Meta Meta   `json:"Meta"`
			Data []User `json:"data"`
		} `json:"data"`
	}

	err := c.Post(ctx, fmt.Sprintf("s/%s/group/user", site), reqBody, &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) != 1 {
		return nil, errors.New("malformed group response")
	}

	// TODO verify if this is still needed cause it's the only place where old-style error handling is used
	if err := respBody.Data[0].Meta.error(); err != nil {
		return nil, err
	}

	if len(respBody.Data[0].Data) != 1 {
		return nil, ErrNotFound
	}

	user := respBody.Data[0].Data[0]

	return &user, nil
}

func (c *client) stamgr(ctx context.Context, site, cmd string, data map[string]interface{}) ([]User, error) {
	reqBody := map[string]interface{}{}

	for k, v := range data {
		reqBody[k] = v
	}

	reqBody["cmd"] = cmd

	var respBody struct {
		Meta Meta   `json:"Meta"`
		Data []User `json:"data"`
	}

	err := c.Post(ctx, fmt.Sprintf("s/%s/cmd/stamgr", site), reqBody, &respBody)
	if err != nil {
		return nil, err
	}

	return respBody.Data, nil
}

func (c *client) BlockUserByMAC(ctx context.Context, site, mac string) error {
	users, err := c.stamgr(ctx, site, "block-sta", map[string]interface{}{
		"mac": mac,
	})
	if err != nil {
		return err
	}
	if len(users) != 1 {
		return ErrNotFound
	}
	return nil
}

func (c *client) UnblockUserByMAC(ctx context.Context, site, mac string) error {
	users, err := c.stamgr(ctx, site, "unblock-sta", map[string]interface{}{
		"mac": mac,
	})
	if err != nil {
		return err
	}
	if len(users) != 1 {
		return ErrNotFound
	}
	return nil
}

func (c *client) DeleteUserByMAC(ctx context.Context, site, mac string) error {
	users, err := c.stamgr(ctx, site, "forget-sta", map[string]interface{}{
		"macs": []string{mac},
	})
	if err != nil {
		return err
	}
	if len(users) != 1 {
		return ErrNotFound
	}
	return nil
}

func (c *client) KickUserByMAC(ctx context.Context, site, mac string) error {
	users, err := c.stamgr(ctx, site, "kick-sta", map[string]interface{}{
		"mac": mac,
	})
	if err != nil {
		return err
	}
	if len(users) != 1 {
		return ErrNotFound
	}
	return nil
}

func (c *client) OverrideUserFingerprint(ctx context.Context, site, mac string, devIdOverride int) error {
	reqBody := map[string]interface{}{
		"mac":             mac,
		"dev_id_override": devIdOverride,
		"search_query":    "",
	}

	var reqMethod string
	if devIdOverride == 0 {
		reqMethod = "DELETE"
	} else {
		reqMethod = "PUT"
	}

	var respBody struct {
		Mac           string `json:"mac"`
		DevIdOverride int    `json:"dev_id_override"`
		SearchQuery   string `json:"search_query"`
	}

	err := c.Do(ctx, reqMethod, fmt.Sprintf("%s/site/%s/station/%s/fingerprint_override", c.apiPaths.ApiV2Path, site, mac), reqBody, &respBody)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) ListUser(ctx context.Context, site string) ([]User, error) {
	return c.listUser(ctx, site)
}

// GetUser returns information about a user from the REST endpoint.
// The GetUserByMAC method returns slightly different information (for
// example the IP) as it uses a different endpoint.
func (c *client) GetUser(ctx context.Context, site, id string) (*User, error) {
	return c.getUser(ctx, site, id)
}

func (c *client) UpdateUser(ctx context.Context, site string, d *User) (*User, error) {
	return c.updateUser(ctx, site, d)
}

func (c *client) DeleteUser(ctx context.Context, site, id string) error {
	return c.deleteUser(ctx, site, id)
}

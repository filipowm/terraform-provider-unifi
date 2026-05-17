package unifi

import (
	"context"
	"encoding/json"
	"fmt"
)

func (dst *Network) MarshalJSON() ([]byte, error) {
	type Alias Network
	aux := &struct {
		*Alias

		WANEgressQOS *emptyStringInt `json:"wan_egress_qos,omitempty"`
	}{
		Alias: (*Alias)(dst),
	}

	if dst.Purpose == "wan" {
		// only send QOS when this is a WAN network
		v := emptyStringInt(dst.WANEgressQOS)
		aux.WANEgressQOS = &v
	}

	b, err := json.Marshal(aux)
	return b, err
}

//func (c *client) DeleteNetwork(ctx context.Context, site, id, name string) error {
//	err := c.Delete(ctx, fmt.Sprintf("s/%s/rest/networkconf/%s", site, id), struct {
//		Name string `json:"name"`
//	}{
//		Name: name,
//	}, nil)
//	if err != nil {
//		return err
//	}
//	return nil
//}

func (c *client) DeleteNetwork(ctx context.Context, site, id string) error {
	err := c.Delete(ctx, fmt.Sprintf("s/%s/rest/networkconf/%s", site, id), nil, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *client) ListNetwork(ctx context.Context, site string) ([]Network, error) {
	return c.listNetwork(ctx, site)
}

func (c *client) GetNetwork(ctx context.Context, site, id string) (*Network, error) {
	return c.getNetwork(ctx, site, id)
}

func (c *client) CreateNetwork(ctx context.Context, site string, d *Network) (*Network, error) {
	return c.createNetwork(ctx, site, d)
}

func (c *client) UpdateNetwork(ctx context.Context, site string, d *Network) (*Network, error) {
	return c.updateNetwork(ctx, site, d)
}

package unifi

import (
	"context"
)

func (c *client) CreateWLAN(ctx context.Context, site string, d *WLAN) (*WLAN, error) {
	if d.Schedule == nil {
		d.Schedule = []string{}
	}

	return c.createWLAN(ctx, site, d)
}

func (c *client) ListWLAN(ctx context.Context, site string) ([]WLAN, error) {
	return c.listWLAN(ctx, site)
}

func (c *client) GetWLAN(ctx context.Context, site, id string) (*WLAN, error) {
	return c.getWLAN(ctx, site, id)
}

func (c *client) DeleteWLAN(ctx context.Context, site, id string) error {
	return c.deleteWLAN(ctx, site, id)
}

func (c *client) UpdateWLAN(ctx context.Context, site string, d *WLAN) (*WLAN, error) {
	return c.updateWLAN(ctx, site, d)
}

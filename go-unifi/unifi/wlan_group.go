package unifi

import (
	"context"
)

func (c *client) ListWLANGroup(ctx context.Context, site string) ([]WLANGroup, error) {
	return c.listWLANGroup(ctx, site)
}

func (c *client) GetWLANGroup(ctx context.Context, site, id string) (*WLANGroup, error) {
	return c.getWLANGroup(ctx, site, id)
}

func (c *client) DeleteWLANGroup(ctx context.Context, site, id string) error {
	return c.deleteWLANGroup(ctx, site, id)
}

func (c *client) CreateWLANGroup(ctx context.Context, site string, d *WLANGroup) (*WLANGroup, error) {
	return c.createWLANGroup(ctx, site, d)
}

func (c *client) UpdateWLANGroup(ctx context.Context, site string, d *WLANGroup) (*WLANGroup, error) {
	return c.updateWLANGroup(ctx, site, d)
}

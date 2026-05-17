package unifi

import (
	"context"
)

func (c *client) ListHotspotOp(ctx context.Context, site string) ([]HotspotOp, error) {
	return c.listHotspotOp(ctx, site)
}

func (c *client) GetHotspotOp(ctx context.Context, site, id string) (*HotspotOp, error) {
	return c.getHotspotOp(ctx, site, id)
}

func (c *client) DeleteHotspotOp(ctx context.Context, site, id string) error {
	return c.deleteHotspotOp(ctx, site, id)
}

func (c *client) CreateHotspotOp(ctx context.Context, site string, d *HotspotOp) (*HotspotOp, error) {
	return c.createHotspotOp(ctx, site, d)
}

func (c *client) UpdateHotspotOp(ctx context.Context, site string, d *HotspotOp) (*HotspotOp, error) {
	return c.updateHotspotOp(ctx, site, d)
}

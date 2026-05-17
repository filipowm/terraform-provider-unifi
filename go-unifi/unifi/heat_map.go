package unifi

import (
	"context"
)

func (c *client) ListHeatMap(ctx context.Context, site string) ([]HeatMap, error) {
	return c.listHeatMap(ctx, site)
}

func (c *client) GetHeatMap(ctx context.Context, site, id string) (*HeatMap, error) {
	return c.getHeatMap(ctx, site, id)
}

func (c *client) DeleteHeatMap(ctx context.Context, site, id string) error {
	return c.deleteHeatMap(ctx, site, id)
}

func (c *client) CreateHeatMap(ctx context.Context, site string, d *HeatMap) (*HeatMap, error) {
	return c.createHeatMap(ctx, site, d)
}

func (c *client) UpdateHeatMap(ctx context.Context, site string, d *HeatMap) (*HeatMap, error) {
	return c.updateHeatMap(ctx, site, d)
}

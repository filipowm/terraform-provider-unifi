package unifi

import (
	"context"
)

func (c *client) ListHeatMapPoint(ctx context.Context, site string) ([]HeatMapPoint, error) {
	return c.listHeatMapPoint(ctx, site)
}

func (c *client) GetHeatMapPoint(ctx context.Context, site, id string) (*HeatMapPoint, error) {
	return c.getHeatMapPoint(ctx, site, id)
}

func (c *client) DeleteHeatMapPoint(ctx context.Context, site, id string) error {
	return c.deleteHeatMapPoint(ctx, site, id)
}

func (c *client) CreateHeatMapPoint(ctx context.Context, site string, d *HeatMapPoint) (*HeatMapPoint, error) {
	return c.createHeatMapPoint(ctx, site, d)
}

func (c *client) UpdateHeatMapPoint(ctx context.Context, site string, d *HeatMapPoint) (*HeatMapPoint, error) {
	return c.updateHeatMapPoint(ctx, site, d)
}

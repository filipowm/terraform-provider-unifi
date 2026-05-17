package unifi

import (
	"context"
)

func (c *client) ListSpatialRecord(ctx context.Context, site string) ([]SpatialRecord, error) {
	return c.listSpatialRecord(ctx, site)
}

func (c *client) GetSpatialRecord(ctx context.Context, site, id string) (*SpatialRecord, error) {
	return c.getSpatialRecord(ctx, site, id)
}

func (c *client) DeleteSpatialRecord(ctx context.Context, site, id string) error {
	return c.deleteSpatialRecord(ctx, site, id)
}

func (c *client) CreateSpatialRecord(ctx context.Context, site string, d *SpatialRecord) (*SpatialRecord, error) {
	return c.createSpatialRecord(ctx, site, d)
}

func (c *client) UpdateSpatialRecord(ctx context.Context, site string, d *SpatialRecord) (*SpatialRecord, error) {
	return c.updateSpatialRecord(ctx, site, d)
}

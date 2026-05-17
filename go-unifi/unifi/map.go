package unifi

import (
	"context"
)

func (c *client) ListMap(ctx context.Context, site string) ([]Map, error) {
	return c.listMap(ctx, site)
}

func (c *client) GetMap(ctx context.Context, site, id string) (*Map, error) {
	return c.getMap(ctx, site, id)
}

func (c *client) DeleteMap(ctx context.Context, site, id string) error {
	return c.deleteMap(ctx, site, id)
}

func (c *client) CreateMap(ctx context.Context, site string, d *Map) (*Map, error) {
	return c.createMap(ctx, site, d)
}

func (c *client) UpdateMap(ctx context.Context, site string, d *Map) (*Map, error) {
	return c.updateMap(ctx, site, d)
}

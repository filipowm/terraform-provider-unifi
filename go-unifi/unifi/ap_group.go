package unifi

import (
	"context"
)

func (c *client) ListAPGroup(ctx context.Context, site string) ([]APGroup, error) {
	return c.listAPGroup(ctx, site)
}

func (c *client) CreateAPGroup(ctx context.Context, site string, d *APGroup) (*APGroup, error) {
	return c.createAPGroup(ctx, site, d)
}

func (c *client) GetAPGroup(ctx context.Context, site, id string) (*APGroup, error) {
	return c.getAPGroup(ctx, site, id)
}

func (c *client) DeleteAPGroup(ctx context.Context, site, id string) error {
	return c.deleteAPGroup(ctx, site, id)
}

func (c *client) UpdateAPGroup(ctx context.Context, site string, d *APGroup) (*APGroup, error) {
	return c.updateAPGroup(ctx, site, d)
}

package unifi

import (
	"context"
)

func (c *client) ListPortProfile(ctx context.Context, site string) ([]PortProfile, error) {
	return c.listPortProfile(ctx, site)
}

func (c *client) GetPortProfile(ctx context.Context, site, id string) (*PortProfile, error) {
	return c.getPortProfile(ctx, site, id)
}

func (c *client) DeletePortProfile(ctx context.Context, site, id string) error {
	return c.deletePortProfile(ctx, site, id)
}

func (c *client) CreatePortProfile(ctx context.Context, site string, d *PortProfile) (*PortProfile, error) {
	return c.createPortProfile(ctx, site, d)
}

func (c *client) UpdatePortProfile(ctx context.Context, site string, d *PortProfile) (*PortProfile, error) {
	return c.updatePortProfile(ctx, site, d)
}

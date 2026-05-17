package unifi

import (
	"context"
)

func (c *client) ListRADIUSProfile(ctx context.Context, site string) ([]RADIUSProfile, error) {
	return c.listRADIUSProfile(ctx, site)
}

func (c *client) GetRADIUSProfile(ctx context.Context, site, id string) (*RADIUSProfile, error) {
	return c.getRADIUSProfile(ctx, site, id)
}

func (c *client) DeleteRADIUSProfile(ctx context.Context, site, id string) error {
	return c.deleteRADIUSProfile(ctx, site, id)
}

func (c *client) CreateRADIUSProfile(ctx context.Context, site string, d *RADIUSProfile) (*RADIUSProfile, error) {
	return c.createRADIUSProfile(ctx, site, d)
}

func (c *client) UpdateRADIUSProfile(ctx context.Context, site string, d *RADIUSProfile) (*RADIUSProfile, error) {
	return c.updateRADIUSProfile(ctx, site, d)
}

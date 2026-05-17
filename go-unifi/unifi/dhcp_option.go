package unifi

import "context"

func (c *client) ListDHCPOption(ctx context.Context, site string) ([]DHCPOption, error) {
	return c.listDHCPOption(ctx, site)
}

func (c *client) CreateDHCPOption(ctx context.Context, site string, d *DHCPOption) (*DHCPOption, error) {
	return c.createDHCPOption(ctx, site, d)
}

func (c *client) GetDHCPOption(ctx context.Context, site, id string) (*DHCPOption, error) {
	return c.getDHCPOption(ctx, site, id)
}

func (c *client) DeleteDHCPOption(ctx context.Context, site, id string) error {
	return c.deleteDHCPOption(ctx, site, id)
}

func (c *client) UpdateDHCPOption(ctx context.Context, site string, d *DHCPOption) (*DHCPOption, error) {
	return c.updateDHCPOption(ctx, site, d)
}

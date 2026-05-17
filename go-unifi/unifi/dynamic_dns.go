package unifi

import "context"

func (c *client) ListDynamicDNS(ctx context.Context, site string) ([]DynamicDNS, error) {
	return c.listDynamicDNS(ctx, site)
}

func (c *client) GetDynamicDNS(ctx context.Context, site, id string) (*DynamicDNS, error) {
	return c.getDynamicDNS(ctx, site, id)
}

func (c *client) DeleteDynamicDNS(ctx context.Context, site, id string) error {
	return c.deleteDynamicDNS(ctx, site, id)
}

func (c *client) CreateDynamicDNS(ctx context.Context, site string, d *DynamicDNS) (*DynamicDNS, error) {
	return c.createDynamicDNS(ctx, site, d)
}

func (c *client) UpdateDynamicDNS(ctx context.Context, site string, d *DynamicDNS) (*DynamicDNS, error) {
	return c.updateDynamicDNS(ctx, site, d)
}

package unifi

import "context"

func (c *client) ListFirewallGroup(ctx context.Context, site string) ([]FirewallGroup, error) {
	return c.listFirewallGroup(ctx, site)
}

func (c *client) GetFirewallGroup(ctx context.Context, site, id string) (*FirewallGroup, error) {
	return c.getFirewallGroup(ctx, site, id)
}

func (c *client) DeleteFirewallGroup(ctx context.Context, site, id string) error {
	return c.deleteFirewallGroup(ctx, site, id)
}

func (c *client) CreateFirewallGroup(ctx context.Context, site string, d *FirewallGroup) (*FirewallGroup, error) {
	return c.createFirewallGroup(ctx, site, d)
}

func (c *client) UpdateFirewallGroup(ctx context.Context, site string, d *FirewallGroup) (*FirewallGroup, error) {
	return c.updateFirewallGroup(ctx, site, d)
}

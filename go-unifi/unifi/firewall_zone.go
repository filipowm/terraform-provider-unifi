package unifi

import "context"

func (c *client) ListFirewallZone(ctx context.Context, site string) ([]FirewallZone, error) {
	return c.listFirewallZone(ctx, site)
}

func (c *client) GetFirewallZone(ctx context.Context, site, id string) (*FirewallZone, error) {
	// client-side filtering is needed, because of lack of endpoint
	zones, err := c.listFirewallZone(ctx, site)
	if err != nil {
		return nil, err
	}

	for _, zone := range zones {
		if zone.ID == id {
			return &zone, nil
		}
	}
	return nil, ErrNotFound
}

func (c *client) DeleteFirewallZone(ctx context.Context, site, id string) error {
	return c.deleteFirewallZone(ctx, site, id)
}

func (c *client) CreateFirewallZone(ctx context.Context, site string, d *FirewallZone) (*FirewallZone, error) {
	return c.createFirewallZone(ctx, site, d)
}

func (c *client) UpdateFirewallZone(ctx context.Context, site string, d *FirewallZone) (*FirewallZone, error) {
	return c.updateFirewallZone(ctx, site, d)
}

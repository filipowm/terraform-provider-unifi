package unifi

import (
	"context"
)

func (c *client) ListFirewallZoneMatrix(ctx context.Context, site string) ([]FirewallZoneMatrix, error) {
	return c.listFirewallZoneMatrix(ctx, site)
}

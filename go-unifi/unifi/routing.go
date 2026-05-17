package unifi

/*
{
    "name": "route1",
    "enabled": true,
    "static-route_network": "10.0.10.0/24",
    "static-route_distance": "2",
    "static-route_type": "nexthop-route",
    "static-route_nexthop": "10.1.2.3",
    "type": "static-route"
}
*/

import (
	"context"
)

func (c *client) ListRouting(ctx context.Context, site string) ([]Routing, error) {
	return c.listRouting(ctx, site)
}

func (c *client) GetRouting(ctx context.Context, site, id string) (*Routing, error) {
	return c.getRouting(ctx, site, id)
}

func (c *client) DeleteRouting(ctx context.Context, site, id string) error {
	return c.deleteRouting(ctx, site, id)
}

func (c *client) CreateRouting(ctx context.Context, site string, d *Routing) (*Routing, error) {
	return c.createRouting(ctx, site, d)
}

func (c *client) UpdateRouting(ctx context.Context, site string, d *Routing) (*Routing, error) {
	return c.updateRouting(ctx, site, d)
}

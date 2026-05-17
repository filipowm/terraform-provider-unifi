package unifi

import (
	"context"
)

func (c *client) ListChannelPlan(ctx context.Context, site string) ([]ChannelPlan, error) {
	return c.listChannelPlan(ctx, site)
}

func (c *client) GetChannelPlan(ctx context.Context, site, id string) (*ChannelPlan, error) {
	return c.getChannelPlan(ctx, site, id)
}

func (c *client) DeleteChannelPlan(ctx context.Context, site, id string) error {
	return c.deleteChannelPlan(ctx, site, id)
}

func (c *client) CreateChannelPlan(ctx context.Context, site string, d *ChannelPlan) (*ChannelPlan, error) {
	return c.createChannelPlan(ctx, site, d)
}

func (c *client) UpdateChannelPlan(ctx context.Context, site string, d *ChannelPlan) (*ChannelPlan, error) {
	return c.updateChannelPlan(ctx, site, d)
}

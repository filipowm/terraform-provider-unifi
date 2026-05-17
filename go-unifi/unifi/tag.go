package unifi

import (
	"context"
)

func (c *client) ListTag(ctx context.Context, site string) ([]Tag, error) {
	return c.listTag(ctx, site)
}

func (c *client) CreateTag(ctx context.Context, site string, d *Tag) (*Tag, error) {
	return c.createTag(ctx, site, d)
}

func (c *client) GetTag(ctx context.Context, site, id string) (*Tag, error) {
	return c.getTag(ctx, site, id)
}

func (c *client) DeleteTag(ctx context.Context, site, id string) error {
	return c.deleteTag(ctx, site, id)
}

func (c *client) UpdateTag(ctx context.Context, site string, d *Tag) (*Tag, error) {
	return c.updateTag(ctx, site, d)
}

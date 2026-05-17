package unifi

import (
	"context"
)

func (c *client) ListMediaFile(ctx context.Context, site string) ([]MediaFile, error) {
	return c.listMediaFile(ctx, site)
}

func (c *client) CreateMediaFile(ctx context.Context, site string, m *MediaFile) (*MediaFile, error) {
	return c.createMediaFile(ctx, site, m)
}

func (c *client) GetMediaFile(ctx context.Context, site, id string) (*MediaFile, error) {
	return c.getMediaFile(ctx, site, id)
}

func (c *client) DeleteMediaFile(ctx context.Context, site, id string) error {
	return c.deleteMediaFile(ctx, site, id)
}

func (c *client) UpdateMediaFile(ctx context.Context, site string, d *MediaFile) (*MediaFile, error) {
	return c.updateMediaFile(ctx, site, d)
}

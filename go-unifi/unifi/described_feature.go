package unifi

import (
	"context"
	"strings"
)

func (c *client) ListFeatures(ctx context.Context, site string) ([]DescribedFeature, error) {
	return c.listDescribedFeature(ctx, site)
}

func (c *client) GetFeature(ctx context.Context, site string, name string) (*DescribedFeature, error) {
	features, err := c.ListFeatures(ctx, site)
	if err != nil {
		return nil, err
	}
	lowerName := strings.ToLower(name)
	for _, f := range features {
		if strings.ToLower(f.Name) == lowerName {
			return &f, nil
		}
	}
	return nil, ErrNotFound
}

func (c *client) IsFeatureEnabled(ctx context.Context, site string, name string) (bool, error) {
	f, err := c.GetFeature(ctx, site, name)
	if err != nil {
		return false, err
	}
	return f.FeatureExists, nil
}

package unifi

import (
	"context"
)

func (c *client) ListVirtualDevice(ctx context.Context, site string) ([]VirtualDevice, error) {
	return c.listVirtualDevice(ctx, site)
}

func (c *client) GetVirtualDevice(ctx context.Context, site, id string) (*VirtualDevice, error) {
	return c.getVirtualDevice(ctx, site, id)
}

func (c *client) DeleteVirtualDevice(ctx context.Context, site, id string) error {
	return c.deleteVirtualDevice(ctx, site, id)
}

func (c *client) CreateVirtualDevice(ctx context.Context, site string, d *VirtualDevice) (*VirtualDevice, error) {
	return c.createVirtualDevice(ctx, site, d)
}

func (c *client) UpdateVirtualDevice(ctx context.Context, site string, d *VirtualDevice) (*VirtualDevice, error) {
	return c.updateVirtualDevice(ctx, site, d)
}

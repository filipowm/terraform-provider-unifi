package unifi

import (
	"context"
)

func (c *client) ListDNSRecord(ctx context.Context, site string) ([]DNSRecord, error) {
	return c.listDNSRecord(ctx, site)
}

func (c *client) GetDNSRecord(ctx context.Context, site, id string) (*DNSRecord, error) {
	// client-side filtering is needed, because of lack of endpoint
	records, err := c.listDNSRecord(ctx, site)
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		if record.ID == id {
			return &record, nil
		}
	}
	return nil, ErrNotFound
}

func (c *client) DeleteDNSRecord(ctx context.Context, site, id string) error {
	return c.deleteDNSRecord(ctx, site, id)
}

func (c *client) CreateDNSRecord(ctx context.Context, site string, d *DNSRecord) (*DNSRecord, error) {
	return c.createDNSRecord(ctx, site, d)
}

func (c *client) UpdateDNSRecord(ctx context.Context, site string, d *DNSRecord) (*DNSRecord, error) {
	return c.updateDNSRecord(ctx, site, d)
}

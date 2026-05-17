package unifi

import (
	"context"
	"fmt"
)

type FirewallPolicyOrderUpdate struct {
	DestinationZoneId   string   `json:"destination_zone_id"`
	SourceZoneId        string   `json:"source_zone_id"`
	AfterPredefinedIds  []string `json:"after_predefined_ids"`
	BeforePredefinedIds []string `json:"before_predefined_ids"`
}

func (c *client) ListFirewallZonePolicy(ctx context.Context, site string) ([]FirewallZonePolicy, error) {
	return c.listFirewallZonePolicy(ctx, site)
}

func (c *client) GetFirewallZonePolicy(ctx context.Context, site, id string) (*FirewallZonePolicy, error) {
	return c.getFirewallZonePolicy(ctx, site, id)
}

func (c *client) DeleteFirewallZonePolicy(ctx context.Context, site, id string) error {
	return c.deleteFirewallZonePolicy(ctx, site, id)
}

func (c *client) CreateFirewallZonePolicy(ctx context.Context, site string, d *FirewallZonePolicy) (*FirewallZonePolicy, error) {
	return c.createFirewallZonePolicy(ctx, site, d)
}

func (c *client) UpdateFirewallZonePolicy(ctx context.Context, site string, d *FirewallZonePolicy) (*FirewallZonePolicy, error) {
	return c.updateFirewallZonePolicy(ctx, site, d)
}

func (c *client) ReorderFirewallPolicies(ctx context.Context, site string, d *FirewallPolicyOrderUpdate) ([]FirewallZonePolicy, error) {
	var res []FirewallZonePolicy
	err := c.Put(ctx, fmt.Sprintf("%s/site/%s/firewall-policies/batch-reorder", c.apiPaths.ApiV2Path, site), d, res)
	if err != nil {
		return nil, err
	}

	// TODO raise error if returned length is not equal to the length of the reordered policies?
	return res, nil
}

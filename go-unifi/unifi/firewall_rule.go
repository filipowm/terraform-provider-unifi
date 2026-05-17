package unifi

import (
	"context"
	"fmt"
)

type FirewallRuleIndexUpdate struct {
	ID        string `json:"_id"`
	RuleIndex int    `json:"rule_index,string"`
}

func (c *client) ListFirewallRule(ctx context.Context, site string) ([]FirewallRule, error) {
	return c.listFirewallRule(ctx, site)
}

func (c *client) GetFirewallRule(ctx context.Context, site, id string) (*FirewallRule, error) {
	return c.getFirewallRule(ctx, site, id)
}

func (c *client) DeleteFirewallRule(ctx context.Context, site, id string) error {
	return c.deleteFirewallRule(ctx, site, id)
}

func (c *client) CreateFirewallRule(ctx context.Context, site string, d *FirewallRule) (*FirewallRule, error) {
	return c.createFirewallRule(ctx, site, d)
}

func (c *client) UpdateFirewallRule(ctx context.Context, site string, d *FirewallRule) (*FirewallRule, error) {
	return c.updateFirewallRule(ctx, site, d)
}

func (c *client) ReorderFirewallRules(ctx context.Context, site, ruleset string, reorder []FirewallRuleIndexUpdate) error {
	reqBody := struct {
		Cmd     string                    `json:"cmd"`
		Ruleset string                    `json:"ruleset"`
		Rules   []FirewallRuleIndexUpdate `json:"rules"`
	}{
		Cmd:     "reorder",
		Ruleset: ruleset,
		Rules:   reorder,
	}
	err := c.Post(ctx, fmt.Sprintf("s/%s/cmd/firewall", site), reqBody, nil)
	if err != nil {
		return err
	}

	return nil
}

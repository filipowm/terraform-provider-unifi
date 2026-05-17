// Code generated from ace.jar fields *.json files
// DO NOT EDIT.

package unifi

import (
	"context"
	"encoding/json"
	"fmt"
)

// just to fix compile issues with the import
var (
	_ context.Context
	_ fmt.Formatter
	_ json.Marshaler
)

type FirewallZonePolicy struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	Hidden   bool   `json:"attr_hidden,omitempty"`
	HiddenID string `json:"attr_hidden_id,omitempty"`
	NoDelete bool   `json:"attr_no_delete,omitempty"`
	NoEdit   bool   `json:"attr_no_edit,omitempty"`

	Action                string                        `json:"action,omitempty" validate:"omitempty,oneof=ALLOW BLOCK REJECT"`                         // ALLOW|BLOCK|REJECT
	ConnectionStateType   string                        `json:"connection_state_type,omitempty" validate:"omitempty,oneof=ALL RESPOND_ONLY CUSTOM"`     // ALL|RESPOND_ONLY|CUSTOM
	ConnectionStates      []string                      `json:"connection_states,omitempty" validate:"omitempty,oneof=ESTABLISHED NEW RELATED INVALID"` // ESTABLISHED|NEW|RELATED|INVALID
	CreateAllowRespond    bool                          `json:"create_allow_respond"`
	Description           string                        `json:"description,omitempty"`
	Destination           FirewallZonePolicyDestination `json:"destination,omitempty"`
	Enabled               bool                          `json:"enabled"`
	IPVersion             string                        `json:"ip_version,omitempty" validate:"omitempty,oneof=BOTH IPV4 IPV6"` // BOTH|IPV4|IPV6
	Index                 int                           `json:"index,omitempty"`                                                // ^[0-9][0-9]?$|^
	Logging               bool                          `json:"logging"`
	MatchIPSec            bool                          `json:"match_ip_sec"`
	MatchIPSecType        string                        `json:"match_ip_sec_type,omitempty" validate:"omitempty,oneof=MATCH_IP_SEC MATCH_NON_IP_SEC"` // MATCH_IP_SEC|MATCH_NON_IP_SEC
	MatchOppositeProtocol bool                          `json:"match_opposite_protocol"`
	Name                  string                        `json:"name,omitempty"`
	Predefined            bool                          `json:"predefined"`
	Protocol              string                        `json:"protocol,omitempty" validate:"omitempty,oneof=all tcp_udp tcp udp ah dccp eigrp esp gre icmp icmpv6 igmp igp ip ipcomp ipip ipv6 isis l2tp manet mobility-header mpls-in-ip number ospf pim pup rdp rohc rspf rcvp sctp shim6 skip st vmtp vrrp wesp xtp"` // all|tcp_udp|tcp|udp|ah|dccp|eigrp|esp|gre|icmp|icmpv6|igmp|igp|ip|ipcomp|ipip|ipv6|isis|l2tp|manet|mobility-header|mpls-in-ip|number|ospf|pim|pup|rdp|rohc|rspf|rcvp|sctp|shim6|skip|st|vmtp|vrrp|wesp|xtp
	Schedule              FirewallZonePolicySchedule    `json:"schedule,omitempty"`
	Source                FirewallZonePolicySource      `json:"source,omitempty"`
}

func (dst *FirewallZonePolicy) UnmarshalJSON(b []byte) error {
	type Alias FirewallZonePolicy
	aux := &struct {
		Index emptyStringInt `json:"index"`

		*Alias
	}{
		Alias: (*Alias)(dst),
	}

	err := json.Unmarshal(b, &aux)
	if err != nil {
		return fmt.Errorf("unable to unmarshal alias: %w", err)
	}
	dst.Index = int(aux.Index)

	return nil
}

type FirewallZonePolicyDestination struct {
	AppCategoryIDs     []string `json:"app_category_ids,omitempty"`
	AppIDs             []string `json:"app_ids,omitempty"`
	IPGroupID          string   `json:"ip_group_id,omitempty"`
	IPs                []string `json:"ips,omitempty" validate:"omitempty,ipv4"` // ^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$|^$
	MatchOppositeIPs   bool     `json:"match_opposite_ips"`
	MatchOppositePorts bool     `json:"match_opposite_ports"`
	MatchingTarget     string   `json:"matching_target,omitempty" validate:"omitempty,oneof=ANY APP APP_CATEGORY IP NETWORK REGION WEB"` // ANY|APP|APP_CATEGORY|IP|NETWORK|REGION|WEB
	MatchingTargetType string   `json:"matching_target_type,omitempty" validate:"omitempty,oneof=ANY OBJECT SPECIFIC"`                   // ANY|OBJECT|SPECIFIC
	NetworkIDs         []string `json:"network_ids,omitempty"`
	Port               int      `json:"port,omitempty"` // ^[0-9][0-9]?$|^
	PortGroupID        string   `json:"port_group_id,omitempty"`
	PortMatchingType   string   `json:"port_matching_type,omitempty" validate:"omitempty,oneof=ANY SPECIFIC OBJECT"` // ANY|SPECIFIC|OBJECT
	Regions            []string `json:"regions,omitempty"`
	WebDomains         []string `json:"web_domains,omitempty"`
	ZoneID             string   `json:"zone_id"`
}

func (dst *FirewallZonePolicyDestination) UnmarshalJSON(b []byte) error {
	type Alias FirewallZonePolicyDestination
	aux := &struct {
		Port emptyStringInt `json:"port"`

		*Alias
	}{
		Alias: (*Alias)(dst),
	}

	err := json.Unmarshal(b, &aux)
	if err != nil {
		return fmt.Errorf("unable to unmarshal alias: %w", err)
	}
	dst.Port = int(aux.Port)

	return nil
}

type FirewallZonePolicySchedule struct {
	Date           string   `json:"date,omitempty"`                                                                             // ^$|^(20[0-9]{2})-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$
	DateEnd        string   `json:"date_end,omitempty"`                                                                         // ^$|^(20[0-9]{2})-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$
	DateStart      string   `json:"date_start,omitempty"`                                                                       // ^$|^(20[0-9]{2})-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$
	Mode           string   `json:"mode,omitempty" validate:"omitempty,oneof=ALWAYS EVERY_DAY EVERY_WEEK ONE_TIME_ONLY CUSTOM"` // ALWAYS|EVERY_DAY|EVERY_WEEK|ONE_TIME_ONLY|CUSTOM
	RepeatOnDays   []string `json:"repeat_on_days,omitempty" validate:"omitempty,oneof=mon tue wed thu fri sat sun"`            // mon|tue|wed|thu|fri|sat|sun
	TimeAllDay     bool     `json:"time_all_day"`
	TimeRangeEnd   string   `json:"time_range_end,omitempty"`   // ^[0-9][0-9]:[0-9][0-9]$
	TimeRangeStart string   `json:"time_range_start,omitempty"` // ^[0-9][0-9]:[0-9][0-9]$
}

func (dst *FirewallZonePolicySchedule) UnmarshalJSON(b []byte) error {
	type Alias FirewallZonePolicySchedule
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(dst),
	}

	err := json.Unmarshal(b, &aux)
	if err != nil {
		return fmt.Errorf("unable to unmarshal alias: %w", err)
	}

	return nil
}

type FirewallZonePolicySource struct {
	ClientMACs            []string `json:"client_macs,omitempty" validate:"omitempty,mac"` // ^([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})$
	IPGroupID             string   `json:"ip_group_id,omitempty"`
	IPs                   []string `json:"ips,omitempty" validate:"omitempty,ipv4"` // ^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$|^$
	MAC                   string   `json:"mac,omitempty" validate:"omitempty,mac"`  // ^([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})$
	MACs                  []string `json:"macs,omitempty" validate:"omitempty,mac"` // ^([0-9A-Fa-f]{2}:){5}([0-9A-Fa-f]{2})$
	MatchMAC              bool     `json:"match_mac"`
	MatchOppositeIPs      bool     `json:"match_opposite_ips"`
	MatchOppositeNetworks bool     `json:"match_opposite_networks"`
	MatchOppositePorts    bool     `json:"match_opposite_ports"`
	MatchingTarget        string   `json:"matching_target,omitempty" validate:"omitempty,oneof=ANY CLIENT NETWORK IP MAC"` // ANY|CLIENT|NETWORK|IP|MAC
	MatchingTargetType    string   `json:"matching_target_type,omitempty" validate:"omitempty,oneof=OBJECT SPECIFIC"`      // OBJECT|SPECIFIC
	NetworkIDs            []string `json:"network_ids,omitempty"`
	Port                  int      `json:"port,omitempty"` // ^[0-9][0-9]?$|^
	PortGroupID           string   `json:"port_group_id,omitempty"`
	PortMatchingType      string   `json:"port_matching_type,omitempty" validate:"omitempty,oneof=ANY SPECIFIC OBJECT"` // ANY|SPECIFIC|OBJECT
	ZoneID                string   `json:"zone_id"`
}

func (dst *FirewallZonePolicySource) UnmarshalJSON(b []byte) error {
	type Alias FirewallZonePolicySource
	aux := &struct {
		Port emptyStringInt `json:"port"`

		*Alias
	}{
		Alias: (*Alias)(dst),
	}

	err := json.Unmarshal(b, &aux)
	if err != nil {
		return fmt.Errorf("unable to unmarshal alias: %w", err)
	}
	dst.Port = int(aux.Port)

	return nil
}

func (c *client) listFirewallZonePolicy(ctx context.Context, site string) ([]FirewallZonePolicy, error) {
	var respBody []FirewallZonePolicy

	err := c.Get(ctx, fmt.Sprintf("%s/site/%s/firewall-policies", c.apiPaths.ApiV2Path, site), nil, &respBody)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

func (c *client) getFirewallZonePolicy(ctx context.Context, site, id string) (*FirewallZonePolicy, error) {
	var respBody FirewallZonePolicy

	err := c.Get(ctx, fmt.Sprintf("%s/site/%s/firewall-policies/%s", c.apiPaths.ApiV2Path, site, id), nil, &respBody)

	if err != nil {
		return nil, err
	}
	if respBody.ID == "" {
		return nil, ErrNotFound
	}
	return &respBody, nil
}

func (c *client) deleteFirewallZonePolicy(ctx context.Context, site, id string) error {
	err := c.Delete(ctx, fmt.Sprintf("%s/site/%s/firewall-policies/%s", c.apiPaths.ApiV2Path, site, id), struct{}{}, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *client) createFirewallZonePolicy(ctx context.Context, site string, d *FirewallZonePolicy) (*FirewallZonePolicy, error) {
	var respBody FirewallZonePolicy

	err := c.Post(ctx, fmt.Sprintf("%s/site/%s/firewall-policies", c.apiPaths.ApiV2Path, site), d, &respBody)
	if err != nil {
		return nil, err
	}

	return &respBody, nil
}

func (c *client) updateFirewallZonePolicy(ctx context.Context, site string, d *FirewallZonePolicy) (*FirewallZonePolicy, error) {
	var respBody FirewallZonePolicy

	err := c.Put(ctx, fmt.Sprintf("%s/site/%s/firewall-policies/%s", c.apiPaths.ApiV2Path, site, d.ID), d, &respBody)
	if err != nil {
		return nil, err
	}
	return &respBody, nil
}

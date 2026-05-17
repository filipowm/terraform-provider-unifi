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

const SettingIpsKey = "ips"

type SettingIps struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	Hidden   bool   `json:"attr_hidden,omitempty"`
	HiddenID string `json:"attr_hidden_id,omitempty"`
	NoDelete bool   `json:"attr_no_delete,omitempty"`
	NoEdit   bool   `json:"attr_no_edit,omitempty"`

	Key string `json:"key"`

	AdBlockingConfigurations    []SettingIpsAdBlockingConfigurations `json:"ad_blocking_configurations,omitempty"`
	AdBlockingEnabled           bool                                 `json:"ad_blocking_enabled"`
	AdvancedFilteringPreference string                               `json:"advanced_filtering_preference,omitempty" validate:"omitempty,oneof=manual disabled"` // |manual|disabled
	DNSFiltering                bool                                 `json:"dns_filtering"`
	DNSFilters                  []SettingIpsDNSFilters               `json:"dns_filters,omitempty"`
	EnabledCategories           []string                             `json:"enabled_categories,omitempty" validate:"omitempty,oneof=emerging-activex emerging-attackresponse botcc emerging-chat ciarmy compromised emerging-dns emerging-dos dshield emerging-exploit emerging-ftp emerging-games emerging-icmp emerging-icmpinfo emerging-imap emerging-inappropriate emerging-info emerging-malware emerging-misc emerging-mobile emerging-netbios emerging-p2p emerging-policy emerging-pop3 emerging-rpc emerging-scada emerging-scan emerging-shellcode emerging-smtp emerging-snmp emerging-sql emerging-telnet emerging-tftp tor emerging-useragent emerging-voip emerging-webapps emerging-webclient emerging-webserver emerging-worm exploit-kit adware-pup botcc-portgrouped phishing threatview-cs-c2 3coresec chat coinminer current-events drop hunting icmp-info inappropriate info ja3 policy scada dark-web-blocker-list malicious-hosts"` // emerging-activex|emerging-attackresponse|botcc|emerging-chat|ciarmy|compromised|emerging-dns|emerging-dos|dshield|emerging-exploit|emerging-ftp|emerging-games|emerging-icmp|emerging-icmpinfo|emerging-imap|emerging-inappropriate|emerging-info|emerging-malware|emerging-misc|emerging-mobile|emerging-netbios|emerging-p2p|emerging-policy|emerging-pop3|emerging-rpc|emerging-scada|emerging-scan|emerging-shellcode|emerging-smtp|emerging-snmp|emerging-sql|emerging-telnet|emerging-tftp|tor|emerging-useragent|emerging-voip|emerging-webapps|emerging-webclient|emerging-webserver|emerging-worm|exploit-kit|adware-pup|botcc-portgrouped|phishing|threatview-cs-c2|3coresec|chat|coinminer|current-events|drop|hunting|icmp-info|inappropriate|info|ja3|policy|scada|dark-web-blocker-list|malicious-hosts
	EnabledNetworks             []string                             `json:"enabled_networks,omitempty"`
	Honeypot                    []SettingIpsHoneypot                 `json:"honeypot"`
	HoneypotEnabled             bool                                 `json:"honeypot_enabled"`
	IPsMode                     string                               `json:"ips_mode,omitempty" validate:"omitempty,oneof=ids ips ipsInline disabled"` // ids|ips|ipsInline|disabled
	MemoryOptimized             bool                                 `json:"memory_optimized"`
	RestrictTorrents            bool                                 `json:"restrict_torrents"`
	Suppression                 SettingIpsSuppression                `json:"suppression,omitempty"`
}

func (dst *SettingIps) UnmarshalJSON(b []byte) error {
	type Alias SettingIps
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

type SettingIpsAdBlockingConfigurations struct {
	NetworkID string `json:"network_id"`
}

func (dst *SettingIpsAdBlockingConfigurations) UnmarshalJSON(b []byte) error {
	type Alias SettingIpsAdBlockingConfigurations
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

type SettingIpsAlerts struct {
	Category  string               `json:"category,omitempty"`
	Gid       int                  `json:"gid,omitempty"`
	ID        int                  `json:"id,omitempty"`
	Signature string               `json:"signature,omitempty"`
	Tracking  []SettingIpsTracking `json:"tracking,omitempty"`
	Type      string               `json:"type,omitempty" validate:"omitempty,oneof=all track"` // all|track
}

func (dst *SettingIpsAlerts) UnmarshalJSON(b []byte) error {
	type Alias SettingIpsAlerts
	aux := &struct {
		Gid emptyStringInt `json:"gid"`
		ID  emptyStringInt `json:"id"`

		*Alias
	}{
		Alias: (*Alias)(dst),
	}

	err := json.Unmarshal(b, &aux)
	if err != nil {
		return fmt.Errorf("unable to unmarshal alias: %w", err)
	}
	dst.Gid = int(aux.Gid)
	dst.ID = int(aux.ID)

	return nil
}

type SettingIpsDNSFilters struct {
	AllowedSites []string `json:"allowed_sites,omitempty"` // ^[a-zA-Z0-9.-]+$|^$
	BlockedSites []string `json:"blocked_sites,omitempty"` // ^[a-zA-Z0-9.-]+$|^$
	BlockedTld   []string `json:"blocked_tld,omitempty"`   // ^[a-zA-Z0-9.-]+$|^$
	Description  string   `json:"description,omitempty"`
	Filter       string   `json:"filter,omitempty" validate:"omitempty,oneof=none work family"` // none|work|family
	Name         string   `json:"name,omitempty"`
	NetworkID    string   `json:"network_id"`
	Version      string   `json:"version,omitempty" validate:"omitempty,oneof=v4 v6"` // v4|v6
}

func (dst *SettingIpsDNSFilters) UnmarshalJSON(b []byte) error {
	type Alias SettingIpsDNSFilters
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

type SettingIpsHoneypot struct {
	IPAddress string `json:"ip_address,omitempty"`
	NetworkID string `json:"network_id"`
	Version   string `json:"version,omitempty" validate:"omitempty,oneof=v4 v6"` // v4|v6
}

func (dst *SettingIpsHoneypot) UnmarshalJSON(b []byte) error {
	type Alias SettingIpsHoneypot
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

type SettingIpsSuppression struct {
	Alerts    []SettingIpsAlerts    `json:"alerts,omitempty"`
	Whitelist []SettingIpsWhitelist `json:"whitelist,omitempty"`
}

func (dst *SettingIpsSuppression) UnmarshalJSON(b []byte) error {
	type Alias SettingIpsSuppression
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

type SettingIpsTracking struct {
	Direction string `json:"direction,omitempty" validate:"omitempty,oneof=both src dest"` // both|src|dest
	Mode      string `json:"mode,omitempty" validate:"omitempty,oneof=ip subnet network"`  // ip|subnet|network
	Value     string `json:"value,omitempty"`
}

func (dst *SettingIpsTracking) UnmarshalJSON(b []byte) error {
	type Alias SettingIpsTracking
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

type SettingIpsWhitelist struct {
	Direction string `json:"direction,omitempty" validate:"omitempty,oneof=both src dest"` // both|src|dest
	Mode      string `json:"mode,omitempty" validate:"omitempty,oneof=ip subnet network"`  // ip|subnet|network
	Value     string `json:"value,omitempty"`
}

func (dst *SettingIpsWhitelist) UnmarshalJSON(b []byte) error {
	type Alias SettingIpsWhitelist
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

// GetSettingIps Experimental! This function is not yet stable and may change in the future.
func (c *client) GetSettingIps(ctx context.Context, site string) (*SettingIps, error) {
	s, f, err := c.GetSetting(ctx, site, SettingIpsKey)
	if err != nil {
		return nil, err
	}
	if s.Key != SettingIpsKey {
		return nil, fmt.Errorf("unexpected setting key received. Requested: %q, received: %q", SettingIpsKey, s.Key)
	}
	return f.(*SettingIps), nil
}

// UpdateSettingIps Experimental! This function is not yet stable and may change in the future.
func (c *client) UpdateSettingIps(ctx context.Context, site string, s *SettingIps) (*SettingIps, error) {
	s.Key = SettingIpsKey
	result, err := c.SetSetting(ctx, site, SettingIpsKey, s)
	if err != nil {
		return nil, err
	}
	return result.(*SettingIps), nil
}

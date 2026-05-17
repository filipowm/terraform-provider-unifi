// Code generated from ace.jar fields *.json files
// DO NOT EDIT.

package unifi

import (
	"context"
	"io"
)

type Client interface {
	Logger

	// BaseURL returns the base URL of the controller.
	BaseURL() string

	// Delete sends a DELETE request to the controller.
	Delete(ctx context.Context, apiPath string, reqBody interface{}, respBody interface{}) error

	// Do sends a request to the controller.
	Do(ctx context.Context, method string, apiPath string, reqBody interface{}, respBody interface{}) error

	// Get sends a GET request to the controller.
	Get(ctx context.Context, apiPath string, reqBody interface{}, respBody interface{}) error

	// Login logs in to the controller. Useful only for user/password authentication.
	Login() error

	// Logout logs out from the controller.
	Logout() error

	// Post sends a POST request to the controller.
	Post(ctx context.Context, apiPath string, reqBody interface{}, respBody interface{}) error

	// Put sends a PUT request to the controller.
	Put(ctx context.Context, apiPath string, reqBody interface{}, respBody interface{}) error

	// Version returns the version of the UniFi Controller API.
	Version() string

	// ==== client methods for APGroup resource ====

	// CreateAPGroup creates a resource
	CreateAPGroup(ctx context.Context, site string, a *APGroup) (*APGroup, error)

	// DeleteAPGroup deletes a resource
	DeleteAPGroup(ctx context.Context, site string, id string) error

	// GetAPGroup retrieves a resource
	GetAPGroup(ctx context.Context, site string, id string) (*APGroup, error)

	// ListAPGroup lists the resources
	ListAPGroup(ctx context.Context, site string) ([]APGroup, error)

	// UpdateAPGroup updates a resource
	UpdateAPGroup(ctx context.Context, site string, a *APGroup) (*APGroup, error)

	// ==== end of client methods for APGroup resource ====

	// ==== client methods for Account resource ====

	// CreateAccount creates a resource
	CreateAccount(ctx context.Context, site string, a *Account) (*Account, error)

	// DeleteAccount deletes a resource
	DeleteAccount(ctx context.Context, site string, id string) error

	// GetAccount retrieves a resource
	GetAccount(ctx context.Context, site string, id string) (*Account, error)

	// ListAccount lists the resources
	ListAccount(ctx context.Context, site string) ([]Account, error)

	// UpdateAccount updates a resource
	UpdateAccount(ctx context.Context, site string, a *Account) (*Account, error)

	// ==== end of client methods for Account resource ====

	// ==== client methods for BroadcastGroup resource ====

	// CreateBroadcastGroup creates a resource
	CreateBroadcastGroup(ctx context.Context, site string, b *BroadcastGroup) (*BroadcastGroup, error)

	// DeleteBroadcastGroup deletes a resource
	DeleteBroadcastGroup(ctx context.Context, site string, id string) error

	// GetBroadcastGroup retrieves a resource
	GetBroadcastGroup(ctx context.Context, site string, id string) (*BroadcastGroup, error)

	// ListBroadcastGroup lists the resources
	ListBroadcastGroup(ctx context.Context, site string) ([]BroadcastGroup, error)

	// UpdateBroadcastGroup updates a resource
	UpdateBroadcastGroup(ctx context.Context, site string, b *BroadcastGroup) (*BroadcastGroup, error)

	// ==== end of client methods for BroadcastGroup resource ====

	// ==== client methods for ChannelPlan resource ====

	// CreateChannelPlan creates a resource
	CreateChannelPlan(ctx context.Context, site string, c *ChannelPlan) (*ChannelPlan, error)

	// DeleteChannelPlan deletes a resource
	DeleteChannelPlan(ctx context.Context, site string, id string) error

	// GetChannelPlan retrieves a resource
	GetChannelPlan(ctx context.Context, site string, id string) (*ChannelPlan, error)

	// ListChannelPlan lists the resources
	ListChannelPlan(ctx context.Context, site string) ([]ChannelPlan, error)

	// UpdateChannelPlan updates a resource
	UpdateChannelPlan(ctx context.Context, site string, c *ChannelPlan) (*ChannelPlan, error)

	// ==== end of client methods for ChannelPlan resource ====

	// ==== client methods for DHCPOption resource ====

	// CreateDHCPOption creates a resource
	CreateDHCPOption(ctx context.Context, site string, d *DHCPOption) (*DHCPOption, error)

	// DeleteDHCPOption deletes a resource
	DeleteDHCPOption(ctx context.Context, site string, id string) error

	// GetDHCPOption retrieves a resource
	GetDHCPOption(ctx context.Context, site string, id string) (*DHCPOption, error)

	// ListDHCPOption lists the resources
	ListDHCPOption(ctx context.Context, site string) ([]DHCPOption, error)

	// UpdateDHCPOption updates a resource
	UpdateDHCPOption(ctx context.Context, site string, d *DHCPOption) (*DHCPOption, error)

	// ==== end of client methods for DHCPOption resource ====

	// ==== client methods for DNSRecord resource ====

	// CreateDNSRecord creates a resource
	CreateDNSRecord(ctx context.Context, site string, d *DNSRecord) (*DNSRecord, error)

	// DeleteDNSRecord deletes a resource
	DeleteDNSRecord(ctx context.Context, site string, id string) error

	// GetDNSRecord retrieves a resource
	GetDNSRecord(ctx context.Context, site string, id string) (*DNSRecord, error)

	// ListDNSRecord lists the resources
	ListDNSRecord(ctx context.Context, site string) ([]DNSRecord, error)

	// UpdateDNSRecord updates a resource
	UpdateDNSRecord(ctx context.Context, site string, d *DNSRecord) (*DNSRecord, error)

	// ==== end of client methods for DNSRecord resource ====

	// ==== client methods for Dashboard resource ====

	// CreateDashboard creates a resource
	CreateDashboard(ctx context.Context, site string, d *Dashboard) (*Dashboard, error)

	// DeleteDashboard deletes a resource
	DeleteDashboard(ctx context.Context, site string, id string) error

	// GetDashboard retrieves a resource
	GetDashboard(ctx context.Context, site string, id string) (*Dashboard, error)

	// ListDashboard lists the resources
	ListDashboard(ctx context.Context, site string) ([]Dashboard, error)

	// UpdateDashboard updates a resource
	UpdateDashboard(ctx context.Context, site string, d *Dashboard) (*Dashboard, error)

	// ==== end of client methods for Dashboard resource ====

	// GetFeature returns a specific feature by it's name. Name is case-insensitive.
	GetFeature(ctx context.Context, site string, name string) (*DescribedFeature, error)

	// IsFeatureEnabled returns if a specific feature is enabled by it's name. Name is case-insensitive.
	IsFeatureEnabled(ctx context.Context, site string, name string) (bool, error)

	// ListFeatures returns all features of the UniFi controller.
	ListFeatures(ctx context.Context, site string) ([]DescribedFeature, error)

	// ==== client methods for Device resource ====

	// AdoptDevice adopts a device by MAC address.
	AdoptDevice(ctx context.Context, site string, mac string) error

	// CreateDevice creates a resource
	CreateDevice(ctx context.Context, site string, d *Device) (*Device, error)

	// DeleteDevice deletes a resource
	DeleteDevice(ctx context.Context, site string, id string) error

	// ForgetDevice forgets a device by MAC address.
	ForgetDevice(ctx context.Context, site string, mac string) error

	// GetDevice retrieves a resource
	GetDevice(ctx context.Context, site string, id string) (*Device, error)

	GetDeviceByMAC(ctx context.Context, site string, mac string) (*Device, error)

	// ListDevice lists the resources
	ListDevice(ctx context.Context, site string) ([]Device, error)

	// UpdateDevice updates a resource
	UpdateDevice(ctx context.Context, site string, d *Device) (*Device, error)

	// ==== end of client methods for Device resource ====

	// ==== client methods for DynamicDNS resource ====

	// CreateDynamicDNS creates a resource
	CreateDynamicDNS(ctx context.Context, site string, d *DynamicDNS) (*DynamicDNS, error)

	// DeleteDynamicDNS deletes a resource
	DeleteDynamicDNS(ctx context.Context, site string, id string) error

	// GetDynamicDNS retrieves a resource
	GetDynamicDNS(ctx context.Context, site string, id string) (*DynamicDNS, error)

	// ListDynamicDNS lists the resources
	ListDynamicDNS(ctx context.Context, site string) ([]DynamicDNS, error)

	// UpdateDynamicDNS updates a resource
	UpdateDynamicDNS(ctx context.Context, site string, d *DynamicDNS) (*DynamicDNS, error)

	// ==== end of client methods for DynamicDNS resource ====

	// ==== client methods for FirewallGroup resource ====

	// CreateFirewallGroup creates a resource
	CreateFirewallGroup(ctx context.Context, site string, f *FirewallGroup) (*FirewallGroup, error)

	// DeleteFirewallGroup deletes a resource
	DeleteFirewallGroup(ctx context.Context, site string, id string) error

	// GetFirewallGroup retrieves a resource
	GetFirewallGroup(ctx context.Context, site string, id string) (*FirewallGroup, error)

	// ListFirewallGroup lists the resources
	ListFirewallGroup(ctx context.Context, site string) ([]FirewallGroup, error)

	// UpdateFirewallGroup updates a resource
	UpdateFirewallGroup(ctx context.Context, site string, f *FirewallGroup) (*FirewallGroup, error)

	// ==== end of client methods for FirewallGroup resource ====

	// ==== client methods for FirewallRule resource ====

	// CreateFirewallRule creates a resource
	CreateFirewallRule(ctx context.Context, site string, f *FirewallRule) (*FirewallRule, error)

	// DeleteFirewallRule deletes a resource
	DeleteFirewallRule(ctx context.Context, site string, id string) error

	// GetFirewallRule retrieves a resource
	GetFirewallRule(ctx context.Context, site string, id string) (*FirewallRule, error)

	// ListFirewallRule lists the resources
	ListFirewallRule(ctx context.Context, site string) ([]FirewallRule, error)

	ReorderFirewallRules(ctx context.Context, site string, ruleset string, reorder []FirewallRuleIndexUpdate) error

	// UpdateFirewallRule updates a resource
	UpdateFirewallRule(ctx context.Context, site string, f *FirewallRule) (*FirewallRule, error)

	// ==== end of client methods for FirewallRule resource ====

	// ==== client methods for FirewallZone resource ====

	// CreateFirewallZone creates a resource
	CreateFirewallZone(ctx context.Context, site string, f *FirewallZone) (*FirewallZone, error)

	// DeleteFirewallZone deletes a resource
	DeleteFirewallZone(ctx context.Context, site string, id string) error

	// GetFirewallZone retrieves a resource
	GetFirewallZone(ctx context.Context, site string, id string) (*FirewallZone, error)

	// ListFirewallZone lists the resources
	ListFirewallZone(ctx context.Context, site string) ([]FirewallZone, error)

	// UpdateFirewallZone updates a resource
	UpdateFirewallZone(ctx context.Context, site string, f *FirewallZone) (*FirewallZone, error)

	ListFirewallZoneMatrix(ctx context.Context, site string) ([]FirewallZoneMatrix, error)

	// ==== client methods for FirewallZonePolicy resource ====

	// CreateFirewallZonePolicy creates a resource
	CreateFirewallZonePolicy(ctx context.Context, site string, f *FirewallZonePolicy) (*FirewallZonePolicy, error)

	// DeleteFirewallZonePolicy deletes a resource
	DeleteFirewallZonePolicy(ctx context.Context, site string, id string) error

	// GetFirewallZonePolicy retrieves a resource
	GetFirewallZonePolicy(ctx context.Context, site string, id string) (*FirewallZonePolicy, error)

	// ListFirewallZonePolicy lists the resources
	ListFirewallZonePolicy(ctx context.Context, site string) ([]FirewallZonePolicy, error)

	// UpdateFirewallZonePolicy updates a resource
	UpdateFirewallZonePolicy(ctx context.Context, site string, f *FirewallZonePolicy) (*FirewallZonePolicy, error)

	// ==== end of client methods for FirewallZonePolicy resource ====

	// ==== end of client methods for FirewallZone resource ====

	// ==== client methods for HeatMap resource ====

	// CreateHeatMap creates a resource
	CreateHeatMap(ctx context.Context, site string, h *HeatMap) (*HeatMap, error)

	// DeleteHeatMap deletes a resource
	DeleteHeatMap(ctx context.Context, site string, id string) error

	// GetHeatMap retrieves a resource
	GetHeatMap(ctx context.Context, site string, id string) (*HeatMap, error)

	// ListHeatMap lists the resources
	ListHeatMap(ctx context.Context, site string) ([]HeatMap, error)

	// UpdateHeatMap updates a resource
	UpdateHeatMap(ctx context.Context, site string, h *HeatMap) (*HeatMap, error)

	// ==== client methods for HeatMapPoint resource ====

	// CreateHeatMapPoint creates a resource
	CreateHeatMapPoint(ctx context.Context, site string, h *HeatMapPoint) (*HeatMapPoint, error)

	// DeleteHeatMapPoint deletes a resource
	DeleteHeatMapPoint(ctx context.Context, site string, id string) error

	// GetHeatMapPoint retrieves a resource
	GetHeatMapPoint(ctx context.Context, site string, id string) (*HeatMapPoint, error)

	// ListHeatMapPoint lists the resources
	ListHeatMapPoint(ctx context.Context, site string) ([]HeatMapPoint, error)

	// UpdateHeatMapPoint updates a resource
	UpdateHeatMapPoint(ctx context.Context, site string, h *HeatMapPoint) (*HeatMapPoint, error)

	// ==== end of client methods for HeatMapPoint resource ====

	// ==== end of client methods for HeatMap resource ====

	// ==== client methods for Hotspot2Conf resource ====

	// CreateHotspot2Conf creates a resource
	CreateHotspot2Conf(ctx context.Context, site string, h *Hotspot2Conf) (*Hotspot2Conf, error)

	// DeleteHotspot2Conf deletes a resource
	DeleteHotspot2Conf(ctx context.Context, site string, id string) error

	// GetHotspot2Conf retrieves a resource
	GetHotspot2Conf(ctx context.Context, site string, id string) (*Hotspot2Conf, error)

	// ListHotspot2Conf lists the resources
	ListHotspot2Conf(ctx context.Context, site string) ([]Hotspot2Conf, error)

	// UpdateHotspot2Conf updates a resource
	UpdateHotspot2Conf(ctx context.Context, site string, h *Hotspot2Conf) (*Hotspot2Conf, error)

	// ==== end of client methods for Hotspot2Conf resource ====

	// ==== client methods for HotspotOp resource ====

	// CreateHotspotOp creates a resource
	CreateHotspotOp(ctx context.Context, site string, h *HotspotOp) (*HotspotOp, error)

	// DeleteHotspotOp deletes a resource
	DeleteHotspotOp(ctx context.Context, site string, id string) error

	// GetHotspotOp retrieves a resource
	GetHotspotOp(ctx context.Context, site string, id string) (*HotspotOp, error)

	// ListHotspotOp lists the resources
	ListHotspotOp(ctx context.Context, site string) ([]HotspotOp, error)

	// UpdateHotspotOp updates a resource
	UpdateHotspotOp(ctx context.Context, site string, h *HotspotOp) (*HotspotOp, error)

	// ==== end of client methods for HotspotOp resource ====

	// ==== client methods for HotspotPackage resource ====

	// CreateHotspotPackage creates a resource
	CreateHotspotPackage(ctx context.Context, site string, h *HotspotPackage) (*HotspotPackage, error)

	// DeleteHotspotPackage deletes a resource
	DeleteHotspotPackage(ctx context.Context, site string, id string) error

	// GetHotspotPackage retrieves a resource
	GetHotspotPackage(ctx context.Context, site string, id string) (*HotspotPackage, error)

	// ListHotspotPackage lists the resources
	ListHotspotPackage(ctx context.Context, site string) ([]HotspotPackage, error)

	// UpdateHotspotPackage updates a resource
	UpdateHotspotPackage(ctx context.Context, site string, h *HotspotPackage) (*HotspotPackage, error)

	// ==== end of client methods for HotspotPackage resource ====

	// ==== client methods for Map resource ====

	// CreateMap creates a resource
	CreateMap(ctx context.Context, site string, m *Map) (*Map, error)

	// DeleteMap deletes a resource
	DeleteMap(ctx context.Context, site string, id string) error

	// GetMap retrieves a resource
	GetMap(ctx context.Context, site string, id string) (*Map, error)

	// ListMap lists the resources
	ListMap(ctx context.Context, site string) ([]Map, error)

	// UpdateMap updates a resource
	UpdateMap(ctx context.Context, site string, m *Map) (*Map, error)

	// ==== end of client methods for Map resource ====

	// ==== client methods for MediaFile resource ====

	// CreateMediaFile creates a resource
	CreateMediaFile(ctx context.Context, site string, m *MediaFile) (*MediaFile, error)

	// DeleteMediaFile deletes a resource
	DeleteMediaFile(ctx context.Context, site string, id string) error

	// GetMediaFile retrieves a resource
	GetMediaFile(ctx context.Context, site string, id string) (*MediaFile, error)

	// ListMediaFile lists the resources
	ListMediaFile(ctx context.Context, site string) ([]MediaFile, error)

	// UpdateMediaFile updates a resource
	UpdateMediaFile(ctx context.Context, site string, m *MediaFile) (*MediaFile, error)

	// ==== end of client methods for MediaFile resource ====

	// ==== client methods for Network resource ====

	// CreateNetwork creates a resource
	CreateNetwork(ctx context.Context, site string, n *Network) (*Network, error)

	// DeleteNetwork deletes a resource
	DeleteNetwork(ctx context.Context, site string, id string) error

	// GetNetwork retrieves a resource
	GetNetwork(ctx context.Context, site string, id string) (*Network, error)

	// ListNetwork lists the resources
	ListNetwork(ctx context.Context, site string) ([]Network, error)

	// UpdateNetwork updates a resource
	UpdateNetwork(ctx context.Context, site string, n *Network) (*Network, error)

	// ==== end of client methods for Network resource ====

	// ==== client methods for PortForward resource ====

	// CreatePortForward creates a resource
	CreatePortForward(ctx context.Context, site string, p *PortForward) (*PortForward, error)

	// DeletePortForward deletes a resource
	DeletePortForward(ctx context.Context, site string, id string) error

	// GetPortForward retrieves a resource
	GetPortForward(ctx context.Context, site string, id string) (*PortForward, error)

	// ListPortForward lists the resources
	ListPortForward(ctx context.Context, site string) ([]PortForward, error)

	// UpdatePortForward updates a resource
	UpdatePortForward(ctx context.Context, site string, p *PortForward) (*PortForward, error)

	// ==== end of client methods for PortForward resource ====

	// ==== client methods for PortProfile resource ====

	// CreatePortProfile creates a resource
	CreatePortProfile(ctx context.Context, site string, p *PortProfile) (*PortProfile, error)

	// DeletePortProfile deletes a resource
	DeletePortProfile(ctx context.Context, site string, id string) error

	// GetPortProfile retrieves a resource
	GetPortProfile(ctx context.Context, site string, id string) (*PortProfile, error)

	// ListPortProfile lists the resources
	ListPortProfile(ctx context.Context, site string) ([]PortProfile, error)

	// UpdatePortProfile updates a resource
	UpdatePortProfile(ctx context.Context, site string, p *PortProfile) (*PortProfile, error)

	// ==== end of client methods for PortProfile resource ====

	// DeletePortalFile deletes a Hotspot Portal file from the controller.
	DeletePortalFile(ctx context.Context, site string, id string) error

	// GetPortalFile returns a specific Hotspot Portal file by it's ID.
	GetPortalFile(ctx context.Context, site string, id string) (*PortalFile, error)

	// ListPortalFiles lists all Hotspot Portal files on the controller.
	ListPortalFiles(ctx context.Context, site string) ([]PortalFile, error)

	// UploadPortalFile uploads a Hotspot Portal file to the controller.
	UploadPortalFile(ctx context.Context, site string, filepath string) (*PortalFile, error)

	// UploadPortalFileFromReader uploads a Hotspot Portal file using io.Reader to the controller.
	UploadPortalFileFromReader(ctx context.Context, site string, reader io.Reader, filename string) (*PortalFile, error)

	// ==== client methods for RADIUSProfile resource ====

	// CreateRADIUSProfile creates a resource
	CreateRADIUSProfile(ctx context.Context, site string, r *RADIUSProfile) (*RADIUSProfile, error)

	// DeleteRADIUSProfile deletes a resource
	DeleteRADIUSProfile(ctx context.Context, site string, id string) error

	// GetRADIUSProfile retrieves a resource
	GetRADIUSProfile(ctx context.Context, site string, id string) (*RADIUSProfile, error)

	// ListRADIUSProfile lists the resources
	ListRADIUSProfile(ctx context.Context, site string) ([]RADIUSProfile, error)

	// UpdateRADIUSProfile updates a resource
	UpdateRADIUSProfile(ctx context.Context, site string, r *RADIUSProfile) (*RADIUSProfile, error)

	// ==== end of client methods for RADIUSProfile resource ====

	// ==== client methods for Routing resource ====

	// CreateRouting creates a resource
	CreateRouting(ctx context.Context, site string, r *Routing) (*Routing, error)

	// DeleteRouting deletes a resource
	DeleteRouting(ctx context.Context, site string, id string) error

	// GetRouting retrieves a resource
	GetRouting(ctx context.Context, site string, id string) (*Routing, error)

	// ListRouting lists the resources
	ListRouting(ctx context.Context, site string) ([]Routing, error)

	// UpdateRouting updates a resource
	UpdateRouting(ctx context.Context, site string, r *Routing) (*Routing, error)

	// ==== end of client methods for Routing resource ====

	// ==== client methods for ScheduleTask resource ====

	// CreateScheduleTask creates a resource
	CreateScheduleTask(ctx context.Context, site string, s *ScheduleTask) (*ScheduleTask, error)

	// DeleteScheduleTask deletes a resource
	DeleteScheduleTask(ctx context.Context, site string, id string) error

	// GetScheduleTask retrieves a resource
	GetScheduleTask(ctx context.Context, site string, id string) (*ScheduleTask, error)

	// ListScheduleTask lists the resources
	ListScheduleTask(ctx context.Context, site string) ([]ScheduleTask, error)

	// UpdateScheduleTask updates a resource
	UpdateScheduleTask(ctx context.Context, site string, s *ScheduleTask) (*ScheduleTask, error)

	// ==== end of client methods for ScheduleTask resource ====

	GetSetting(ctx context.Context, site string, key string) (*Setting, interface{}, error)

	// ==== client methods for SettingAutoSpeedtest resource ====

	// GetSettingAutoSpeedtest retrieves the settings for a resource
	GetSettingAutoSpeedtest(ctx context.Context, site string) (*SettingAutoSpeedtest, error)

	// UpdateSettingAutoSpeedtest updates a resource
	UpdateSettingAutoSpeedtest(ctx context.Context, site string, s *SettingAutoSpeedtest) (*SettingAutoSpeedtest, error)

	// ==== client methods for SettingBaresip resource ====

	// GetSettingBaresip retrieves the settings for a resource
	GetSettingBaresip(ctx context.Context, site string) (*SettingBaresip, error)

	// UpdateSettingBaresip updates a resource
	UpdateSettingBaresip(ctx context.Context, site string, s *SettingBaresip) (*SettingBaresip, error)

	// ==== client methods for SettingBroadcast resource ====

	// GetSettingBroadcast retrieves the settings for a resource
	GetSettingBroadcast(ctx context.Context, site string) (*SettingBroadcast, error)

	// UpdateSettingBroadcast updates a resource
	UpdateSettingBroadcast(ctx context.Context, site string, s *SettingBroadcast) (*SettingBroadcast, error)

	// ==== client methods for SettingConnectivity resource ====

	// GetSettingConnectivity retrieves the settings for a resource
	GetSettingConnectivity(ctx context.Context, site string) (*SettingConnectivity, error)

	// UpdateSettingConnectivity updates a resource
	UpdateSettingConnectivity(ctx context.Context, site string, s *SettingConnectivity) (*SettingConnectivity, error)

	// ==== client methods for SettingCountry resource ====

	// GetSettingCountry retrieves the settings for a resource
	GetSettingCountry(ctx context.Context, site string) (*SettingCountry, error)

	// UpdateSettingCountry updates a resource
	UpdateSettingCountry(ctx context.Context, site string, s *SettingCountry) (*SettingCountry, error)

	// ==== client methods for SettingDashboard resource ====

	// GetSettingDashboard retrieves the settings for a resource
	GetSettingDashboard(ctx context.Context, site string) (*SettingDashboard, error)

	// UpdateSettingDashboard updates a resource
	UpdateSettingDashboard(ctx context.Context, site string, s *SettingDashboard) (*SettingDashboard, error)

	// ==== client methods for SettingDoh resource ====

	// GetSettingDoh retrieves the settings for a resource
	GetSettingDoh(ctx context.Context, site string) (*SettingDoh, error)

	// UpdateSettingDoh updates a resource
	UpdateSettingDoh(ctx context.Context, site string, s *SettingDoh) (*SettingDoh, error)

	// ==== client methods for SettingDpi resource ====

	// GetSettingDpi retrieves the settings for a resource
	GetSettingDpi(ctx context.Context, site string) (*SettingDpi, error)

	// UpdateSettingDpi updates a resource
	UpdateSettingDpi(ctx context.Context, site string, s *SettingDpi) (*SettingDpi, error)

	// ==== client methods for SettingElementAdopt resource ====

	// GetSettingElementAdopt retrieves the settings for a resource
	GetSettingElementAdopt(ctx context.Context, site string) (*SettingElementAdopt, error)

	// UpdateSettingElementAdopt updates a resource
	UpdateSettingElementAdopt(ctx context.Context, site string, s *SettingElementAdopt) (*SettingElementAdopt, error)

	// ==== client methods for SettingEtherLighting resource ====

	// GetSettingEtherLighting retrieves the settings for a resource
	GetSettingEtherLighting(ctx context.Context, site string) (*SettingEtherLighting, error)

	// UpdateSettingEtherLighting updates a resource
	UpdateSettingEtherLighting(ctx context.Context, site string, s *SettingEtherLighting) (*SettingEtherLighting, error)

	// ==== client methods for SettingEvaluationScore resource ====

	// GetSettingEvaluationScore retrieves the settings for a resource
	GetSettingEvaluationScore(ctx context.Context, site string) (*SettingEvaluationScore, error)

	// UpdateSettingEvaluationScore updates a resource
	UpdateSettingEvaluationScore(ctx context.Context, site string, s *SettingEvaluationScore) (*SettingEvaluationScore, error)

	// ==== client methods for SettingGlobalAp resource ====

	// GetSettingGlobalAp retrieves the settings for a resource
	GetSettingGlobalAp(ctx context.Context, site string) (*SettingGlobalAp, error)

	// UpdateSettingGlobalAp updates a resource
	UpdateSettingGlobalAp(ctx context.Context, site string, s *SettingGlobalAp) (*SettingGlobalAp, error)

	// ==== client methods for SettingGlobalNat resource ====

	// GetSettingGlobalNat retrieves the settings for a resource
	GetSettingGlobalNat(ctx context.Context, site string) (*SettingGlobalNat, error)

	// UpdateSettingGlobalNat updates a resource
	UpdateSettingGlobalNat(ctx context.Context, site string, s *SettingGlobalNat) (*SettingGlobalNat, error)

	// ==== client methods for SettingGlobalSwitch resource ====

	// GetSettingGlobalSwitch retrieves the settings for a resource
	GetSettingGlobalSwitch(ctx context.Context, site string) (*SettingGlobalSwitch, error)

	// UpdateSettingGlobalSwitch updates a resource
	UpdateSettingGlobalSwitch(ctx context.Context, site string, s *SettingGlobalSwitch) (*SettingGlobalSwitch, error)

	// ==== client methods for SettingGuestAccess resource ====

	// GetSettingGuestAccess retrieves the settings for a resource
	GetSettingGuestAccess(ctx context.Context, site string) (*SettingGuestAccess, error)

	// UpdateSettingGuestAccess updates a resource
	UpdateSettingGuestAccess(ctx context.Context, site string, s *SettingGuestAccess) (*SettingGuestAccess, error)

	// ==== client methods for SettingIps resource ====

	// GetSettingIps retrieves the settings for a resource
	GetSettingIps(ctx context.Context, site string) (*SettingIps, error)

	// UpdateSettingIps updates a resource
	UpdateSettingIps(ctx context.Context, site string, s *SettingIps) (*SettingIps, error)

	// ==== client methods for SettingLcm resource ====

	// GetSettingLcm retrieves the settings for a resource
	GetSettingLcm(ctx context.Context, site string) (*SettingLcm, error)

	// UpdateSettingLcm updates a resource
	UpdateSettingLcm(ctx context.Context, site string, s *SettingLcm) (*SettingLcm, error)

	// ==== client methods for SettingLocale resource ====

	// GetSettingLocale retrieves the settings for a resource
	GetSettingLocale(ctx context.Context, site string) (*SettingLocale, error)

	// UpdateSettingLocale updates a resource
	UpdateSettingLocale(ctx context.Context, site string, s *SettingLocale) (*SettingLocale, error)

	// ==== client methods for SettingMagicSiteToSiteVpn resource ====

	// GetSettingMagicSiteToSiteVpn retrieves the settings for a resource
	GetSettingMagicSiteToSiteVpn(ctx context.Context, site string) (*SettingMagicSiteToSiteVpn, error)

	// UpdateSettingMagicSiteToSiteVpn updates a resource
	UpdateSettingMagicSiteToSiteVpn(ctx context.Context, site string, s *SettingMagicSiteToSiteVpn) (*SettingMagicSiteToSiteVpn, error)

	// ==== client methods for SettingMgmt resource ====

	// GetSettingMgmt retrieves the settings for a resource
	GetSettingMgmt(ctx context.Context, site string) (*SettingMgmt, error)

	// UpdateSettingMgmt updates a resource
	UpdateSettingMgmt(ctx context.Context, site string, s *SettingMgmt) (*SettingMgmt, error)

	// ==== client methods for SettingNetflow resource ====

	// GetSettingNetflow retrieves the settings for a resource
	GetSettingNetflow(ctx context.Context, site string) (*SettingNetflow, error)

	// UpdateSettingNetflow updates a resource
	UpdateSettingNetflow(ctx context.Context, site string, s *SettingNetflow) (*SettingNetflow, error)

	// ==== client methods for SettingNetworkOptimization resource ====

	// GetSettingNetworkOptimization retrieves the settings for a resource
	GetSettingNetworkOptimization(ctx context.Context, site string) (*SettingNetworkOptimization, error)

	// UpdateSettingNetworkOptimization updates a resource
	UpdateSettingNetworkOptimization(ctx context.Context, site string, s *SettingNetworkOptimization) (*SettingNetworkOptimization, error)

	// ==== client methods for SettingNtp resource ====

	// GetSettingNtp retrieves the settings for a resource
	GetSettingNtp(ctx context.Context, site string) (*SettingNtp, error)

	// UpdateSettingNtp updates a resource
	UpdateSettingNtp(ctx context.Context, site string, s *SettingNtp) (*SettingNtp, error)

	// ==== client methods for SettingPorta resource ====

	// GetSettingPorta retrieves the settings for a resource
	GetSettingPorta(ctx context.Context, site string) (*SettingPorta, error)

	// UpdateSettingPorta updates a resource
	UpdateSettingPorta(ctx context.Context, site string, s *SettingPorta) (*SettingPorta, error)

	// ==== client methods for SettingRadioAi resource ====

	// GetSettingRadioAi retrieves the settings for a resource
	GetSettingRadioAi(ctx context.Context, site string) (*SettingRadioAi, error)

	// UpdateSettingRadioAi updates a resource
	UpdateSettingRadioAi(ctx context.Context, site string, s *SettingRadioAi) (*SettingRadioAi, error)

	// ==== client methods for SettingRadius resource ====

	// GetSettingRadius retrieves the settings for a resource
	GetSettingRadius(ctx context.Context, site string) (*SettingRadius, error)

	// UpdateSettingRadius updates a resource
	UpdateSettingRadius(ctx context.Context, site string, s *SettingRadius) (*SettingRadius, error)

	// ==== client methods for SettingRsyslogd resource ====

	// GetSettingRsyslogd retrieves the settings for a resource
	GetSettingRsyslogd(ctx context.Context, site string) (*SettingRsyslogd, error)

	// UpdateSettingRsyslogd updates a resource
	UpdateSettingRsyslogd(ctx context.Context, site string, s *SettingRsyslogd) (*SettingRsyslogd, error)

	// ==== client methods for SettingSnmp resource ====

	// GetSettingSnmp retrieves the settings for a resource
	GetSettingSnmp(ctx context.Context, site string) (*SettingSnmp, error)

	// UpdateSettingSnmp updates a resource
	UpdateSettingSnmp(ctx context.Context, site string, s *SettingSnmp) (*SettingSnmp, error)

	// ==== client methods for SettingSslInspection resource ====

	// GetSettingSslInspection retrieves the settings for a resource
	GetSettingSslInspection(ctx context.Context, site string) (*SettingSslInspection, error)

	// UpdateSettingSslInspection updates a resource
	UpdateSettingSslInspection(ctx context.Context, site string, s *SettingSslInspection) (*SettingSslInspection, error)

	// ==== client methods for SettingSuperCloudaccess resource ====

	// GetSettingSuperCloudaccess retrieves the settings for a resource
	GetSettingSuperCloudaccess(ctx context.Context, site string) (*SettingSuperCloudaccess, error)

	// UpdateSettingSuperCloudaccess updates a resource
	UpdateSettingSuperCloudaccess(ctx context.Context, site string, s *SettingSuperCloudaccess) (*SettingSuperCloudaccess, error)

	// ==== client methods for SettingSuperEvents resource ====

	// GetSettingSuperEvents retrieves the settings for a resource
	GetSettingSuperEvents(ctx context.Context, site string) (*SettingSuperEvents, error)

	// UpdateSettingSuperEvents updates a resource
	UpdateSettingSuperEvents(ctx context.Context, site string, s *SettingSuperEvents) (*SettingSuperEvents, error)

	// ==== client methods for SettingSuperFwupdate resource ====

	// GetSettingSuperFwupdate retrieves the settings for a resource
	GetSettingSuperFwupdate(ctx context.Context, site string) (*SettingSuperFwupdate, error)

	// UpdateSettingSuperFwupdate updates a resource
	UpdateSettingSuperFwupdate(ctx context.Context, site string, s *SettingSuperFwupdate) (*SettingSuperFwupdate, error)

	// ==== client methods for SettingSuperIdentity resource ====

	// GetSettingSuperIdentity retrieves the settings for a resource
	GetSettingSuperIdentity(ctx context.Context, site string) (*SettingSuperIdentity, error)

	// UpdateSettingSuperIdentity updates a resource
	UpdateSettingSuperIdentity(ctx context.Context, site string, s *SettingSuperIdentity) (*SettingSuperIdentity, error)

	// ==== client methods for SettingSuperMail resource ====

	// GetSettingSuperMail retrieves the settings for a resource
	GetSettingSuperMail(ctx context.Context, site string) (*SettingSuperMail, error)

	// UpdateSettingSuperMail updates a resource
	UpdateSettingSuperMail(ctx context.Context, site string, s *SettingSuperMail) (*SettingSuperMail, error)

	// ==== client methods for SettingSuperMgmt resource ====

	// GetSettingSuperMgmt retrieves the settings for a resource
	GetSettingSuperMgmt(ctx context.Context, site string) (*SettingSuperMgmt, error)

	// UpdateSettingSuperMgmt updates a resource
	UpdateSettingSuperMgmt(ctx context.Context, site string, s *SettingSuperMgmt) (*SettingSuperMgmt, error)

	// ==== client methods for SettingSuperSdn resource ====

	// GetSettingSuperSdn retrieves the settings for a resource
	GetSettingSuperSdn(ctx context.Context, site string) (*SettingSuperSdn, error)

	// UpdateSettingSuperSdn updates a resource
	UpdateSettingSuperSdn(ctx context.Context, site string, s *SettingSuperSdn) (*SettingSuperSdn, error)

	// ==== client methods for SettingSuperSmtp resource ====

	// GetSettingSuperSmtp retrieves the settings for a resource
	GetSettingSuperSmtp(ctx context.Context, site string) (*SettingSuperSmtp, error)

	// UpdateSettingSuperSmtp updates a resource
	UpdateSettingSuperSmtp(ctx context.Context, site string, s *SettingSuperSmtp) (*SettingSuperSmtp, error)

	// ==== client methods for SettingTeleport resource ====

	// GetSettingTeleport retrieves the settings for a resource
	GetSettingTeleport(ctx context.Context, site string) (*SettingTeleport, error)

	// UpdateSettingTeleport updates a resource
	UpdateSettingTeleport(ctx context.Context, site string, s *SettingTeleport) (*SettingTeleport, error)

	// ==== client methods for SettingUsg resource ====

	// GetSettingUsg retrieves the settings for a resource
	GetSettingUsg(ctx context.Context, site string) (*SettingUsg, error)

	// UpdateSettingUsg updates a resource
	UpdateSettingUsg(ctx context.Context, site string, s *SettingUsg) (*SettingUsg, error)

	// ==== client methods for SettingUsw resource ====

	// GetSettingUsw retrieves the settings for a resource
	GetSettingUsw(ctx context.Context, site string) (*SettingUsw, error)

	// UpdateSettingUsw updates a resource
	UpdateSettingUsw(ctx context.Context, site string, s *SettingUsw) (*SettingUsw, error)

	CreateSite(ctx context.Context, description string) ([]Site, error)

	DeleteSite(ctx context.Context, id string) ([]Site, error)

	GetSite(ctx context.Context, id string) (*Site, error)

	ListSites(ctx context.Context) ([]Site, error)

	UpdateSite(ctx context.Context, name string, description string) ([]Site, error)

	// ==== client methods for SpatialRecord resource ====

	// CreateSpatialRecord creates a resource
	CreateSpatialRecord(ctx context.Context, site string, s *SpatialRecord) (*SpatialRecord, error)

	// DeleteSpatialRecord deletes a resource
	DeleteSpatialRecord(ctx context.Context, site string, id string) error

	// GetSpatialRecord retrieves a resource
	GetSpatialRecord(ctx context.Context, site string, id string) (*SpatialRecord, error)

	// ListSpatialRecord lists the resources
	ListSpatialRecord(ctx context.Context, site string) ([]SpatialRecord, error)

	// UpdateSpatialRecord updates a resource
	UpdateSpatialRecord(ctx context.Context, site string, s *SpatialRecord) (*SpatialRecord, error)

	// ==== end of client methods for SpatialRecord resource ====

	GetSystemInfo(ctx context.Context, id string) (*SysInfo, error)

	GetSystemInformation() (*SysInfo, error)

	// ==== client methods for Tag resource ====

	// CreateTag creates a resource
	CreateTag(ctx context.Context, site string, t *Tag) (*Tag, error)

	// DeleteTag deletes a resource
	DeleteTag(ctx context.Context, site string, id string) error

	// GetTag retrieves a resource
	GetTag(ctx context.Context, site string, id string) (*Tag, error)

	// ListTag lists the resources
	ListTag(ctx context.Context, site string) ([]Tag, error)

	// UpdateTag updates a resource
	UpdateTag(ctx context.Context, site string, t *Tag) (*Tag, error)

	// ==== end of client methods for Tag resource ====

	// ==== client methods for User resource ====

	BlockUserByMAC(ctx context.Context, site string, mac string) error

	// CreateUser creates a resource
	CreateUser(ctx context.Context, site string, u *User) (*User, error)

	// DeleteUser deletes a resource
	DeleteUser(ctx context.Context, site string, id string) error

	DeleteUserByMAC(ctx context.Context, site string, mac string) error

	// GetUser retrieves a resource
	GetUser(ctx context.Context, site string, id string) (*User, error)

	GetUserByMAC(ctx context.Context, site string, mac string) (*User, error)

	KickUserByMAC(ctx context.Context, site string, mac string) error

	// ListUser lists the resources
	ListUser(ctx context.Context, site string) ([]User, error)

	OverrideUserFingerprint(ctx context.Context, site string, mac string, devIdOverride int) error

	UnblockUserByMAC(ctx context.Context, site string, mac string) error

	// UpdateUser updates a resource
	UpdateUser(ctx context.Context, site string, u *User) (*User, error)

	// ==== client methods for UserGroup resource ====

	// CreateUserGroup creates a resource
	CreateUserGroup(ctx context.Context, site string, u *UserGroup) (*UserGroup, error)

	// DeleteUserGroup deletes a resource
	DeleteUserGroup(ctx context.Context, site string, id string) error

	// GetUserGroup retrieves a resource
	GetUserGroup(ctx context.Context, site string, id string) (*UserGroup, error)

	// ListUserGroup lists the resources
	ListUserGroup(ctx context.Context, site string) ([]UserGroup, error)

	// UpdateUserGroup updates a resource
	UpdateUserGroup(ctx context.Context, site string, u *UserGroup) (*UserGroup, error)

	// ==== end of client methods for UserGroup resource ====

	// ==== end of client methods for User resource ====

	// ==== client methods for VirtualDevice resource ====

	// CreateVirtualDevice creates a resource
	CreateVirtualDevice(ctx context.Context, site string, v *VirtualDevice) (*VirtualDevice, error)

	// DeleteVirtualDevice deletes a resource
	DeleteVirtualDevice(ctx context.Context, site string, id string) error

	// GetVirtualDevice retrieves a resource
	GetVirtualDevice(ctx context.Context, site string, id string) (*VirtualDevice, error)

	// ListVirtualDevice lists the resources
	ListVirtualDevice(ctx context.Context, site string) ([]VirtualDevice, error)

	// UpdateVirtualDevice updates a resource
	UpdateVirtualDevice(ctx context.Context, site string, v *VirtualDevice) (*VirtualDevice, error)

	// ==== end of client methods for VirtualDevice resource ====

	// ==== client methods for WLAN resource ====

	// CreateWLAN creates a resource
	CreateWLAN(ctx context.Context, site string, w *WLAN) (*WLAN, error)

	// DeleteWLAN deletes a resource
	DeleteWLAN(ctx context.Context, site string, id string) error

	// GetWLAN retrieves a resource
	GetWLAN(ctx context.Context, site string, id string) (*WLAN, error)

	// ListWLAN lists the resources
	ListWLAN(ctx context.Context, site string) ([]WLAN, error)

	// UpdateWLAN updates a resource
	UpdateWLAN(ctx context.Context, site string, w *WLAN) (*WLAN, error)

	// ==== client methods for WLANGroup resource ====

	// CreateWLANGroup creates a resource
	CreateWLANGroup(ctx context.Context, site string, w *WLANGroup) (*WLANGroup, error)

	// DeleteWLANGroup deletes a resource
	DeleteWLANGroup(ctx context.Context, site string, id string) error

	// GetWLANGroup retrieves a resource
	GetWLANGroup(ctx context.Context, site string, id string) (*WLANGroup, error)

	// ListWLANGroup lists the resources
	ListWLANGroup(ctx context.Context, site string) ([]WLANGroup, error)

	// UpdateWLANGroup updates a resource
	UpdateWLANGroup(ctx context.Context, site string, w *WLANGroup) (*WLANGroup, error)

	// ==== end of client methods for WLANGroup resource ====

	// ==== end of client methods for WLAN resource ====

}

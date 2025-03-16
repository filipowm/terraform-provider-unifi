package settings

import (
	"context"
	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/go-unifi/unifi/features"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
	"github.com/filipowm/terraform-provider-unifi/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// DNS Filter model
type DNSFilterModel struct {
	AllowedSites types.List   `tfsdk:"allowed_sites"`
	BlockedSites types.List   `tfsdk:"blocked_sites"`
	BlockedTld   types.List   `tfsdk:"blocked_tld"`
	Description  types.String `tfsdk:"description"`
	Filter       types.String `tfsdk:"filter"`
	Name         types.String `tfsdk:"name"`
	NetworkID    types.String `tfsdk:"network_id"`
}

func (m *DNSFilterModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"allowed_sites": types.ListType{
			ElemType: types.StringType,
		},
		"blocked_sites": types.ListType{
			ElemType: types.StringType,
		},
		"blocked_tld": types.ListType{
			ElemType: types.StringType,
		},
		"description": types.StringType,
		"filter":      types.StringType,
		"name":        types.StringType,
		"network_id":  types.StringType,
	}
}

// Honeypots model
type HoneypotModel struct {
	IPAddress types.String `tfsdk:"ip_address"`
	NetworkID types.String `tfsdk:"network_id"`
}

func (m *HoneypotModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"ip_address": types.StringType,
		"network_id": types.StringType,
	}
}

// Tracking model
type TrackingModel struct {
	Direction types.String `tfsdk:"direction"`
	Mode      types.String `tfsdk:"mode"`
	Value     types.String `tfsdk:"value"`
}

func (m *TrackingModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"direction": types.StringType,
		"mode":      types.StringType,
		"value":     types.StringType,
	}
}

// Alerts model
type AlertsModel struct {
	Category  types.String `tfsdk:"category"`
	Signature types.String `tfsdk:"signature"`
	Tracking  types.List   `tfsdk:"tracking"`
	Type      types.String `tfsdk:"type"`
}

func (m *AlertsModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"category":  types.StringType,
		"signature": types.StringType,
		"tracking": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: (&TrackingModel{}).AttributeTypes(),
			},
		},
		"type": types.StringType,
	}
}

// Whitelist model
type WhitelistModel struct {
	Direction types.String `tfsdk:"direction"`
	Mode      types.String `tfsdk:"mode"`
	Value     types.String `tfsdk:"value"`
}

func (m *WhitelistModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"direction": types.StringType,
		"mode":      types.StringType,
		"value":     types.StringType,
	}
}

// Suppression model
type SuppressionModel struct {
	Alerts    types.List `tfsdk:"alerts"`
	Whitelist types.List `tfsdk:"whitelist"`
}

func (m *SuppressionModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"alerts": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: (&AlertsModel{}).AttributeTypes(),
			},
		},
		"whitelist": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: (&WhitelistModel{}).AttributeTypes(),
			},
		},
	}
}

// Main IPS model
type ipsModel struct {
	base.Model
	AdBlockedNetworks           types.List   `tfsdk:"ad_blocked_networks"`
	AdvancedFilteringPreference types.String `tfsdk:"advanced_filtering_preference"`
	DNSFilters                  types.List   `tfsdk:"dns_filters"`
	EnabledCategories           types.List   `tfsdk:"enabled_categories"`
	EnabledNetworks             types.List   `tfsdk:"enabled_networks"`
	Honeypots                   types.List   `tfsdk:"honeypots"`
	Mode                        types.String `tfsdk:"ips_mode"`
	MemoryOptimized             types.Bool   `tfsdk:"memory_optimized"`
	RestrictTorrents            types.Bool   `tfsdk:"restrict_torrents"`
	Suppression                 types.Object `tfsdk:"suppression"`
}

func (d *ipsModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingIps{
		AdvancedFilteringPreference: d.AdvancedFilteringPreference.ValueString(),
		IPsMode:                     d.Mode.ValueString(),
		MemoryOptimized:             d.MemoryOptimized.ValueBool(),
		RestrictTorrents:            d.RestrictTorrents.ValueBool(),
		// Initialize empty slices for arrays to avoid null values in JSON
		AdBlockingConfigurations: []unifi.SettingIpsAdBlockingConfigurations{},
		DNSFilters:               []unifi.SettingIpsDNSFilters{},
		EnabledCategories:        []string{},
		EnabledNetworks:          []string{},
		Honeypot:                 []unifi.SettingIpsHoneypot{},
		// Initialize suppression with empty arrays
		Suppression: unifi.SettingIpsSuppression{
			Alerts:    []unifi.SettingIpsAlerts{},
			Whitelist: []unifi.SettingIpsWhitelist{},
		},
	}

	var enabledCategories []string
	diags.Append(utils.ListElementsAs(d.EnabledCategories, &enabledCategories)...)
	if diags.HasError() {
		return nil, diags
	}
	model.EnabledCategories = enabledCategories

	var enabledNetworks []string
	diags.Append(utils.ListElementsAs(d.EnabledNetworks, &enabledNetworks)...)
	if diags.HasError() {
		return nil, diags
	}
	model.EnabledNetworks = enabledNetworks

	// Handle AdBlockedNetworks - if any networks are configured, set AdBlockingEnabled to true
	if base.IsDefined(d.AdBlockedNetworks) {
		var adBlockedNetworks []string
		diags.Append(utils.ListElementsAs(d.AdBlockedNetworks, &adBlockedNetworks)...)
		if diags.HasError() {
			return nil, diags
		}

		if len(adBlockedNetworks) > 0 {
			model.AdBlockingEnabled = true
			model.AdBlockingConfigurations = make([]unifi.SettingIpsAdBlockingConfigurations, 0, len(adBlockedNetworks))
			for _, networkID := range adBlockedNetworks {
				model.AdBlockingConfigurations = append(model.AdBlockingConfigurations, unifi.SettingIpsAdBlockingConfigurations{
					NetworkID: networkID,
				})
			}
		} else {
			model.AdBlockingEnabled = false
			model.AdBlockingConfigurations = []unifi.SettingIpsAdBlockingConfigurations{}
		}
	}

	// Handle DNSFilters - if any filters are configured, set DNSFiltering to true
	if base.IsDefined(d.DNSFilters) {
		var dnsFiltersObjects []DNSFilterModel
		diags.Append(utils.ListElementsAs(d.DNSFilters, &dnsFiltersObjects)...)
		if diags.HasError() {
			return nil, diags
		}

		if len(dnsFiltersObjects) > 0 {
			model.DNSFiltering = true
			model.DNSFilters = make([]unifi.SettingIpsDNSFilters, 0, len(dnsFiltersObjects))

			for _, filterObj := range dnsFiltersObjects {
				version := "v4"
				if utils.IsIPv6(filterObj.NetworkID.ValueString()) {
					version = "v6"
				}
				filter := unifi.SettingIpsDNSFilters{
					Description: filterObj.Description.ValueString(),
					Filter:      filterObj.Filter.ValueString(),
					Name:        filterObj.Name.ValueString(),
					NetworkID:   filterObj.NetworkID.ValueString(),
					Version:     version,
				}

				// Handle allowed sites

				var allowedSites, blockedSites, blockedTlds []string
				diags.Append(utils.ListElementsAs(filterObj.AllowedSites, &allowedSites)...)
				diags.Append(utils.ListElementsAs(filterObj.BlockedSites, &blockedSites)...)
				diags.Append(utils.ListElementsAs(filterObj.BlockedTld, &blockedTlds)...)
				if diags.HasError() {
					return nil, diags
				}
				filter.AllowedSites = allowedSites
				filter.BlockedSites = blockedSites
				filter.BlockedTld = blockedTlds
				model.DNSFilters = append(model.DNSFilters, filter)
			}
		} else {
			model.DNSFiltering = false
			model.DNSFilters = []unifi.SettingIpsDNSFilters{}
		}
	}

	// Handle honeypot
	if base.IsDefined(d.Honeypots) {
		var honeypotObjects []HoneypotModel
		diags.Append(utils.ListElementsAs(d.Honeypots, &honeypotObjects)...)
		if diags.HasError() {
			return nil, diags
		}

		model.Honeypot = make([]unifi.SettingIpsHoneypot, 0)
		for _, honeypotObj := range honeypotObjects {
			version := "v4"
			if utils.IsIPv6(honeypotObj.IPAddress.ValueString()) {
				version = "v6"
			}
			model.Honeypot = append(model.Honeypot, unifi.SettingIpsHoneypot{
				IPAddress: honeypotObj.IPAddress.ValueString(),
				NetworkID: honeypotObj.NetworkID.ValueString(),
				Version:   version,
			})
		}
	}
	if len(model.Honeypot) > 0 {
		model.HoneypotEnabled = true
	} else {
		model.HoneypotEnabled = false
	}

	// Handle suppression
	if base.IsDefined(d.Suppression) {
		var suppressionObj SuppressionModel
		diags.Append(d.Suppression.As(ctx, &suppressionObj, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}

		var alerts []AlertsModel
		diags.Append(utils.ListElementsAs(suppressionObj.Alerts, &alerts)...)
		if diags.HasError() {
			return nil, diags
		}
		model.Suppression.Alerts = make([]unifi.SettingIpsAlerts, 0)
		for idx, alertObj := range alerts {
			alert := unifi.SettingIpsAlerts{
				Category:  alertObj.Category.ValueString(),
				Signature: alertObj.Signature.ValueString(),
				Type:      alertObj.Type.ValueString(),
				ID:        100 + idx,
				Gid:       200 + idx,
			}
			// Handle tracking

			var trackings []TrackingModel
			diags.Append(utils.ListElementsAs(alertObj.Tracking, &trackings)...)
			if diags.HasError() {
				return nil, diags
			}
			alert.Tracking = make([]unifi.SettingIpsTracking, 0)
			for _, trackingObj := range trackings {
				if base.IsEmptyString(trackingObj.Direction) || base.IsEmptyString(trackingObj.Mode) || base.IsEmptyString(trackingObj.Value) {
					continue
				}
				alert.Tracking = append(alert.Tracking, unifi.SettingIpsTracking{
					Direction: trackingObj.Direction.ValueString(),
					Mode:      trackingObj.Mode.ValueString(),
					Value:     trackingObj.Value.ValueString()})
			}
			model.Suppression.Alerts = append(model.Suppression.Alerts, alert)
		}

		var whitelists []WhitelistModel
		diags.Append(utils.ListElementsAs(suppressionObj.Whitelist, &whitelists)...)
		if diags.HasError() {
			return nil, diags
		}
		model.Suppression.Whitelist = make([]unifi.SettingIpsWhitelist, 0, len(whitelists))
		for _, whitelistObj := range whitelists {
			model.Suppression.Whitelist = append(model.Suppression.Whitelist, unifi.SettingIpsWhitelist{
				Direction: whitelistObj.Direction.ValueString(),
				Mode:      whitelistObj.Mode.ValueString(),
				Value:     whitelistObj.Value.ValueString(),
			})
		}
	}

	return model, diags
}

func (d *ipsModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingIps)
	if !ok {
		diags.AddError("Invalid model type", "Expected *unifi.SettingIps")
		return diags
	}

	d.ID = types.StringValue(model.ID)

	// Only set values for fields that were explicitly set in the configuration
	// or returned by the API with non-default values

	// Set basic fields if they were defined in the plan
	d.AdvancedFilteringPreference = types.StringValue(model.AdvancedFilteringPreference)

	d.Mode = types.StringValue(model.IPsMode)

	d.MemoryOptimized = types.BoolValue(model.MemoryOptimized)
	d.RestrictTorrents = types.BoolValue(model.RestrictTorrents)

	// Handle enabled categories
	enabledCategoriesList, diags := types.ListValueFrom(ctx, types.StringType, model.EnabledCategories)
	if diags.HasError() {
		return diags
	}
	if base.IsDefined(enabledCategoriesList) {
		d.EnabledCategories = enabledCategoriesList
	} else {
		d.EnabledCategories = utils.EmptyList(types.StringType)
	}

	// Handle enabled networks
	enabledNetworksList, diags := types.ListValueFrom(ctx, types.StringType, model.EnabledNetworks)
	if diags.HasError() {
		return diags
	}
	if base.IsDefined(enabledNetworksList) {
		d.EnabledNetworks = enabledNetworksList
	} else {
		d.EnabledNetworks = utils.EmptyList(types.StringType)
	}

	//Handle AdBlockedNetworks - extract network IDs from AdBlockingConfigurations
	adBlockedNetworks := make([]string, 0, len(model.AdBlockingConfigurations))
	for _, config := range model.AdBlockingConfigurations {
		adBlockedNetworks = append(adBlockedNetworks, config.NetworkID)
	}

	adBlockedNetworksList, diags := types.ListValueFrom(ctx, types.StringType, adBlockedNetworks)
	if diags.HasError() {
		return diags
	}
	d.AdBlockedNetworks = adBlockedNetworksList

	// Handle DNSFilters
	dnsFilters := make([]DNSFilterModel, 0)

	for _, filter := range model.DNSFilters {
		dnsFilter := DNSFilterModel{
			Description: types.StringValue(filter.Description),
			Filter:      types.StringValue(filter.Filter),
			Name:        types.StringValue(filter.Name),
			NetworkID:   types.StringValue(filter.NetworkID),
		}

		allowedSites, diags := types.ListValueFrom(ctx, types.StringType, filter.AllowedSites)
		if diags.HasError() {
			return diags
		}
		dnsFilter.AllowedSites = allowedSites

		blockedSites, diags := types.ListValueFrom(ctx, types.StringType, filter.BlockedSites)
		if diags.HasError() {
			return diags
		}
		dnsFilter.BlockedSites = blockedSites

		blockedTlds, diags := types.ListValueFrom(ctx, types.StringType, filter.BlockedTld)
		if diags.HasError() {
			return diags
		}
		dnsFilter.BlockedTld = blockedTlds

		dnsFilters = append(dnsFilters, dnsFilter)
	}

	dnsFiltersList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: (&DNSFilterModel{}).AttributeTypes(),
	}, dnsFilters)
	if diags.HasError() {
		return diags
	}
	d.DNSFilters = dnsFiltersList

	// Handle honeypot
	honeypotModels := make([]HoneypotModel, 0, len(model.Honeypot))
	for _, honeypot := range model.Honeypot {
		honeypotModels = append(honeypotModels, HoneypotModel{
			IPAddress: types.StringValue(honeypot.IPAddress),
			NetworkID: types.StringValue(honeypot.NetworkID),
		})
	}

	honeypotList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: (&HoneypotModel{}).AttributeTypes()}, honeypotModels)
	if diags.HasError() {
		return diags
	}
	d.Honeypots = honeypotList

	// Handle suppression
	suppression := SuppressionModel{}

	// Handle alerts
	alertModels := make([]AlertsModel, 0)
	for _, alert := range model.Suppression.Alerts {
		// Skip alerts with ID 0, because they may come as default values from the API
		if alert.ID == 0 && alert.Category == "" && alert.Signature == "" && alert.Type == "" {
			continue
		}
		alertModel := AlertsModel{
			Category:  types.StringValue(alert.Category),
			Signature: types.StringValue(alert.Signature),
			Type:      types.StringValue(alert.Type),
		}

		// Handle tracking
		trackingModels := make([]TrackingModel, 0)
		for _, tracking := range alert.Tracking {
			trackingModels = append(trackingModels, TrackingModel{
				Direction: types.StringValue(tracking.Direction),
				Mode:      types.StringValue(tracking.Mode),
				Value:     types.StringValue(tracking.Value),
			})
		}
		trackings, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: (&TrackingModel{}).AttributeTypes()}, trackingModels)
		if diags.HasError() {
			return diags
		}
		alertModel.Tracking = trackings
		alertModels = append(alertModels, alertModel)
	}
	alerts, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: (&AlertsModel{}).AttributeTypes()}, alertModels)
	if diags.HasError() {
		return diags
	}
	suppression.Alerts = alerts

	// Handle whitelist
	whitelistModels := make([]WhitelistModel, 0)
	for _, whitelist := range model.Suppression.Whitelist {
		whitelistModels = append(whitelistModels, WhitelistModel{
			Direction: types.StringValue(whitelist.Direction),
			Mode:      types.StringValue(whitelist.Mode),
			Value:     types.StringValue(whitelist.Value),
		})
	}
	whitelist, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: (&WhitelistModel{}).AttributeTypes()}, whitelistModels)
	if diags.HasError() {
		return diags
	}
	suppression.Whitelist = whitelist

	suppressionObj, diags := types.ObjectValueFrom(ctx, (&SuppressionModel{}).AttributeTypes(), suppression)
	if diags.HasError() {
		return diags
	}
	d.Suppression = suppressionObj

	return diags
}

type ipsResource struct {
	*base.GenericResource[*ipsModel]
}

func (r *ipsResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	resp.Diagnostics.Append(r.RequireMinVersionForPath("9.0", path.Root("memory_optimized"), req.Config)...)
	site, diags := r.GetClient().ResolveSiteFromConfig(ctx, req.Config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	resp.Diagnostics.Append(r.RequireFeaturesEnabled(ctx, site, features.Ips)...)
	//resp.Diagnostics.Append(r.RequireFeaturesEnabledForPath(ctx, site, path.Root("enabled_categories"), req.Config, features.LimitIpsCategories)...)
}

func (r *ipsResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		validators.RequiredTogetherIf(path.MatchRoot("ips_mode"), types.StringValue("ips"), path.MatchRoot("enabled_networks")),
		validators.RequiredTogetherIf(path.MatchRoot("ips_mode"), types.StringValue("ids"), path.MatchRoot("enabled_networks")),
		validators.RequiredTogetherIf(path.MatchRoot("ips_mode"), types.StringValue("ipsInline"), path.MatchRoot("enabled_networks")),
	}
}

func (r *ipsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_ips` resource allows you to configure the Intrusion Prevention System (IPS) settings for your UniFi network. IPS provides network threat protection by monitoring, detecting, and preventing malicious traffic based on configured rules and policies.",
		Attributes: map[string]schema.Attribute{
			"id":   base.ID(),
			"site": base.SiteAttribute(),
			"ad_blocked_networks": schema.ListAttribute{
				MarkdownDescription: "List of network IDs to enable ad blocking for. If any networks are configured, ad blocking will be automatically enabled. Each entry should be a valid network ID from your UniFi configuration. Leave empty to disable ad blocking.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"advanced_filtering_preference": schema.StringAttribute{
				MarkdownDescription: "The advanced filtering preference for IPS. Valid values are:\n" +
					"  * `disabled` - Advanced filtering is disabled\n" +
					"  * `manual` - Advanced filtering is enabled and manually configured",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("manual"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("disabled", "manual"),
				},
			},
			"dns_filters": schema.ListNestedAttribute{
				MarkdownDescription: "DNS filters configuration. If any filters are configured, DNS filtering will be automatically enabled. Each filter can be applied to a specific network and provides content filtering capabilities.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"allowed_sites": schema.ListAttribute{
							MarkdownDescription: "List of allowed sites for this DNS filter. These domains will always be accessible regardless of other filtering rules. Each entry should be a valid domain name (e.g., `example.com`).",
							ElementType:         types.StringType,
							Optional:            true,
						},
						"blocked_sites": schema.ListAttribute{
							MarkdownDescription: "List of blocked sites for this DNS filter. These domains will be blocked regardless of other filtering rules. Each entry should be a valid domain name (e.g., `example.com`).",
							ElementType:         types.StringType,
							Optional:            true,
						},
						"blocked_tld": schema.ListAttribute{
							MarkdownDescription: "List of blocked top-level domains (TLDs) for this DNS filter. All domains with these TLDs will be blocked. Each entry should be a valid TLD without the dot prefix (e.g., `xyz`, `info`).",
							ElementType:         types.StringType,
							Optional:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description of the DNS filter. This is used for documentation purposes only and does not affect functionality.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"filter": schema.StringAttribute{
							MarkdownDescription: "Filter type that determines the predefined filtering level. Valid values are:\n" +
								"  * `none` - No predefined filtering\n" +
								"  * `work` - Work-appropriate filtering that blocks adult content\n" +
								"  * `family` - Family-friendly filtering that blocks adult content and other inappropriate sites",
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf("none", "work", "family"),
							},
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the DNS filter. This is used to identify the filter in the UniFi interface.",
							Required:            true,
						},
						"network_id": schema.StringAttribute{
							MarkdownDescription: "Network ID this filter applies to. This should be a valid network ID from your UniFi configuration.",
							Required:            true,
						},
					},
				},
			},
			"enabled_categories": schema.ListAttribute{
				MarkdownDescription: "List of enabled IPS threat categories. Each entry enables detection and prevention for a specific type of threat. The list of valid categories includes common threats like malware, exploits, scanning, and policy violations. See the validator for the complete list of available categories.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				//Default: utils.DefaultEmptyList(types.StringType),
				Validators: []validator.List{
					listvalidator.ValueStringsAre(stringvalidator.OneOf("emerging-activex", "emerging-attackresponse", "botcc", "emerging-chat", "ciarmy", "compromised", "emerging-dns", "emerging-dos", "dshield", "emerging-exploit", "emerging-ftp", "emerging-games", "emerging-icmp", "emerging-icmpinfo", "emerging-imap", "emerging-inappropriate", "emerging-info", "emerging-malware", "emerging-misc", "emerging-mobile", "emerging-netbios", "emerging-p2p", "emerging-policy", "emerging-pop3", "emerging-rpc", "emerging-scada", "emerging-scan", "emerging-shellcode", "emerging-smtp", "emerging-snmp", "emerging-sql", "emerging-telnet", "emerging-tftp", "tor", "emerging-useragent", "emerging-voip", "emerging-webapps", "emerging-webclient", "emerging-webserver", "emerging-worm", "exploit-kit", "adware-pup", "botcc-portgrouped", "phishing", "threatview-cs-c2", "3coresec", "chat", "coinminer", "current-events", "drop", "hunting", "icmp-info", "inappropriate", "info", "ja3", "policy", "scada", "dark-web-blocker-list", "malicious-hosts")),
				},
			},
			"enabled_networks": schema.ListAttribute{
				MarkdownDescription: "List of network IDs to enable IPS protection for. Each entry should be a valid network ID from your UniFi configuration. IPS will only monitor and protect traffic on these networks.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"honeypots": schema.ListNestedAttribute{
				MarkdownDescription: "Honeypots configuration. Honeypots are decoy systems designed to detect, deflect, or study hacking attempts. They appear as legitimate parts of the network but are isolated and monitored.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip_address": schema.StringAttribute{
							MarkdownDescription: "IP address for the honeypot. This should be an unused IPv4 address within your network range that will be used as a decoy system.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.Any(validators.IPv4(), validators.IPv6()),
							},
						},
						"network_id": schema.StringAttribute{
							MarkdownDescription: "Network ID for the honeypot. This should be a valid network ID from your UniFi configuration where the honeypot will be deployed.",
							Required:            true,
						},
					},
				},
			},
			"ips_mode": schema.StringAttribute{
				MarkdownDescription: "The IPS operation mode. Valid values are:\n" +
					"  * `ids` - Intrusion Detection System mode (detect and log threats only)\n" +
					"  * `ips` - Intrusion Prevention System mode (detect and block threats)\n" +
					"  * `ipsInline` - Inline Intrusion Prevention System mode (more aggressive blocking)\n" +
					"  * `disabled` - IPS functionality is completely disabled",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("ids", "ips", "ipsInline", "disabled"),
				},
			},
			"memory_optimized": schema.BoolAttribute{
				MarkdownDescription: "Whether memory optimization is enabled for IPS. When set to `true`, the system will use less memory at the cost of potentially reduced detection capabilities. Useful for devices with limited resources. Defaults to `false`. Requires controller version 9.0 or later.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"restrict_torrents": schema.BoolAttribute{
				MarkdownDescription: "Whether to restrict BitTorrent and other peer-to-peer file sharing traffic. When set to `true`, the system will block P2P traffic across the network. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"suppression": schema.SingleNestedAttribute{
				MarkdownDescription: "Suppression configuration for IPS. This allows you to customize which alerts are suppressed or tracked, and define whitelisted traffic that should never trigger IPS alerts.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"alerts": schema.ListNestedAttribute{
						MarkdownDescription: "Alert suppressions. Each entry defines a specific IPS alert that should be suppressed or tracked differently from the default behavior.",
						Optional:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"category": schema.StringAttribute{
									MarkdownDescription: "Category of the alert to suppress. This should match one of the categories from the enabled_categories list.",
									Required:            true,
								},
								//"gid": schema.Int64Attribute{
								//	MarkdownDescription: "Group ID of the alert to suppress. This is a numeric identifier for the alert group in the IPS ruleset.",
								//	Required:            true,
								//},
								//"id": schema.Int64Attribute{
								//	MarkdownDescription: "ID of the alert to suppress. This is a numeric identifier for the specific alert in the IPS ruleset.",
								//	Required:            true,
								//},
								"signature": schema.StringAttribute{
									MarkdownDescription: "Signature name of the alert to suppress. This is a human-readable identifier for the alert in the IPS ruleset.",
									Required:            true,
								},
								"tracking": schema.ListNestedAttribute{
									MarkdownDescription: "Tracking configuration for the alert. This defines how the system should track occurrences of this alert based on source/destination addresses.",
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.List{
										listplanmodifier.UseStateForUnknown(),
									},
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"direction": schema.StringAttribute{
												MarkdownDescription: "Direction for tracking. Valid values are:\n" +
													"  * `src` - Track by source address\n" +
													"  * `dest` - Track by destination address\n" +
													"  * `both` - Track by both source and destination addresses",
												Required: true,
												Validators: []validator.String{
													stringvalidator.OneOf("src", "dest", "both"),
												},
											},
											"mode": schema.StringAttribute{
												MarkdownDescription: "Mode for tracking. Valid values are:\n" +
													"  * `ip` - Track by individual IP address\n" +
													"  * `subnet` - Track by subnet\n" +
													"  * `network` - Track by network ID",
												Required: true,
												Validators: []validator.String{
													stringvalidator.OneOf("ip", "subnet", "network"),
												},
											},
											"value": schema.StringAttribute{
												MarkdownDescription: "Value for tracking. The meaning depends on the mode:\n" +
													"  * For `ip` mode: An IP address (e.g., `192.168.1.100`)\n" +
													"  * For `subnet` mode: A CIDR notation subnet (e.g., `192.168.1.0/24`)\n" +
													"  * For `network` mode: A network ID from your UniFi configuration",
												Required: true,
											},
										},
									},
								},
								"type": schema.StringAttribute{
									MarkdownDescription: "Type of suppression. Valid values are:\n" +
										"  * `all` - Suppress all occurrences of this alert\n" +
										"  * `track` - Only track this alert according to the tracking configuration",
									Required: true,
									Validators: []validator.String{
										stringvalidator.OneOf("all", "track"),
									},
								},
							},
						},
					},
					"whitelist": schema.ListNestedAttribute{
						MarkdownDescription: "Whitelist configuration. Each entry defines traffic that should never trigger IPS alerts, regardless of other rules.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"direction": schema.StringAttribute{
									MarkdownDescription: "Direction for whitelist. Valid values are:\n" +
										"  * `src` - Whitelist by source address\n" +
										"  * `dst` - Whitelist by destination address\n" +
										"  * `both` - Whitelist by both source and destination addresses",
									Required: true,
									Validators: []validator.String{
										stringvalidator.OneOf("src", "dst", "both"),
									},
								},
								"mode": schema.StringAttribute{
									MarkdownDescription: "Mode for whitelist. Valid values are:\n" +
										"  * `ip` - Whitelist by individual IP address\n" +
										"  * `subnet` - Whitelist by subnet\n" +
										"  * `network` - Whitelist by network ID",
									Required: true,
									Validators: []validator.String{
										stringvalidator.OneOf("ip", "subnet", "network"),
									},
								},
								"value": schema.StringAttribute{
									MarkdownDescription: "Value for whitelist. The meaning depends on the mode:\n" +
										"  * For `ip` mode: An IP address (e.g., `192.168.1.100`)\n" +
										"  * For `subnet` mode: A CIDR notation subnet (e.g., `192.168.1.0/24`)\n" +
										"  * For `network` mode: A network ID from your UniFi configuration",
									Required: true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func NewIpsResource() resource.Resource {
	r := &ipsResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_ips",
		func() *ipsModel { return &ipsModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingIps(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			return client.UpdateSettingIps(ctx, site, body.(*unifi.SettingIps))
		},
	)
	return r
}

var (
	_ base.ResourceModel                    = &ipsModel{}
	_ resource.Resource                     = &ipsResource{}
	_ resource.ResourceWithConfigure        = &ipsResource{}
	_ resource.ResourceWithConfigValidators = &ipsResource{}
	_ resource.ResourceWithModifyPlan       = &ipsResource{}
)

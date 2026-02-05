package firewall

import (
	"context"
	"fmt"
	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/go-unifi/unifi/features"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	ut "github.com/filipowm/terraform-provider-unifi/internal/provider/types"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/utils"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"maps"
)

var (
	_ resource.Resource                     = &firewallZonePolicyResource{}
	_ resource.ResourceWithConfigure        = &firewallZonePolicyResource{}
	_ resource.ResourceWithConfigValidators = &firewallZonePolicyResource{}
	_ resource.ResourceWithImportState      = &firewallZonePolicyResource{}
	_ resource.ResourceWithModifyPlan       = &firewallZonePolicyResource{}
	_ base.Resource                         = &firewallZonePolicyResource{}
)

func mergedTargetAttributes(additional map[string]schema.Attribute) map[string]schema.Attribute {
	attrs := map[string]schema.Attribute{
		"ip_group_id": schema.StringAttribute{
			MarkdownDescription: "ID of the source IP group.",
			Optional:            true,
		},
		"ips": schema.ListAttribute{
			MarkdownDescription: "List of source IPs.",
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.List{
				listvalidator.ValueStringsAre(
					validators.IPv4(),
				),
			},
		},
		"match_opposite_ips": schema.BoolAttribute{
			MarkdownDescription: "Whether to match opposite IPs.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"match_opposite_ports": schema.BoolAttribute{
			MarkdownDescription: "Whether to match opposite ports.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"port": schema.Int32Attribute{
			MarkdownDescription: "Source port.",
			Optional:            true,
			Validators: []validator.Int32{
				int32validator.Between(1, 65535),
			},
		},
		"port_group_id": schema.StringAttribute{
			MarkdownDescription: "ID of the source port group.",
			Optional:            true,
		},
		"zone_id": schema.StringAttribute{
			MarkdownDescription: "ID of the firewall zone.",
			Required:            true,
		},
	}
	maps.Copy(attrs, additional)
	return attrs
}

type FirewallPolicyTargetModel struct {
	IPGroupID          types.String `tfsdk:"ip_group_id"`
	IPs                types.List   `tfsdk:"ips"`
	MatchOppositeIPs   types.Bool   `tfsdk:"match_opposite_ips"`
	MatchOppositePorts types.Bool   `tfsdk:"match_opposite_ports"`
	Port               types.Int32  `tfsdk:"port"`
	PortGroupID        types.String `tfsdk:"port_group_id"`
	ZoneID             types.String `tfsdk:"zone_id"`
}

func (m *FirewallPolicyTargetModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"ip_group_id":          types.StringType,
		"ips":                  types.ListType{ElemType: types.StringType},
		"match_opposite_ips":   types.BoolType,
		"match_opposite_ports": types.BoolType,
		"port":                 types.Int32Type,
		"port_group_id":        types.StringType,
		"zone_id":              types.StringType,
	}
}

func NewFirewallPolicyTargetModel(ipGroupId string, ips []string, matchOppositeIps, matchOppositePorts bool, port int, portGroupId, zoneId string) *FirewallPolicyTargetModel {
	diags := diag.Diagnostics{}
	m := &FirewallPolicyTargetModel{
		IPGroupID:          ut.StringOrNull(ipGroupId),
		IPs:                types.ListNull(types.StringType),
		MatchOppositeIPs:   types.BoolValue(matchOppositeIps),
		MatchOppositePorts: types.BoolValue(matchOppositePorts),
		Port:               ut.Int32OrNull(port),
		PortGroupID:        ut.StringOrNull(portGroupId),
		ZoneID:             types.StringValue(zoneId),
	}

	// Handle IPs list
	if len(ips) > 0 {
		lIps, d := types.ListValueFrom(context.Background(), types.StringType, ips)
		diags.Append(d...)
		m.IPs = lIps
	}
	return m
}

// FirewallZonePolicySourceModel represents the source configuration for a firewall zone policy
type FirewallZonePolicySourceModel struct {
	FirewallPolicyTargetModel
	ClientMACs            types.List   `tfsdk:"client_macs"`
	MAC                   types.String `tfsdk:"mac"`
	MACs                  types.List   `tfsdk:"macs"`
	MatchOppositeNetworks types.Bool   `tfsdk:"match_opposite_networks"`
	NetworkIDs            types.List   `tfsdk:"network_ids"`
}

func (m *FirewallZonePolicySourceModel) AttributeTypes() map[string]attr.Type {
	attrs := map[string]attr.Type{
		"client_macs": types.ListType{
			ElemType: types.StringType,
		},
		"mac": types.StringType,
		"macs": types.ListType{
			ElemType: types.StringType,
		},
		"match_opposite_networks": types.BoolType,
		"network_ids": types.ListType{
			ElemType: types.StringType,
		},
	}
	maps.Copy(attrs, m.FirewallPolicyTargetModel.AttributeTypes())
	return attrs
}

// FirewallZonePolicyDestinationModel represents the destination configuration for a firewall zone policy
type FirewallZonePolicyDestinationModel struct {
	FirewallPolicyTargetModel
	AppCategoryIDs types.List `tfsdk:"app_category_ids"`
	AppIDs         types.List `tfsdk:"app_ids"`
	Regions        types.List `tfsdk:"regions"`
	WebDomains     types.List `tfsdk:"web_domains"`
}

func (m *FirewallZonePolicyDestinationModel) AttributeTypes() map[string]attr.Type {
	attrs := map[string]attr.Type{
		"app_category_ids": types.ListType{
			ElemType: types.StringType,
		},
		"app_ids": types.ListType{
			ElemType: types.StringType,
		},
		"regions": types.ListType{
			ElemType: types.StringType,
		},
		"web_domains": types.ListType{
			ElemType: types.StringType,
		},
	}
	maps.Copy(attrs, m.FirewallPolicyTargetModel.AttributeTypes())
	return attrs
}

// FirewallZonePolicyScheduleModel represents the schedule configuration for a firewall zone policy
type FirewallZonePolicyScheduleModel struct {
	Date         types.String `tfsdk:"date"`
	DateEnd      types.String `tfsdk:"date_end"`
	DateStart    types.String `tfsdk:"date_start"`
	Mode         types.String `tfsdk:"mode"`
	RepeatOnDays types.List   `tfsdk:"repeat_on_days"`
	TimeAllDay   types.Bool   `tfsdk:"time_all_day"`
	TimeTo       types.String `tfsdk:"time_to"`
	TimeFrom     types.String `tfsdk:"time_from"`
}

func (m *FirewallZonePolicyScheduleModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"date":       types.StringType,
		"date_end":   types.StringType,
		"date_start": types.StringType,
		"mode":       types.StringType,
		"repeat_on_days": types.ListType{
			ElemType: types.StringType,
		},
		"time_all_day": types.BoolType,
		"time_to":      types.StringType,
		"time_from":    types.StringType,
	}
}

// FirewallZonePolicyModel represents the data model for firewall zone policies in the UniFi controller
type FirewallZonePolicyModel struct {
	base.Model
	Action                 types.String `tfsdk:"action"`
	AutoAllowReturnTraffic types.Bool   `tfsdk:"auto_allow_return_traffic"`
	ConnectionStateType    types.String `tfsdk:"connection_state_type"`
	ConnectionStates       types.List   `tfsdk:"connection_states"`
	Description            types.String `tfsdk:"description"`
	Destination            types.Object `tfsdk:"destination"`
	Enabled                types.Bool   `tfsdk:"enabled"`
	IPVersion              types.String `tfsdk:"ip_version"`
	Index                  types.Int64  `tfsdk:"index"`
	Logging                types.Bool   `tfsdk:"logging"`
	MatchIPSecType         types.String `tfsdk:"match_ip_sec_type"`
	MatchOppositeProtocol  types.Bool   `tfsdk:"match_opposite_protocol"`
	Name                   types.String `tfsdk:"name"`
	Protocol               types.String `tfsdk:"protocol"`
	Schedule               types.Object `tfsdk:"schedule"`
	Source                 types.Object `tfsdk:"source"`
}

func (m *FirewallZonePolicyModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.FirewallZonePolicy{
		ID:                    m.ID.ValueString(),
		SiteID:                m.Site.ValueString(),
		Action:                m.Action.ValueString(),
		CreateAllowRespond:    m.AutoAllowReturnTraffic.ValueBool(),
		ConnectionStateType:   m.ConnectionStateType.ValueString(),
		Description:           m.Description.ValueString(),
		Enabled:               m.Enabled.ValueBool(),
		IPVersion:             m.IPVersion.ValueString(),
		Index:                 int(m.Index.ValueInt64()),
		Logging:               m.Logging.ValueBool(),
		MatchIPSecType:        m.MatchIPSecType.ValueString(),
		MatchOppositeProtocol: m.MatchOppositeProtocol.ValueBool(),
		Name:                  m.Name.ValueString(),
		Protocol:              m.Protocol.ValueString(),
	}
	diags.Append(m.ConnectionStates.ElementsAs(ctx, &model.ConnectionStates, false)...)

	if !ut.IsEmptyString(m.MatchIPSecType) {
		model.MatchIPSec = true
	} else {
		model.MatchIPSec = false
	}
	// Handle Source object
	if ut.IsDefined(m.Source) {
		var source FirewallZonePolicySourceModel
		diags.Append(m.Source.As(ctx, &source, basetypes.ObjectAsOptions{})...)

		unifiSource := &unifi.FirewallZonePolicySource{
			MatchOppositeIPs:      source.MatchOppositeIPs.ValueBool(),
			MatchOppositeNetworks: source.MatchOppositeNetworks.ValueBool(),
			MatchOppositePorts:    source.MatchOppositePorts.ValueBool(),
			MatchingTarget:        "ANY",
			PortMatchingType:      "ANY",
			ZoneID:                source.ZoneID.ValueString(),
		}

		if ut.IsDefined(source.MAC) {
			unifiSource.MAC = source.MAC.ValueString()
			unifiSource.MatchMAC = true
		} else {
			unifiSource.MatchMAC = false
		}

		if ut.IsDefined(source.PortGroupID) {
			unifiSource.PortMatchingType = "OBJECT"
			unifiSource.PortGroupID = source.PortGroupID.ValueString()
		}

		if ut.IsDefined(source.Port) {
			unifiSource.PortMatchingType = "SPECIFIC"
			unifiSource.Port = int(source.Port.ValueInt32())
		}

		if len(source.ClientMACs.Elements()) > 0 {
			diags.Append(ut.ListElementsAs(source.ClientMACs, &unifiSource.ClientMACs)...)
			unifiSource.MatchingTarget = "CLIENT"
			unifiSource.MatchingTargetType = "SPECIFIC"
		}

		if len(source.IPs.Elements()) > 0 {
			diags.Append(ut.ListElementsAs(source.IPs, &unifiSource.IPs)...)
			unifiSource.MatchingTarget = "IP"
			unifiSource.MatchingTargetType = "SPECIFIC"
		}
		if ut.IsDefined(source.IPGroupID) {
			unifiSource.IPGroupID = source.IPGroupID.ValueString()
			unifiSource.MatchingTarget = "IP"
			unifiSource.MatchingTargetType = "OBJECT"
		}
		if len(source.MACs.Elements()) > 0 {
			diags.Append(ut.ListElementsAs(source.MACs, &unifiSource.MACs)...)
			unifiSource.MatchingTarget = "MAC"
			unifiSource.MatchingTargetType = "SPECIFIC"
		}
		if len(source.NetworkIDs.Elements()) > 0 {
			diags.Append(ut.ListElementsAs(source.NetworkIDs, &unifiSource.NetworkIDs)...)
			unifiSource.MatchingTarget = "NETWORK"
			unifiSource.MatchingTargetType = "SPECIFIC"
		}
		model.Source = *unifiSource
	}

	// Handle Destination object
	if ut.IsDefined(m.Destination) {
		var destination FirewallZonePolicyDestinationModel
		diags.Append(m.Destination.As(ctx, &destination, basetypes.ObjectAsOptions{})...)

		unifiDestination := &unifi.FirewallZonePolicyDestination{
			MatchOppositeIPs:   destination.MatchOppositeIPs.ValueBool(),
			MatchOppositePorts: destination.MatchOppositePorts.ValueBool(),
			MatchingTarget:     "ANY",
			PortMatchingType:   "ANY",
			ZoneID:             destination.ZoneID.ValueString(),
		}

		if ut.IsDefined(destination.PortGroupID) {
			unifiDestination.PortMatchingType = "OBJECT"
			unifiDestination.PortGroupID = destination.PortGroupID.ValueString()
		}

		if ut.IsDefined(destination.Port) {
			unifiDestination.PortMatchingType = "SPECIFIC"
			unifiDestination.Port = int(destination.Port.ValueInt32())
		}

		if len(destination.AppCategoryIDs.Elements()) > 0 {
			diags.Append(ut.ListElementsAs(destination.AppCategoryIDs, &unifiDestination.AppCategoryIDs)...)
			unifiDestination.MatchingTarget = "APP_CATEGORY"
		}

		if len(destination.AppIDs.Elements()) > 0 {
			diags.Append(ut.ListElementsAs(destination.AppIDs, &unifiDestination.AppIDs)...)
			unifiDestination.MatchingTarget = "APP"
		}

		if len(destination.IPs.Elements()) > 0 {
			diags.Append(ut.ListElementsAs(destination.IPs, &unifiDestination.IPs)...)
			unifiDestination.MatchingTarget = "IP"
			unifiDestination.MatchingTargetType = "SPECIFIC"
		}
		if ut.IsDefined(destination.IPGroupID) {
			unifiDestination.IPGroupID = destination.IPGroupID.ValueString()
			unifiDestination.MatchingTarget = "IP"
			unifiDestination.MatchingTargetType = "OBJECT"
		}
		if len(destination.Regions.Elements()) > 0 {
			diags.Append(ut.ListElementsAs(destination.Regions, &unifiDestination.Regions)...)
			unifiDestination.MatchingTarget = "REGION"
			unifiDestination.MatchingTargetType = "SPECIFIC"
		}
		if len(destination.WebDomains.Elements()) > 0 {
			diags.Append(ut.ListElementsAs(destination.WebDomains, &unifiDestination.WebDomains)...)
			unifiDestination.MatchingTarget = "WEB"
			unifiDestination.MatchingTargetType = "SPECIFIC"
		}
		model.Destination = *unifiDestination
	}

	// Handle Schedule object
	if ut.IsDefined(m.Schedule) {
		var schedule FirewallZonePolicyScheduleModel
		diags.Append(m.Schedule.As(ctx, &schedule, basetypes.ObjectAsOptions{})...)

		unifiSchedule := &unifi.FirewallZonePolicySchedule{
			Date:           schedule.Date.ValueString(),
			DateEnd:        schedule.DateEnd.ValueString(),
			DateStart:      schedule.DateStart.ValueString(),
			Mode:           schedule.Mode.ValueString(),
			TimeAllDay:     schedule.TimeAllDay.ValueBool(),
			TimeRangeEnd:   schedule.TimeTo.ValueString(),
			TimeRangeStart: schedule.TimeFrom.ValueString(),
		}
		if len(schedule.RepeatOnDays.Elements()) > 0 {
			diags.Append(ut.ListElementsAs(schedule.RepeatOnDays, &unifiSchedule.RepeatOnDays)...)
		}
		model.Schedule = *unifiSchedule
	}

	return model, diags
}

func (m *FirewallZonePolicyModel) mergeSource(ctx context.Context, model *unifi.FirewallZonePolicy) diag.Diagnostics {
	diags := diag.Diagnostics{}

	sourceModel := &FirewallZonePolicySourceModel{
		FirewallPolicyTargetModel: *NewFirewallPolicyTargetModel(model.Source.IPGroupID, model.Source.IPs, model.Source.MatchOppositeIPs, model.Source.MatchOppositePorts, model.Source.Port, model.Source.PortGroupID, model.Source.ZoneID),
		MAC:                       ut.StringOrNull(model.Source.MAC),
		MatchOppositeNetworks:     types.BoolValue(model.Source.MatchOppositeNetworks),
		MACs:                      types.ListNull(types.StringType),
		ClientMACs:                types.ListNull(types.StringType),
		NetworkIDs:                types.ListNull(types.StringType),
	}

	switch model.Source.MatchingTarget {
	// TODO !
	case "MAC":
		macs, d := types.ListValueFrom(ctx, types.StringType, model.Source.MACs)
		diags.Append(d...)
		sourceModel.MACs = macs
	case "NETWORK":
		networks, d := types.ListValueFrom(ctx, types.StringType, model.Source.NetworkIDs)
		diags.Append(d...)
		sourceModel.NetworkIDs = networks
	case "CLIENT":
		clientMACs, d := types.ListValueFrom(ctx, types.StringType, model.Source.ClientMACs)
		diags.Append(d...)
		sourceModel.ClientMACs = clientMACs
	case "IP":
	case "ANY":
		// do nothing as handled commonly
	default:
		diags.AddWarning("Unexpected matching target", fmt.Sprintf("Source matching target is %s, which is not supported by the provider", model.Source.MatchingTarget))
	}

	// Create object value from source model
	sourceObject, d := types.ObjectValueFrom(ctx, sourceModel.AttributeTypes(), &sourceModel)
	diags.Append(d...)
	m.Source = sourceObject

	return diags
}

func (m *FirewallZonePolicyModel) mergeDestination(ctx context.Context, model *unifi.FirewallZonePolicy) diag.Diagnostics {
	diags := diag.Diagnostics{}
	destModel := &FirewallZonePolicyDestinationModel{
		FirewallPolicyTargetModel: *NewFirewallPolicyTargetModel(model.Destination.IPGroupID, model.Destination.IPs, model.Destination.MatchOppositeIPs, model.Destination.MatchOppositePorts, model.Destination.Port, model.Destination.PortGroupID, model.Destination.ZoneID),
		AppCategoryIDs:            types.ListNull(types.StringType),
		AppIDs:                    types.ListNull(types.StringType),
		Regions:                   types.ListNull(types.StringType),
		WebDomains:                types.ListNull(types.StringType),
	}
	switch model.Destination.MatchingTarget {
	case "APP_CATEGORY":
		appCategories, d := types.ListValueFrom(ctx, types.StringType, model.Destination.AppCategoryIDs)
		diags.Append(d...)
		destModel.AppCategoryIDs = appCategories
	case "APP":
		apps, d := types.ListValueFrom(ctx, types.StringType, model.Destination.AppIDs)
		diags.Append(d...)
		destModel.AppIDs = apps
	case "REGION":
		regions, d := types.ListValueFrom(ctx, types.StringType, model.Destination.Regions)
		diags.Append(d...)
		destModel.Regions = regions
	case "WEB":
		webs, d := types.ListValueFrom(ctx, types.StringType, model.Destination.WebDomains)
		diags.Append(d...)
		destModel.WebDomains = webs
	case "IP":
	case "ANY":
		// do nothing as handled commonly
	default:
		diags.AddWarning("Unexpected matching target", fmt.Sprintf("Destination matching target is %s, which is not supported by the provider", model.Source.MatchingTarget))
	}

	// Create object value from source model
	destObject, d := types.ObjectValueFrom(ctx, destModel.AttributeTypes(), destModel)
	diags.Append(d...)
	m.Destination = destObject

	return diags
}

func (m *FirewallZonePolicyModel) mergeSchedule(ctx context.Context, model *unifi.FirewallZonePolicy) diag.Diagnostics {
	diags := diag.Diagnostics{}
	// Set Schedule object
	scheduleModel := &FirewallZonePolicyScheduleModel{
		Date:         ut.StringOrNull(model.Schedule.Date),
		DateEnd:      ut.StringOrNull(model.Schedule.DateEnd),
		DateStart:    ut.StringOrNull(model.Schedule.DateStart),
		Mode:         ut.StringOrNull(model.Schedule.Mode),
		TimeAllDay:   types.BoolValue(model.Schedule.TimeAllDay),
		TimeTo:       ut.StringOrNull(model.Schedule.TimeRangeEnd),
		TimeFrom:     ut.StringOrNull(model.Schedule.TimeRangeStart),
		RepeatOnDays: types.ListNull(types.StringType),
	}
	if model.Schedule.Mode == "EVERY_WEEK" || model.Schedule.Mode == "CUSTOM" {
		days, d := types.ListValueFrom(ctx, types.StringType, model.Schedule.RepeatOnDays)
		diags.Append(d...)
		scheduleModel.RepeatOnDays = days
	}
	// `always`, `every_day` (start, end T), `every_week` (days, start / end T, all day),
	// `one_time_only` (date, start / end T), or `custom` (start / end D, start / end T, days, all day).
	// Create object value from schedule model
	scheduleObject, scheduleDiags := types.ObjectValueFrom(ctx, scheduleModel.AttributeTypes(), scheduleModel)
	diags.Append(scheduleDiags...)
	m.Schedule = scheduleObject
	return diags
}

func (m *FirewallZonePolicyModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.FirewallZonePolicy)
	if !ok {
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.Action = types.StringValue(model.Action)
	m.AutoAllowReturnTraffic = types.BoolValue(model.CreateAllowRespond)
	m.ConnectionStateType = types.StringValue(model.ConnectionStateType)
	if model.Description != "" {
		m.Description = types.StringValue(model.Description)
	}
	m.Enabled = types.BoolValue(model.Enabled)
	m.IPVersion = types.StringValue(model.IPVersion)
	m.Index = types.Int64Value(int64(model.Index))
	m.Logging = types.BoolValue(model.Logging)
	m.MatchOppositeProtocol = types.BoolValue(model.MatchOppositeProtocol)
	m.Name = types.StringValue(model.Name)
	m.Protocol = types.StringValue(model.Protocol)

	if model.MatchIPSecType != "" {
		m.MatchIPSecType = types.StringValue(model.MatchIPSecType)
	} else {
		m.MatchIPSecType = types.StringNull()
	}

	diags.Append(m.mergeSource(ctx, model)...)
	diags.Append(m.mergeDestination(ctx, model)...)
	diags.Append(m.mergeSchedule(ctx, model)...)

	// Set ConnectionStates
	if model.ConnectionStateType == "custom" {
		connectionStates, d := types.ListValueFrom(ctx, types.StringType, model.ConnectionStates)
		diags.Append(d...)
		m.ConnectionStates = connectionStates
	} else {
		m.ConnectionStates = types.ListNull(types.StringType)
	}

	return diags
}

type firewallZonePolicyResource struct {
	*base.GenericResource[*FirewallZonePolicyModel]
}

func (r *firewallZonePolicyResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		validators.RequiredSimpleTogetherIf("connection_state_type", types.StringValue("CUSTOM"), "connection_states"),
		validators.RequiredSimpleTogetherIf("connection_state_type", types.StringValue("CUSTOM"), "connection_states"),
	}
}

func (r *firewallZonePolicyResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	site, diags := r.GetClient().ResolveSiteFromConfig(ctx, req.Config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	resp.Diagnostics.Append(r.RequireFeaturesEnabled(ctx, site, features.ZoneBasedFirewall, features.ZoneBasedFirewallMigration)...)
}

// NewFirewallZonePolicyResource creates a new instance of the firewall zone policy resource
func NewFirewallZonePolicyResource() resource.Resource {
	return &firewallZonePolicyResource{
		GenericResource: base.NewGenericResource(
			"unifi_firewall_zone_policy",
			func() *FirewallZonePolicyModel { return &FirewallZonePolicyModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetFirewallZonePolicy(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					return client.CreateFirewallZonePolicy(ctx, site, model.(*unifi.FirewallZonePolicy))
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					return client.UpdateFirewallZonePolicy(ctx, site, model.(*unifi.FirewallZonePolicy))
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteFirewallZonePolicy(ctx, site, id)
				},
			},
		),
	}
}

// Schema defines the schema for the resource
func (r *firewallZonePolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_firewall_zone_policy` resource manages firewall policies between zones in the UniFi controller. " +
			"This resource allows you to create, update, and delete policies that define allowed or blocked traffic between zones.\n\n" +
			"!> This is experimental feature, that requires UniFi OS 9.0.0 or later and Zone Based Firewall feature enabled. " +
			"Check [official documentation](https://help.ui.com/hc/en-us/articles/28223082254743-Migrating-to-Zone-Based-Firewalls-in-UniFi) how to migate to Zone-Based firewalls.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the firewall zone policy.",
				Required:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable the policy",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the firewall zone policy.",
				Optional:            true,
			},
			"action": schema.StringAttribute{
				MarkdownDescription: "Determines which action to take on matching traffic. Must be one of `BLOCK`, `ALLOW`, or `REJECT`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("BLOCK", "ALLOW", "REJECT"),
				},
			},
			"auto_allow_return_traffic": schema.BoolAttribute{
				MarkdownDescription: "Creates a built-in policy for the opposite Zone Pair to automatically allow the return traffic. If disabled, return traffic must be manually allowed",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"index": schema.Int64Attribute{
				MarkdownDescription: "Priority index for the policy. This value is assigned by the UniFi controller and cannot be set directly. " +
					"To control policy ordering, use the `unifi_firewall_zone_policy_order` resource (planned for future release).",
				Computed: true,
			},
			"logging": schema.BoolAttribute{
				MarkdownDescription: "Enable to generate syslog entries when traffic is matched.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "Optionally match a specific protocol. Valid values include: `all`, `tcp_udp`, `tcp`, `udp`, etc.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("all"),
				Validators: []validator.String{
					stringvalidator.OneOf(
						"all", "tcp_udp", "tcp", "udp", "ah", "dccp", "eigrp", "esp", "gre",
						"icmp", "icmpv6", "igmp", "igp", "ip", "ipcomp", "ipip", "ipv6",
						"isis", "l2tp", "manet", "mobility-header", "mpls-in-ip", "number",
						"ospf", "pim", "pup", "rdp", "rohc", "rspf", "rcvp", "sctp", "shim6",
						"skip", "st", "vmtp", "vrrp", "wesp", "xtp",
					),
				},
			},
			"ip_version": schema.StringAttribute{
				MarkdownDescription: "Optionally match on only IPv4 or IPv6. Valid values are `BOTH`, `IPV4`, or `IPV6`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("BOTH"),
				Validators: []validator.String{
					stringvalidator.OneOf("BOTH", "IPV4", "IPV6"),
				},
			},
			"connection_state_type": schema.StringAttribute{
				MarkdownDescription: "Optionally match on a firewall connection state such as traffic associated with an already existing connection. Valid values are `ALL`, `RESPOND_ONLY`, or `CUSTOM`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("ALL"),
				Validators: []validator.String{
					stringvalidator.OneOf("ALL", "RESPOND_ONLY", "CUSTOM"),
				},
			},
			"connection_states": schema.ListAttribute{
				MarkdownDescription: "Connection states to match when `connection_state_type` is `CUSTOM`. Valid values include `ESTABLISHED`, `NEW`, `RELATED`, and `INVALID`.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.UniqueValues(),
					listvalidator.ValueStringsAre(
						stringvalidator.OneOf("ESTABLISHED", "NEW", "RELATED", "INVALID"),
					),
				},
			},
			"match_ip_sec_type": schema.StringAttribute{
				MarkdownDescription: "Optionally match on traffic encrypted by IPsec. This is typically used for Ipsec Policy-Based VPNs. Valid values are `MATCH_IP_SEC` or `MATCH_NON_IP_SEC`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("MATCH_IP_SEC", "MATCH_NON_IP_SEC"),
				},
			},
			"match_opposite_protocol": schema.BoolAttribute{
				MarkdownDescription: "Whether to match the opposite protocol.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"source": schema.SingleNestedAttribute{
				MarkdownDescription: "The zone matching the source of the traffic. Optionally match on a specific source inside the zone.",
				Required:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: mergedTargetAttributes(map[string]schema.Attribute{
					"mac": schema.StringAttribute{
						MarkdownDescription: "Source MAC address.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(utils.MacAddressRegexp, "must be a valid MAC address"),
							stringvalidator.Any(
								stringvalidator.AlsoRequires(path.MatchRoot("source").AtName("ips")),
								stringvalidator.AlsoRequires(path.MatchRoot("source").AtName("network_ids")),
							),
						},
					},
					"macs": schema.ListAttribute{
						MarkdownDescription: "List of MAC addresses.",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
							listvalidator.ValueStringsAre(
								stringvalidator.RegexMatches(utils.MacAddressRegexp, "must be a valid MAC address"),
							),
							listvalidator.ConflictsWith(
								path.MatchRoot("source").AtName("client_macs"),
								path.MatchRoot("source").AtName("ips"),
								path.MatchRoot("source").AtName("mac"),
								path.MatchRoot("source").AtName("network_ids"),
							),
						},
					},
					"client_macs": schema.ListAttribute{
						MarkdownDescription: "List of client MAC addresses.",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
							listvalidator.ValueStringsAre(
								stringvalidator.RegexMatches(utils.MacAddressRegexp, "must be a valid MAC address"),
							),
							listvalidator.ConflictsWith(
								path.MatchRoot("source").AtName("ips"),
								path.MatchRoot("source").AtName("mac"),
								path.MatchRoot("source").AtName("macs"),
								path.MatchRoot("source").AtName("network_ids"),
							),
						},
					},
					"network_ids": schema.ListAttribute{
						MarkdownDescription: "List of network IDs.",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
							listvalidator.ConflictsWith(
								path.MatchRoot("source").AtName("client_macs"),
								path.MatchRoot("source").AtName("ips"),
								path.MatchRoot("source").AtName("macs"),
							),
						},
					},
					"match_opposite_networks": schema.BoolAttribute{
						MarkdownDescription: "Whether to match opposite networks.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
				}),
			},
			"destination": schema.SingleNestedAttribute{
				MarkdownDescription: "The zone matching the destination of the traffic. Optionally match on a specific destination inside the zone.",
				Required:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: mergedTargetAttributes(map[string]schema.Attribute{
					"app_category_ids": schema.ListAttribute{
						MarkdownDescription: "List of application category IDs.",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
							listvalidator.ConflictsWith(
								path.MatchRoot("destination").AtName("app_ids"),
								path.MatchRoot("destination").AtName("ips"),
								path.MatchRoot("destination").AtName("regions"),
								path.MatchRoot("destination").AtName("web_domains"),
							),
						},
					},
					"app_ids": schema.ListAttribute{
						MarkdownDescription: "List of application IDs.",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
							listvalidator.ConflictsWith(
								path.MatchRoot("destination").AtName("app_category_ids"),
								path.MatchRoot("destination").AtName("ips"),
								path.MatchRoot("destination").AtName("regions"),
								path.MatchRoot("destination").AtName("web_domains"),
							),
						},
					},
					"regions": schema.ListAttribute{
						MarkdownDescription: "List of regions.",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
							listvalidator.ValueStringsAre(validators.CountryCodeAlpha2()),
							listvalidator.ConflictsWith(
								path.MatchRoot("destination").AtName("app_category_ids"),
								path.MatchRoot("destination").AtName("app_ids"),
								path.MatchRoot("destination").AtName("ips"),
								path.MatchRoot("destination").AtName("web_domains"),
							),
						},
					},
					"web_domains": schema.ListAttribute{
						MarkdownDescription: "List of web domains.",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
							listvalidator.ValueStringsAre(
								validators.Hostname(),
							),
							listvalidator.ConflictsWith(
								path.MatchRoot("destination").AtName("app_category_ids"),
								path.MatchRoot("destination").AtName("app_ids"),
								path.MatchRoot("destination").AtName("ips"),
								path.MatchRoot("destination").AtName("regions"),
							),
						},
					},
				}),
			},
			"schedule": schema.SingleNestedAttribute{
				MarkdownDescription: "Enforce this policy at specific times.",
				Optional:            true,
				Computed:            true,
				Default: objectdefault.StaticValue(ut.ObjectValueMust(ctx, &FirewallZonePolicyScheduleModel{
					Mode:         types.StringValue("ALWAYS"),
					TimeAllDay:   types.BoolValue(false),
					RepeatOnDays: types.ListNull(types.StringType),
				})),
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"date": schema.StringAttribute{
						MarkdownDescription: "Date for the schedule.",
						Optional:            true,
						Validators: []validator.String{
							validators.DateFormat,
						},
					},
					"date_end": schema.StringAttribute{
						MarkdownDescription: "End date for the schedule.",
						Optional:            true,
						Validators: []validator.String{
							validators.DateFormat,
						},
					},
					"date_start": schema.StringAttribute{
						MarkdownDescription: "Start date for the schedule.",
						Optional:            true,
						Validators: []validator.String{
							validators.DateFormat,
						},
					},
					"mode": schema.StringAttribute{
						MarkdownDescription: "Schedule mode. Valid values are `ALWAYS`, `EVERY_DAY`, `EVERY_WEEK`, `ONE_TIME_ONLY`, or `CUSTOM`.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("ALWAYS"),
						Validators: []validator.String{
							stringvalidator.OneOf("ALWAYS", "EVERY_DAY", "EVERY_WEEK", "ONE_TIME_ONLY", "CUSTOM"),
						},
					},
					"repeat_on_days": schema.ListAttribute{
						MarkdownDescription: "Days of the week when schedule repeats. Valid values include `mon`, `tue`, `wed`, `thu`, `fri`, `sat`, and `sun`.",
						Optional:            true,
						ElementType:         types.StringType,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
							listvalidator.UniqueValues(),
							listvalidator.ValueStringsAre(
								stringvalidator.OneOf("mon", "tue", "wed", "thu", "fri", "sat", "sun"),
							),
						},
					},
					"time_all_day": schema.BoolAttribute{
						MarkdownDescription: "Whether the schedule applies all day.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"time_from": schema.StringAttribute{
						MarkdownDescription: "Schedule starting time in 24-hour format (HH:MM).",
						Optional:            true,
						Validators: []validator.String{
							validators.TimeFormat,
						},
					},
					"time_to": schema.StringAttribute{
						MarkdownDescription: "Schedule ending time in 24-hour format (HH:MM).",
						Optional:            true,
						Validators: []validator.String{
							validators.TimeFormat,
						},
					},
				},
			},
		},
	}
}

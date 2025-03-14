package settings

import (
	"context"
	"fmt"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SshKeyModel represents an SSH key configuration
type SshKeyModel struct {
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Key     types.String `tfsdk:"key"`
	Comment types.String `tfsdk:"comment"`
}

func (m *SshKeyModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":    types.StringType,
		"type":    types.StringType,
		"key":     types.StringType,
		"comment": types.StringType,
	}
}

// mgmtModel represents the data model for management settings.
type mgmtModel struct {
	base.Model
	AutoUpgrade types.Bool `tfsdk:"auto_upgrade"`
	SshEnabled  types.Bool `tfsdk:"ssh_enabled"`
	SshKeys     types.List `tfsdk:"ssh_key"`
}

func (m *mgmtModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	sshKeys, d := m.getSshKeys(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	return &unifi.SettingMgmt{
		ID:          m.ID.ValueString(),
		Key:         unifi.SettingMgmtKey,
		AutoUpgrade: m.AutoUpgrade.ValueBool(),
		XSshEnabled: m.SshEnabled.ValueBool(),
		XSshKeys:    sshKeys,
	}, diags
}

func (m *mgmtModel) getSshKeys(ctx context.Context) ([]unifi.SettingMgmtXSshKeys, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	var sshKeys []unifi.SettingMgmtXSshKeys

	if m.SshKeys.IsNull() || m.SshKeys.IsUnknown() {
		return sshKeys, diags
	}

	var sshKeyElements []SshKeyModel
	diags.Append(m.SshKeys.ElementsAs(ctx, &sshKeyElements, false)...)
	if diags.HasError() {
		return nil, diags
	}

	for _, sshKey := range sshKeyElements {
		sshKeys = append(sshKeys, unifi.SettingMgmtXSshKeys{
			Name:    sshKey.Name.ValueString(),
			KeyType: sshKey.Type.ValueString(),
			Key:     sshKey.Key.ValueString(),
			Comment: sshKey.Comment.ValueString(),
		})
	}

	return sshKeys, diags
}

func (m *mgmtModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	resp, ok := other.(*unifi.SettingMgmt)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.SettingMgmt, got: %T", other))
		return diags
	}

	m.ID = types.StringValue(resp.ID)
	m.AutoUpgrade = types.BoolValue(resp.AutoUpgrade)
	m.SshEnabled = types.BoolValue(resp.XSshEnabled)

	// Convert SSH keys
	if len(resp.XSshKeys) > 0 {
		sshKeyElements := make([]SshKeyModel, 0, len(resp.XSshKeys))
		for _, sshKey := range resp.XSshKeys {
			sshKeyElements = append(sshKeyElements, SshKeyModel{
				Name:    types.StringValue(sshKey.Name),
				Type:    types.StringValue(sshKey.KeyType),
				Key:     types.StringValue(sshKey.Key),
				Comment: types.StringValue(sshKey.Comment),
			})
		}
		sshKeys, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: (&SshKeyModel{}).AttributeTypes()}, sshKeyElements)
		diags.Append(d...)
		if !diags.HasError() {
			m.SshKeys = sshKeys
		}
	} else {
		m.SshKeys = types.ListNull(types.ObjectType{AttrTypes: (&SshKeyModel{}).AttributeTypes()})
	}

	return diags
}

// NewMgmtResource creates a new instance of the management settings resource.
func NewMgmtResource() resource.Resource {
	return &mgmtResource{
		GenericResource: NewSettingResource(
			"unifi_setting_mgmt",
			func() *mgmtModel { return &mgmtModel{} },
			func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
				return client.GetSettingMgmt(ctx, site)
			},
			func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
				return client.UpdateSettingMgmt(ctx, site, body.(*unifi.SettingMgmt))
			},
		),
	}
}

var (
	_ base.ResourceModel             = &mgmtModel{}
	_ resource.Resource              = &mgmtResource{}
	_ resource.ResourceWithConfigure = &mgmtResource{}
)

type mgmtResource struct {
	*base.GenericResource[*mgmtModel]
}

func (r *mgmtResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_mgmt` resource manages site-wide management settings in the UniFi controller.\n\n" +
			"This resource allows you to configure important management features including:\n" +
			"  * Automatic firmware upgrades for UniFi devices\n" +
			"  * SSH access for advanced configuration and troubleshooting\n" +
			"  * SSH key management for secure remote access\n\n" +
			"These settings affect how the UniFi controller manages devices at the site level. " +
			"They are particularly important for:\n" +
			"  * Maintaining device security through automatic updates\n" +
			"  * Enabling secure remote administration\n" +
			"  * Implementing SSH key-based authentication",
		Attributes: map[string]schema.Attribute{
			"id":   base.ID(),
			"site": base.SiteAttribute(),
			"auto_upgrade": schema.BoolAttribute{
				MarkdownDescription: "Enable automatic firmware upgrades for all UniFi devices at this site. When enabled, devices will automatically " +
					"update to the latest stable firmware version approved for your controller version.",
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"ssh_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable SSH access to UniFi devices at this site. When enabled, you can connect to devices using SSH for advanced " +
					"configuration and troubleshooting. It's recommended to only enable this temporarily when needed.",
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"ssh_key": schema.ListNestedBlock{
				MarkdownDescription: "List of SSH public keys that are allowed to connect to UniFi devices when SSH is enabled. Using SSH keys is more " +
					"secure than password authentication.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "A friendly name for the SSH key to help identify its owner or purpose (e.g., 'admin-laptop' or 'backup-server').",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of SSH key. Common values include:\n" +
								"  * `ssh-rsa` - RSA key (most common)\n" +
								"  * `ssh-ed25519` - Ed25519 key (more secure)\n" +
								"  * `ecdsa-sha2-nistp256` - ECDSA key",
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"key": schema.StringAttribute{
							MarkdownDescription: "The public key string. This is the content that would normally go in an authorized_keys file, " +
								"excluding the type and comment (e.g., 'AAAAB3NzaC1yc2EA...').",
							Optional: true,
						},
						"comment": schema.StringAttribute{
							MarkdownDescription: "An optional comment to provide additional context about the key (e.g., 'generated on 2024-01-01' or 'expires 2025-12-31').",
							Optional:            true,
						},
					},
				},
			},
		},
	}
}

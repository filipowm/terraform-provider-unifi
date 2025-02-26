package settings

import (
	"context"
	"errors"
	"fmt"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/utils"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TODO: probably need to update this to be more like setting_usg,
// using locking, and upsert, more computed, etc.

func ResourceSettingMgmt() *schema.Resource {
	return &schema.Resource{
		Description: "The `unifi_setting_mgmt` resource manages site-wide management settings in the UniFi controller.\n\n" +
			"This resource allows you to configure important management features including:\n" +
			"  * Automatic firmware upgrades for UniFi devices\n" +
			"  * SSH access for advanced configuration and troubleshooting\n" +
			"  * SSH key management for secure remote access\n\n" +
			"These settings affect how the UniFi controller manages devices at the site level. " +
			"They are particularly important for:\n" +
			"  * Maintaining device security through automatic updates\n" +
			"  * Enabling secure remote administration\n" +
			"  * Implementing SSH key-based authentication",

		CreateContext: resourceSettingMgmtCreate,
		ReadContext:   resourceSettingMgmtRead,
		UpdateContext: resourceSettingMgmtUpdate,
		DeleteContext: resourceSettingMgmtDelete,
		Importer: &schema.ResourceImporter{
			StateContext: utils.ImportSiteAndID,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the management settings configuration in the UniFi controller.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"site": {
				Description: "The name of the UniFi site where these management settings should be applied. If not specified, the default site will be used.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},
			"auto_upgrade": {
				Description: "Enable automatic firmware upgrades for all UniFi devices at this site. When enabled, devices will automatically " +
					"update to the latest stable firmware version approved for your controller version.",
				Type:     schema.TypeBool,
				Optional: true,
			},
			"ssh_enabled": {
				Description: "Enable SSH access to UniFi devices at this site. When enabled, you can connect to devices using SSH for advanced " +
					"configuration and troubleshooting. It's recommended to only enable this temporarily when needed.",
				Type:     schema.TypeBool,
				Optional: true,
			},
			"ssh_key": {
				Description: "List of SSH public keys that are allowed to connect to UniFi devices when SSH is enabled. Using SSH keys is more " +
					"secure than password authentication.",
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "A friendly name for the SSH key to help identify its owner or purpose (e.g., 'admin-laptop' or 'backup-server').",
							Type:        schema.TypeString,
							Required:    true,
						},
						"type": {
							Description: "The type of SSH key. Common values include:\n" +
								"  * `ssh-rsa` - RSA key (most common)\n" +
								"  * `ssh-ed25519` - Ed25519 key (more secure)\n" +
								"  * `ecdsa-sha2-nistp256` - ECDSA key",
							Type:     schema.TypeString,
							Required: true,
						},
						"key": {
							Description: "The public key string. This is the content that would normally go in an authorized_keys file, " +
								"excluding the type and comment (e.g., 'AAAAB3NzaC1yc2EA...').",
							Type:     schema.TypeString,
							Optional: true,
						},
						"comment": {
							Description: "An optional comment to provide additional context about the key (e.g., 'generated on 2024-01-01' or 'expires 2025-12-31').",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func setToSshKeys(set *schema.Set) ([]unifi.SettingMgmtXSshKeys, error) {
	var sshKeys []unifi.SettingMgmtXSshKeys
	for _, item := range set.List() {
		data, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected data in block")
		}
		sshKey, err := toSshKey(data)
		if err != nil {
			return nil, fmt.Errorf("unable to create port override: %w", err)
		}
		sshKeys = append(sshKeys, sshKey)
	}
	return sshKeys, nil
}

func toSshKey(data map[string]interface{}) (unifi.SettingMgmtXSshKeys, error) {
	return unifi.SettingMgmtXSshKeys{
		Name:    data["name"].(string),
		KeyType: data["type"].(string),
		Key:     data["key"].(string),
		Comment: data["comment"].(string),
	}, nil
}

func setFromSshKeys(sshKeys []unifi.SettingMgmtXSshKeys) ([]map[string]interface{}, error) {
	list := make([]map[string]interface{}, 0, len(sshKeys))
	for _, sshKey := range sshKeys {
		v, err := fromSshKey(sshKey)
		if err != nil {
			return nil, fmt.Errorf("unable to parse ssh key: %w", err)
		}
		list = append(list, v)
	}
	return list, nil
}

func fromSshKey(sshKey unifi.SettingMgmtXSshKeys) (map[string]interface{}, error) {
	return map[string]interface{}{
		"name":    sshKey.Name,
		"type":    sshKey.KeyType,
		"key":     sshKey.Key,
		"comment": sshKey.Comment,
	}, nil
}

func resourceSettingMgmtGetResourceData(d *schema.ResourceData, meta interface{}) (*unifi.SettingMgmt, error) {
	sshKeys, err := setToSshKeys(d.Get("ssh_key").(*schema.Set))
	if err != nil {
		return nil, fmt.Errorf("unable to process ssh_key block: %w", err)
	}

	return &unifi.SettingMgmt{
		AutoUpgrade: d.Get("auto_upgrade").(bool),
		XSshEnabled: d.Get("ssh_enabled").(bool),
		XSshKeys:    sshKeys,
	}, nil
}

func resourceSettingMgmtCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	req, err := resourceSettingMgmtGetResourceData(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.UpdateSettingMgmt(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return resourceSettingMgmtSetResourceData(resp, d, meta, site)
}

func resourceSettingMgmtSetResourceData(resp *unifi.SettingMgmt, d *schema.ResourceData, meta interface{}, site string) diag.Diagnostics {
	sshKeys, err := setFromSshKeys(resp.XSshKeys)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("site", site)
	d.Set("auto_upgrade", resp.AutoUpgrade)
	d.Set("ssh_enabled", resp.XSshEnabled)
	d.Set("ssh_key", sshKeys)
	return nil
}

func resourceSettingMgmtRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.GetSettingMgmt(ctx, site)
	if errors.Is(err, unifi.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSettingMgmtSetResourceData(resp, d, meta, site)
}

func resourceSettingMgmtUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*base.Client)

	req, err := resourceSettingMgmtGetResourceData(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	req.ID = d.Id()
	site := d.Get("site").(string)
	if site == "" {
		site = c.Site
	}

	resp, err := c.UpdateSettingMgmt(ctx, site, req)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSettingMgmtSetResourceData(resp, d, meta, site)
}

func resourceSettingMgmtDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "unifi_setting_teleport Resource - terraform-provider-unifi"
subcategory: ""
description: |-
  Manages Teleport settings for a UniFi site. Teleport is a secure remote access technology that allows authorized users to connect to UniFi devices from anywhere.
---

# unifi_setting_teleport (Resource)

Manages Teleport settings for a UniFi site. Teleport is a secure remote access technology that allows authorized users to connect to UniFi devices from anywhere.

## Example Usage

```terraform
resource "unifi_setting_teleport" "example" {
  # Enable Teleport remote access functionality
  enabled = true
  
  # Optional subnet configuration for Teleport
  # Specify a CIDR notation subnet for Teleport to use
  subnet = "192.168.100.0/24"
  
  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `enabled` (Boolean) Whether Teleport is enabled.

### Optional

- `site` (String) The name of the UniFi site where this resource should be applied. If not specified, the default site will be used.
- `subnet` (String) The subnet CIDR for Teleport (e.g., `192.168.1.0/24`). Can be empty but must be set explicitly.

### Read-Only

- `id` (String) The unique identifier of this resource.

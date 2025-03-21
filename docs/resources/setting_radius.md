---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "unifi_setting_radius Resource - terraform-provider-unifi"
subcategory: ""
description: |-
  The unifi_setting_radius resource manages the built-in RADIUS server configuration in the UniFi controller.
  This resource allows you to configure:
  Authentication settings for network access controlAccounting settings for tracking user sessionsSecurity features like tunneled replies
  The RADIUS server is commonly used for:
  Enterprise WPA2/WPA3-Enterprise wireless networks802.1X port-based network access controlCentralized user authentication and accounting
  When enabled, the RADIUS server can authenticate clients using the UniFi user database or external authentication sources.
---

# unifi_setting_radius (Resource)

The `unifi_setting_radius` resource manages the built-in RADIUS server configuration in the UniFi controller.

This resource allows you to configure:
  * Authentication settings for network access control
  * Accounting settings for tracking user sessions
  * Security features like tunneled replies

The RADIUS server is commonly used for:
  * Enterprise WPA2/WPA3-Enterprise wireless networks
  * 802.1X port-based network access control
  * Centralized user authentication and accounting

When enabled, the RADIUS server can authenticate clients using the UniFi user database or external authentication sources.

## Example Usage

```terraform
resource "unifi_setting_radius" "example" {
  # Enable RADIUS functionality
  enabled = true
  
  # RADIUS server secret
  secret = "your-secure-secret"
  
  # Optional: Enable RADIUS accounting
  accounting_enabled = true
  
  # Optional: Configure custom ports
  auth_port = 1812
  accounting_port = 1813
  
  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `accounting_enabled` (Boolean) Enable RADIUS accounting to track user sessions, including connection time, data usage, and other metrics. This information can be useful for billing, capacity planning, and security auditing. Defaults to `false`.
- `accounting_port` (Number) The UDP port number for RADIUS accounting communications. The standard port is 1813. Only change this if you need to avoid port conflicts or match specific network requirements. Defaults to `1813`.
- `auth_port` (Number) The UDP port number for RADIUS authentication communications. The standard port is 1812. Only change this if you need to avoid port conflicts or match specific network requirements. Defaults to `1812`.
- `enabled` (Boolean) Enable or disable the built-in RADIUS server. When disabled, no RADIUS authentication or accounting services will be provided, affecting any network services that rely on RADIUS (like WPA2-Enterprise networks). Defaults to `true`.
- `interim_update_interval` (Number) The interval (in seconds) at which the RADIUS server collects and updates statistics from connected clients. Default is 3600 seconds (1 hour). Lower values provide more frequent updates but increase server load. Defaults to `3600`.
- `secret` (String, Sensitive) The shared secret passphrase used to authenticate RADIUS clients (like wireless access points) with the RADIUS server. This should be a strong, random string known only to the server and its clients. Defaults to ``.
- `site` (String) The name of the UniFi site where these RADIUS settings should be applied. If not specified, the default site will be used.
- `tunneled_reply` (Boolean) Enable encrypted communication between the RADIUS server and clients using RADIUS tunneling. This adds an extra layer of security by protecting RADIUS attributes in transit. Defaults to `true`.

### Read-Only

- `id` (String) The unique identifier of the RADIUS settings configuration in the UniFi controller.

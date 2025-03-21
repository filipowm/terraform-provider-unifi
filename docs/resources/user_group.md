---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "unifi_user_group Resource - terraform-provider-unifi"
subcategory: ""
description: |-
  The unifi_user_group resource manages client groups in the UniFi controller, which allow you to apply common settings and restrictions to multiple network clients.
  User groups are primarily used for:
  Implementing Quality of Service (QoS) policiesSetting bandwidth limits for different types of usersOrganizing clients into logical groups (e.g., Staff, Guests, IoT devices)
  Key features include:
  Download rate limitingUpload rate limitingGroup-based policy application
  User groups are particularly useful in:
  Educational environments (different policies for staff and students)Guest networks (limiting guest bandwidth)Shared office spaces (managing different tenant groups)
---

# unifi_user_group (Resource)

The `unifi_user_group` resource manages client groups in the UniFi controller, which allow you to apply common settings and restrictions to multiple network clients.

User groups are primarily used for:
  * Implementing Quality of Service (QoS) policies
  * Setting bandwidth limits for different types of users
  * Organizing clients into logical groups (e.g., Staff, Guests, IoT devices)

Key features include:
  * Download rate limiting
  * Upload rate limiting
  * Group-based policy application

User groups are particularly useful in:
  * Educational environments (different policies for staff and students)
  * Guest networks (limiting guest bandwidth)
  * Shared office spaces (managing different tenant groups)

## Example Usage

```terraform
resource "unifi_user_group" "wifi" {
  name = "wifi"

  qos_rate_max_down = 2000 # 2mbps
  qos_rate_max_up   = 10   # 10kbps
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) A descriptive name for the user group (e.g., 'Staff', 'Guests', 'IoT Devices'). This name will be displayed in the UniFi controller interface and used when assigning clients to the group.

### Optional

- `qos_rate_max_down` (Number) The maximum allowed download speed in Kbps (kilobits per second) for clients in this group. Set to -1 for unlimited. Note: Values of 0 or 1 are not allowed. Defaults to `-1`.
- `qos_rate_max_up` (Number) The maximum allowed upload speed in Kbps (kilobits per second) for clients in this group. Set to -1 for unlimited. Note: Values of 0 or 1 are not allowed. Defaults to `-1`.
- `site` (String) The name of the UniFi site where this user group should be created. If not specified, the default site will be used.

### Read-Only

- `id` (String) The unique identifier of the user group in the UniFi controller. This is automatically assigned.

## Import

Import is supported using the following syntax:

```shell
# import using the ID
terraform import unifi_user_group.wifi 5fe6261995fe130013456a36
```

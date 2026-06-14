resource "unifi_setting_connectivity" "example" {
  # Wireless meshing on/off for the site. Disable on a fully wired
  # network to free the standby mesh radio on each access point.
  enabled = false

  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}

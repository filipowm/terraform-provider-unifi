resource "unifi_setting_teleport" "example" {
  # Enable Teleport remote access functionality
  enabled = true
  
  # Optional subnet configuration for Teleport
  # Specify a CIDR notation subnet for Teleport to use
  subnet = "192.168.100.0/24"
  
  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}

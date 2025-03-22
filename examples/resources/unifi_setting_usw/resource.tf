resource "unifi_setting_usw" "example" {
  # Enable DHCP snooping to protect against rogue DHCP servers
  dhcp_snoop = true
  
  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}

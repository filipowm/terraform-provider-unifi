resource "unifi_setting_magic_site_to_site_vpn" "example" {
  # Enable Magic Site-to-Site VPN functionality
  enabled = true
  
  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}

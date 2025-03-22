resource "unifi_setting_dpi" "example" {
  # Enable Deep Packet Inspection
  enabled = true
  
  # Enable DPI fingerprinting for more accurate application identification
  fingerprinting_enabled = true
  
  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}

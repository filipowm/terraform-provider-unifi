resource "unifi_setting_network_optimization" "example" {
  # Enable network optimization features
  enabled = true
  
  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}

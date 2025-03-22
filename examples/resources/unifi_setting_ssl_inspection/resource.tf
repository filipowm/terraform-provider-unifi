resource "unifi_setting_ssl_inspection" "example" {
  # Configure SSL inspection state
  # Valid options: "off", "simple", "advanced"
  state = "advanced"
  
  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}

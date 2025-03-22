resource "unifi_setting_locale" "example" {
  # Set the timezone using IANA timezone identifier format
  timezone = "America/New_York"
  
  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}

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

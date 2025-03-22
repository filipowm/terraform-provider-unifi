resource "unifi_site" "example" {
  description = "example"
}

resource "unifi_setting_mgmt" "example" {
  # Reference a specific site (optional, defaults to site configured in provider, otherwise "default")
  site = unifi_site.example.name
  
  # Auto upgrade settings
  auto_upgrade = true
  auto_upgrade_hour = 3
  
  # Device management settings
  advanced_feature_enabled = true
  alert_enabled = true
  boot_sound = false
  debug_tools_enabled = true
  direct_connect_enabled = false
  led_enabled = true
  outdoor_mode_enabled = false
  unifi_idp_enabled = false
  wifiman_enabled = true
  
  # SSH access configuration
  ssh_enabled = true
  ssh_auth_password_enabled = true
  ssh_bind_wildcard = false
  ssh_username = "admin"
  
  # Optional: SSH key configuration
  ssh_key = [
    {
      name = "Admin Key"
      type = "ssh-rsa"
      key = "AAAAB3NzaC1yc2EAAAADAQABAAABAQCxxx..."
      comment = "admin@example.com"
    }
  ]
}

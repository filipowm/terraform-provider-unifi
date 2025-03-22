resource "unifi_setting_ntp" "example" {
  # Set NTP mode to manual to specify custom NTP servers
  # Valid options: "auto" or "manual"
  mode = "manual"
  
  # Configure up to four NTP servers
  ntp_server_1 = "time.cloudflare.com"
  ntp_server_2 = "pool.ntp.org"
  ntp_server_3 = "time.google.com"
  ntp_server_4 = "0.pool.ntp.org"
  
  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}

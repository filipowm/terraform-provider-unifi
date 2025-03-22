resource "unifi_setting_auto_speedtest" "example" {
  # Enable automatic speedtest functionality
  enabled = true
  
  # Schedule for running speedtests using cron syntax
  # This example runs at midnight every day
  cron = "0 0 * * *"
  
  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}

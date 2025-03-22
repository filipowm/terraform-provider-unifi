resource "unifi_setting_rsyslogd" "example" {
  # Enable remote syslog functionality
  enabled = true
  
  # Remote syslog server IP address
  ip = "192.168.1.200"
  
  # Remote syslog server port
  port = 514
  
  # Types of log content to send
  # Valid options: "device", "client", "admin_activity"
  contents = ["device", "client", "admin_activity"]
  
  # Enable debug logging
  debug = true
  
  # Netconsole configuration (optional)
  netconsole_enabled = true
  netconsole_host = "192.168.1.150"
  netconsole_port = 1514
  
  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}

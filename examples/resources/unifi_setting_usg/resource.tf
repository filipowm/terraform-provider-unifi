resource "unifi_setting_usg" "example" {
  # Geo IP Filtering Configuration
  geo_ip_filtering = {
    block            = "block"  # Options: "block" or "allow"
    countries        = ["UK", "CN", "AU"]
    traffic_direction = "both"  # Options: "both", "ingress", or "egress"
  }

  # UPNP Configuration
  upnp = {
    nat_pmp_enabled = true
    secure_mode     = true
    wan_interface   = "WAN"
  }

  # DNS Verification Settings
  dns_verification = {
    domain              = "example.com"
    primary_dns_server  = "1.1.1.1"
    secondary_dns_server = "1.0.0.1"
    setting_preference  = "manual"  # Options: "auto" or "manual"
  }

  # TCP Timeout Settings
  tcp_timeouts = {
    close_timeout       = 10
    established_timeout = 3600
    close_wait_timeout  = 20
    fin_wait_timeout    = 30
    last_ack_timeout    = 30
    syn_recv_timeout    = 60
    syn_sent_timeout    = 120
    time_wait_timeout   = 120
  }

  # ARP Cache Configuration
  arp_cache_timeout = "custom"  # Options: "auto" or "custom"
  arp_cache_base_reachable = 60

  # DHCP Configuration
  broadcast_ping = true
  dhcpd_hostfile_update = true
  dhcpd_use_dnsmasq = true
  dnsmasq_all_servers = true

  # DHCP Relay Configuration
  dhcp_relay = {
    agents_packets = "forward"  # Options: "forward" or "replace"
    hop_count = 5
  }
  dhcp_relay_servers = ["10.1.2.3", "10.1.2.4"]

  # Network Tools
  echo_server = "echo.example.com"

  # Protocol Modules
  ftp_module = true
  gre_module = true
  tftp_module = true

  # ICMP & LLDP Settings
  icmp_timeout = 20
  lldp_enable_all = true

  # MSS Clamp Settings
  mss_clamp = "auto"  # Options: "auto" or "custom"
  mss_clamp_mss = 1452

  # Offload Settings
  offload_accounting = true
  offload_l2_blocking = true
  offload_scheduling = false

  # Timeout Settings
  other_timeout = 600
  timeout_setting_preference = "auto"  # Options: "auto" or "custom"

  # Security Settings
  receive_redirects = false
  send_redirects = true
  syn_cookies = true

  # UDP Timeout Settings
  udp_other_timeout = 30
  udp_stream_timeout = 120

  # Specify the site (optional)
  # site = "default"
}

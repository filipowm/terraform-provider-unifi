
resource "unifi_network" "test" {
  name = "My Network"
  purpose = "corporate"
  subnet = "192.168.1.0/24"
  vlan_id = 10
}

resource "unifi_setting_ips" "example" {
  # Set IPS mode to "ips" (Intrusion Prevention System)
  # Other valid options: "ids" (Intrusion Detection System) or "disabled"
  ips_mode = "ips"
  
  # Networks on which IPS/IDS should be enabled
  enabled_networks = [unifi_network.test.id]
  
  # Advanced filtering preference
  # Valid options: "disabled", "manual", or "auto"
  advanced_filtering_preference = "manual"
  
  # Categories of threats to detect/prevent
  enabled_categories = [
    "emerging-dos",
    "emerging-exploit",
    "emerging-malware"
  ]
  
  # Ad blocking configuration
  ad_blocked_networks = [unifi_network.test.id]
  
  # Honeypot configuration
  honeypots = [
    {
      ip_address = "192.168.1.10"
      network_id = unifi_network.test.id
    }
  ]
  
  # DNS filtering configuration
  dns_filters = [
    {
      name        = "Work Filter"
      filter      = "work"
      description = "Block non-work related sites"
      
      # Sites that are always allowed
      allowed_sites = [
        "example.com",
        "company.com"
      ]
      
      # Sites that are always blocked
      blocked_sites = [
        "gaming.example.com",
        "social.example.com"
      ]
      
      # Top-level domains to block
      blocked_tld = [
        "xyz"
      ]
    }
  ]
  
  # Specify the site (optional, defaults to site configured in provider, otherwise "default")
  # site = "default"
}

variable "vlan_id" {
  default = 10
}

resource "unifi_network" "vlan" {
  name    = "wifi-vlan"
  purpose = "corporate"

  subnet       = "10.0.0.1/24"
  vlan_id      = var.vlan_id
  dhcp_start   = "10.0.0.6"
  dhcp_stop    = "10.0.0.254"
  dhcp_enabled = true
}

# Override the default gateway advertised to DHCP clients (DHCP option 3).
# Useful, for example, to point LAN clients at a Tailscale subnet-router node
# (10.0.1.5 here) for site-to-site routing instead of the gateway's own IP.
resource "unifi_network" "custom_gateway" {
  name    = "lan-via-subnet-router"
  purpose = "corporate"

  subnet       = "10.0.1.1/24"
  vlan_id      = 20
  dhcp_start   = "10.0.1.6"
  dhcp_stop    = "10.0.1.254"
  dhcp_enabled = true

  dhcpd_gateway_enabled = true
  dhcpd_gateway         = "10.0.1.5"
}

resource "unifi_network" "wan" {
  name    = "wan"
  purpose = "wan"

  wan_networkgroup = "WAN"
  wan_type         = "pppoe"
  wan_ip           = "192.168.1.1"
  wan_egress_qos   = 1
  wan_username     = "username"
  x_wan_password   = "password"
}

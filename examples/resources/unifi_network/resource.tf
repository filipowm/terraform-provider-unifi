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

# Zone-Based Firewall (UniFi OS 9.x): pin a network to a firewall zone from the
# network side. Use EITHER this `firewall_zone_id` lever OR the zone-side
# `unifi_firewall_zone.networks` argument for a given network — not both, or the two
# resources will fight over the association.
resource "unifi_firewall_zone" "iot" {
  name     = "iot"
  networks = []
}

resource "unifi_network" "iot" {
  name    = "iot-vlan"
  purpose = "corporate"

  subnet  = "10.0.20.1/24"
  vlan_id = 20

  firewall_zone_id = unifi_firewall_zone.iot.id
}

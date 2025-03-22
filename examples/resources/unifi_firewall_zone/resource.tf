resource "unifi_network" "network" {
    name    = "my-network"
    purpose = "corporate"
    subnet  = "10.0.10.0/24"
    vlan_id = "400"
}

resource "unifi_firewall_zone" "zone" {
    name     = "my-zone"
    networks = [unifi_network.network.id]
}
resource "unifi_network" "network" {
    name    = "my-network"
    purpose = "corporate"
    subnet  = "10.0.10.0/24"
    vlan_id = "400"
}

resource "unifi_firewall_zone" "src" {
    name = "my-source-zone"
    networks = [unifi_network.network.id]
}

resource "unifi_firewall_zone" "dst" {
    name = "my-destination-zone"
}

# Allow TCP/UDP traffic from any ip and port other than 192.168.1.1 and 443 in `src` zone to `dst` zone
resource "unifi_firewall_zone_policy" "policy" {
    name     = "my-zone-policy"
    action   = "ALLOW"
    protocol = "tcp_udp"

    source = {
        zone_id              = unifi_firewall_zone.src.id
        ips = ["192.168.1.1"]
        port                 = "443"
        match_opposite_ips   = true
        match_opposite_ports = true
    }

    destination = {
        zone_id = unifi_firewall_zone.dst.id
    }

    schedule = {
        mode         = "EVERY_DAY"
        time_all_day = false
        time_from    = "08:00"
        time_to      = "17:00"
    }
}

resource "unifi_firewall_group" "web-ports" {
    name = "web-apps"
    type = "port-group"
    members = ["80", "443"]
}

# Block TCP/UDP traffic from any ip and port in `src` zone to `dst` zone ports 80 and 443 defined in port group
resource "unifi_firewall_zone_policy" "policy2" {
    name     = "my-policy-2"
    action   = "BLOCK"
    protocol = "tcp_udp"

    source = {
        zone_id = unifi_firewall_zone.src.id
    }

    destination = {
        zone_id       = unifi_firewall_zone.dst.id
        port_group_id = unifi_firewall_group.web-ports.id
    }
}
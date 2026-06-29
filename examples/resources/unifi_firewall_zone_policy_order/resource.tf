resource "unifi_network" "network" {
    name    = "my-network"
    purpose = "corporate"
    subnet  = "10.0.10.0/24"
    vlan_id = "400"
}

resource "unifi_firewall_zone" "src" {
    name     = "my-source-zone"
    networks = [unifi_network.network.id]
}

resource "unifi_firewall_zone" "dst" {
    name = "my-destination-zone"
}

resource "unifi_firewall_zone_policy" "allow_web" {
    name     = "allow-web"
    action   = "ALLOW"
    protocol = "tcp_udp"

    source = {
        zone_id = unifi_firewall_zone.src.id
    }

    destination = {
        zone_id = unifi_firewall_zone.dst.id
    }
}

resource "unifi_firewall_zone_policy" "block_rest" {
    name     = "block-rest"
    action   = "BLOCK"
    protocol = "all"

    source = {
        zone_id = unifi_firewall_zone.src.id
    }

    destination = {
        zone_id = unifi_firewall_zone.dst.id
    }
}

# Control the evaluation order of the custom policies in the src -> dst zone pair.
# `allow_web` runs before the predefined policies; `block_rest` runs after them.
# Order within each list is significant.
resource "unifi_firewall_zone_policy_order" "order" {
    source_zone_id      = unifi_firewall_zone.src.id
    destination_zone_id = unifi_firewall_zone.dst.id

    before_predefined_ids = [
        unifi_firewall_zone_policy.allow_web.id,
    ]

    after_predefined_ids = [
        unifi_firewall_zone_policy.block_rest.id,
    ]
}

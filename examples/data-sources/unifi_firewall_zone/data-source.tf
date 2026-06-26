data "unifi_firewall_zone" "vpn" {
    name = "Vpn"
}

data "unifi_firewall_zone" "gateway" {
    name = "Gateway"
}

data "unifi_firewall_zone" "Internal" {
    name = "Internal"
}

data "unifi_firewall_zone" "External" {
    name = "External"
}

data "unifi_firewall_zone" "hotspot" {
    name = "Hotspot"
}

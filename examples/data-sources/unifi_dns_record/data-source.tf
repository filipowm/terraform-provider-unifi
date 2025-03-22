data "unifi_dns_record" "by_name" {
    name = "example.mydomain.com"
}

data "unifi_dns_record" "by_record" {
    record = "192.168.0.1"
}
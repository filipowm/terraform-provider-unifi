data "unifi_dns_record" "by_name" {
    filter {
        name = "example.mydomain.com"
    }
}

data "unifi_dns_record" "by_record" {
    filter {
        record = "192.168.0.1"
    }
}
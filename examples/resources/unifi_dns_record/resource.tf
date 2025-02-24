resource "unifi_dns_record" "a_record" {
    name   = "example.mydomain.com"
    type   = "A"
    record = "192.168.1.190"
}

resource "unifi_dns_record" "cname_record" {
    name   = "example.mydomain.com"
    type   = "CNAME"
    record = "example.com"
}

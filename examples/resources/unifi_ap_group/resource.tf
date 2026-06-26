resource "unifi_ap_group" "example" {
  name        = "my-ap-group"
  device_macs = ["00:11:22:33:44:55"]
}

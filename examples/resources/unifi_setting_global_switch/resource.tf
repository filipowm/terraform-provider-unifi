# Define the networks used by the layer-3 isolation rules. The isolation
# settings reference networks by their UniFi network ID (the `id` attribute),
# not by name or CIDR.
resource "unifi_network" "engineering" {
  name    = "Engineering"
  purpose = "corporate"
  subnet  = "10.0.10.1/24"
  vlan_id = 10
}

resource "unifi_network" "guest" {
  name    = "Guest"
  purpose = "corporate"
  subnet  = "10.0.20.1/24"
  vlan_id = 20
}

# Manage the site's switch isolation settings
# (Settings -> Network -> Switch Isolation Settings).
#
# Only the isolation-related fields are managed; all other global switch
# settings (DHCP snooping, STP, jumbo frames, etc.) are preserved untouched.
resource "unifi_setting_global_switch" "example" {
  # Layer-3 (network-to-network) isolation: isolate the guest network from the
  # engineering network. Values are UniFi network IDs.
  acl_l3_isolation = [
    {
      source_network       = unifi_network.guest.id
      destination_networks = [unifi_network.engineering.id]
    },
  ]

  # Switch MAC addresses excluded from isolation enforcement. MACs are
  # normalized to lowercase, colon-separated form.
  switch_exclusions = ["00:11:22:33:44:55"]

  # Specify the site (optional, defaults to the site configured in the provider,
  # otherwise "default").
  # site = "default"
}

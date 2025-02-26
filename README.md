![Latest](https://img.shields.io/github/v/release/filipowm/terraform-provider-unifi)
[![Docs](https://img.shields.io/badge/docs-reference-blue)](https://registry.terraform.io/providers/filipowm/unifi/latest)
![Acceptance Tests](https://github.com/filipowm/terraform-provider-unifi/workflows/Acceptance%20Tests/badge.svg?event=push)
![License](https://img.shields.io/github/license/filipowm/terraform-provider-unifi)

# UniFi Terraform Provider

A Terraform provider for managing Ubiquiti's UniFi network controller. This provider allows you to automate 
the configuration of your UniFi network infrastructure using Terraform.

**Note:** You can't configure your network while connected to something that may disconnect (like WiFi). 
Use a hard-wired connection to your controller to use this provider.

## Features

- Manage UniFi network resources using Infrastructure as Code
- Support for UniFi Controller version 6.x and later
- Compatible with UDM, UDM-Pro, UCG, and standard controller deployments
- Comprehensive resource management including:
    - Network/WLAN configuration
    - Firewall rules
    - Port forwarding
    - DNS records
    - User management
    - And more...

## Installation

The provider is available in the [Terraform Registry](https://registry.terraform.io/providers/filipowm/unifi/latest). To use it in your Terraform configuration:

```hcl
terraform {
    required_providers {
        unifi = {
            source  = "filipowm/unifi"
            version = "~> 0.0.1"  # Use the latest version
        }
    }
}
```

## Authentication

The provider supports two authentication methods:

1. Username/Password authentication
2. API Key authentication (requires controller version 9.0.108 or later)

```hcl
# Using environment variables (recommended)
# Export these variables:
# export UNIFI_API_KEY="my-api-key"
# export UNIFI_USERNAME="admin"
# export UNIFI_PASSWORD="password"
# export UNIFI_API="https://unifi.example.com:8443"
# export UNIFI_INSECURE=true  # Only if using self-signed certificates

# Or configure directly in the provider block
provider "unifi" {
    api_key = "my-api-key"
    # username = "admin" # Use either username/password or API key
    # password = "password"
    api_url = "https://unifi.example.com:8443"

    # Optional settings
    allow_insecure = true  # For self-signed certificates
    site = "default"       # Specify non-default site
}
```

## Example Usage

Here's a basic example of creating a wireless network:

```hcl
resource "unifi_wlan" "wifi" {
    name       = "My WiFi Network"
    security   = "wpa2"
    passphrase = "mystrongpassword"
    network_id = unifi_network.vlan_50.id
}

resource "unifi_network" "vlan_50" {
    name    = "VLAN 50"
    purpose = "corporate"
    subnet  = "10.0.50.0/24"
    vlan_id = 50
}
```

More examples can be found in the [documentation](https://registry.terraform.io/providers/filipowm/unifi/latest/docs).

## Documentation

Comprehensive documentation is available on the [Terraform Registry](https://registry.terraform.io/providers/filipowm/unifi/latest/docs)

## Supported Platforms

* UniFi Controller version 6.x and later
* UniFi Dream Machine (UDM)
* UniFi Dream Machine Pro (UDM-Pro)
* UniFi Cloud Gateway (UCG)
* Standard UniFi Controller deployments

## Plans

- [ ] Add support for zone-based firewalls
- [x] Add support for DNS records
- [ ] Add support for traffic management
- [ ] Add support for more setting resources (eg. guest access, connectivity, country, IPS/IDS, etc.)
- [x] Support API key authentication
- [x] Improve documentation
- [ ] Switch to Terraform Plugin Framework from Terraform Plugin SDK v2
- [x] Fix flaky tests
- [ ] Implement validation for fields and structures
- [ ] Increase test coverage and make tests more reliable
- [x] Update all dependencies to the latest versions
- [x] Switch to [filipowm/go-unifi](https://github.com/filipowm/go-unifi) SDK
- [x] Support newest controller versions (9.x)

## Contributing

Contributions are welcome! Please follow [contributing guide](./.github/CONTRIBUTING.md).

The provider is built on top of the [go-unifi](https://github.com/filipowm/go-unifi) SDK.

## License

This provider is licensed under the [LICENSE](./LICENSE) file.

## Acknowledgements

This project is a fork of [paultyng/terraform-provider-unifi](https://github.com/paultyng/terraform-provider-unifi). We extend our heartfelt gratitude to Paul Tyng and all the contributors of the original provider for their outstanding work. Their efforts have laid a solid foundation for this fork.

Our goal with this fork is to build upon their excellent work by:

1. Keeping the provider up-to-date with the latest UniFi Controller versions
2. Expanding support for new resources
3. Enhancing and improving documentation
4. Migrating to the Terraform Plugin Framework

We are committed to maintaining a stable, current, and reliable Terraform Provider for UniFi Networks & Devices, ensuring that users have the best possible tools for managing their infrastructure.

We thank the original authors for their invaluable contribution to the UniFi and Terraform communities, and we look forward to continuing this important work.


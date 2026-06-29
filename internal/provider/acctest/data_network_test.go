package acctest

import (
	"fmt"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccDataNetwork_byName(t *testing.T) {
	defaultName := "Default"
	v, err := version.NewVersion(testClient.Version())
	if err != nil {
		t.Fatalf("error parsing version: %s", err)
	}
	if v.LessThan(base.ControllerV7) {
		defaultName = "LAN"
	}
	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccDataNetworkConfig_byName(defaultName),
				Check:  resource.ComposeTestCheckFunc(
				// testCheckNetworkExists(t, "name"),
				),
			},
		},
	})
}

func TestAccDataNetwork_byID(t *testing.T) {
	defaultName := "Default"
	v, err := version.NewVersion(testClient.Version())
	if err != nil {
		t.Fatalf("error parsing version: %s", err)
	}
	if v.LessThan(base.ControllerV7) {
		defaultName = "LAN"
	}

	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccDataNetworkConfig_byID(defaultName),
				Check:  resource.ComposeTestCheckFunc(
				// testCheckNetworkExists(t, "name"),
				),
			},
		},
	})
}

// TestAccDataNetwork_defaultGateway creates a network with the DHCP default-gateway
// override and reads it back through the data source to prove the new computed
// fields surface non-empty (the Default/LAN network has no override).
func TestAccDataNetwork_defaultGateway(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	subnet, vlan := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccDataNetworkConfig_defaultGateway(name, subnet.String(), vlan),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_network.test", "dhcpd_gateway_enabled", "true"),
					resource.TestCheckResourceAttrSet("data.unifi_network.test", "dhcpd_gateway"),
				),
			},
		},
	})
}

func testAccDataNetworkConfig_defaultGateway(name, subnet string, vlan int) string {
	return fmt.Sprintf(`
locals {
	subnet  = "%[2]s"
	vlan_id = %[3]d
}

resource "unifi_network" "test" {
	name    = "%[1]s"
	purpose = "corporate"

	subnet       = local.subnet
	vlan_id      = local.vlan_id
	dhcp_start   = cidrhost(local.subnet, 6)
	dhcp_stop    = cidrhost(local.subnet, 254)
	dhcp_enabled = true

	dhcpd_gateway_enabled = true
	dhcpd_gateway         = cidrhost(local.subnet, 5)
}

data "unifi_network" "test" {
	name = unifi_network.test.name
}
`, name, subnet, vlan)
}

func testAccDataNetworkConfig_byName(name string) string {
	return fmt.Sprintf(`
data "unifi_network" "lan" {
	name = %q
}
`, name)
}

func testAccDataNetworkConfig_byID(name string) string {
	return fmt.Sprintf(`
data "unifi_network" "lan" {
	name = %q
}

data "unifi_network" "lan_id" {
	id = data.unifi_network.lan.id
}
`, name)
}

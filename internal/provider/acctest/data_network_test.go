package acctest

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
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
				Config: testAccDataNetworkConfigByName(defaultName),
				Check: resource.ComposeTestCheckFunc(
					// testCheckNetworkExists(t, "name"),
					// dhcp_guarding is Computed on the data source; assert it is set to a
					// known bool so resource/data-source coverage stays in lockstep (#123).
					resource.TestCheckResourceAttrSet("data.unifi_network.lan", "dhcp_guarding"),
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
				Config: testAccDataNetworkConfigByID(defaultName),
				Check:  resource.ComposeTestCheckFunc(
				// testCheckNetworkExists(t, "name"),
				),
			},
		},
	})
}

// TestAccDataNetwork_firewallZoneID verifies the data source surfaces the computed
// firewall_zone_id with read parity to the resource. Gated on 9.x ZBF controllers and
// skipped in the Dockerized harness, mirroring the resource zone tests.
func TestAccDataNetwork_firewallZoneID(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc")
	zoneName := acctest.RandomWithPrefix("tfacc-zone")
	subnet, vlan := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZoneLock,
		Steps: []resource.TestStep{
			{
				Config: testAccDataNetworkConfigFirewallZoneID(name, subnet.String(), vlan, zoneName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.unifi_network.test", "firewall_zone_id", "unifi_network.test", "firewall_zone_id"),
					resource.TestCheckResourceAttrPair("data.unifi_network.test", "firewall_zone_id", "unifi_firewall_zone.test", "id"),
				),
			},
		},
		CheckDestroy: testAccCheckFirewallZoneDestroy,
	})
}

func testAccDataNetworkConfigFirewallZoneID(name, subnet string, vlan int, zoneName string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone" "test" {
	name = %[4]q
	# networks intentionally omitted — managed from the network side below.
}

resource "unifi_network" "test" {
	name             = %[1]q
	purpose          = "corporate"
	subnet           = %[2]q
	vlan_id          = %[3]d
	firewall_zone_id = unifi_firewall_zone.test.id
}

data "unifi_network" "test" {
	id = unifi_network.test.id
}
`, name, subnet, vlan, zoneName)
}

// TestAccDataNetwork_defaultGateway proves the data source surfaces the computed
// default-gateway override fields (issue #120). It creates a dedicated network with the
// override enabled rather than reusing the pre-existing Default network, which has no
// override, so the computed values would otherwise be empty/false.
func TestAccDataNetwork_defaultGateway(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	subnet, vlan := pt.GetTestVLAN(t)
	gw := mustHost(t, subnet, 100)

	AcceptanceTest(t, AcceptanceTestCase{
		CheckDestroy: testAccCheckNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataNetworkConfigDefaultGateway(name, subnet.String(), vlan, gw),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_network.test", "dhcpd_gateway_enabled", "true"),
					resource.TestCheckResourceAttr("data.unifi_network.test", "dhcpd_gateway", gw),
				),
			},
		},
	})
}

func testAccDataNetworkConfigDefaultGateway(name, subnet string, vlan int, gateway string) string {
	return fmt.Sprintf(`
resource "unifi_network" "test" {
	name    = %[1]q
	purpose = "corporate"
	subnet  = %[2]q
	vlan_id = %[3]d

	dhcp_enabled          = true
	dhcp_start            = cidrhost(%[2]q, 6)
	dhcp_stop             = cidrhost(%[2]q, 254)
	dhcpd_gateway_enabled = true
	dhcpd_gateway         = %[4]q
}

data "unifi_network" "test" {
	id = unifi_network.test.id
}
`, name, subnet, vlan, gateway)
}

func testAccDataNetworkConfigByName(name string) string {
	return fmt.Sprintf(`
data "unifi_network" "lan" {
	name = %q
}
`, name)
}

func testAccDataNetworkConfigByID(name string) string {
	return fmt.Sprintf(`
data "unifi_network" "lan" {
	name = %q
}

data "unifi_network" "lan_id" {
	id = data.unifi_network.lan.id
}
`, name)
}

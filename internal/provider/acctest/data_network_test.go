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
				Config: testAccDataNetworkConfig_firewallZoneID(name, subnet.String(), vlan, zoneName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.unifi_network.test", "firewall_zone_id", "unifi_network.test", "firewall_zone_id"),
					resource.TestCheckResourceAttrPair("data.unifi_network.test", "firewall_zone_id", "unifi_firewall_zone.test", "id"),
				),
			},
		},
		CheckDestroy: testAccCheckFirewallZoneDestroy,
	})
}

func testAccDataNetworkConfig_firewallZoneID(name, subnet string, vlan int, zoneName string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone" "test" {
	name     = %[4]q
	networks = []
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

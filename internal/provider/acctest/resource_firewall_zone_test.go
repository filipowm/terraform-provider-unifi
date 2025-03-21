package acctest

import (
	"context"
	"fmt"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"strings"
	"sync"
	"testing"
)

const testFirewallZoneResourceName = "unifi_firewall_zone.test"

var firewallZoneLock = &sync.Mutex{}

// TODO make tests runnable on test environment hosted on container
// to enable those tests, set TF_ACC_LOCAL=1

func TestAccFirewallZone_withNetworks(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet, and no idea how to enable it")
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZoneLock,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallZoneConfig(t, "test_zone_networks", acctest.RandomWithPrefix("tfacc-")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZoneResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZoneResourceName, "site", "default"),
					resource.TestCheckResourceAttr(testFirewallZoneResourceName, "name", "test_zone_networks"),
					resource.TestCheckResourceAttr(testFirewallZoneResourceName, "networks.#", "1"),
				),
				ConfigPlanChecks: pt.CheckResourceActions(testFirewallZoneResourceName, plancheck.ResourceActionCreate),
			},
			pt.ImportStepWithSite(testFirewallZoneResourceName),
		},
		CheckDestroy: testAccCheckFirewallZoneDestroy,
	})
}

func TestAccFirewallZone_update(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet, and no idea how to enable it")
	network1 := acctest.RandomWithPrefix("tfacc-")
	network2 := acctest.RandomWithPrefix("tfacc-")
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZoneLock,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallZoneConfig(t, "initial_zone", network1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZoneResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZoneResourceName, "name", "initial_zone"),
					resource.TestCheckResourceAttr(testFirewallZoneResourceName, "networks.#", "1"),
					resource.TestCheckResourceAttrSet(testFirewallZoneResourceName, "networks.0"),
				),
			},
			{
				Config: testAccFirewallZoneConfig(t, "updated_zone", network2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZoneResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZoneResourceName, "name", "updated_zone"),
					resource.TestCheckResourceAttr(testFirewallZoneResourceName, "networks.#", "1"),
					resource.TestCheckResourceAttrSet(testFirewallZoneResourceName, "networks.0"),
				),
				ConfigPlanChecks: pt.CheckResourceActions(testFirewallZoneResourceName, plancheck.ResourceActionUpdate),
			},
		},
		CheckDestroy: testAccCheckFirewallZoneDestroy,
	})
}

func TestAccFirewallZone_missingName(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet, and no idea how to enable it")
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZoneLock,
		Steps: []resource.TestStep{
			{
				Config:      testAccFirewallZoneConfigMissingName(),
				ExpectError: pt.MissingArgumentErrorRegex("name"),
			},
		},
	})
}

func testAccFirewallZoneConfig(t *testing.T, name string, network string) string {
	subnet, vlanId := pt.GetTestVLAN(t)
	return fmt.Sprintf(`

resource "unifi_network" "test" {
	name    = %[2]q
	purpose = "corporate"
	subnet  = %[3]q
	vlan_id = "%[4]d"
}

resource "unifi_firewall_zone" "test" {
	name     = %[1]q
	networks = [unifi_network.test.id]
}
`, name, network, subnet.String(), vlanId)
}

func testAccFirewallZoneConfigMissingName() string {
	return `
resource "unifi_firewall_zone" "test" {
	networks = []
}
`
}

func testAccCheckFirewallZoneDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "unifi_firewall_zone" {
			continue
		}

		_, err := testClient.GetFirewallZone(context.Background(), "default", rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Firewall Zone %s still exists", rs.Primary.ID)
		}

		// If we get a 404 error, that means the resource was deleted
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			continue
		}

		// For any other error, return it
		return err
	}

	return nil
}

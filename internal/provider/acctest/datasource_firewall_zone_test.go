package acctest

import (
	"fmt"
	"regexp"
	"testing"

	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testFirewallZoneDataSourceName = "data.unifi_firewall_zone.test"

func TestFirewallZoneDataSource_basic(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet, and no idea how to enable it")
	name := acctest.RandomWithPrefix("tfacc-")

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZoneLock,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallZoneDataSourceConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testFirewallZoneDataSourceName, "name", name),
					resource.TestCheckResourceAttr(testFirewallZoneResourceName, "networks.#", "0"),
				),
			},
		},
	})
}

func TestFirewallZoneDataSource_nonExistent(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet, and no idea how to enable it")
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Steps: []resource.TestStep{
			{
				Config:      testAccFirewallZoneDataSourceConfig_nonExistent(),
				ExpectError: regexp.MustCompile(`No firewall zone with name`),
			},
		},
	})
}

func TestFirewallZoneDataSource_missingName(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Steps: []resource.TestStep{
			{
				Config:      testAccFirewallZoneDataSourceConfigMissingName(),
				ExpectError: pt.MissingArgumentErrorRegex("name"),
			},
		},
	})
}

func testAccFirewallZoneDataSourceConfig(name string) string {
	return fmt.Sprintf(`

resource "unifi_firewall_zone" "test" {
	name = %[1]q
}

data "unifi_firewall_zone" "test" {
	name = %[1]q
	depends_on = [unifi_firewall_zone.test]
}`, name)
}

func testAccFirewallZoneDataSourceConfig_nonExistent() string {
	return `

data "unifi_firewall_zone" "test" {
	name = "non-existent"
}`
}

func testAccFirewallZoneDataSourceConfigMissingName() string {
	return `
data "unifi_firewall_zone" "test" {
}`
}

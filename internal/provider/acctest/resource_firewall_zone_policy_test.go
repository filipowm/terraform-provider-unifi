package acctest

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"testing"

	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var firewallZonePolicyLock = &sync.Mutex{}

const testFirewallZonePolicyResourceName = "unifi_firewall_zone_policy.test"

// TestAccFirewallZonePolicy_basic tests the basic configuration of a firewall zone policy
func TestAccFirewallZonePolicy_basic(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc-zone-policy")
	subnet, vlanId := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZonePolicyLock,
		Steps: []resource.TestStep{
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicyBasicConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "site", "default"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "name", name),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "action", "BLOCK"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "protocol", "all"),
				),
				ConfigPlanChecks: pt.CheckResourceActions(testFirewallZonePolicyResourceName, plancheck.ResourceActionCreate),
			},
			pt.ImportStepWithSite(testFirewallZonePolicyResourceName),
		},
		CheckDestroy: testAccCheckFirewallZonePolicyDestroy,
	})
}

// TestAccFirewallZonePolicy_update tests updating a firewall zone policy
func TestAccFirewallZonePolicy_update(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc-zone-policy")
	pt.GetTestVLAN(t)
	subnet, vlanId := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZonePolicyLock,
		Steps: []resource.TestStep{
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicyBasicConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "name", name),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "action", "BLOCK"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "enabled", "true"),
				),
			},
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicyUpdatedConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "name", name),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "action", "ALLOW"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "description", "Updated zone policy"),
				),
				ConfigPlanChecks: pt.CheckResourceActions(testFirewallZonePolicyResourceName, plancheck.ResourceActionUpdate),
			},
		},
		CheckDestroy: testAccCheckFirewallZonePolicyDestroy,
	})
}

// TestAccFirewallZonePolicy_matchOppositeProtocol tests match opposite protocol setting
func TestAccFirewallZonePolicy_matchOppositeProtocol(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc-zone-policy")
	subnet, vlanId := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZonePolicyLock,
		Steps: []resource.TestStep{
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicyMatchOppositeProtocolConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "match_opposite_protocol", "true"),
				),
			},
			pt.ImportStepWithSite(testFirewallZonePolicyResourceName),
		},
	})
}

// TestAccFirewallZonePolicy_scheduledPolicy tests all schedule modes
func TestAccFirewallZonePolicy_scheduledPolicy(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc-zone-policy")
	pt.GetTestVLAN(t)
	subnet, vlanId := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZonePolicyLock,
		Steps: []resource.TestStep{
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicyScheduleAlwaysConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "schedule.mode", "ALWAYS"),
				),
			},
			pt.ImportStepWithSite(testFirewallZonePolicyResourceName),
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicyScheduleEveryDayConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "schedule.mode", "EVERY_DAY"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "schedule.time_all_day", "false"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "schedule.time_from", "08:00"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "schedule.time_to", "17:00"),
				),
			},
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicyScheduleEveryWeekConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "schedule.mode", "EVERY_WEEK"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "schedule.repeat_on_days.#", "3"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "schedule.time_all_day", "true"),
				),
			},
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicyScheduleOneTimeOnlyConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "schedule.mode", "ONE_TIME_ONLY"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "schedule.date", "2025-12-31"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "schedule.time_all_day", "false"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "schedule.time_from", "09:00"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "schedule.time_to", "18:00"),
				),
			},
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicyScheduleCustomConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "schedule.mode", "CUSTOM"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "schedule.date_start", "2025-06-01"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "schedule.date_end", "2025-12-31"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "schedule.repeat_on_days.#", "2"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "schedule.time_all_day", "false"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "schedule.time_from", "10:00"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "schedule.time_to", "16:00"),
				),
			},
		},
	})
}

// TestAccFirewallZonePolicy_invalidConfig tests validation failures
func TestAccFirewallZonePolicy_invalidConfig(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc-zone-policy")
	subnet, vlanId := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZonePolicyLock,
		Steps: []resource.TestStep{
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicyInvalidProtocolConfig(name),
				),
				ExpectError: regexp.MustCompile(`Attribute protocol value must be one of`),
			},
		},
	})
}

// TestAccFirewallZonePolicy_sourceDestinationConfig tests source and destination configuration with basic IP settings
func TestAccFirewallZonePolicy_sourceDestinationConfig(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc-zone-policy")
	subnet, vlanId := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZonePolicyLock,
		Steps: []resource.TestStep{
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicySourceDestinationConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "source.ips.#", "2"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "source.port", "80"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "destination.ips.#", "1"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "destination.port", "443"),
				),
			},
			pt.ImportStepWithSite(testFirewallZonePolicyResourceName),
		},
	})
}

// TestAccFirewallZonePolicy_sourceIPGroup tests source configuration with IP groups
func TestAccFirewallZonePolicy_sourceIPGroup(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc-zone-policy")
	subnet, vlanId := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZonePolicyLock,
		Steps: []resource.TestStep{
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicySourceIPGroupConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "source.ip_group_id"),
				),
			},
			pt.ImportStepWithSite(testFirewallZonePolicyResourceName),
		},
	})
}

// TestAccFirewallZonePolicy_webDomainsPolicy tests policy with web domains configuration
func TestAccFirewallZonePolicy_sourceIpsPolicy(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc-zone-policy")
	subnet, vlanId := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZonePolicyLock,
		Steps: []resource.TestStep{
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicyIpsConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "destination.ips.#", "1"),
				),
			},
			pt.ImportStepWithSite(testFirewallZonePolicyResourceName),
		},
	})
}

// TestAccFirewallZonePolicy_sourcePortGroup tests source configuration with port groups
func TestAccFirewallZonePolicy_sourcePortGroup(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc-zone-policy")
	subnet, vlanId := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZonePolicyLock,
		Steps: []resource.TestStep{
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicySourcePortGroupConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "source.port_group_id"),
				),
			},
			pt.ImportStepWithSite(testFirewallZonePolicyResourceName),
		},
	})
}

// TestAccFirewallZonePolicy_sourceMACs tests source configuration with MAC addresses
func TestAccFirewallZonePolicy_sourceMACs(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc-zone-policy")
	subnet, vlanId := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZonePolicyLock,
		Steps: []resource.TestStep{
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicySourceMACsConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "source.macs.#", "2"),
				),
			},
			pt.ImportStepWithSite(testFirewallZonePolicyResourceName),
		},
	})
}

// TestAccFirewallZonePolicy_sourceClientMACs tests source configuration with client MAC addresses
func TestAccFirewallZonePolicy_sourceClientMACs(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc-zone-policy")
	subnet, vlanId := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZonePolicyLock,
		Steps: []resource.TestStep{
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicySourceClientMACsConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "source.client_macs.#", "2"),
				),
			},
			pt.ImportStepWithSite(testFirewallZonePolicyResourceName),
		},
	})
}

// TestAccFirewallZonePolicy_sourceNetworkIDs tests source configuration with network IDs
func TestAccFirewallZonePolicy_sourceNetworkIDs(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc-zone-policy")
	subnet, vlanId := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZonePolicyLock,
		Steps: []resource.TestStep{
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicySourceNetworkIDsConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "source.network_ids.#", "1"),
				),
			},
			pt.ImportStepWithSite(testFirewallZonePolicyResourceName),
		},
	})
}

// TestAccFirewallZonePolicy_sourceSingleMAC tests source configuration with a single MAC address
func TestAccFirewallZonePolicy_sourceSingleMAC(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc-zone-policy")
	subnet, vlanId := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZonePolicyLock,
		Steps: []resource.TestStep{
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicySourceSingleMACConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "source.mac", "00:11:22:33:44:55"),
				),
			},
			pt.ImportStepWithSite(testFirewallZonePolicyResourceName),
		},
	})
}

// TestAccFirewallZonePolicy_sourceMatchOpposite tests source configuration with match opposite settings
func TestAccFirewallZonePolicy_sourceMatchOpposite(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc-zone-policy")
	subnet, vlanId := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZonePolicyLock,
		Steps: []resource.TestStep{
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicySourceMatchOppositeConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "source.match_opposite_ips", "true"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "source.match_opposite_ports", "true"),
				),
			},
			pt.ImportStepWithSite(testFirewallZonePolicyResourceName),
		},
	})
}

// TestAccFirewallZonePolicy_destinationIPGroup tests destination configuration with IP groups
func TestAccFirewallZonePolicy_destinationIPGroup(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc-zone-policy")
	subnet, vlanId := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZonePolicyLock,
		Steps: []resource.TestStep{
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicyDestinationIPGroupConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "destination.ip_group_id"),
				),
			},
			pt.ImportStepWithSite(testFirewallZonePolicyResourceName),
		},
	})
}

// TestAccFirewallZonePolicy_destinationPortGroup tests destination configuration with port groups
func TestAccFirewallZonePolicy_destinationPortGroup(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc-zone-policy")
	subnet, vlanId := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZonePolicyLock,
		Steps: []resource.TestStep{
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicyDestinationPortGroupConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "destination.port_group_id"),
				),
			},
			pt.ImportStepWithSite(testFirewallZonePolicyResourceName),
		},
	})
}

// TestAccFirewallZonePolicy_destinationRegions tests destination configuration with regions
func TestAccFirewallZonePolicy_destinationRegions(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc-zone-policy")
	subnet, vlanId := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZonePolicyLock,
		Steps: []resource.TestStep{
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicyDestinationRegionsConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "destination.regions.#", "2"),
				),
			},
			pt.ImportStepWithSite(testFirewallZonePolicyResourceName),
		},
	})
}

// TestAccFirewallZonePolicy_destinationMatchOpposite tests destination configuration with match opposite settings
func TestAccFirewallZonePolicy_destinationMatchOpposite(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc-zone-policy")
	subnet, vlanId := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZonePolicyLock,
		Steps: []resource.TestStep{
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicyDestinationMatchOppositeConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "destination.match_opposite_ips", "true"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "destination.match_opposite_ports", "true"),
				),
			},
			pt.ImportStepWithSite(testFirewallZonePolicyResourceName),
		},
	})
}

// TestAccFirewallZonePolicy_destinationAppIDs tests destination configuration with app IDs
func TestAccFirewallZonePolicy_destinationAppIDs(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc-zone-policy")
	subnet, vlanId := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZonePolicyLock,
		Steps: []resource.TestStep{
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicyDestinationAppIDsConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "destination.app_ids.#", "2"),
				),
			},
			pt.ImportStepWithSite(testFirewallZonePolicyResourceName),
		},
	})
}

// TestAccFirewallZonePolicy_destinationAppCategoryIDs tests destination configuration with app category IDs
func TestAccFirewallZonePolicy_destinationAppCategoryIDs(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc-zone-policy")
	subnet, vlanId := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZonePolicyLock,
		Steps: []resource.TestStep{
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicyDestinationAppCategoryIDsConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "destination.app_category_ids.#", "2"),
				),
			},
			pt.ImportStepWithSite(testFirewallZonePolicyResourceName),
		},
	})
}

// TestAccFirewallZonePolicy_ipSecPolicy tests policy with IPSec configuration
func TestAccFirewallZonePolicy_ipSecPolicy(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc-zone-policy")
	subnet, vlanId := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZonePolicyLock,
		Steps: []resource.TestStep{
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicyIPSecConfig(name),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyResourceName, "match_ip_sec_type", "MATCH_IP_SEC"),
				),
			},
			pt.ImportStepWithSite(testFirewallZonePolicyResourceName),
		},
	})
}

func testAccCheckFirewallZonePolicyDestroy(s *terraform.State) error {
	return pt.CheckDestroy("unifi_firewall_zone_policy", func(ctx context.Context, site, id string) error {
		_, err := testClient.GetFirewallZonePolicy(ctx, site, id)
		return err
	})(s)
}

func testAccFirewallZonePolicyPreConfig(name, subnet string, vlanId int) string {
	return fmt.Sprintf(`
resource "unifi_network" "test" {
	name    = %[1]q
	purpose = "corporate"
	subnet  = %[2]q
	vlan_id = "%[3]d"
}

resource "unifi_firewall_zone" "test" {
	name     = %[1]q
	networks = [unifi_network.test.id]
}
`, name, subnet, vlanId)
}

// Test configurations
func testAccFirewallZonePolicyBasicConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy" "test" {
	name   = %[1]q
	action = "BLOCK"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
	}
}
`, name)
}

func testAccFirewallZonePolicyUpdatedConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy" "test" {
	name        = %[1]q
	action      = "ALLOW"
	auto_allow_return_traffic = true
	enabled     = false
	description = "Updated zone policy"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
		macs = ["00:11:22:33:44:55"]
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
		ips = ["192.168.1.2", "192.168.1.3"]
	}
}
`, name)
}

func testAccFirewallZonePolicyInvalidProtocolConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy" "test" {
	name     = %[1]q
	action   = "BLOCK"
	protocol = "invalid_protocol"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
	}
}
`, name)
}

func testAccFirewallZonePolicySourceDestinationConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy" "test" {
	name     = %[1]q
	action   = "BLOCK"
	protocol = "tcp"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
		ips = ["192.168.1.10", "192.168.1.11"]
		port = 80
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
		ips = ["192.168.2.10"]
		port = 443
	}
}
`, name)
}

func testAccFirewallZonePolicyScheduledConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy" "test" {
	name     = %[1]q
	action   = "BLOCK"
	protocol = "all"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
	}
	
	schedule = {
		mode = "EVERY_DAY"
		time_all_day = false
		time_from = "08:00"
		time_to = "17:00"
	}
}
`, name)
}

func testAccFirewallZonePolicyIPSecConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy" "test" {
	name             = %[1]q
	action           = "BLOCK"
	protocol         = "all"
	match_ip_sec_type = "MATCH_IP_SEC"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
	}
}
`, name)
}

func testAccFirewallZonePolicyIpsConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy" "test" {
	name     = %[1]q
	action   = "BLOCK"
	protocol = "all"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
		ips = ["192.168.1.2"]
	}
}
`, name)
}

func testAccFirewallZonePolicySourceIPGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_group" "test" {
	name    = %[1]q
	type    = "address-group"
	members = ["192.168.1.1", "10.0.0.0"]
}

resource "unifi_firewall_zone_policy" "test" {
	name     = %[1]q
	action   = "ALLOW"
	protocol = "tcp_udp"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
		ip_group_id = unifi_firewall_group.test.id
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
	}
}
`, name)
}

func testAccFirewallZonePolicySourcePortGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_group" "test" {
	name    = %[1]q
	type    = "port-group"
	members = ["80", "443"]
}

resource "unifi_firewall_zone_policy" "test" {
	name     = %[1]q
	action   = "ALLOW"
	protocol = "tcp_udp"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
		port_group_id = unifi_firewall_group.test.id
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
	}
}
`, name)
}

func testAccFirewallZonePolicySourceMACsConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy" "test" {
	name     = %[1]q
	action   = "ALLOW"
	protocol = "tcp_udp"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
		macs = ["00:11:22:33:44:55", "66:77:88:99:AA:BB"]
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
	}
}
`, name)
}

func testAccFirewallZonePolicySourceSingleMACConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy" "test" {
	name     = %[1]q
	action   = "ALLOW"
	protocol = "tcp_udp"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
		mac = "00:11:22:33:44:55"
		ips = ["192.168.1.1"] 
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
	}
}
`, name)
}

func testAccFirewallZonePolicySourceClientMACsConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy" "test" {
	name     = %[1]q
	action   = "ALLOW"
	protocol = "tcp_udp"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
		client_macs = ["00:11:22:33:44:55", "66:77:88:99:AA:BB"]
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
	}
}
`, name)
}

func testAccFirewallZonePolicySourceNetworkIDsConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy" "test" {
	name     = %[1]q
	action   = "ALLOW"
	protocol = "tcp_udp"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
		network_ids = [unifi_network.test.id]
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
	}
}
`, name)
}

func testAccFirewallZonePolicySourceMatchOppositeConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy" "test" {
	name     = %[1]q
	action   = "ALLOW"
	protocol = "tcp_udp"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
		ips = ["192.168.1.1"]
		port = "80"
		match_opposite_ips = true
		match_opposite_ports = true
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
	}
}
`, name)
}

func testAccFirewallZonePolicyDestinationIPGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_group" "test" {
	name    = %[1]q
	type    = "address-group"
	members = ["192.168.1.1", "10.0.0.0"]
}

resource "unifi_firewall_zone_policy" "test" {
	name     = %[1]q
	action   = "ALLOW"
	protocol = "tcp_udp"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
		ip_group_id = unifi_firewall_group.test.id
	}
}
`, name)
}

func testAccFirewallZonePolicyDestinationPortGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_group" "test" {
	name    = %[1]q
	type    = "port-group"
	members = ["80", "443"]
}

resource "unifi_firewall_zone_policy" "test" {
	name     = %[1]q
	action   = "ALLOW"
	protocol = "tcp_udp"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
		port_group_id = unifi_firewall_group.test.id
	}
}
`, name)
}

func testAccFirewallZonePolicyDestinationAppIDsConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy" "test" {
	name     = %[1]q
	action   = "ALLOW"
	protocol = "tcp_udp"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
		app_ids = ["1", "2"]
	}
}
`, name)
}

func testAccFirewallZonePolicyDestinationAppCategoryIDsConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy" "test" {
	name     = %[1]q
	action   = "ALLOW"
	enabled  = true
	protocol = "tcp_udp"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
		app_category_ids = ["1", "2"]
	}
}
`, name)
}

func testAccFirewallZonePolicyDestinationRegionsConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy" "test" {
	name     = %[1]q
	action   = "ALLOW"
	protocol = "tcp_udp"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
		regions = ["US", "CA"]
	}
}
`, name)
}

func testAccFirewallZonePolicyDestinationMatchOppositeConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy" "test" {
	name     = %[1]q
	action   = "ALLOW"
	protocol = "tcp_udp"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
		ips = ["192.168.1.1"]
		port = "443"
		match_opposite_ips = true
		match_opposite_ports = true
	}
}
`, name)
}

func testAccFirewallZonePolicyMatchOppositeProtocolConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy" "test" {
	name     = %[1]q
	action   = "ALLOW"
	protocol = "tcp"
	match_opposite_protocol = true
	
	source = {
		zone_id = unifi_firewall_zone.test.id
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
	}
}
`, name)
}

func testAccFirewallZonePolicyScheduleAlwaysConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy" "test" {
	name     = %[1]q
	action   = "BLOCK"
	protocol = "all"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
	}
	
	schedule = {
		mode = "ALWAYS"
	}
}
`, name)
}

func testAccFirewallZonePolicyScheduleEveryDayConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy" "test" {
	name     = %[1]q
	action   = "BLOCK"
	protocol = "all"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
	}
	
	schedule = {
		mode = "EVERY_DAY"
		time_all_day = false
		time_from = "08:00"
		time_to = "17:00"
	}
}
`, name)
}

func testAccFirewallZonePolicyScheduleEveryWeekConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy" "test" {
	name     = %[1]q
	action   = "BLOCK"
	protocol = "all"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
	}
	
	schedule = {
		mode = "EVERY_WEEK"
		time_all_day = true
		repeat_on_days = ["mon", "wed", "fri"]
	}
}
`, name)
}

func testAccFirewallZonePolicyScheduleOneTimeOnlyConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy" "test" {
	name     = %[1]q
	action   = "BLOCK"
	protocol = "all"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
	}
	
	schedule = {
		mode = "ONE_TIME_ONLY"
		date = "2025-12-31"
		time_all_day = false
		time_from = "09:00"
		time_to = "18:00"
	}
}
`, name)
}

func testAccFirewallZonePolicyScheduleCustomConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy" "test" {
	name     = %[1]q
	action   = "BLOCK"
	protocol = "all"
	
	source = {
		zone_id = unifi_firewall_zone.test.id
	}
	
	destination = {
		zone_id = unifi_firewall_zone.test.id
	}
	
	schedule = {
		mode = "CUSTOM"
		date_start = "2025-06-01"
		date_end = "2025-12-31"
		repeat_on_days = ["tue", "thu"]
		time_all_day = false
		time_from = "10:00"
		time_to = "16:00"
	}
}
`, name)
}

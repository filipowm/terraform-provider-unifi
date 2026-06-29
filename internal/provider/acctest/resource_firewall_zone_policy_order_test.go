package acctest

import (
	"fmt"

	"testing"

	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

const testFirewallZonePolicyOrderResourceName = "unifi_firewall_zone_policy_order.order"

// TestAccFirewallZonePolicyOrder_basic creates three policies in one zone pair
// and a unifi_firewall_zone_policy_order that orders all three after the
// predefined policies, then verifies the ordering round-trips through import.
func TestAccFirewallZonePolicyOrder_basic(t *testing.T) {
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
					testAccFirewallZonePolicyMultiConfig(name, false),
					testAccFirewallZonePolicyOrderConfig(orderedAfterIDs123),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyOrderResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyOrderResourceName, "site", "default"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyOrderResourceName, "after_predefined_ids.#", "3"),
				),
			},
			// id is `<source>:<dest>`, so SiteAndIDImportStateIDFunc yields the
			// expected `<site>:<source>:<dest>` import id.
			pt.ImportStepWithSite(testFirewallZonePolicyOrderResourceName),
		},
		CheckDestroy: testAccCheckFirewallZonePolicyDestroy,
	})
}

// TestAccFirewallZonePolicyOrder_update permutes the ordering and asserts the
// resource is updated (not replaced).
func TestAccFirewallZonePolicyOrder_update(t *testing.T) {
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
					testAccFirewallZonePolicyMultiConfig(name, false),
					testAccFirewallZonePolicyOrderConfig(orderedAfterIDs123),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testFirewallZonePolicyOrderResourceName, "after_predefined_ids.#", "3"),
				),
			},
			{
				Config: pt.ComposeConfig(
					testAccFirewallZonePolicyPreConfig(name, subnet.String(), vlanId),
					testAccFirewallZonePolicyMultiConfig(name, false),
					testAccFirewallZonePolicyOrderConfig(orderedAfterIDs312),
				),
				ConfigPlanChecks: pt.CheckResourceActions(testFirewallZonePolicyOrderResourceName, plancheck.ResourceActionUpdate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testFirewallZonePolicyOrderResourceName, "after_predefined_ids.#", "3"),
				),
			},
		},
		CheckDestroy: testAccCheckFirewallZonePolicyDestroy,
	})
}

// TestAccFirewallZonePolicyOrder_beforeAndAfter exercises the before/after split
// end-to-end: one policy is placed before the predefined policies and the other
// two after them. The implicit post-apply refresh-and-plan (and the import step)
// MUST be empty — that is what validates both that the controller honored the
// before/after placement and that partitionZonePairOrder reconstructs it back to
// the configured lists (the `Index < minPredefinedIndex` boundary inference). No
// ExpectNonEmptyPlan is set on purpose: a wrong inference fails here.
func TestAccFirewallZonePolicyOrder_beforeAndAfter(t *testing.T) {
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
					testAccFirewallZonePolicyMultiConfig(name, false),
					testAccFirewallZonePolicyOrderBeforeAfterConfig(orderedBeforeIDs1, orderedAfterIDs23),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testFirewallZonePolicyOrderResourceName, "id"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyOrderResourceName, "site", "default"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyOrderResourceName, "before_predefined_ids.#", "1"),
					resource.TestCheckResourceAttr(testFirewallZonePolicyOrderResourceName, "after_predefined_ids.#", "2"),
				),
			},
			pt.ImportStepWithSite(testFirewallZonePolicyOrderResourceName),
		},
		CheckDestroy: testAccCheckFirewallZonePolicyDestroy,
	})
}

const (
	orderedAfterIDs123 = `[
		unifi_firewall_zone_policy.test1.id,
		unifi_firewall_zone_policy.test2.id,
		unifi_firewall_zone_policy.test3.id,
	]`
	orderedAfterIDs312 = `[
		unifi_firewall_zone_policy.test3.id,
		unifi_firewall_zone_policy.test1.id,
		unifi_firewall_zone_policy.test2.id,
	]`
	orderedBeforeIDs1 = `[
		unifi_firewall_zone_policy.test1.id,
	]`
	orderedAfterIDs23 = `[
		unifi_firewall_zone_policy.test2.id,
		unifi_firewall_zone_policy.test3.id,
	]`
)

func testAccFirewallZonePolicyOrderConfig(afterIDs string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy_order" "order" {
	source_zone_id      = unifi_firewall_zone.test.id
	destination_zone_id = unifi_firewall_zone.test.id

	after_predefined_ids = %[1]s
}
`, afterIDs)
}

func testAccFirewallZonePolicyOrderBeforeAfterConfig(beforeIDs, afterIDs string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone_policy_order" "order" {
	source_zone_id      = unifi_firewall_zone.test.id
	destination_zone_id = unifi_firewall_zone.test.id

	before_predefined_ids = %[1]s
	after_predefined_ids  = %[2]s
}
`, beforeIDs, afterIDs)
}

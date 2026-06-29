package acctest

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// global_switch is a site-global singleton; serialize all tests that mutate it.
var settingGlobalSwitchLock = &sync.Mutex{}

const settingGlobalSwitchResourceName = "unifi_setting_global_switch.test"

// TestAccSettingGlobalSwitch_basic covers create + import + update of the
// switch_exclusions collection.
func TestAccSettingGlobalSwitch_basic(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGlobalSwitchLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGlobalSwitchExclusions("00:11:22:33:44:55"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(settingGlobalSwitchResourceName, "id"),
					resource.TestCheckResourceAttr(settingGlobalSwitchResourceName, "site", "default"),
					pt.TestCheckSetResourceAttr(settingGlobalSwitchResourceName, "switch_exclusions", "00:11:22:33:44:55"),
				),
				ConfigPlanChecks: pt.CheckResourceActions(settingGlobalSwitchResourceName, plancheck.ResourceActionCreate),
			},
			pt.ImportStepWithSite(settingGlobalSwitchResourceName),
			{
				Config: testAccSettingGlobalSwitchExclusions("00:11:22:33:44:55", "aa:bb:cc:dd:ee:ff"),
				Check: resource.ComposeTestCheckFunc(
					pt.TestCheckSetResourceAttr(settingGlobalSwitchResourceName, "switch_exclusions",
						"00:11:22:33:44:55", "aa:bb:cc:dd:ee:ff"),
				),
				ConfigPlanChecks: pt.CheckResourceActions(settingGlobalSwitchResourceName, plancheck.ResourceActionUpdate),
			},
			// Re-apply identical config -> no-op (idempotency).
			{
				Config:   testAccSettingGlobalSwitchExclusions("00:11:22:33:44:55", "aa:bb:cc:dd:ee:ff"),
				PlanOnly: true,
			},
		},
	})
}

// TestAccSettingGlobalSwitch_macNormalization proves that uppercase/hyphenated
// MACs in config are canonicalized at plan time and produce no perpetual diff.
func TestAccSettingGlobalSwitch_macNormalization(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGlobalSwitchLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGlobalSwitchExclusions("AA-BB-CC-DD-EE-FF"),
				Check: resource.ComposeTestCheckFunc(
					pt.TestCheckSetResourceAttr(settingGlobalSwitchResourceName, "switch_exclusions", "aa:bb:cc:dd:ee:ff"),
				),
			},
			// Same MAC, different casing/separators -> no diff.
			{
				Config:   testAccSettingGlobalSwitchExclusions("aa:bb:cc:dd:ee:ff"),
				PlanOnly: true,
			},
		},
	})
}

// TestAccSettingGlobalSwitch_clobberGuard is the flagship regression: it seeds a
// non-modeled field (JumboframeEnabled) out-of-band, then manages only an
// isolation field. The read-modify-write path must preserve the seeded toggle
// across both Create and Update.
func TestAccSettingGlobalSwitch_clobberGuard(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGlobalSwitchLock,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					ctx := context.Background()
					cur, err := testClient.GetSettingGlobalSwitch(ctx, "default")
					if err != nil {
						cur = &unifi.SettingGlobalSwitch{}
					}
					cur.JumboframeEnabled = true
					if _, err := testClient.UpdateSettingGlobalSwitch(ctx, "default", cur); err != nil {
						t.Fatalf("seeding global_switch failed: %s", err)
					}
				},
				Config: testAccSettingGlobalSwitchDeviceIsolation("dev-1"),
				Check: resource.ComposeTestCheckFunc(
					pt.TestCheckSetResourceAttr(settingGlobalSwitchResourceName, "acl_device_isolation", "dev-1"),
					checkGlobalSwitchJumboframe(t, true),
				),
			},
			// Update a managed field; the seeded toggle must still survive.
			{
				Config: testAccSettingGlobalSwitchDeviceIsolation("dev-2"),
				Check: resource.ComposeTestCheckFunc(
					pt.TestCheckSetResourceAttr(settingGlobalSwitchResourceName, "acl_device_isolation", "dev-2"),
					checkGlobalSwitchJumboframe(t, true),
				),
			},
		},
	})
}

// TestAccSettingGlobalSwitch_adoptWithoutManaging applies the resource with no
// isolation attributes configured; every controller field must be preserved.
func TestAccSettingGlobalSwitch_adoptWithoutManaging(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGlobalSwitchLock,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					ctx := context.Background()
					cur, err := testClient.GetSettingGlobalSwitch(ctx, "default")
					if err != nil {
						cur = &unifi.SettingGlobalSwitch{}
					}
					cur.JumboframeEnabled = true
					if _, err := testClient.UpdateSettingGlobalSwitch(ctx, "default", cur); err != nil {
						t.Fatalf("seeding global_switch failed: %s", err)
					}
				},
				Config: `resource "unifi_setting_global_switch" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(settingGlobalSwitchResourceName, "id"),
					checkGlobalSwitchJumboframe(t, true),
				),
			},
		},
	})
}

// TestAccSettingGlobalSwitch_aclL3Isolation creates a layer-3 isolation rule
// wired to real networks.
func TestAccSettingGlobalSwitch_aclL3Isolation(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGlobalSwitchLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGlobalSwitchL3Config(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(settingGlobalSwitchResourceName, "acl_l3_isolation.#", "1"),
				),
			},
		},
	})
}

// TestAccSettingGlobalSwitch_emptyCollectionRejected verifies the plan-time
// empty-collection guard (SizeAtLeast(1)).
func TestAccSettingGlobalSwitch_emptyCollectionRejected(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGlobalSwitchLock,
		Steps: []resource.TestStep{
			{
				Config: `resource "unifi_setting_global_switch" "test" {
	switch_exclusions = []
}`,
				ExpectError: regexp.MustCompile(`(?i)at least 1`),
			},
		},
	})
}

// TestAccSettingGlobalSwitch_duplicateSourceNetwork verifies the plan-time
// source_network uniqueness validator.
func TestAccSettingGlobalSwitch_duplicateSourceNetwork(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGlobalSwitchLock,
		Steps: []resource.TestStep{
			{
				Config: `resource "unifi_setting_global_switch" "test" {
	acl_l3_isolation = [
		{ source_network = "net-a", destination_networks = ["net-b"] },
		{ source_network = "net-a", destination_networks = ["net-c"] },
	]
}`,
				ExpectError: regexp.MustCompile(`(?i)duplicate source_network`),
			},
		},
	})
}

// TestAccSettingGlobalSwitch_drift mutates a managed field out-of-band and
// asserts the next plan is non-empty (drift detection).
func TestAccSettingGlobalSwitch_drift(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGlobalSwitchLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGlobalSwitchExclusions("00:11:22:33:44:55"),
			},
			{
				PreConfig: func() {
					ctx := context.Background()
					cur, err := testClient.GetSettingGlobalSwitch(ctx, "default")
					if err != nil {
						t.Fatalf("reading global_switch failed: %s", err)
					}
					cur.SwitchExclusions = []string{"99:99:99:99:99:99"}
					if _, err := testClient.UpdateSettingGlobalSwitch(ctx, "default", cur); err != nil {
						t.Fatalf("mutating global_switch failed: %s", err)
					}
				},
				Config:             testAccSettingGlobalSwitchExclusions("00:11:22:33:44:55"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func checkGlobalSwitchJumboframe(t *testing.T, want bool) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		cur, err := testClient.GetSettingGlobalSwitch(context.Background(), "default")
		if err != nil {
			return fmt.Errorf("reading global_switch: %w", err)
		}
		if cur.JumboframeEnabled != want {
			return fmt.Errorf("jumboframe_enabled = %t, want %t (unmanaged field was clobbered)", cur.JumboframeEnabled, want)
		}
		return nil
	}
}

func testAccSettingGlobalSwitchExclusions(macs ...string) string {
	quoted := ""
	for i, m := range macs {
		if i > 0 {
			quoted += ", "
		}
		quoted += fmt.Sprintf("%q", m)
	}
	return fmt.Sprintf(`
resource "unifi_setting_global_switch" "test" {
	switch_exclusions = [%s]
}
`, quoted)
}

func testAccSettingGlobalSwitchDeviceIsolation(ids ...string) string {
	quoted := ""
	for i, id := range ids {
		if i > 0 {
			quoted += ", "
		}
		quoted += fmt.Sprintf("%q", id)
	}
	return fmt.Sprintf(`
resource "unifi_setting_global_switch" "test" {
	acl_device_isolation = [%s]
}
`, quoted)
}

func testAccSettingGlobalSwitchL3Config() string {
	return `
resource "unifi_network" "src" {
	name    = "tfacc-gs-src"
	purpose = "corporate"
	subnet  = "10.42.10.1/24"
	vlan_id = 4210
}

resource "unifi_network" "dst" {
	name    = "tfacc-gs-dst"
	purpose = "corporate"
	subnet  = "10.42.11.1/24"
	vlan_id = 4211
}

resource "unifi_setting_global_switch" "test" {
	acl_l3_isolation = [
		{
			source_network       = unifi_network.src.id
			destination_networks = [unifi_network.dst.id]
		},
	]
}
`
}

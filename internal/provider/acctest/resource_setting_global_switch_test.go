package acctest

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/filipowm/go-unifi/unifi"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// global_switch is a site-global singleton; serialize all tests that mutate it.
var settingGlobalSwitchLock = &sync.Mutex{}

const settingGlobalSwitchResourceName = "unifi_setting_global_switch.test"

// nonPoolSwitchMACs returns up to n switch MAC addresses that the unifi_device
// resource test pool (allocateDevice) never selects. switch_exclusions only
// accepts MACs of switches the controller manages, so these demo switches must
// be adopted (via the unifi_device resources wired into each config) before the
// global_switch write. Using switches outside the allocateDevice pool means the
// global_switch tests never race the device-resource tests for a device; and
// because all global_switch tests are serialized via settingGlobalSwitchLock,
// reusing the same MACs across them is safe.
func nonPoolSwitchMACs(t *testing.T, n int) []string {
	t.Helper()
	pt.MarkAccTest(t)
	ctx := context.Background()

	var macs []string
	err := retry.RetryContext(ctx, 90*time.Second, func() *retry.RetryError {
		devices, err := testClient.ListDevice(ctx, "default")
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("listing devices: %w", err))
		}
		macs = macs[:0]
		for _, d := range devices {
			if d.Type != "usw" {
				continue
			}
			// Skip devices the unifi_device pool may adopt.
			if isBroadcomSwitch(d) || isMicrosemiSwitch(d) || isNephosSwitch(d) {
				continue
			}
			// Skip PSU/PDU pseudo-switches that aren't real switches.
			if d.Model == "USPRPS" || d.Model == "USPRPSP" || d.Model == "USPPDUHD" || d.Model == "USPPDUP" {
				continue
			}
			macs = append(macs, d.MAC)
			if len(macs) == n {
				break
			}
		}
		if len(macs) < n {
			return retry.RetryableError(fmt.Errorf("need %d non-pool switches, found %d", n, len(macs)))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return macs
}

// gsAdoptSwitches renders unifi_device resources (sw0, sw1, ...) that adopt the
// given demo switch MACs, so their MACs can be referenced from switch_exclusions
// (e.g. unifi_device.sw0.mac) and are guaranteed adopted before the global_switch
// write via the implicit dependency.
func gsAdoptSwitches(macs ...string) string {
	var b strings.Builder
	for i, m := range macs {
		fmt.Fprintf(&b, `
resource "unifi_device" "sw%d" {
	mac               = %q
	allow_adoption    = true
	forget_on_destroy = true
}
`, i, m)
	}
	return b.String()
}

// gsSeedJumboframe sets a non-modeled field (jumboframe_enabled) out-of-band so
// clobber-guard / adopt tests can assert the read-modify-write path preserves it.
// Invoked from a step PreConfig (under settingGlobalSwitchLock). Note: the
// controller's switch_exclusions field is omitempty, so it cannot be cleared via
// the API (mirroring the resource's documented limitation); a leftover value from
// a prior serialized test is harmless because tests that manage switch_exclusions
// overwrite it, and re-sending an existing value is tolerated by the controller.
func gsSeedJumboframe(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	cur, err := testClient.GetSettingGlobalSwitch(ctx, "default")
	if err != nil {
		// Only treat a genuinely-absent setting as "start from scratch"; any other
		// read error must fail loudly rather than silently writing a blank object
		// that would clobber unrelated controller fields on the seed PUT.
		if !errors.Is(err, unifi.ErrNotFound) {
			t.Fatalf("reading global_switch for seeding failed: %s", err)
		}
		cur = &unifi.SettingGlobalSwitch{}
	}
	cur.JumboframeEnabled = true
	if _, err := testClient.UpdateSettingGlobalSwitch(ctx, "default", cur); err != nil {
		t.Fatalf("seeding global_switch failed: %s", err)
	}
}

// TestAccSettingGlobalSwitch_basic covers create + import + update of the
// switch_exclusions collection against real, adopted switch devices.
func TestAccSettingGlobalSwitch_basic(t *testing.T) {
	macs := nonPoolSwitchMACs(t, 2)
	devices := gsAdoptSwitches(macs[0], macs[1])

	one := devices + `
resource "unifi_setting_global_switch" "test" {
	switch_exclusions = [unifi_device.sw0.mac]
}
`
	two := devices + `
resource "unifi_setting_global_switch" "test" {
	switch_exclusions = [unifi_device.sw0.mac, unifi_device.sw1.mac]
}
`

	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGlobalSwitchLock,
		Steps: []resource.TestStep{
			{
				Config: one,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(settingGlobalSwitchResourceName, "id"),
					resource.TestCheckResourceAttr(settingGlobalSwitchResourceName, "site", "default"),
					resource.TestCheckResourceAttr(settingGlobalSwitchResourceName, "switch_exclusions.#", "1"),
					resource.TestCheckTypeSetElemAttr(settingGlobalSwitchResourceName, "switch_exclusions.*", macs[0]),
				),
				ConfigPlanChecks: pt.CheckResourceActions(settingGlobalSwitchResourceName, plancheck.ResourceActionCreate),
			},
			pt.ImportStepWithSite(settingGlobalSwitchResourceName),
			{
				Config: two,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(settingGlobalSwitchResourceName, "switch_exclusions.#", "2"),
					resource.TestCheckTypeSetElemAttr(settingGlobalSwitchResourceName, "switch_exclusions.*", macs[0]),
					resource.TestCheckTypeSetElemAttr(settingGlobalSwitchResourceName, "switch_exclusions.*", macs[1]),
				),
				ConfigPlanChecks: pt.CheckResourceActions(settingGlobalSwitchResourceName, plancheck.ResourceActionUpdate),
			},
			// Re-apply identical config -> no-op (idempotency).
			{
				Config:   two,
				PlanOnly: true,
			},
		},
	})
}

// TestAccSettingGlobalSwitch_macNormalization proves the MAC-normalization fix:
// a non-canonical (uppercase/hyphenated) MAC is accepted on create without a
// "provider produced an invalid plan" error (the previous plan-modifier approach
// failed here), the value is kept verbatim in state, and the controller's
// canonical echo produces no perpetual diff on a no-op re-apply (MAC semantic
// equality reconciles the refresh).
func TestAccSettingGlobalSwitch_macNormalization(t *testing.T) {
	macs := nonPoolSwitchMACs(t, 1)
	devices := gsAdoptSwitches(macs[0])

	// Reference the adopted device's MAC but rewrite it to a hyphen-separated,
	// uppercase form. Keeps the implicit dependency (adopt before write) while
	// feeding a non-canonical form through switch_exclusions.
	noncanonical := devices + `
resource "unifi_setting_global_switch" "test" {
	switch_exclusions = [upper(replace(unifi_device.sw0.mac, ":", "-"))]
}
`
	wantMAC := strings.ToUpper(strings.ReplaceAll(macs[0], ":", "-"))

	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGlobalSwitchLock,
		Steps: []resource.TestStep{
			{
				Config: noncanonical,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(settingGlobalSwitchResourceName, "switch_exclusions.#", "1"),
					// The non-canonical form is preserved verbatim in state (the
					// controller's canonical echo is reconciled away by MAC semantic
					// equality), so a refresh/re-apply produces no diff.
					resource.TestCheckTypeSetElemAttr(settingGlobalSwitchResourceName, "switch_exclusions.*", wantMAC),
				),
			},
			// The controller stores the MAC canonically (lowercase, colon); the
			// refresh of the same non-canonical config must reconcile to no diff.
			{
				Config:   noncanonical,
				PlanOnly: true,
			},
		},
	})
}

// TestAccSettingGlobalSwitch_clobberGuard is the flagship read-modify-write
// regression: it seeds a non-modeled field (jumboframe_enabled) out-of-band, then
// manages only switch_exclusions. The seeded toggle must survive both Create and
// Update (when the managed value changes to a different adopted switch).
func TestAccSettingGlobalSwitch_clobberGuard(t *testing.T) {
	macs := nonPoolSwitchMACs(t, 2)
	devices := gsAdoptSwitches(macs[0], macs[1])

	first := devices + `
resource "unifi_setting_global_switch" "test" {
	switch_exclusions = [unifi_device.sw0.mac]
}
`
	second := devices + `
resource "unifi_setting_global_switch" "test" {
	switch_exclusions = [unifi_device.sw1.mac]
}
`

	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGlobalSwitchLock,
		Steps: []resource.TestStep{
			{
				PreConfig: func() { gsSeedJumboframe(t) },
				Config:    first,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(settingGlobalSwitchResourceName, "switch_exclusions.#", "1"),
					resource.TestCheckTypeSetElemAttr(settingGlobalSwitchResourceName, "switch_exclusions.*", macs[0]),
					checkGlobalSwitchJumboframe(t, true),
				),
			},
			// Update the managed field (swap to the other adopted switch); the
			// seeded toggle must still survive.
			{
				Config: second,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(settingGlobalSwitchResourceName, "switch_exclusions.#", "1"),
					resource.TestCheckTypeSetElemAttr(settingGlobalSwitchResourceName, "switch_exclusions.*", macs[1]),
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
				PreConfig: func() { gsSeedJumboframe(t) },
				Config:    `resource "unifi_setting_global_switch" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(settingGlobalSwitchResourceName, "id"),
					checkGlobalSwitchJumboframe(t, true),
				),
			},
		},
	})
}

// TestAccSettingGlobalSwitch_aclL3Isolation creates a layer-3 isolation rule
// wired to real networks. The controller only accepts global ACL entries when an
// adopted gateway with L3-isolation support is present, which the Dockerized test
// controller is not (it returns api.err.OverMaxEntriesOfGlobalAcl), so this is
// gated to local runs with real hardware.
func TestAccSettingGlobalSwitch_aclL3Isolation(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping: acl_l3_isolation requires an adopted gateway with L3-isolation support not available on the Docker test controller")
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGlobalSwitchLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGlobalSwitchL3("unifi_network.src.id", "unifi_network.dst.id"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(settingGlobalSwitchResourceName, "acl_l3_isolation.#", "1"),
				),
			},
		},
	})
}

// TestAccSettingGlobalSwitch_duplicateMacRejected verifies the plan-time
// UniqueMACs validator: two spellings of the same MAC (differing only in case
// and separator) are rejected before any controller call.
func TestAccSettingGlobalSwitch_duplicateMacRejected(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGlobalSwitchLock,
		Steps: []resource.TestStep{
			{
				Config: `resource "unifi_setting_global_switch" "test" {
	switch_exclusions = ["AA-BB-CC-DD-EE-FF", "aa:bb:cc:dd:ee:ff"]
}`,
				ExpectError: regexp.MustCompile(`(?i)same MAC address`),
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

// TestAccSettingGlobalSwitch_drift mutates the managed field out-of-band (to a
// different adopted switch, since the field cannot be cleared) and asserts the
// next plan is non-empty (drift detection).
func TestAccSettingGlobalSwitch_drift(t *testing.T) {
	macs := nonPoolSwitchMACs(t, 2)
	config := gsAdoptSwitches(macs[0], macs[1]) + `
resource "unifi_setting_global_switch" "test" {
	switch_exclusions = [unifi_device.sw0.mac]
}
`

	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGlobalSwitchLock,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				// Swap switch_exclusions to the other adopted switch out-of-band to
				// create drift; both switches stay adopted (still in config), so the
				// next plan reverts it back to sw0.
				PreConfig: func() {
					ctx := context.Background()
					cur, err := testClient.GetSettingGlobalSwitch(ctx, "default")
					if err != nil {
						t.Fatalf("reading global_switch failed: %s", err)
					}
					cur.SwitchExclusions = []string{macs[1]}
					if _, err := testClient.UpdateSettingGlobalSwitch(ctx, "default", cur); err != nil {
						t.Fatalf("mutating global_switch failed: %s", err)
					}
				},
				Config:             config,
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

// testAccSettingGlobalSwitchL3 builds a config with two real networks (src/dst)
// and a single acl_l3_isolation rule wiring the given source/destination network
// references (e.g. "unifi_network.src.id"). Using real unifi_network IDs keeps
// the layer-3 isolation tests on a verified value format.
func testAccSettingGlobalSwitchL3(source, dest string) string {
	return fmt.Sprintf(`
resource "unifi_network" "src" {
	name    = "tfacc-gs-src"
	purpose = "corporate"
	subnet  = "10.42.10.1/24"
	vlan_id = 4000
}

resource "unifi_network" "dst" {
	name    = "tfacc-gs-dst"
	purpose = "corporate"
	subnet  = "10.42.11.1/24"
	vlan_id = 4001
}

resource "unifi_setting_global_switch" "test" {
	acl_l3_isolation = [
		{
			source_network       = %s
			destination_networks = [%s]
		},
	]
}
`, source, dest)
}

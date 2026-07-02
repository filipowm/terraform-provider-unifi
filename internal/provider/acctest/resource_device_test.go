package acctest

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
)

var (
	deviceInit sync.Once
	devicePool mapset.Set[*unifi.Device] = mapset.NewSet[*unifi.Device]()
)

func allocateDevice(t *testing.T) (*unifi.Device, func()) {
	t.Helper()
	pt.MarkAccTest(t)
	ctx := context.Background()

	deviceInit.Do(func() {
		// The demo devices don't appear instantly when the controller starts.
		err := retry.RetryContext(ctx, 1*time.Minute, func() *retry.RetryError {
			devices, err := testClient.ListDevice(ctx, "default")
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("Error listing devices: %w", err))
			}

			if len(devices) == 0 {
				return retry.RetryableError(errors.New("No devices found"))
			}

			for _, device := range devices {
				if device.Type != "usw" {
					continue
				}

				// These devices aren't really switches.
				if device.Model == "USPRPS" || device.Model == "USPRPSP" || device.Model == "USPPDUHD" || device.Model == "USPPDUP" {
					continue
				}

				// The USW-Leaf is an EOL product and consistently fails to be adopted.
				if device.Model == "UDC48X6" {
					continue
				}

				// Only switches with these chipsets support both port mirroring ang aggregation.
				if !isBroadcomSwitch(device) && !isMicrosemiSwitch(device) && !isNephosSwitch(device) {
					continue
				}

				d := device
				if ok := devicePool.Add(&d); !ok {
					return retry.NonRetryableError(errors.New("Failed to add device to pool"))
				}
			}

			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	var device *unifi.Device

	err := retry.RetryContext(ctx, 1*time.Minute, func() *retry.RetryError {
		var ok bool
		device, ok = devicePool.Pop()

		if device == nil || !ok {
			return retry.RetryableError(errors.New("Unable to allocate test device"))
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	unallocate := func() {
		if ok := devicePool.Add(device); !ok {
			t.Fatal("Failed to add device to pool")
		}
	}

	return device, unallocate
}

func isBroadcomSwitch(device unifi.Device) bool {
	if device.Type != "usw" {
		return false
	}

	switch device.Model {
	// US-8 variants
	case "US8", "US8P60", "US8P150", "S28150":
		return true

	// US-16 variants
	case "US16P150", "S216150", "USXG":
		return true

	// US-24 variants
	case "US24", "US24P250", "S224250", "US24P500", "S224500", "US24PL2":
		return true

	// US-48 variants
	case "US48", "US48P500", "S248500", "US48P750", "S248750", "US48PL2":
		return true

	// USW-Pro
	case "US24PRO", "US24PRO2", "US48PRO", "US48PRO2", "USAGGPRO":
		return true

		// USW-Enterprise
	case "US624P", "US648P", "USXG24":
		return true

	// US-XG-6PoE
	case "US6XG150":
		return true
	}

	return false
}

func isMicrosemiSwitch(device unifi.Device) bool {
	if device.Type != "usw" {
		return false
	}

	switch device.Model {
	// US-8 variants
	case "USC8", "USC8P60", "USC8P150":
		return true

	// USW-Industrial
	case "USC8P450":
		return true
	}

	return false
}

func isNephosSwitch(device unifi.Device) bool {
	if device.Type != "usw" {
		return false
	}

	switch device.Model {
	// USW-Leaf
	case "UDC48X6":
		return true
	}

	return false
}

func preCheckDeviceExists(t *testing.T, site, mac string) {
	t.Helper()
	_, err := testClient.GetDeviceByMAC(context.Background(), site, mac)

	if errors.Is(err, unifi.ErrNotFound) {
		t.Fatal("Test device not found")
	}
}

func TestAccDevice_empty(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		CheckDestroy: testAccCheckDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccDeviceConfigEmpty(),
				ExpectError: regexp.MustCompile(`no MAC address specified, please import the device using terraform import`),
			},
		},
	})
}

func TestAccDevice_switch_basic(t *testing.T) {
	// t.Skip("FIXME")
	resourceName := "unifi_device.test"
	site := "default"

	device, unallocateDevice := allocateDevice(t)
	defer unallocateDevice()

	importStateVerifyIgnore := []string{"allow_adoption", "forget_on_destroy", "name"}

	AcceptanceTest(t, AcceptanceTestCase{
		PreCheck: func() {
			preCheckDeviceExists(t, site, device.MAC)
		},
		CheckDestroy: testAccCheckDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceConfig(device.MAC),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "site", site),
					resource.TestCheckResourceAttr(resourceName, "mac", device.MAC),
					resource.TestCheckResourceAttr(resourceName, "name", ""),
				),
			},

			// Import with ID
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: importStateVerifyIgnore,
			},

			// Import with MAC
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateId:           device.MAC,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: importStateVerifyIgnore,
			},

			{
				Config: testAccDeviceConfigWithName(device.MAC, "Test Switch"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "Test Switch"),
				),
			},
		},
	})
}

// TestAccDevice_switch_portOverrides covers the port_override attributes the
// Dockerized demo switches reliably accept and persist: per-port name, op_mode,
// and poe_mode. Advanced overrides that the demo controller cannot faithfully
// persist — LAG aggregation and the inline per-port VLAN cluster — are covered
// by TestAccDevice_switch_portOverrides_inlineVLAN (gated on TF_ACC_LOCAL) and by
// the offline unit tests in the device package. See the note on that test.
func TestAccDevice_switch_portOverrides(t *testing.T) {
	resourceName := "unifi_device.test"
	site := "default"

	device, unallocateDevice := allocateDevice(t)
	defer unallocateDevice()

	importStateVerifyIgnore := []string{"allow_adoption", "forget_on_destroy", "name"}

	AcceptanceTest(t, AcceptanceTestCase{
		PreCheck: func() {
			preCheckDeviceExists(t, site, device.MAC)
		},
		CheckDestroy: testAccCheckDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceConfigWithPortOverridesBasic(device.MAC),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "port_override.#", "3"),

					// TypeSet membership assertions (order-independent): the
					// element index reshuffles when the schema changes, so match
					// on the nested attribute values instead of positional keys.
					// op_mode is intentionally not asserted for port 2: the
					// controller drops op_mode="switch" (the omitempty default), so
					// it reads back as "" (see the DiffSuppressFunc on op_mode).
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_override.*", map[string]string{
						"number": "1",
						"name":   "Port 1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_override.*", map[string]string{
						"number": "2",
						"name":   "Port 2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_override.*", map[string]string{
						"number":   "4",
						"poe_mode": "pasv24",
					}),
				),
			},
			// Merge gate: the same config must produce no further plan, proving
			// these overrides persisted on the controller with no perpetual diff.
			{
				Config:   testAccDeviceConfigWithPortOverridesBasic(device.MAC),
				PlanOnly: true,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: importStateVerifyIgnore,
			},
			{
				Config: testAccDeviceConfig(device.MAC),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "port_override.#", "0"),
				),
			},
		},
	})
}

// TestAccDevice_switch_portOverrides_inlineVLAN verifies the LAG aggregation and
// inline per-port VLAN cluster (native_networkconf_id, forward, tagged_vlan_mgmt,
// excluded_network_ids, setting_preference) end-to-end against a real controller.
//
// It is gated on TF_ACC_LOCAL because the Dockerized demo switches do NOT
// faithfully persist these fields: v9.x controllers reject the LAG member range
// with api.err.InvalidAggregateRange, and older controllers silently drop both
// the aggregate members and the inline VLAN fields (they read back empty). The
// provider-side conversion is covered offline by the device-package unit tests
// (TestToPortOverride_VLANFields, TestPortOverride_VLANRoundTrip,
// TestToPortOverrideAggregateTranslation, …); this test proves the round-trip
// against real switch hardware where the controller actually persists them.
//
// Note: the LAG members (ports 3-4) must not carry their own port_override — a
// port cannot be both an aggregate member and individually configured. Real LAG
// port ranges are switch-model specific; adjust the port numbers to match the
// hardware under test if needed.
func TestAccDevice_switch_portOverrides_inlineVLAN(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "inline per-port VLAN overrides and LAG aggregation require real switch hardware not available on the Docker test controller")

	resourceName := "unifi_device.test"
	site := "default"

	device, unallocateDevice := allocateDevice(t)
	defer unallocateDevice()

	importStateVerifyIgnore := []string{"allow_adoption", "forget_on_destroy", "name"}

	AcceptanceTest(t, AcceptanceTestCase{
		PreCheck: func() {
			preCheckDeviceExists(t, site, device.MAC)
		},
		CheckDestroy: testAccCheckDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceConfigWithPortOverrides(device.MAC),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "port_override.#", "3"),

					// LAG aggregation: port 3 aggregates the contiguous range
					// [3, 4]; port 4 is deliberately not declared separately.
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_override.*", map[string]string{
						"number":              "3",
						"op_mode":             "aggregate",
						"aggregate_num_ports": "2",
					}),
					// Inline per-port VLAN overrides: a native (access) port and a
					// customized trunk that excludes one network. Identify each
					// element by its declared, deterministic attributes only. The
					// real native_networkconf_id is a computed network ID, so it is
					// asserted to be non-empty via the dedicated state check below,
					// and the PlanOnly merge gate proves it round-trips.
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_override.*", map[string]string{
						"number":             "5",
						"forward":            "customize",
						"setting_preference": "manual",
					}),
					testAccCheckPortOverrideNativeNetworkSet(resourceName, 5),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_override.*", map[string]string{
						"number":             "6",
						"forward":            "customize",
						"tagged_vlan_mgmt":   "custom",
						"setting_preference": "manual",
					}),
				),
			},
			// Merge gate: the same config must produce no further plan, proving
			// the inline VLAN overrides actually persisted on the controller (and
			// that setting_preference=manual is sufficient for persistence).
			{
				Config:   testAccDeviceConfigWithPortOverrides(device.MAC),
				PlanOnly: true,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: importStateVerifyIgnore,
			},
			{
				Config: testAccDeviceConfig(device.MAC),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "port_override.#", "0"),
				),
			},
		},
	})
}

func testAccDeviceConfigEmpty() string {
	return `
resource "unifi_device" "test" {}
`
}

func testAccDeviceConfig(mac string) string {
	return fmt.Sprintf(`
resource "unifi_device" "test" {
	mac = %q
}
`, mac)
}

func testAccDeviceConfigWithName(mac, name string) string {
	return fmt.Sprintf(`
resource "unifi_device" "test" {
	mac  = %q
	name = %q
}
`, mac, name)
}

// testAccDeviceConfigWithPortOverridesBasic renders the port_override fields the
// Dockerized demo switches reliably persist (name, op_mode, poe_mode). Used by the
// always-on TestAccDevice_switch_portOverrides.
func testAccDeviceConfigWithPortOverridesBasic(mac string) string {
	return fmt.Sprintf(`
resource "unifi_device" "test" {
	mac = %q

	port_override {
		number = 1
		name   = "Port 1"
	}

	port_override {
		number  = 2
		name    = "Port 2"
		op_mode = "switch"
	}

	port_override {
		number   = 4
		poe_mode = "pasv24"
	}
}
`, mac)
}

// testAccDeviceConfigWithPortOverrides renders the LAG aggregation and inline
// per-port VLAN cluster. Used only by the TF_ACC_LOCAL-gated
// TestAccDevice_switch_portOverrides_inlineVLAN, because the Dockerized demo
// switches do not persist these fields. Port 4 is intentionally NOT declared: it
// is an aggregate member of port 3 (a port cannot be both a LAG member and
// individually overridden — that is what triggers api.err.InvalidAggregateRange).
func testAccDeviceConfigWithPortOverrides(mac string) string {
	return fmt.Sprintf(`
resource "unifi_network" "test_native" {
	name    = "tfacc-device-native"
	purpose = "corporate"

	subnet       = "10.97.0.1/24"
	vlan_id      = 97
	dhcp_start   = "10.97.0.6"
	dhcp_stop    = "10.97.0.254"
	dhcp_enabled = true
}

resource "unifi_network" "test_excluded" {
	name    = "tfacc-device-excluded"
	purpose = "corporate"

	subnet       = "10.98.0.1/24"
	vlan_id      = 98
	dhcp_start   = "10.98.0.6"
	dhcp_stop    = "10.98.0.254"
	dhcp_enabled = true
}

resource "unifi_device" "test" {
	mac = %q

	# LAG aggregation over the contiguous range [3, 4]. The member ports must
	# not carry their own port_override.
	port_override {
		number              = 3
		op_mode             = "aggregate"
		aggregate_num_ports = 2
	}

	# Inline access port: untagged on the native network. The controller
	# canonicalizes any port that pins a custom native network to
	# forward = "customize" (it only stores "all" or "customize"), so use
	# that here to keep the config drift-free on the merge-gate re-plan.
	port_override {
		number                = 5
		name                  = "Access VLAN 97"
		forward               = "customize"
		native_networkconf_id = unifi_network.test_native.id
		setting_preference    = "manual"
	}

	# Inline customized trunk: tag everything except the excluded network.
	port_override {
		number               = 6
		name                 = "Trunk except VLAN 98"
		forward              = "customize"
		tagged_vlan_mgmt     = "custom"
		excluded_network_ids = [unifi_network.test_excluded.id]
		setting_preference   = "manual"
	}
}
`, mac)
}

// testAccCheckPortOverrideNativeNetworkSet asserts that the port_override block
// for the given port number has a non-empty native_networkconf_id. The element's
// set index is a hash we don't want to hard-code, so locate it by matching the
// `number` attribute and then inspect that element's native_networkconf_id. This
// proves the computed network ID actually persisted (paired with the PlanOnly
// merge-gate step that proves it round-trips with no diff).
func testAccCheckPortOverrideNativeNetworkSet(resourceName string, number int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		want := strconv.Itoa(number)
		for k, v := range rs.Primary.Attributes {
			if !strings.HasPrefix(k, "port_override.") || !strings.HasSuffix(k, ".number") || v != want {
				continue
			}
			hash := strings.TrimSuffix(strings.TrimPrefix(k, "port_override."), ".number")
			native := rs.Primary.Attributes["port_override."+hash+".native_networkconf_id"]
			if native == "" {
				return fmt.Errorf("port_override number %d: expected non-empty native_networkconf_id, got empty", number)
			}
			return nil
		}
		return fmt.Errorf("port_override with number %d not found in state", number)
	}
}

func testAccCheckDeviceDestroy(s *terraform.State) error {
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "unifi_device" {
			continue
		}

		device, err := testClient.GetDevice(ctx, rs.Primary.Attributes["site"], rs.Primary.ID)
		if device != nil {
			return fmt.Errorf("Device still exists with ID %v", rs.Primary.ID)
		}
		if !errors.Is(err, unifi.ErrNotFound) {
			return err
		}
	}

	return nil
}

func testAccCheckDeviceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No ID is set")
		}

		id := rs.Primary.ID
		site := rs.Primary.Attributes["site"]

		device, err := testClient.GetDevice(context.Background(), site, id)
		if device == nil {
			return fmt.Errorf("Device not found with ID %v", id)
		}
		if !errors.Is(err, unifi.ErrNotFound) {
			return err
		}

		return nil
	}
}

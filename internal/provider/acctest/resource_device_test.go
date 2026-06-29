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
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	deviceInit sync.Once
	devicePool mapset.Set[*unifi.Device] = mapset.NewSet[*unifi.Device]()
)

func allocateDevice(t *testing.T) (*unifi.Device, func()) {
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
				return retry.RetryableError(fmt.Errorf("No devices found"))
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
				if !(isBroadcomSwitch(device) || isMicrosemiSwitch(device) || isNephosSwitch(device)) {
					continue
				}

				d := device
				if ok := devicePool.Add(&d); !ok {
					return retry.NonRetryableError(fmt.Errorf("Failed to add device to pool"))
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
			return retry.RetryableError(fmt.Errorf("Unable to allocate test device"))
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
	//t.Skip("FIXME")
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
				Config: testAccDeviceConfig_withName(device.MAC, "Test Switch"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "Test Switch"),
				),
			},
		},
	})
}

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
				Config: testAccDeviceConfig_withPortOverrides(device.MAC),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "port_override.#", "6"),

					// TypeSet membership assertions (order-independent): the
					// element index reshuffles when the schema changes, so match
					// on the nested attribute values instead of positional keys.
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_override.*", map[string]string{
						"number":              "3",
						"op_mode":             "aggregate",
						"aggregate_num_ports": "2",
					}),
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
					// Inline per-port VLAN overrides: a native (access) port and a
					// customized trunk that excludes one network. Identify each
					// element by its declared, deterministic attributes only. The
					// real native_networkconf_id is a computed network ID, so it is
					// asserted to be non-empty via the dedicated state check below,
					// and the PlanOnly merge gate proves it round-trips.
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_override.*", map[string]string{
						"number":             "5",
						"forward":            "native",
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
				Config:   testAccDeviceConfig_withPortOverrides(device.MAC),
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

func testAccDeviceConfig_withName(mac, name string) string {
	return fmt.Sprintf(`
resource "unifi_device" "test" {
	mac  = %q
	name = %q
}
`, mac, name)
}

func testAccDeviceConfig_withPortOverrides(mac string) string {
	return fmt.Sprintf(`
data "unifi_port_profile" "all" {}

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

	port_override {
		number = 1
		name   = "Port 1"
	}

	port_override {
		number          = 2
		name            = "Port 2"
		port_profile_id = data.unifi_port_profile.all.id
		op_mode         = "switch"
	}

	port_override {
		number              = 3
		op_mode             = "aggregate"
		aggregate_num_ports = 2
	}

	port_override {
		number   = 4
		poe_mode = "pasv24"
	}

	# Inline access port: untagged on the native network.
	port_override {
		number                = 5
		name                  = "Access VLAN 97"
		forward               = "native"
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
			return fmt.Errorf("No ID is set")
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

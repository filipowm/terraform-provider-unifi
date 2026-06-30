package acctest

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/apparentlymart/go-cidr/cidr"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccNetwork_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	subnet1, vlan1 := pt.GetTestVLAN(t)
	subnet2, vlan2 := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkConfig(name, subnet1, vlan1, true, nil),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "domain_name", "foo.local"),
					resource.TestCheckResourceAttr("unifi_network.test", "vlan_id", strconv.Itoa(vlan1)),
					resource.TestCheckResourceAttr("unifi_network.test", "igmp_snooping", "true"),
				),
			},
			pt.ImportStep("unifi_network.test"),
			{
				Config: testAccNetworkConfig(name, subnet2, vlan2, false, nil),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "vlan_id", strconv.Itoa(vlan2)),
					resource.TestCheckResourceAttr("unifi_network.test", "igmp_snooping", "false"),
				),
			},
			pt.ImportStep("unifi_network.test"),
			// re-test import here with default site, but full ID string
			{
				ResourceName:      "unifi_network.test",
				ImportState:       true,
				ImportStateIdFunc: pt.SiteAndIDImportStateIDFunc("unifi_network.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetwork_weird_cidr(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	subnet, vlan := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkConfig(name, subnet, vlan, true, nil),
				Check:  resource.ComposeTestCheckFunc(
				// TODO: ...
				),
			},
			pt.ImportStep("unifi_network.test"),
		},
	})
}

func TestAccNetwork_dhcp_dns(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	subnet, vlan := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkConfig(name, subnet, vlan, true, []string{"192.168.1.101"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_dns.0", "192.168.1.101"),
				),
			},
			pt.ImportStep("unifi_network.test"),
			{
				Config: testAccNetworkConfig(name, subnet, vlan, true, []string{"192.168.1.101", "192.168.1.102"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_dns.0", "192.168.1.101"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_dns.1", "192.168.1.102"),
				),
			},
			pt.ImportStep("unifi_network.test"),
			{
				Config: testAccNetworkConfig(name, subnet, vlan, true, nil),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_dns.#", "0"),
				),
			},
			{
				Config: testAccNetworkConfig(name, subnet, vlan, true, []string{"192.168.1.101"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_dns.0", "192.168.1.101"),
				),
			},
		},
	})
}

func TestAccNetwork_dhcp_boot(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	subnet, vlan := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkConfigDHCPBoot(name, subnet, vlan),
				Check:  resource.ComposeTestCheckFunc(
				// TODO: ...
				),
			},
			pt.ImportStep("unifi_network.test"),
		},
	})
}

func TestAccNetwork_v6(t *testing.T) {
	t.Skip("FIXME")

	name := acctest.RandomWithPrefix("tfacc")
	subnet1, vlan1 := pt.GetTestVLAN(t)
	subnet2, vlan2 := pt.GetTestVLAN(t)
	subnet3, vlan3 := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkConfigV6(name, subnet1, vlan1, "static", "fd6a:37be:e362::1/64"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "domain_name", "foo.local"),
					resource.TestCheckResourceAttr("unifi_network.test", "vlan_id", strconv.Itoa(vlan1)),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6_static_subnet", "fd6a:37be:e362::1/64"),
				),
			},
			pt.ImportStep("unifi_network.test"),
			{
				Config: testAccNetworkConfigV6(name, subnet2, vlan2, "static", "fd6a:37be:e363::1/64"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "vlan_id", strconv.Itoa(vlan2)),
					resource.TestCheckResourceAttr("unifi_network.test", "ipv6_static_subnet", "fd6a:37be:e363::1/64"),
				),
			},
			pt.ImportStep("unifi_network.test"),
			{
				Config: testAccNetworkConfigDhcpV6(
					name,
					subnet3,
					vlan3,
					"fd6a:37be:e364::1/64",
					"fd6a:37be:e364::2",
					"fd6a:37be:e364::7d1",
					[]string{"2001:4860:4860::8888", "2001:4860:4860::8844"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "vlan_id", strconv.Itoa(vlan3)),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_v6_start", "fd6a:37be:e364::2"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_v6_stop", "fd6a:37be:e364::7d1"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_v6_lease", strconv.Itoa(12*60*60)),
				),
			},
			{
				Config: testAccNetworkConfigDhcpV6(
					name,
					subnet3,
					vlan3,
					"fd6a:37be:e365::1/64",
					"fd6a:37be:e364::2",
					"fd6a:37be:e364::7d1",
					[]string{"2001:4860:4860::8888", "2001:4860:4860::8844"}),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("api.err.InvalidDHCPv6Range")),
			},
		},
	})
}

func TestAccNetwork_wan(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")

	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testWanNetworkConfig(name, "WAN", "pppoe", "192.168.1.1", 1, "username", "password", "8.8.8.8", "4.4.4.4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.wan_test", "wan_networkgroup", "WAN"),
					resource.TestCheckResourceAttr("unifi_network.wan_test", "wan_type", "pppoe"),
					resource.TestCheckResourceAttr("unifi_network.wan_test", "wan_ip", "192.168.1.1"),
					resource.TestCheckResourceAttr("unifi_network.wan_test", "wan_egress_qos", "1"),
					resource.TestCheckResourceAttr("unifi_network.wan_test", "wan_username", "username"),
					resource.TestCheckResourceAttr("unifi_network.wan_test", "x_wan_password", "password"),

					resource.TestCheckOutput("wan_dns1", "8.8.8.8"),
					resource.TestCheckOutput("wan_dns2", "4.4.4.4"),
				),
			},
			pt.ImportStep("unifi_network.wan_test"),
			// remove qos
			{
				Config: testWanNetworkConfig(name, "WAN", "pppoe", "192.168.1.1", 0, "username", "password", "8.8.8.8", "4.4.4.4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.wan_test", "wan_networkgroup", "WAN"),
					resource.TestCheckResourceAttr("unifi_network.wan_test", "wan_type", "pppoe"),
					resource.TestCheckResourceAttr("unifi_network.wan_test", "wan_ip", "192.168.1.1"),
					resource.TestCheckResourceAttr("unifi_network.wan_test", "wan_egress_qos", "0"),
					resource.TestCheckResourceAttr("unifi_network.wan_test", "wan_username", "username"),
					resource.TestCheckResourceAttr("unifi_network.wan_test", "x_wan_password", "password"),

					resource.TestCheckOutput("wan_dns1", "8.8.8.8"),
					resource.TestCheckOutput("wan_dns2", "4.4.4.4"),
				),
			},
			pt.ImportStep("unifi_network.wan_test"),
			{
				Config: testWanNetworkConfig(name, "WAN", "pppoe", "192.168.1.1", 1, "username", "password", "8.8.8.8", "4.4.4.4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.wan_test", "wan_networkgroup", "WAN"),
					resource.TestCheckResourceAttr("unifi_network.wan_test", "wan_type", "pppoe"),
					resource.TestCheckResourceAttr("unifi_network.wan_test", "wan_ip", "192.168.1.1"),
					resource.TestCheckResourceAttr("unifi_network.wan_test", "wan_egress_qos", "1"),
					resource.TestCheckResourceAttr("unifi_network.wan_test", "wan_username", "username"),
					resource.TestCheckResourceAttr("unifi_network.wan_test", "x_wan_password", "password"),

					resource.TestCheckOutput("wan_dns1", "8.8.8.8"),
					resource.TestCheckOutput("wan_dns2", "4.4.4.4"),
				),
			},
			pt.ImportStep("unifi_network.wan_test"),
			{
				Config:      testWanV6NetworkConfig(name, "dhcpv6", 47),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("expected wan_dhcp_v6_pd_size to be in the range (48 - 64)")),
			},
			{
				Config:      testWanV6NetworkConfig(name, "invalid", 48),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("invalid value for wan_type_v6")),
			},
			{
				Config: testWanV6NetworkConfig(name, "dhcpv6", 48),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.wan_test", "wan_type_v6", "dhcpv6"),
					resource.TestCheckResourceAttr("unifi_network.wan_test", "wan_dhcp_v6_pd_size", "48"),
				),
			},
		},
	})
}

func TestAccNetwork_differentSite(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	subnet1, vlan1 := pt.GetTestVLAN(t)
	subnet2, vlan2 := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkWithSiteConfig(name, subnet1, vlan1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("unifi_network.test", "site", "unifi_site.test", "name"),
				),
			},
			{
				ResourceName:      "unifi_network.test",
				ImportState:       true,
				ImportStateIdFunc: pt.SiteAndIDImportStateIDFunc("unifi_network.test"),
				ImportStateVerify: true,
			},
			{
				Config: testAccNetworkWithSiteConfig(name, subnet2, vlan2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("unifi_network.test", "site", "unifi_site.test", "name"),
				),
			},
			{
				ResourceName:      "unifi_network.test",
				ImportState:       true,
				ImportStateIdFunc: pt.SiteAndIDImportStateIDFunc("unifi_network.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetwork_importByName(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	subnet1, vlan1 := pt.GetTestVLAN(t)
	subnet2, vlan2 := pt.GetTestVLAN(t)
	subnet3, vlan3 := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		Steps: []resource.TestStep{
			// Apply and import network by name.
			{
				Config: testAccNetworkConfig(name, subnet1, vlan1, true, nil),
			},
			{
				Config:            testAccNetworkConfig(name, subnet1, vlan1, true, nil),
				ResourceName:      "unifi_network.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     fmt.Sprintf("name=%s", name),
			},
			// Apply and test errors.
			{
				Config: testAccNetworkWithDuplicateNames(subnet2, vlan2, subnet3, vlan3, "DUPLICATE_NAME"),
			},
			// Test error on name that doesn't exist.
			{
				Config:            testAccNetworkWithDuplicateNames(subnet2, vlan2, subnet3, vlan3, "DUPLICATE_NAME"),
				ResourceName:      "unifi_network.test1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "name=BAD_NAME",
				ExpectError:       regexp.MustCompile("BAD_NAME"),
			},
			// Test error on multiple matches.
			{
				Config:            testAccNetworkWithDuplicateNames(subnet2, vlan2, subnet3, vlan3, "DUPLICATE_NAME"),
				ResourceName:      "unifi_network.test1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "name=DUPLICATE_NAME",
				ExpectError:       regexp.MustCompile("DUPLICATE_NAME"),
			},
		},
	})
}

func TestAccNetwork_dhcpRelay(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	subnet, vlan := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkConfigDHCPRelay(name, subnet, vlan, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_relay_enabled", "true"),
				),
			},
			pt.ImportStep("unifi_network.test"),
			{
				Config: testAccNetworkConfigDHCPRelay(name, subnet, vlan, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_relay_enabled", "false"),
				),
			},
			pt.ImportStep("unifi_network.test"),
		},
	})
}

// TestAccNetwork_dhcpGuarding is the regression test for issue #123: a value
// enabled for DHCP Guarding must not be silently cleared when an unrelated change
// triggers an Update with dhcp_guarding omitted from config.
func TestAccNetwork_dhcpGuarding(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	subnet, vlan := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			// 1. Enable DHCP Guarding explicitly (with the required trusted server).
			{
				Config: testAccNetworkConfigDHCPGuarding(name, subnet, vlan, "foo.local", true, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_guarding", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_guarding_trusted_servers.#", "1"),
				),
			},
			// 2. Import round-trip.
			pt.ImportStep("unifi_network.test"),
			// 3. Decisive #123 guard: omit dhcp_guarding from config while changing an
			// unrelated attribute (domain_name) so a real Update fires; the previously
			// enabled value must be preserved (Optional+Computed), not reset to false.
			{
				Config:           testAccNetworkConfigDHCPGuarding(name, subnet, vlan, "bar.local", false, false),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_network.test", plancheck.ResourceActionUpdate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "domain_name", "bar.local"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_guarding", "true"),
					// The companion trusted-server list must be preserved too, otherwise
					// the inherited guarding=true would round-trip to MissingIPAddress.
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_guarding_trusted_servers.#", "1"),
				),
			},
			pt.ImportStep("unifi_network.test"),
			// 4. Disable path (hard gate): explicit false must be honored.
			{
				Config: testAccNetworkConfigDHCPGuarding(name, subnet, vlan, "bar.local", true, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_guarding", "false"),
				),
			},
			pt.ImportStep("unifi_network.test"),
		},
	})
}

// TestAccNetwork_dhcpGuardingVlanOnly covers dhcp_guarding on a non-corporate
// (vlan-only / L2-only) purpose. A vlan-only network runs no DHCP server, so
// whether the controller honors an *enabled* value is controller-dependent and
// unverified — this test deliberately does NOT assert that true is persisted.
// Instead it asserts the deterministic round-trip we can rely on: the provider
// already serializes dhcpguard_enabled on every network PUT regardless of
// purpose (the very mechanism behind issue #123), so an explicit false is
// accepted and read back as false (and an ignored field also reads back false),
// giving a stable state with no perpetual diff and a clean import-verify.
func TestAccNetwork_dhcpGuardingVlanOnly(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	_, vlan := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkVlanOnlyDHCPGuarding(name, vlan),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "dhcp_guarding", "false"),
				),
			},
			{
				ResourceName:      "unifi_network.test",
				ImportState:       true,
				ImportStateIdFunc: pt.SiteAndIDImportStateIDFunc("unifi_network.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetwork_vlanOnly(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	_, vlan := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkVlanOnly(name, vlan),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "vlan_id", strconv.Itoa(vlan)),
				),
			},
			{
				ResourceName:      "unifi_network.test",
				ImportState:       true,
				ImportStateIdFunc: pt.SiteAndIDImportStateIDFunc("unifi_network.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetwork_mdns(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	subnet, vlan := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: "> 7.0",
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkConfigMDNS(name, subnet, vlan, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "multicast_dns", "true"),
				),
			},
			pt.ImportStep("unifi_network.test"),
			{
				Config: testAccNetworkConfigMDNS(name, subnet, vlan, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "multicast_dns", "false"),
				),
			},
			pt.ImportStep("unifi_network.test"),
		},
	})
}

// TODO: ipv6 prefix delegation test

func quoteStrings(src []string) []string {
	dst := make([]string, 0, len(src))
	for _, s := range src {
		dst = append(dst, fmt.Sprintf("%q", s))
	}
	return dst
}

func testAccNetworkConfigDHCPBoot(name string, subnet *net.IPNet, vlan int) string {
	return fmt.Sprintf(`
locals {
	subnet  = "%[2]s"
	vlan_id = %[3]d
}

resource "unifi_network" "test" {
	name     = "%[1]s"
	purpose = "corporate"

	subnet        = local.subnet
	vlan_id       = local.vlan_id
	dhcp_start    = cidrhost(local.subnet, 6)
	dhcp_stop     = cidrhost(local.subnet, 254)
	dhcp_enabled  = true
	domain_name   = "foo.local"

	dhcpd_boot_enabled  = true
	dhcpd_boot_server   = "192.168.1.180"
	dhcpd_boot_filename = "test.boot"

	dhcp_dns = ["192.168.1.101", "192.168.1.102"]
}
`, name, subnet, vlan)
}

func testAccNetworkConfig(name string, subnet *net.IPNet, vlan int, igmpSnoop bool, dhcpDNS []string) string {
	return fmt.Sprintf(`
locals {
	subnet  = "%[2]s"
	vlan_id = %[3]d
}

resource "unifi_network" "test" {
	name    = "%[1]s"
	purpose = "corporate"

	subnet        = local.subnet
	vlan_id       = local.vlan_id
	dhcp_start    = cidrhost(local.subnet, 6)
	dhcp_stop     = cidrhost(local.subnet, 254)
	dhcp_enabled  = true
	domain_name   = "foo.local"
	igmp_snooping = %[4]t

	dhcp_dns = [%[5]s]
}
`, name, subnet, vlan, igmpSnoop, strings.Join(quoteStrings(dhcpDNS), ","))
}

func testAccNetworkConfigV6(name string, subnet *net.IPNet, vlan int, ipv6Type string, ipv6Subnet string) string {
	return fmt.Sprintf(`
locals {
	subnet  = "%[2]s"
	vlan_id = %[3]d
}
	
resource "unifi_network" "test" {
	name    = "%[1]s"
	purpose = "corporate"

	subnet        = local.subnet
	vlan_id       = local.vlan_id
	dhcp_start    = cidrhost(local.subnet, 6)
	dhcp_stop     = cidrhost(local.subnet, 254)
	dhcp_enabled  = true
	domain_name   = "foo.local"

	ipv6_interface_type = "%[4]s"
	ipv6_static_subnet  = "%[5]s"
	ipv6_ra_enable      = true
}
`, name, subnet, vlan, ipv6Type, ipv6Subnet)
}

func testWanNetworkConfig(name string, networkGroup string, wanType string, wanIP string, wanEgressQOS int, wanUsername string, wanPassword string, wanDNS1 string, wanDNS2 string) string {
	return fmt.Sprintf(`
resource "unifi_network" "wan_test" {
	name             = "%s"
	purpose          = "wan"
	wan_networkgroup = "%s"
	wan_type         = "%s"
	wan_ip           = "%s"
	wan_egress_qos   = %d
	wan_username     = "%s"
	x_wan_password   = "%s"

	wan_dns = ["%s", "%s"]
}

output "wan_dns1" {
	value = unifi_network.wan_test.wan_dns[0]
}

output "wan_dns2" {
	value = unifi_network.wan_test.wan_dns[1]
}
`, name, networkGroup, wanType, wanIP, wanEgressQOS, wanUsername, wanPassword, wanDNS1, wanDNS2)
}

func testWanV6NetworkConfig(name string, wanTypeV6 string, wanDhcpV6PdSize int) string {
	return fmt.Sprintf(`
resource "unifi_network" "wan_test" {
	name             = "%s"
	purpose          = "wan"
	wan_networkgroup = "WAN"
	wan_type         = "pppoe"
	wan_ip           = "192.168.1.1"
	wan_egress_qos   = 1
	wan_username     = "username"
	x_wan_password   = "password"

	wan_dns = ["8.8.8.8", "4.4.4.4"]

	wan_type_v6 = "%s"
	wan_dhcp_v6_pd_size = %d
}
`, name, wanTypeV6, wanDhcpV6PdSize)
}

func testAccNetworkWithSiteConfig(name string, subnet *net.IPNet, vlan int) string {
	return fmt.Sprintf(`
locals {
	subnet  = "%[2]s"
	vlan_id = %[3]d
}

resource "unifi_site" "test" {
  description = "%[1]s"
}

resource "unifi_network" "test" {
	site    = unifi_site.test.name
	name    = "%[1]s"
	purpose = "corporate"

	subnet        = local.subnet
	vlan_id       = local.vlan_id
	dhcp_start    = cidrhost(local.subnet, 6)
	dhcp_stop     = cidrhost(local.subnet, 254)
	dhcp_enabled  = true
	domain_name   = "foo.local"
	igmp_snooping = true
}
`, name, subnet, vlan)
}

func testAccNetworkWithDuplicateNames(subnet1 *net.IPNet, vlan1 int, subnet2 *net.IPNet, vlan2 int, networkName string) string {
	return fmt.Sprintf(`
locals {
	subnet1  = "%[1]s"
	vlan_id1 = %[2]d
	subnet2  = "%[3]s"
	vlan_id2 = %[4]d
}

resource "unifi_network" "test1" {
	name    = "%[5]s"
	purpose = "corporate"

	subnet  = local.subnet1
	vlan_id = local.vlan_id1
}

resource "unifi_network" "test2" {
	name    = "%[5]s"
	purpose = "corporate"

	subnet  = local.subnet2
	vlan_id = local.vlan_id2
}
`, subnet1, vlan1, subnet2, vlan2, networkName)
}

func testAccNetworkConfigDHCPRelay(name string, subnet *net.IPNet, vlan int, dhcpRelay bool) string {
	return fmt.Sprintf(`
locals {
	subnet  = "%[2]s"
	vlan_id = %[3]d
}

resource "unifi_network" "test" {
	name    = "%[1]s"
	purpose = "corporate"

	subnet      = local.subnet
	vlan_id     = local.vlan_id
	domain_name = "foo.local"
	
	dhcp_relay_enabled = %[4]t
}
`, name, subnet, vlan, dhcpRelay)
}

// testAccNetworkConfigDHCPGuarding renders a corporate network. When guardingSet is
// false the dhcp_guarding attribute is omitted entirely (exercising the
// Optional+Computed inherit-from-controller path that protects issue #123).
func testAccNetworkConfigDHCPGuarding(name string, subnet *net.IPNet, vlan int, domainName string, guardingSet bool, guarding bool) string {
	guardingLine := ""
	if guardingSet {
		guardingLine = fmt.Sprintf("dhcp_guarding = %t", guarding)
		if guarding {
			// DHCP Guarding requires at least one trusted DHCP server IP; the
			// network's own gateway (first host of the subnet) is the natural one.
			guardingLine += "\n\tdhcp_guarding_trusted_servers = [cidrhost(local.subnet, 1)]"
		}
	}
	return fmt.Sprintf(`
locals {
	subnet  = "%[2]s"
	vlan_id = %[3]d
}

resource "unifi_network" "test" {
	name    = "%[1]s"
	purpose = "corporate"

	subnet      = local.subnet
	vlan_id     = local.vlan_id
	domain_name = "%[4]s"

	%[5]s
}
`, name, subnet, vlan, domainName, guardingLine)
}

// testAccNetworkVlanOnlyDHCPGuarding renders a vlan-only (L2-only) network with
// dhcp_guarding explicitly disabled. false is used deliberately: it is the value
// the provider already sends for such networks today, so the round-trip is
// deterministic regardless of how the controller treats DHCP guarding on a
// purpose that runs no DHCP server.
func testAccNetworkVlanOnlyDHCPGuarding(name string, vlan int) string {
	return fmt.Sprintf(`
resource "unifi_site" "test" {
  description = "%[1]s"
}

resource "unifi_network" "test" {
  site          = unifi_site.test.name
  name          = "test"
  purpose       = "vlan-only"
  vlan_id       = %[2]d
  dhcp_guarding = false
}
`, name, vlan)
}

func testAccNetworkVlanOnly(name string, vlan int) string {
	return fmt.Sprintf(`
resource "unifi_site" "test" {
  description = "%[1]s"
}

resource "unifi_network" "test" {
  site    = unifi_site.test.name
  name    = "test"
  purpose = "vlan-only"
  vlan_id = %[2]d
}
`, name, vlan)
}

func testAccNetworkConfigDhcpV6(name string, subnet *net.IPNet, vlan int, gatewayIP string, dhcpdV6Start string, dhcpdV6Stop string, dhcpV6DNS []string) string {
	return fmt.Sprintf(`
locals {
  subnet  = "%[2]s"
	vlan_id = %[3]d
}

resource "unifi_network" "test" {
	name    = "%[1]s"
	purpose = "corporate"

	subnet        = local.subnet
	vlan_id       = local.vlan_id

	ipv6_static_subnet  = "%[4]s"

	dhcp_v6_dns_auto = false
	dhcp_v6_dns = [%[7]s]
	dhcp_v6_enabled = true
	dhcp_v6_start = "%[5]s"
	dhcp_v6_stop = "%[6]s"
	dhcp_v6_lease = 12 * 60 * 60
}
`, name, subnet, vlan, gatewayIP, dhcpdV6Start, dhcpdV6Stop, strings.Join(quoteStrings(dhcpV6DNS), ","))
}

func testAccNetworkConfigMDNS(name string, subnet *net.IPNet, vlan int, mdns bool) string {
	return fmt.Sprintf(`
resource "unifi_network" "test" {
	name    = "%[1]s"
	purpose = "corporate"
	subnet  = "%[2]s"
	vlan_id = %[3]d

	multicast_dns = %[4]t
}
`, name, subnet, vlan, mdns)
}

// TestAccNetwork_defaultGateway exercises the DHCP default-gateway override (issue
// #120): create with the override enabled + an in-subnet gateway IP, import
// round-trip (including the two new attributes), a perpetual-diff guard, an in-place
// update of the gateway IP, and the disable path. The gateway IP is computed from the
// allocated test subnet rather than hardcoded so the assertion settles polarity (an
// enabled override stores the exact IP) against the live controller.
func TestAccNetwork_defaultGateway(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	subnet, vlan := pt.GetTestVLAN(t)
	gw1 := mustHost(t, subnet, 100)
	gw2 := mustHost(t, subnet, 150)

	AcceptanceTest(t, AcceptanceTestCase{
		CheckDestroy: testAccCheckNetworkDestroy,
		Steps: []resource.TestStep{
			// 1. Create with the override enabled + an in-subnet gateway IP.
			{
				Config: testAccNetworkConfigDefaultGateway(name, subnet, vlan, true, gw1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "dhcpd_gateway_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcpd_gateway", gw1),
				),
			},
			// 2. Import round-trip including the two new fields.
			pt.ImportStep("unifi_network.test"),
			// 3. Perpetual-diff guard: re-apply the identical config, expect an empty plan
			//    (Optional+Computed must absorb any controller echo in either field).
			{
				Config:   testAccNetworkConfigDefaultGateway(name, subnet, vlan, true, gw1),
				PlanOnly: true,
			},
			// 4. In-place update: change the gateway IP, assert it applies.
			{
				Config: testAccNetworkConfigDefaultGateway(name, subnet, vlan, true, gw2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "dhcpd_gateway_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_network.test", "dhcpd_gateway", gw2),
				),
			},
			pt.ImportStep("unifi_network.test"),
			// 5. Disable: flip the override off (gateway omitted, inherited), assert false.
			{
				Config: testAccNetworkConfigDefaultGateway(name, subnet, vlan, false, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.test", "dhcpd_gateway_enabled", "false"),
				),
			},
			pt.ImportStep("unifi_network.test"),
		},
	})
}

// TestAccNetwork_defaultGatewayValidation covers the plan-time guards for the override:
// the IPv4 schema validator and the two cross-field rules. Each step expects an error,
// so nothing is created.
func TestAccNetwork_defaultGatewayValidation(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	subnet, vlan := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		Steps: []resource.TestStep{
			// Non-IPv4 gateway → schema IsIPv4Address validator.
			{
				Config:      testAccNetworkConfigDefaultGateway(name, subnet, vlan, true, "not-an-ip"),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("expected dhcpd_gateway to contain a valid IPv4 address")),
			},
			// Gateway set while the override is explicitly disabled → cross-field error.
			{
				Config:      testAccNetworkConfigDefaultGateway(name, subnet, vlan, false, "10.0.0.1"),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta(`"dhcpd_gateway_enabled" must be true`)),
			},
			// Override enabled with no gateway IP → cross-field error.
			{
				Config:      testAccNetworkConfigDefaultGateway(name, subnet, vlan, true, ""),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta(`"dhcpd_gateway" is required`)),
			},
		},
	})
}

// TestAccNetwork_disappears deletes the network out-of-band and asserts the next plan
// is a non-empty re-create. It also wires up testAccCheckNetworkDestroy, the first
// CheckDestroy in the network suite (existing tests carry `// TODO: CheckDestroy`).
func TestAccNetwork_disappears(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	subnet, vlan := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		CheckDestroy: testAccCheckNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkConfig(name, subnet, vlan, true, nil),
			},
			{
				// Delete the network behind Terraform's back, then plan with the same
				// config: the start-of-step refresh drops it from state and the plan
				// must show a re-create.
				PreConfig: func() {
					ctx := context.Background()
					id := mustNetworkIDByName(t, "default", name)
					if err := testClient.DeleteNetwork(ctx, "default", id); err != nil {
						t.Fatalf("out-of-band DeleteNetwork: %s", err)
					}
				},
				Config:             testAccNetworkConfig(name, subnet, vlan, true, nil),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// testAccNetworkConfigDefaultGateway renders a corporate DHCP network with the
// default-gateway override. When gateway is "" the dhcpd_gateway attribute is omitted
// entirely (exercising the inherit-from-controller / disable paths).
func testAccNetworkConfigDefaultGateway(name string, subnet *net.IPNet, vlan int, enabled bool, gateway string) string {
	gatewayLine := ""
	if gateway != "" {
		gatewayLine = fmt.Sprintf("dhcpd_gateway = %q", gateway)
	}
	return fmt.Sprintf(`
locals {
	subnet  = "%[2]s"
	vlan_id = %[3]d
}

resource "unifi_network" "test" {
	name    = "%[1]s"
	purpose = "corporate"

	subnet       = local.subnet
	vlan_id      = local.vlan_id
	dhcp_start   = cidrhost(local.subnet, 6)
	dhcp_stop    = cidrhost(local.subnet, 254)
	dhcp_enabled = true

	dhcpd_gateway_enabled = %[4]t
	%[5]s
}
`, name, subnet, vlan, enabled, gatewayLine)
}

// mustHost returns the Nth host address of subnet as a string, failing the test on
// error. Used to compute an in-subnet gateway IP from the allocated test subnet
// instead of hardcoding one.
func mustHost(t *testing.T, subnet *net.IPNet, n int) string {
	t.Helper()
	ip, err := cidr.Host(subnet, n)
	if err != nil {
		t.Fatalf("cidr.Host(%s, %d): %s", subnet, n, err)
	}
	return ip.String()
}

// testAccCheckNetworkDestroy asserts every unifi_network in state is gone from the
// controller. Mirrors testAccCheckDNSRecordDestroy; networks created on a non-default
// site read back as not-found on "default", which also counts as destroyed.
func testAccCheckNetworkDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "unifi_network" {
			continue
		}
		_, err := testClient.GetNetwork(context.Background(), "default", rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("network %s still exists", rs.Primary.ID)
		}
		// A 404 / not-found means the network was deleted.
		if strings.Contains(err.Error(), "404") || strings.Contains(strings.ToLower(err.Error()), "not found") {
			continue
		}
		return err
	}
	return nil
}

func TestAccNetwork_wireguardVPNClient(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")

	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testWireguardVPNClientNetworkConfig(name, "203.0.113.10", 51820, "0WvUlUyZZ0yTUibNCAdBrQ6XJd+8V37zmk/j8y/V9g4="),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.wg_test", "purpose", "vpn-client"),
					resource.TestCheckResourceAttr("unifi_network.wg_test", "vpn_type", "wireguard-client"),
					resource.TestCheckResourceAttr("unifi_network.wg_test", "wireguard_interface", "wan"),
					resource.TestCheckResourceAttr("unifi_network.wg_test", "wireguard_client_peer_ip", "203.0.113.10"),
					resource.TestCheckResourceAttr("unifi_network.wg_test", "wireguard_client_peer_port", "51820"),
					resource.TestCheckResourceAttr("unifi_network.wg_test", "vpn_client_default_route", "false"),
					resource.TestCheckResourceAttr("unifi_network.wg_test", "uid_vpn_custom_routing.0", "10.200.0.0/24"),
					resource.TestCheckResourceAttr("unifi_network.wg_test", "uid_vpn_custom_routing.1", "10.0.0.0/16"),
					// The tunnel interface address must round-trip exactly (no LAN +1 offset).
					resource.TestCheckResourceAttr("unifi_network.wg_test", "subnet", "10.255.255.2/32"),
					// The provider derives the gateway's public key from the generated private
					// key (the controller returns null), so it must be present.
					resource.TestCheckResourceAttrSet("unifi_network.wg_test", "wireguard_public_key"),
				),
			},
			// Sensitive, controller-managed secrets are not guaranteed to be returned on read.
			pt.ImportStep("unifi_network.wg_test", "x_wireguard_private_key", "wireguard_client_preshared_key"),
		},
	})
}

// --- Zone-Based Firewall (issue #94) -----------------------------------------
//
// These exercise the new unifi_network.firewall_zone_id attribute. They require a
// real UniFi OS 9.x controller with Zone-Based Firewall enabled and therefore skip
// in the Dockerized make testacc run (the harness "does not support firewall zones
// yet"), mirroring acctest/resource_firewall_zone_test.go. Existing unifi_network
// tests carry `// TODO: CheckDestroy` placeholders; these do not regress that — the
// zone resource is destroy-checked via testAccCheckFirewallZoneDestroy where a zone
// is created.
//
// CAVEAT (must be validated manually on a live 9.x ZBF controller before relying on
// these): whether the controller honors Network.FirewallZoneID on POST/PUT at all,
// and whether a full-object PUT with firewall_zone_id omitted PRESERVES vs CLEARS the
// existing zone (TestAccNetwork_unsetZoneDoesNotClobber guards the latter assumption).

// TestAccNetwork_explicitFirewallZone creates a zone, pins a network to it via
// firewall_zone_id, and asserts it sticks, imports cleanly, and that the zone matrix
// endpoint does not 500 (the exact UI-blank failure path from #94).
func TestAccNetwork_explicitFirewallZone(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc")
	zoneName := acctest.RandomWithPrefix("tfacc-zone")
	subnet, vlan := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZoneLock,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkExplicitZoneConfig(name, subnet.String(), vlan, zoneName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_network.test", "firewall_zone_id"),
					resource.TestCheckResourceAttrPair("unifi_network.test", "firewall_zone_id", "unifi_firewall_zone.test", "id"),
					testAccCheckFirewallZoneMatrixNoError("default"),
				),
			},
			pt.ImportStep("unifi_network.test"),
		},
		CheckDestroy: testAccCheckFirewallZoneDestroy,
	})
}

// TestAccNetwork_moveBetweenZones flips a network's firewall_zone_id from zone A to
// zone B and asserts the change applies and persists.
func TestAccNetwork_moveBetweenZones(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc")
	zoneA := acctest.RandomWithPrefix("tfacc-zonea")
	zoneB := acctest.RandomWithPrefix("tfacc-zoneb")
	subnet, vlan := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZoneLock,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkTwoZonesConfig(name, subnet.String(), vlan, zoneA, zoneB, "unifi_firewall_zone.a.id"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("unifi_network.test", "firewall_zone_id", "unifi_firewall_zone.a", "id"),
				),
			},
			pt.ImportStep("unifi_network.test"),
			{
				Config: testAccNetworkTwoZonesConfig(name, subnet.String(), vlan, zoneA, zoneB, "unifi_firewall_zone.b.id"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("unifi_network.test", "firewall_zone_id", "unifi_firewall_zone.b", "id"),
				),
			},
			pt.ImportStep("unifi_network.test"),
		},
		CheckDestroy: testAccCheckFirewallZoneDestroy,
	})
}

// TestAccNetwork_externalDrift moves a network's zone out-of-band and asserts the
// next refresh surfaces a non-empty plan — proving the primary value of the fix
// (drift visibility) works.
func TestAccNetwork_externalDrift(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc")
	zoneA := acctest.RandomWithPrefix("tfacc-zonea")
	zoneB := acctest.RandomWithPrefix("tfacc-zoneb")
	subnet, vlan := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZoneLock,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkTwoZonesConfig(name, subnet.String(), vlan, zoneA, zoneB, "unifi_firewall_zone.a.id"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("unifi_network.test", "firewall_zone_id", "unifi_firewall_zone.a", "id"),
				),
			},
			{
				// Move the network to zone B behind Terraform's back, then plan with the
				// original (zone A) config. Refresh must detect the drift.
				PreConfig: func() {
					ctx := context.Background()
					netID := mustNetworkIDByName(t, "default", name)
					zoneBID := mustFirewallZoneIDByName(t, "default", zoneB)
					n, err := testClient.GetNetwork(ctx, "default", netID)
					if err != nil {
						t.Fatalf("GetNetwork: %s", err)
					}
					n.FirewallZoneID = zoneBID
					if _, err := testClient.UpdateNetwork(ctx, "default", n); err != nil {
						t.Fatalf("out-of-band UpdateNetwork: %s", err)
					}
				},
				Config:             testAccNetworkTwoZonesConfig(name, subnet.String(), vlan, zoneA, zoneB, "unifi_firewall_zone.a.id"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
		CheckDestroy: testAccCheckFirewallZoneDestroy,
	})
}

// TestAccNetwork_unsetZoneDoesNotClobber assigns a network to a zone from the zone
// side (unifi_firewall_zone.networks) while leaving firewall_zone_id unset on the
// network, and asserts the apply produces no diff and preserves membership. This
// guards the dueling-writes resolution AND the PUT-omit-preserves-zone assumption:
// the network resource must not send firewall_zone_id when unconfigured.
func TestAccNetwork_unsetZoneDoesNotClobber(t *testing.T) {
	pt.SkipIfEnvLocalMissing(t, "Skipping, because test environment does not support firewall zones yet")
	name := acctest.RandomWithPrefix("tfacc")
	zoneName := acctest.RandomWithPrefix("tfacc-zone")
	subnet, vlan := pt.GetTestVLAN(t)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              firewallZoneLock,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkZoneFromZoneSideConfig(name, subnet.String(), vlan, zoneName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_zone.test", "networks.#", "1"),
					// NOTE: we deliberately do NOT assert unifi_network.firewall_zone_id here.
					// Because the zone references the network's id, Terraform creates the
					// network FIRST (read back while no zone exists yet -> firewall_zone_id is
					// empty or a controller default) and only then creates the zone that
					// assigns membership. The network is not re-read within the same apply, so
					// the zone-side assignment is not yet visible in the network's post-apply
					// state. It surfaces at the start-of-step refresh in the next step, where
					// the pair check below correctly passes.
				),
			},
			// Re-apply the identical config. The start-of-step refresh first reads the
			// network back (now surfacing the zone-side assignment via the computed
			// read-back), and the framework then fails on a non-empty post-apply plan — so a
			// clean second apply proves the network's omitted firewall_zone_id does not fight
			// the zone-side assignment.
			{
				Config: testAccNetworkZoneFromZoneSideConfig(name, subnet.String(), vlan, zoneName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("unifi_network.test", "firewall_zone_id", "unifi_firewall_zone.test", "id"),
				),
			},
		},
		CheckDestroy: testAccCheckFirewallZoneDestroy,
	})
}

func testAccNetworkExplicitZoneConfig(name, subnet string, vlan int, zoneName string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone" "test" {
	name = %[4]q
	# networks intentionally omitted — managed from the network side below.
}

resource "unifi_network" "test" {
	name    = %[1]q
	purpose = "corporate"
	subnet  = %[2]q
	vlan_id = %[3]d

	# Manage the zone association from the network side only; the zone above
	# intentionally does not list this network (the two representations are
	# mutually exclusive — see the example).
	firewall_zone_id = unifi_firewall_zone.test.id
}
`, name, subnet, vlan, zoneName)
}

func testAccNetworkTwoZonesConfig(name, subnet string, vlan int, zoneA, zoneB, selected string) string {
	return fmt.Sprintf(`
resource "unifi_firewall_zone" "a" {
	name = %[4]q
	# networks intentionally omitted — managed from the network side below.
}

resource "unifi_firewall_zone" "b" {
	name = %[5]q
	# networks intentionally omitted — managed from the network side below.
}

resource "unifi_network" "test" {
	name    = %[1]q
	purpose = "corporate"
	subnet  = %[2]q
	vlan_id = %[3]d

	firewall_zone_id = %[6]s
}
`, name, subnet, vlan, zoneA, zoneB, selected)
}

func testAccNetworkZoneFromZoneSideConfig(name, subnet string, vlan int, zoneName string) string {
	return fmt.Sprintf(`
resource "unifi_network" "test" {
	name    = %[1]q
	purpose = "corporate"
	subnet  = %[2]q
	vlan_id = %[3]d

	# firewall_zone_id intentionally unset — membership is managed from the zone side.
}

resource "unifi_firewall_zone" "test" {
	name     = %[4]q
	networks = [unifi_network.test.id]
}
`, name, subnet, vlan, zoneName)
}

// testAccCheckFirewallZoneMatrixNoError reproduces the #94 failure path: an unzoned
// network made the zone-matrix endpoint return HTTP 500, blanking the ZBF Rules UI.
// A successful (non-error) list proves explicit assignment repairs it. The SDK
// surfaces an error, not the raw HTTP status.
func testAccCheckFirewallZoneMatrixNoError(site string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if _, err := testClient.ListFirewallZoneMatrix(context.Background(), site); err != nil {
			return fmt.Errorf("ListFirewallZoneMatrix(%q) returned an error (reproduces the #94 zone-matrix 500): %w", site, err)
		}
		return nil
	}
}

func mustNetworkIDByName(t *testing.T, site, name string) string {
	t.Helper()
	networks, err := testClient.ListNetwork(context.Background(), site)
	if err != nil {
		t.Fatalf("ListNetwork: %s", err)
	}
	for _, n := range networks {
		if n.Name == name {
			return n.ID
		}
	}
	t.Fatalf("no network found with name %q", name)
	return ""
}

func mustFirewallZoneIDByName(t *testing.T, site, name string) string {
	t.Helper()
	zones, err := testClient.ListFirewallZone(context.Background(), site)
	if err != nil {
		t.Fatalf("ListFirewallZone: %s", err)
	}
	for _, z := range zones {
		if z.Name == name {
			return z.ID
		}
	}
	t.Fatalf("no firewall zone found with name %q", name)
	return ""
}

func testWireguardVPNClientNetworkConfig(name string, peerIP string, peerPort int, peerPublicKey string) string {
	return fmt.Sprintf(`
resource "unifi_network" "wg_test" {
	name     = "%[1]s"
	purpose  = "vpn-client"
	vpn_type = "wireguard-client"

	# The controller requires a tunnel interface address and interface DNS for a
	# WireGuard VPN client; it rejects the create otherwise.
	subnet   = "10.255.255.2/32"
	dhcp_dns = ["1.1.1.1"]

	wireguard_interface              = "wan"
	wireguard_client_mode            = "manual"
	wireguard_client_peer_ip         = "%[2]s"
	wireguard_client_peer_port       = %[3]d
	wireguard_client_peer_public_key = "%[4]s"

	vpn_client_default_route = false
	uid_vpn_custom_routing   = ["10.200.0.0/24", "10.0.0.0/16"]
}
`, name, peerIP, peerPort, peerPublicKey)
}

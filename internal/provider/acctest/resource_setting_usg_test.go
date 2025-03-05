package acctest

import (
	"fmt"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"regexp"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// using an additional lock to the one around the resource to avoid deadlocking accidentally
var settingUsgLock = sync.Mutex{}

func TestAccSettingUsg_mdns_v6(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: "< 7",
		Lock:              &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgConfig_mdns(true),
				Check:  resource.ComposeTestCheckFunc(),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgConfig_mdns(false),
				Check:  resource.ComposeTestCheckFunc(),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgConfig_mdns(true),
				Check:  resource.ComposeTestCheckFunc(),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
		},
	})
}

func TestAccSettingUsg_mdns_v7(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 7",
		Lock:              &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config:      testAccSettingUsgConfig_mdns(true),
				ExpectError: regexp.MustCompile("multicast_dns_enabled is not supported"),
			},
		},
	})
}

func TestAccSettingUsg_dhcpRelay(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgConfig_dhcpRelay(),
				Check:  resource.ComposeTestCheckFunc(),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
		},
	})
}

func TestAccSettingUsg_site(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgConfig_site(),
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				ResourceName:      "unifi_setting_usg.test",
				ImportState:       true,
				ImportStateIdFunc: pt.SiteAndIDImportStateIDFunc("unifi_setting_usg.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSettingUsg_geoIpFiltering(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgConfig_geoIpFilteringBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.block", "block"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.traffic_direction", "both"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.#", "3"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "RU"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "CN"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "KP"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgConfig_geoIpFilteringAllow(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.block", "allow"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.traffic_direction", "both"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.#", "3"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "US"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "CA"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "GB"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgConfig_geoIpFilteringDirections(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.block", "block"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.traffic_direction", "ingress"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.#", "2"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "RU"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "CN"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgConfig_geoIpFilteringDisabled(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.enabled", "false"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgConfig_geoIpFilteringBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.block", "block"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.traffic_direction", "both"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.#", "3"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "RU"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "CN"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_usg.test", "geo_ip_filtering.countries.*", "KP"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
		},
	})
}

func TestAccSettingUsg_upnp(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingUsgLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgConfig_upnpBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "upnp.enabled", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgConfig_upnpAdvanced(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "upnp.enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "upnp.nat_pmp_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "upnp.secure_mode", "true"),
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "upnp.wan_interface", "WAN"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgConfig_upnpDisabled(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_usg.test", "upnp.enabled", "false"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_usg.test"),
		},
	})
}

func testAccSettingUsgConfig_mdns(mdns bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_usg" "test" {
	multicast_dns_enabled = %t
}
`, mdns)
}

func testAccSettingUsgConfig_dhcpRelay() string {
	return `
resource "unifi_setting_usg" "test" {
	dhcp_relay_servers = [
		"10.1.2.3",
		"10.1.2.4",
	]
}
`
}

func testAccSettingUsgConfig_site() string {
	return `
resource "unifi_site" "test" {
	description = "test"
}

resource "unifi_setting_usg" "test" {
	site = unifi_site.test.name
}
`
}

func testAccSettingUsgConfig_geoIpFilteringBasic() string {
	return `
resource "unifi_setting_usg" "test" {
	geo_ip_filtering = {
		enabled = true
		countries = ["RU", "CN", "KP"]
	}
}
`
}

func testAccSettingUsgConfig_geoIpFilteringAllow() string {
	return `
resource "unifi_setting_usg" "test" {
	geo_ip_filtering = {
		enabled = true
		block = "allow"
		countries = ["US", "CA", "GB"]
	}
}
`
}

func testAccSettingUsgConfig_geoIpFilteringDirections() string {
	return `
resource "unifi_setting_usg" "test" {
	geo_ip_filtering = {
		enabled = true
		traffic_direction = "ingress"
		countries = ["RU", "CN"]
	}
}
`
}

func testAccSettingUsgConfig_geoIpFilteringDisabled() string {
	return `
resource "unifi_setting_usg" "test" {
	geo_ip_filtering = {
		enabled = false
	}
}
`
}

func testAccSettingUsgConfig_upnpBasic() string {
	return `
resource "unifi_setting_usg" "test" {
	upnp = {
		enabled = true
	}
}
`
}

func testAccSettingUsgConfig_upnpAdvanced() string {
	return `
resource "unifi_setting_usg" "test" {
	upnp = {
		enabled = true
		nat_pmp_enabled = true
		secure_mode = true
		wan_interface = "WAN"
	}
}
`
}

func testAccSettingUsgConfig_upnpDisabled() string {
	return `
resource "unifi_setting_usg" "test" {
	upnp = {
		enabled = false
	}
}
`
}

package acctest

import (
	"fmt"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"sync"
	"testing"
)

// Using dedicated lock for IPS settings to avoid interference with other tests
var settingIpsLock = &sync.Mutex{}

func TestAccSettingIps_basic(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 8.0",
		Lock:              settingIpsLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingIpsConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ips.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "ips_mode", "ips"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "enabled_networks.#", "1"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_ips.test", "enabled_networks.*", "LAN"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_ips.test", plancheck.ResourceActionCreate),
			},
			pt.ImportStepWithSite("unifi_setting_ips.test"),
			{
				Config: testAccSettingIpsConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ips.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "ips_mode", "ids"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "enabled_networks.#", "1"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_ips.test", "enabled_networks.*", "LAN"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_ips.test", plancheck.ResourceActionUpdate),
			},
			pt.ImportStepWithSite("unifi_setting_ips.test"),
		},
	})
}

func TestAccSettingIps_enabledCategories(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 8.0",
		Lock:              settingIpsLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingIpsConfig_enabledCategories(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ips.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "enabled_categories.#", "3"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_ips.test", "enabled_categories.*", "emerging-dos"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_ips.test", "enabled_categories.*", "emerging-exploit"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_ips.test", "enabled_categories.*", "emerging-malware"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_ips.test"),
			{
				Config: testAccSettingIpsConfig_enabledCategoriesUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ips.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "enabled_categories.#", "2"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_ips.test", "enabled_categories.*", "emerging-scan"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_ips.test", "enabled_categories.*", "emerging-worm"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_ips.test"),
		},
	})
}

func TestAccSettingIps_adBlocking(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 8.0",
		Lock:              settingIpsLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingIpsConfig_adBlocking(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ips.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "ad_blocked_networks.#", "2"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_ips.test", "ad_blocked_networks.*", "network1"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_ips.test", "ad_blocked_networks.*", "network2"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_ips.test"),
			{
				Config: testAccSettingIpsConfig_adBlockingUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ips.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "ad_blocked_networks.#", "1"),
					resource.TestCheckTypeSetElemAttr("unifi_setting_ips.test", "ad_blocked_networks.*", "network3"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_ips.test"),
		},
	})
}

func TestAccSettingIps_honeypot(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 8.0",
		Lock:              settingIpsLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingIpsConfig_honeypot(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ips.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "honeypots.#", "1"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "honeypots.0.ip_address", "192.168.1.10"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "honeypots.0.network_id", "network1"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_ips.test"),
			{
				Config: testAccSettingIpsConfig_honeypotUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ips.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "honeypots.#", "1"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "honeypots.0.ip_address", "192.168.2.20"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "honeypots.0.network_id", "network2"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_ips.test"),
			{
				Config: testAccSettingIpsConfig_honeypotDisabled(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ips.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "honeypots.#", "0"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_ips.test"),
		},
	})
}

func TestAccSettingIps_dnsFilters(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 8.0",
		Lock:              settingIpsLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingIpsConfig_dnsFilters(t),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ips.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "dns_filters.#", "1"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "dns_filters.0.name", "Test Filter"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "dns_filters.0.filter", "work"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "dns_filters.0.description", "Test description"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "dns_filters.0.allowed_sites.#", "2"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "dns_filters.0.blocked_sites.#", "2"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "dns_filters.0.blocked_tld.#", "1"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_ips.test"),
			{
				Config: testAccSettingIpsConfig_dnsFiltersUpdated(t),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ips.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "dns_filters.#", "2"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "dns_filters.0.name", "Test Filter Updated"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "dns_filters.0.filter", "family"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "dns_filters.1.name", "Second Filter"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "dns_filters.1.filter", "none"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_ips.test"),
		},
	})
}

func TestAccSettingIps_suppression(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 8.0",
		Lock:              settingIpsLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingIpsConfig_suppression(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ips.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "suppression.alerts.#", "1"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "suppression.alerts.0.category", "emerging-dos"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "suppression.alerts.0.signature", "Test Signature"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "suppression.alerts.0.type", "all"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "suppression.whitelist.#", "1"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "suppression.whitelist.0.direction", "src"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "suppression.whitelist.0.mode", "ip"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "suppression.whitelist.0.value", "192.168.1.100"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_ips.test"),
			{
				Config: testAccSettingIpsConfig_suppressionUpdated(t),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ips.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "suppression.alerts.#", "2"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "suppression.alerts.0.type", "track"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "suppression.alerts.0.tracking.#", "1"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "suppression.alerts.0.tracking.0.direction", "dest"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "suppression.alerts.0.tracking.0.mode", "subnet"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "suppression.alerts.0.tracking.0.value", "192.168.0.0/24"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "suppression.whitelist.#", "2"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_ips.test"),
		},
	})
}

func TestAccSettingIps_comprehensive(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 8.0",
		Lock:              settingIpsLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingIpsConfig_comprehensive(t),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ips.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "ips_mode", "ids"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "restrict_torrents", "true"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "advanced_filtering_preference", "manual"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "enabled_categories.#", "2"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "enabled_networks.#", "2"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "ad_blocked_networks.#", "1"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "honeypots.#", "1"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "dns_filters.#", "1"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "suppression.alerts.#", "1"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "suppression.whitelist.#", "1"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_ips.test"),
		},
	})
}

func TestAccSettingIps_comprehensiveBefore8(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: "< 8.0",
		MinVersion:        version.Must(version.NewVersion("7.4")),
		Lock:              settingIpsLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingIpsConfig_comprehensiveBefore8(t),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ips.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "ips_mode", "ids"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "restrict_torrents", "true"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "enabled_categories.#", "2"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "ad_blocked_networks.#", "1"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "honeypots.#", "1"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "dns_filters.#", "1"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "suppression.alerts.#", "1"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "suppression.whitelist.#", "1"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_ips.test"),
		},
	})
}

func TestAccSettingIps_restrictTorrents(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 8.0",
		Lock:              settingIpsLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingIpsConfig_restrictTorrents(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ips.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "restrict_torrents", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_ips.test"),
			{
				Config: testAccSettingIpsConfig_restrictTorrents(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ips.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "restrict_torrents", "false"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_ips.test"),
		},
	})
}

func TestAccSettingIps_memoryOptimized(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0",
		Lock:              settingIpsLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingIpsConfig_memoryOptimized(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ips.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "memory_optimized", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_ips.test"),
			{
				Config: testAccSettingIpsConfig_memoryOptimized(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ips.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ips.test", "memory_optimized", "false"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_ips.test"),
		},
	})
}

func testAccSettingIpsConfig_basic() string {
	return `
resource "unifi_setting_ips" "test" {
  ips_mode      = "ips"
  enabled_networks = ["LAN"]
}
`
}

func testAccSettingIpsConfig_updated() string {
	return `
resource "unifi_setting_ips" "test" {
  ips_mode      = "ids"
  enabled_networks = ["LAN"]
}
`
}

func testAccSettingIpsConfig_enabledCategories() string {
	return `
resource "unifi_setting_ips" "test" {
  ips_mode      = "ids"
  enabled_networks = ["network1"]
  enabled_categories = [
    "emerging-dos",
    "emerging-exploit",
    "emerging-malware"
  ]
}
`
}

func testAccSettingIpsConfig_enabledCategoriesUpdated() string {
	return `
resource "unifi_setting_ips" "test" {
  ips_mode      = "ids"
  enabled_networks = ["network1"]
  enabled_categories = [
    "emerging-scan",
    "emerging-worm",
  ]
}
`
}

func testAccSettingIpsConfig_adBlocking() string {
	return `
resource "unifi_setting_ips" "test" {
  ips_mode      = "ids"
  enabled_networks = ["network1"]
  ad_blocked_networks = [
    "network1",
    "network2"
  ]
}
`
}

func testAccSettingIpsConfig_adBlockingUpdated() string {
	return `
resource "unifi_setting_ips" "test" {
  ips_mode      = "ids"
  enabled_networks = ["network1"]
  ad_blocked_networks = [
    "network3"
  ]
}
`
}

func testAccSettingIpsConfig_honeypot() string {
	return `
resource "unifi_setting_ips" "test" {
  ips_mode      = "ids"
  enabled_networks = ["network1"]
  honeypots = [{
    ip_address = "192.168.1.10"
    network_id = "network1"
  }]
}
`
}

func testAccSettingIpsConfig_honeypotUpdated() string {
	return `
resource "unifi_setting_ips" "test" {
  ips_mode      = "ids"
  enabled_networks = ["network1"]
  honeypots = [{
    ip_address = "192.168.2.20"
    network_id = "network2"
  }]
}
`
}

func testAccSettingIpsConfig_honeypotDisabled() string {
	return `
resource "unifi_setting_ips" "test" {
  ips_mode      = "ids"
  enabled_networks = ["network1"]
  honeypots = []
}
`
}

func testAccSettingIpsConfig_dnsFilters(t *testing.T) string {
	subnet, vlanId := pt.GetTestVLAN(t)
	return fmt.Sprintf(`

resource "unifi_network" "test" {
  name = "Test"
  purpose = "corporate"
  subnet = %q
  vlan_id = %d
}

resource "unifi_setting_ips" "test" {
  ips_mode      = "ids"
  enabled_networks = ["network1"]
  dns_filters = [{
    name = "Test Filter"
    filter = "work"
    description = "Test description"
    network_id = unifi_network.test.id
    allowed_sites = [
      "example.com",
      "allowed.org"
    ]
    blocked_sites = [
      "blocked1.com",
      "blocked2.com"
    ]
    blocked_tld = [
      "xyz"
    ]
  }]
}
`, subnet.String(), vlanId)
}

func testAccSettingIpsConfig_dnsFiltersUpdated(t *testing.T) string {
	subnet, vlanId := pt.GetTestVLAN(t)
	subnet2, vlanId2 := pt.GetTestVLAN(t)
	return fmt.Sprintf(`

resource "unifi_network" "test" {
  name = "Test"
  purpose = "corporate"
  subnet = %q
  vlan_id = %d
}


resource "unifi_network" "test2" {
  name = "Test"
  purpose = "corporate"
  subnet = %q
  vlan_id = %d
}

resource "unifi_setting_ips" "test" {
  ips_mode      = "ids"
  enabled_networks = ["network1"]
  dns_filters = [
	{
      name = "Test Filter Updated"
      filter = "family"
      description = "Updated description"
      network_id = unifi_network.test.id
      allowed_sites = [
        "example.com",
        "allowed.org",
        "new-allowed.com"
      ]
      blocked_sites = [
        "blocked1.com"
      ]
    },
    {
      name = "Second Filter"
      filter = "none"
      network_id = unifi_network.test2.id
    }
  ]
}
`, subnet.String(), vlanId, subnet2.String(), vlanId2)
}

func testAccSettingIpsConfig_suppression() string {
	return `
resource "unifi_setting_ips" "test" {
  ips_mode      = "ids"
  enabled_networks = ["network1"]
  suppression = {
    alerts = [{
      category = "emerging-dos"
      signature = "Test Signature"
      type = "all"
    }]
    whitelist = [{
      direction = "src"
      mode = "ip"
      value = "192.168.1.100"
    }]
  }
}
`
}

func testAccSettingIpsConfig_suppressionUpdated(t *testing.T) string {
	subnet, vlanId := pt.GetTestVLAN(t)
	return fmt.Sprintf(`
resource "unifi_network" "test" {
	  name = "Test"
	  purpose = "corporate"
	  subnet = %q
	  vlan_id = %d
}

resource "unifi_setting_ips" "test" {
  ips_mode = "ids"
  enabled_networks = ["network1"]
  suppression = {
    alerts = [
	  {
        category = "emerging-dos"
        signature = "Test Signature"
        type = "track"
        tracking = [{
         direction = "dest"
         mode = "subnet"
         value = "192.168.0.0/24"
        }]
      },
      {
        category = "emerging-exploit"
        signature = "Another Signature"
        type = "track"
      }
	]
    whitelist = [
 	  {
        direction = "src"
        mode = "subnet"
        value = "192.168.1.0/24"
      },
      {
        direction = "both"
        mode = "network"
        value = unifi_network.test.id
      }
    ]
  }
}
`, subnet.String(), vlanId)
}

func testAccSettingIpsConfig_comprehensive(t *testing.T) string {
	subnet, vlanId := pt.GetTestVLAN(t)
	return fmt.Sprintf(`
resource "unifi_network" "test" {
	name = "Test"
	purpose = "corporate"
	subnet = %q
	vlan_id = %d
}

resource "unifi_setting_ips" "test" {
  ips_mode = "ids"
  restrict_torrents = true
  advanced_filtering_preference = "manual"
  
  enabled_categories = [
    "emerging-dos",
    "emerging-exploit"
  ]
  
  enabled_networks = [
    "network1",
    "network2"
  ]
  
  ad_blocked_networks = [
    "network1"
  ]
  
  honeypots = [{
    ip_address = "192.168.1.10"
    network_id = "network1"
  }]
  
  dns_filters = [{
    name = "Comprehensive Filter"
    filter = "work"
    description = "Comprehensive test filter"
    network_id = unifi_network.test.id
    allowed_sites = ["allowed.com"]
    blocked_sites = ["blocked.com"]
  }]
  
  suppression = {
    alerts = [{
      category = "emerging-dos"
      signature = "Test Signature"
      type = "all"
    }]
    whitelist = [{
      direction = "src"
      mode = "ip"
      value = "192.168.1.100"
    }]
  }
}
`, subnet.String(), vlanId)
}

func testAccSettingIpsConfig_comprehensiveBefore8(t *testing.T) string {
	subnet, vlanId := pt.GetTestVLAN(t)
	return fmt.Sprintf(`
resource "unifi_network" "test" {
	name = "Test"
	purpose = "corporate"
	subnet = %q
	vlan_id = %d
}

resource "unifi_setting_ips" "test" {
  ips_mode = "ids"
  restrict_torrents = true
  
  enabled_categories = [
    "emerging-dos",
    "emerging-exploit"
  ]
  
  ad_blocked_networks = [
    "network1"
  ]
  
  honeypots = [{
    ip_address = "192.168.1.10"
    network_id = "network1"
  }]
  
  dns_filters = [{
    name = "Comprehensive Filter"
    filter = "work"
    description = "Comprehensive test filter"
    network_id = unifi_network.test.id
    allowed_sites = ["allowed.com"]
    blocked_sites = ["blocked.com"]
  }]
  
  suppression = {
    alerts = [{
      category = "emerging-dos"
      signature = "Test Signature"
      type = "all"
    }]
    whitelist = [{
      direction = "src"
      mode = "ip"
      value = "192.168.1.100"
    }]
  }
}
`, subnet.String(), vlanId)
}

func testAccSettingIpsConfig_restrictTorrents(enabled bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_ips" "test" {
  ips_mode      = "ids"
  enabled_networks = ["network1"]
  restrict_torrents = %t
}
`, enabled)
}

func testAccSettingIpsConfig_memoryOptimized(enabled bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_ips" "test" {
  ips_mode      = "ids"
  enabled_networks = ["network1"]
  memory_optimized = %t
}
`, enabled)
}

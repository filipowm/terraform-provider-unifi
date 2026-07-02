package acctest

import (
	"sync"
	"testing"

	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var settingRadiusLock = sync.Mutex{}

func TestAccSettingRadius_basic(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingRadiusLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingRadiusConfigBasic(),
				Check:  resource.ComposeTestCheckFunc(),
			},
			pt.ImportStep("unifi_setting_radius.test"),
		},
	})
}

func TestAccSettingRadius_site(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingRadiusLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingRadiusConfigSite(),
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				ResourceName:      "unifi_setting_radius.test",
				ImportState:       true,
				ImportStateIdFunc: pt.SiteAndIDImportStateIDFunc("unifi_setting_radius.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSettingRadius_full(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingRadiusLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingRadiusConfigFull(),
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				ResourceName:      "unifi_setting_radius.test",
				ImportState:       true,
				ImportStateIdFunc: pt.SiteAndIDImportStateIDFunc("unifi_setting_radius.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSettingRadius_vlan(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingRadiusLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingRadiusConfigVlan(),
				Check:  resource.ComposeTestCheckFunc(),
			},
			pt.ImportStep("unifi_setting_radius.test"),
		},
	})
}

func testAccSettingRadiusConfigBasic() string {
	return `
resource "unifi_setting_radius" "test" {
	enabled = true
	secret = "securepw"
}
`
}

func testAccSettingRadiusConfigSite() string {
	return `
resource "unifi_site" "test" {
	description = "test"
}

resource "unifi_setting_radius" "test" {
	site = unifi_site.test.name
	enabled = true
	secret = "securepw"
}
`
}

func testAccSettingRadiusConfigFull() string {
	return `
resource "unifi_setting_radius" "test" {
	enabled = true
	secret = "securepw"
	accounting_port = "9999"
	auth_port = "8888"
}
`
}

func testAccSettingRadiusConfigVlan() string {
	return `
resource "unifi_setting_radius" "test" {
	enabled = true
	secret = "securepw"
	accounting_enabled = true
}
`
}

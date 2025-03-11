package acctest

import (
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"regexp"
	"sync"
	"testing"
)

var settingRsyslogdLock = &sync.Mutex{}

// TestAccSettingRsyslogdBasic tests the basic creation and import of the rsyslogd settings
func TestAccSettingRsyslogdBasic(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 8.5",
		Lock:              settingRsyslogdLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingRsyslogdConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_rsyslogd.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_rsyslogd.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_rsyslogd.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_rsyslogd.test", "ip", "192.168.1.100"),
					resource.TestCheckResourceAttr("unifi_setting_rsyslogd.test", "port", "514"),
					resource.TestCheckResourceAttr("unifi_setting_rsyslogd.test", "contents.#", "2"),
					resource.TestCheckResourceAttr("unifi_setting_rsyslogd.test", "contents.0", "device"),
					resource.TestCheckResourceAttr("unifi_setting_rsyslogd.test", "contents.1", "client"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_rsyslogd.test", plancheck.ResourceActionCreate),
			},
			pt.ImportStepWithSite("unifi_setting_rsyslogd.test"),
		},
	})
}

// TestAccSettingRsyslogdUpdate tests updating the rsyslogd settings with different values
func TestAccSettingRsyslogdUpdate(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 8.5",
		Lock:              settingRsyslogdLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingRsyslogdConfigBasic(),
			},
			{
				Config: testAccSettingRsyslogdConfigUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_rsyslogd.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_rsyslogd.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_rsyslogd.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_rsyslogd.test", "ip", "192.168.1.200"),
					resource.TestCheckResourceAttr("unifi_setting_rsyslogd.test", "port", "1514"),
					resource.TestCheckResourceAttr("unifi_setting_rsyslogd.test", "contents.#", "3"),
					resource.TestCheckResourceAttr("unifi_setting_rsyslogd.test", "contents.0", "device"),
					resource.TestCheckResourceAttr("unifi_setting_rsyslogd.test", "contents.1", "client"),
					resource.TestCheckResourceAttr("unifi_setting_rsyslogd.test", "contents.2", "admin_activity"),
					resource.TestCheckResourceAttr("unifi_setting_rsyslogd.test", "debug", "true"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_rsyslogd.test", plancheck.ResourceActionUpdate),
			},
		},
	})
}

// TestAccSettingRsyslogdDisable tests disabling the rsyslogd settings
func TestAccSettingRsyslogdDisable(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 8.5",
		Lock:              settingRsyslogdLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingRsyslogdConfigBasic(),
			},
			{
				Config: testAccSettingRsyslogdConfigDisabled(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_rsyslogd.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_rsyslogd.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_rsyslogd.test", "enabled", "false"),
					resource.TestCheckNoResourceAttr("unifi_setting_rsyslogd.test", "ip"),
					resource.TestCheckNoResourceAttr("unifi_setting_rsyslogd.test", "port"),
					resource.TestCheckNoResourceAttr("unifi_setting_rsyslogd.test", "contents.#"),
					resource.TestCheckNoResourceAttr("unifi_setting_rsyslogd.test", "debug"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_rsyslogd.test", plancheck.ResourceActionUpdate),
			},
		},
	})
}

// TestAccSettingRsyslogdReEnable tests re-enabling the rsyslogd settings with different values
func TestAccSettingRsyslogdReEnable(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 8.5",
		Lock:              settingRsyslogdLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingRsyslogdConfigDisabled(),
			},
			{
				Config: testAccSettingRsyslogdConfigReEnabled(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_rsyslogd.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_rsyslogd.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_rsyslogd.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_rsyslogd.test", "netconsole_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_rsyslogd.test", "netconsole_host", "192.168.1.150"),
					resource.TestCheckResourceAttr("unifi_setting_rsyslogd.test", "netconsole_port", "1514"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_rsyslogd.test", plancheck.ResourceActionUpdate),
			},
		},
	})
}

// TestAccSettingRsyslogdValidation tests validation errors when trying to set fields with rsyslogd disabled
func TestAccSettingRsyslogdValidation(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 8.5",
		Lock:              settingRsyslogdLock,
		Steps: []resource.TestStep{
			{
				Config:      testAccSettingRsyslogdConfigInvalid(),
				ExpectError: regexp.MustCompile(`any of those attributes must not be configured`),
			},
		},
	})
}

// TestAccSettingRsyslogdPortValidation tests validation errors for invalid port numbers
func TestAccSettingRsyslogdPortValidation(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 8.5",
		Lock:              settingRsyslogdLock,
		Steps: []resource.TestStep{
			{
				Config:      testAccSettingRsyslogdConfigInvalidPort(),
				ExpectError: regexp.MustCompile(`value must be between`),
			},
		},
	})
}

func testAccSettingRsyslogdConfigBasic() string {
	return `
resource "unifi_setting_rsyslogd" "test" {
  enabled  = true
  ip       = "192.168.1.100"
  port     = 514
  contents = ["device", "client"]
}
`
}

func testAccSettingRsyslogdConfigUpdate() string {
	return `
resource "unifi_setting_rsyslogd" "test" {
  enabled  = true
  ip       = "192.168.1.200"
  port     = 1514
  contents = ["device", "client", "admin_activity"]
  debug    = true
}
`
}

func testAccSettingRsyslogdConfigDisabled() string {
	return `
resource "unifi_setting_rsyslogd" "test" {
  enabled = false
}
`
}

func testAccSettingRsyslogdConfigReEnabled() string {
	return `
resource "unifi_setting_rsyslogd" "test" {
  enabled           = true
  contents          = ["device", "client"]
  ip       = "192.168.1.200"
  netconsole_enabled = true
  netconsole_host   = "192.168.1.150"
  netconsole_port   = 1514
}
`
}

func testAccSettingRsyslogdConfigInvalid() string {
	return `
resource "unifi_setting_rsyslogd" "test" {
  enabled = false
  ip      = "192.168.1.100"
}
`
}

func testAccSettingRsyslogdConfigInvalidPort() string {
	return `
resource "unifi_setting_rsyslogd" "test" {
  enabled = true
  ip      = "192.168.1.100"
  port    = 70000
}
`
}

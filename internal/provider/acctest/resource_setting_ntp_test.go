package acctest

import (
	"fmt"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"regexp"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var settingNtpLock = &sync.Mutex{}

func TestAccSettingNtp(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 7.3",
		Lock:              settingNtpLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingNtpModeOnly("auto"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ntp.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ntp.test", "site", "default"),
					resource.TestCheckResourceAttrSet("unifi_setting_ntp.test", "ntp_server_1"),
					resource.TestCheckResourceAttrSet("unifi_setting_ntp.test", "ntp_server_2"),
					resource.TestCheckResourceAttrSet("unifi_setting_ntp.test", "ntp_server_3"),
					resource.TestCheckResourceAttrSet("unifi_setting_ntp.test", "ntp_server_4"),
					resource.TestCheckResourceAttr("unifi_setting_ntp.test", "mode", "auto"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_ntp.test", plancheck.ResourceActionCreate),
			},
			pt.ImportStepWithSite("unifi_setting_ntp.test"),
			{
				Config: testAccSettingNtpConfig2Servers("time.google.com", "pool.ntp.org", "manual"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ntp.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ntp.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_ntp.test", "ntp_server_1", "time.google.com"),
					resource.TestCheckResourceAttr("unifi_setting_ntp.test", "ntp_server_2", "pool.ntp.org"),
					resource.TestCheckResourceAttrSet("unifi_setting_ntp.test", "ntp_server_3"),
					resource.TestCheckResourceAttrSet("unifi_setting_ntp.test", "ntp_server_4"),
					resource.TestCheckResourceAttr("unifi_setting_ntp.test", "mode", "manual"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_ntp.test", plancheck.ResourceActionUpdate),
			},
			pt.ImportStepWithSite("unifi_setting_ntp.test"),
			{
				Config: testAccSettingNtpConfig2Servers("0.pool.ntp.org", "1.pool.ntp.org", "manual"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ntp.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ntp.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_ntp.test", "ntp_server_1", "0.pool.ntp.org"),
					resource.TestCheckResourceAttr("unifi_setting_ntp.test", "ntp_server_2", "1.pool.ntp.org"),
					resource.TestCheckResourceAttrSet("unifi_setting_ntp.test", "ntp_server_3"),
					resource.TestCheckResourceAttrSet("unifi_setting_ntp.test", "ntp_server_4"),
					resource.TestCheckResourceAttr("unifi_setting_ntp.test", "mode", "manual"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_ntp.test", plancheck.ResourceActionUpdate),
			},
			{
				Config: testAccSettingNtpConfig2Servers("192.168.1.10", "10.0.0.1", "manual"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ntp.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ntp.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_ntp.test", "ntp_server_1", "192.168.1.10"),
					resource.TestCheckResourceAttr("unifi_setting_ntp.test", "ntp_server_2", "10.0.0.1"),
					resource.TestCheckResourceAttrSet("unifi_setting_ntp.test", "ntp_server_3"),
					resource.TestCheckResourceAttrSet("unifi_setting_ntp.test", "ntp_server_4"),
					resource.TestCheckResourceAttr("unifi_setting_ntp.test", "mode", "manual"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_ntp.test", plancheck.ResourceActionUpdate),
			},
			{
				Config: testAccSettingNtpConfig4Servers("time.cloudflare.com", "8.8.8.8", "1.1.1.1", "2.2.2.2", "manual"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ntp.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ntp.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_ntp.test", "ntp_server_1", "time.cloudflare.com"),
					resource.TestCheckResourceAttr("unifi_setting_ntp.test", "ntp_server_2", "8.8.8.8"),
					resource.TestCheckResourceAttr("unifi_setting_ntp.test", "ntp_server_3", "1.1.1.1"),
					resource.TestCheckResourceAttr("unifi_setting_ntp.test", "ntp_server_4", "2.2.2.2"),
					resource.TestCheckResourceAttr("unifi_setting_ntp.test", "mode", "manual"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_ntp.test", plancheck.ResourceActionUpdate),
			},
		},
	})
}

func TestAccSettingNtpInvalid(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 7.3",
		Lock:              settingNtpLock,
		Steps: []resource.TestStep{
			{
				Config:      testAccSettingNtpSimpleConfig("http://invalid-server.com", "auto"),
				ExpectError: regexp.MustCompile("is not a valid"),
			},
			{
				Config:      testAccSettingNtpSimpleConfig("192.168.1", "auto"),
				ExpectError: regexp.MustCompile("is not a valid"),
			},
			{
				Config:      testAccSettingNtpSimpleConfig("time.google.com", "invalid"),
				ExpectError: regexp.MustCompile(`must be one of`),
			},
			{
				Config:      testAccSettingNtpSimpleConfig("time.google.com", "auto"),
				ExpectError: regexp.MustCompile(`must not be configured`),
			},
			{
				Config:      testAccSettingNtpModeOnly("manual"),
				ExpectError: regexp.MustCompile(`At least one of`),
			},
		},
	})
}

func testAccSettingNtpModeOnly(mode string) string {
	return fmt.Sprintf(`
resource "unifi_setting_ntp" "test" {
	mode = %q
}
`, mode)
}

func testAccSettingNtpConfig2Servers(server1, server2, mode string) string {
	return fmt.Sprintf(`
resource "unifi_setting_ntp" "test" {
	ntp_server_1 = %q
	ntp_server_2 = %q
	mode = %q
}
`, server1, server2, mode)
}

func testAccSettingNtpConfig4Servers(server1, server2, server3, server4, mode string) string {
	return fmt.Sprintf(`
resource "unifi_setting_ntp" "test" {
	ntp_server_1 = %q
	ntp_server_2 = %q
	ntp_server_3 = %q
	ntp_server_4 = %q
	mode = %q
}
`, server1, server2, server3, server4, mode)
}

func testAccSettingNtpSimpleConfig(server string, mode string) string {
	return fmt.Sprintf(`
resource "unifi_setting_ntp" "test" {
	ntp_server_1 = %q
	mode = %q
}
`, server, mode)
}

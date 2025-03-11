package acctest

import (
	"fmt"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"sync"
	"testing"
)

var settingDpiLock = &sync.Mutex{}

func TestAccSettingDpi(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingDpiLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingDpiConfig(true, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_dpi.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_dpi.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_dpi.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_dpi.test", "fingerprinting_enabled", "true"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_dpi.test", plancheck.ResourceActionCreate),
			},
			pt.ImportStepWithSite("unifi_setting_dpi.test"),
			{
				Config: testAccSettingDpiConfig(false, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_dpi.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_dpi.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_dpi.test", "enabled", "false"),
					resource.TestCheckResourceAttr("unifi_setting_dpi.test", "fingerprinting_enabled", "true"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_dpi.test", plancheck.ResourceActionUpdate),
			},
			{
				Config: testAccSettingDpiConfig(false, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_dpi.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_dpi.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_dpi.test", "enabled", "false"),
					resource.TestCheckResourceAttr("unifi_setting_dpi.test", "fingerprinting_enabled", "false"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_dpi.test", plancheck.ResourceActionUpdate),
			},
			{
				Config: testAccSettingDpiConfig(true, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_dpi.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_dpi.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_dpi.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_dpi.test", "fingerprinting_enabled", "false"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_dpi.test", plancheck.ResourceActionUpdate),
			},
		},
	})
}

func testAccSettingDpiConfig(enabled, fingerprintingEnabled bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_dpi" "test" {
	enabled = %t
	fingerprinting_enabled = %t
}
`, enabled, fingerprintingEnabled)
}

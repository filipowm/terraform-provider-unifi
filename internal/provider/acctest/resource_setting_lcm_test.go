package acctest

import (
	"fmt"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"regexp"
	"sync"
	"testing"
)

var settingLcmLock = &sync.Mutex{}

func TestAccSettingLcm(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingLcmLock,
		Steps: []resource.TestStep{
			{
				// Test creating with LCM enabled and all optional fields set
				Config: testAccSettingLcmConfig(true, 75, 300, true, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_lcd_monitor.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_lcd_monitor.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_lcd_monitor.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_lcd_monitor.test", "brightness", "75"),
					resource.TestCheckResourceAttr("unifi_setting_lcd_monitor.test", "idle_timeout", "300"),
					resource.TestCheckResourceAttr("unifi_setting_lcd_monitor.test", "sync", "true"),
					resource.TestCheckResourceAttr("unifi_setting_lcd_monitor.test", "touch_event", "true"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_lcd_monitor.test", plancheck.ResourceActionCreate),
			},
			pt.ImportStepWithSite("unifi_setting_lcd_monitor.test"),
			{
				// Test updating with different values
				Config: testAccSettingLcmConfig(true, 50, 600, false, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_lcd_monitor.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_lcd_monitor.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_lcd_monitor.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_lcd_monitor.test", "brightness", "50"),
					resource.TestCheckResourceAttr("unifi_setting_lcd_monitor.test", "idle_timeout", "600"),
					resource.TestCheckResourceAttr("unifi_setting_lcd_monitor.test", "sync", "false"),
					resource.TestCheckResourceAttr("unifi_setting_lcd_monitor.test", "touch_event", "true"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_lcd_monitor.test", plancheck.ResourceActionUpdate),
			},
			{
				// Test disabling LCM (all optional fields should be removed)
				Config: testAccSettingLcmConfigDisabled(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_lcd_monitor.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_lcd_monitor.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_lcd_monitor.test", "enabled", "false"),
					resource.TestCheckNoResourceAttr("unifi_setting_lcd_monitor.test", "brightness"),
					resource.TestCheckNoResourceAttr("unifi_setting_lcd_monitor.test", "idle_timeout"),
					resource.TestCheckNoResourceAttr("unifi_setting_lcd_monitor.test", "sync"),
					resource.TestCheckNoResourceAttr("unifi_setting_lcd_monitor.test", "touch_event"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_lcd_monitor.test", plancheck.ResourceActionUpdate),
			},
			{
				// Test re-enabling LCM with different values
				Config: testAccSettingLcmConfig(true, 100, 3600, true, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_lcd_monitor.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_lcd_monitor.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_lcd_monitor.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_lcd_monitor.test", "brightness", "100"),
					resource.TestCheckResourceAttr("unifi_setting_lcd_monitor.test", "idle_timeout", "3600"),
					resource.TestCheckResourceAttr("unifi_setting_lcd_monitor.test", "sync", "true"),
					resource.TestCheckResourceAttr("unifi_setting_lcd_monitor.test", "touch_event", "false"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_lcd_monitor.test", plancheck.ResourceActionUpdate),
			},
		},
	})
}

// Test that validation errors are raised when trying to set fields with LCM disabled
func TestAccSettingLcmValidation(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingLcmLock,
		Steps: []resource.TestStep{
			{
				Config:      testAccSettingLcmConfigInvalid(),
				ExpectError: regexp.MustCompile(`any of those attributes must not be configured`),
			},
		},
	})
}

func testAccSettingLcmConfig(enabled bool, brightness, idleTimeout int, sync, touchEvent bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_lcd_monitor" "test" {
	enabled = %t
	brightness = %d
	idle_timeout = %d
	sync = %t
	touch_event = %t
}
`, enabled, brightness, idleTimeout, sync, touchEvent)
}

func testAccSettingLcmConfigDisabled() string {
	return `
resource "unifi_setting_lcd_monitor" "test" {
	enabled = false
}
`
}

func testAccSettingLcmConfigInvalid() string {
	return `
resource "unifi_setting_lcd_monitor" "test" {
	enabled = false
	brightness = 50
}
`
}

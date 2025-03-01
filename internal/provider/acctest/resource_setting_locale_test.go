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

var settingLocaleLock = &sync.Mutex{}

func TestAccSettingLocale(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 7.3",
		Lock:              settingLocaleLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingLocaleConfig("America/New_York"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_locale.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_locale.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_locale.test", "timezone", "America/New_York"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_locale.test", plancheck.ResourceActionCreate),
			},
			pt.ImportStepWithSite("unifi_setting_locale.test"),
			{
				Config: testAccSettingLocaleConfig("Europe/London"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_locale.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_locale.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_locale.test", "timezone", "Europe/London"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_locale.test", plancheck.ResourceActionUpdate),
			},
			{
				Config: testAccSettingLocaleConfig("UTC"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_locale.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_locale.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_locale.test", "timezone", "UTC"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_locale.test", plancheck.ResourceActionUpdate),
			},
		},
	})
}
func TestAccSettingLocaleInvalid(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 7.3",
		Lock:              settingLocaleLock,
		Steps: []resource.TestStep{
			{
				Config:      testAccSettingLocaleConfig("Invalid/Timezone"),
				ExpectError: regexp.MustCompile("must be a valid IANA timezone identifier"),
			},
		},
	})
}

func testAccSettingLocaleConfig(timezone string) string {
	return fmt.Sprintf(`
resource "unifi_setting_locale" "test" {
	timezone = %q
}
`, timezone)
}

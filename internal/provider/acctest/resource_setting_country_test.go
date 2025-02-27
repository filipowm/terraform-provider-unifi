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

var settingCountryLock = &sync.Mutex{}

func TestAccSettingCountry(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingCountryLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingCountryConfig("US"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_country.test", "code", "US"),
					resource.TestCheckResourceAttr("unifi_setting_country.test", "code_numeric", "840"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_country.test", plancheck.ResourceActionCreate),
			},
			pt.ImportStepWithSite("unifi_setting_country.test"),
			{
				Config: testAccSettingCountryConfig("PL"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_country.test", "code", "PL"),
					resource.TestCheckResourceAttr("unifi_setting_country.test", "code_numeric", "616"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_country.test", plancheck.ResourceActionUpdate),
			},
		},
	})
}

var invalidCountryCodeErrorRegex = regexp.MustCompile("ISO 3166-1 alpha-2")
var stringLengthExactly2Regex = regexp.MustCompile("string length must be exactly 2")

func TestAccSettingCountry_invalidCode(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingCountryLock,
		Steps: []resource.TestStep{
			{
				Config:      testAccSettingCountryConfig("WP"),
				ExpectError: invalidCountryCodeErrorRegex,
			},
			{
				Config:      testAccSettingCountryConfig("Too long"),
				ExpectError: stringLengthExactly2Regex,
			},
			{
				Config:      testAccSettingCountryConfig(""),
				ExpectError: stringLengthExactly2Regex,
			},
		},
	})
}

func testAccSettingCountryConfig(code string) string {
	return fmt.Sprintf(`
resource "unifi_setting_country" "test" {
	code = %q
}
`, code)
}

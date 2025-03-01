package acctest

import (
	"fmt"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"regexp"
	"testing"
)

func TestAccSettingAutoSpeedtest(t *testing.T) {
	t.Skip("Auto Speedtest is not supported on test controller")
	AcceptanceTest(t, AcceptanceTestCase{
		MinVersion:        version.Must(version.NewVersion("7.2")),
		VersionConstraint: "< 7.4",
		Steps: []resource.TestStep{
			{
				Config: testAccSettingAutoSpeedtestConfig(true, "0 0 * * *"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_auto_speedtest.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_auto_speedtest.test", "cron", "0 0 * * *"),
				),
			},
			pt.ImportStep("unifi_setting_auto_speedtest.test"),
			{
				Config: testAccSettingAutoSpeedtestConfig(false, "0 0 * * *"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_auto_speedtest.test", "enabled", "false"),
					resource.TestCheckResourceAttr("unifi_setting_auto_speedtest.test", "cron", "0 0 * * *"),
				),
			},
			pt.ImportStep("unifi_setting_auto_speedtest.test"),
			{
				Config: testAccSettingAutoSpeedtestConfig(true, "5 0 * * *"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_auto_speedtest.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_auto_speedtest.test", "cron", "5 0 * * *"),
				),
			},
		},
	})
}

// TODO remove when controller changed from USG, which has removed support of speedtest since 7.4. Other controllers still have it.
func TestAccSettingAutoSpeedtest_unsupported(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 7.4",
		Steps: []resource.TestStep{
			{
				Config:      testAccSettingAutoSpeedtestConfig(true, "0 0 * * *"),
				ExpectError: regexp.MustCompile("Auto Speedtest is not supported on this controller"),
			},
		},
	})
}

func testAccSettingAutoSpeedtestConfig(enabled bool, cron string) string {
	return fmt.Sprintf(`
resource "unifi_setting_auto_speedtest" "test" {
	enabled = %t
	cron = %q
}
`, enabled, cron)
}

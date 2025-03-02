package acctest

import (
	"fmt"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"sync"
	"testing"
)

var settingNetworkOptimizationLock = &sync.Mutex{}

func TestAccSettingNetworkOptimization(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingNetworkOptimizationLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingNetworkOptimizationConfig(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_network_optimization.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_network_optimization.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_network_optimization.test", "enabled", "true"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_network_optimization.test", plancheck.ResourceActionCreate),
			},
			pt.ImportStepWithSite("unifi_setting_network_optimization.test"),
			{
				Config: testAccSettingNetworkOptimizationConfig(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_network_optimization.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_network_optimization.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_network_optimization.test", "enabled", "false"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_network_optimization.test", plancheck.ResourceActionUpdate),
			},
		},
	})
}

func testAccSettingNetworkOptimizationConfig(enabled bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_network_optimization" "test" {
	enabled = %t
}
`, enabled)
}

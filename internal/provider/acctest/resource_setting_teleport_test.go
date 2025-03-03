package acctest

import (
	"fmt"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"sync"
	"testing"
)

var settingTeleportLock = &sync.Mutex{}

func TestAccSettingTeleport(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 7.1",
		Lock:              settingTeleportLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingTeleportConfig(true, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_teleport.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_teleport.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_teleport.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_teleport.test", "subnet", ""),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_teleport.test", plancheck.ResourceActionCreate),
			},
			pt.ImportStepWithSite("unifi_setting_teleport.test"),
			{
				Config: testAccSettingTeleportConfig(true, "192.168.100.0/24"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_teleport.test", "enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_teleport.test", "subnet", "192.168.100.0/24"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_teleport.test", plancheck.ResourceActionUpdate),
			},
			{
				Config: testAccSettingTeleportConfigWithoutSubnet(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_teleport.test", "enabled", "false"),
					resource.TestCheckResourceAttr("unifi_setting_teleport.test", "subnet", ""),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_teleport.test", plancheck.ResourceActionUpdate),
			},
		},
	})
}

func testAccSettingTeleportConfig(enabled bool, subnetCidr string) string {
	return fmt.Sprintf(`
resource "unifi_setting_teleport" "test" {
	enabled = %t
	subnet  = %q
}
`, enabled, subnetCidr)
}

func testAccSettingTeleportConfigWithoutSubnet(enabled bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_teleport" "test" {
	enabled     = %t
}
`, enabled)
}

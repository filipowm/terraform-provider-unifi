package acctest

import (
	"fmt"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"sync"
	"testing"
)

var settingUswLock = &sync.Mutex{}

func TestAccSettingUsw(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingUswLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUswConfig(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_usw.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_usw.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_usw.test", "dhcp_snoop", "true"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_usw.test", plancheck.ResourceActionCreate),
			},
			pt.ImportStepWithSite("unifi_setting_usw.test"),
			{
				Config: testAccSettingUswConfig(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_usw.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_usw.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_usw.test", "dhcp_snoop", "false"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_usw.test", plancheck.ResourceActionUpdate),
			},
		},
	})
}

func testAccSettingUswConfig(dhcpSnoop bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_usw" "test" {
	dhcp_snoop = %t
}
`, dhcpSnoop)
}

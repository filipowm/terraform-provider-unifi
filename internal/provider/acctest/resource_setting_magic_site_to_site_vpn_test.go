package acctest

import (
	"fmt"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"sync"
	"testing"
)

var settingMagicSiteToSiteVpnLock = &sync.Mutex{}

func TestAccSettingMagicSiteToSiteVpn(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingMagicSiteToSiteVpnLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingMagicSiteToSiteVpnConfig(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_magic_site_to_site_vpn.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_magic_site_to_site_vpn.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_magic_site_to_site_vpn.test", "enabled", "true"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_magic_site_to_site_vpn.test", plancheck.ResourceActionCreate),
			},
			pt.ImportStepWithSite("unifi_setting_magic_site_to_site_vpn.test"),
			{
				Config: testAccSettingMagicSiteToSiteVpnConfig(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_magic_site_to_site_vpn.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_magic_site_to_site_vpn.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_magic_site_to_site_vpn.test", "enabled", "false"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_magic_site_to_site_vpn.test", plancheck.ResourceActionUpdate),
			},
		},
	})
}

func testAccSettingMagicSiteToSiteVpnConfig(enabled bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_magic_site_to_site_vpn" "test" {
	enabled = %t
}
`, enabled)
}

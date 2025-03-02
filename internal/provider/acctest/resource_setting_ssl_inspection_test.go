package acctest

import (
	"fmt"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"sync"
	"testing"
)

var settingSslInspectionLock = &sync.Mutex{}

func TestAccSettingSslInspection(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 8.2",
		Lock:              settingSslInspectionLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingSslInspectionConfig("off"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ssl_inspection.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ssl_inspection.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_ssl_inspection.test", "state", "off"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_ssl_inspection.test", plancheck.ResourceActionCreate),
			},
			pt.ImportStepWithSite("unifi_setting_ssl_inspection.test"),
			{
				Config: testAccSettingSslInspectionConfig("simple"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ssl_inspection.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ssl_inspection.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_ssl_inspection.test", "state", "simple"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_ssl_inspection.test", plancheck.ResourceActionUpdate),
			},
			{
				Config: testAccSettingSslInspectionConfig("advanced"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("unifi_setting_ssl_inspection.test", "id"),
					resource.TestCheckResourceAttr("unifi_setting_ssl_inspection.test", "site", "default"),
					resource.TestCheckResourceAttr("unifi_setting_ssl_inspection.test", "state", "advanced"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_setting_ssl_inspection.test", plancheck.ResourceActionUpdate),
			},
		},
	})
}

func testAccSettingSslInspectionConfig(state string) string {
	return fmt.Sprintf(`
resource "unifi_setting_ssl_inspection" "test" {
	state = "%s"
}
`, state)
}

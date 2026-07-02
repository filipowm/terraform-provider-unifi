package acctest

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"

	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserGroup_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupConfig(name),
				// Check:  resource.ComposeTestCheckFunc(
				// // testCheckUserGroupExists(t, "name"),
				// ),
			},
			{
				Config: testAccUserGroupConfigQos(name),
			},
			pt.ImportStep("unifi_user_group.test"),
			{
				Config: testAccUserGroupConfig(name),
			},
			pt.ImportStep("unifi_user_group.test"),
		},
	})
}

func testAccUserGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_user_group" "test" {
	name = "%s"
}
`, name)
}

func testAccUserGroupConfigQos(name string) string {
	return fmt.Sprintf(`
resource "unifi_user_group" "test" {
	name = "%s"

	qos_rate_max_up   = 2000
	qos_rate_max_down = 50
}
`, name)
}

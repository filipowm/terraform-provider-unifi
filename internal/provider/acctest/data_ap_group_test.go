package acctest

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

const (
	testAPGroupDatasourceName = "data.unifi_ap_group.test"
	defaultAPGroupName        = "All APs"
)

func TestAccDataAPGroup_default(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccDataAPGroupConfig_default,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testAPGroupDatasourceName, "name", defaultAPGroupName),
				),
			},
		},
	})
}

func TestAccDataAPGroup_byName(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccDataAPGroupConfig(defaultAPGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testAPGroupDatasourceName, "name", defaultAPGroupName),
				),
			},
		},
	})
}

const testAccDataAPGroupConfig_default = `
data "unifi_ap_group" "test" {
}
`

func testAccDataAPGroupConfig(name string) string {
	return fmt.Sprintf(`
data "unifi_ap_group" "test" {
	name = %q
}
`, name)
}

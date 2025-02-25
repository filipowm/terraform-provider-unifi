package acctest

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataPortProfile_default(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: "< 7.4",
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccDataPortProfileConfig_default,
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

const testAccDataPortProfileConfig_default = `
data "unifi_port_profile" "default" {
}
`

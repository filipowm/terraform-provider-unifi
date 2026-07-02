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
				Config: testAccDataPortProfileConfigDefault,
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

const testAccDataPortProfileConfigDefault = `
data "unifi_port_profile" "default" {
}
`

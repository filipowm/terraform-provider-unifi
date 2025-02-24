package v1

import (
	"fmt"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPortProfile_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			pt.PreCheck(t)
			pt.PreCheckVersionConstraint(t, "< 7.4")
		},
		ProtoV6ProviderFactories: MuxProviders(t),
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccPortProfileConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_profile.test", "poe_mode", "off"),
					resource.TestCheckResourceAttr("unifi_port_profile.test", "name", name),
				),
			},
			pt.ImportStep("unifi_port_profile.test"),
		},
	})
}

func testAccPortProfileConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_port_profile" "test" {
	name = "%s"

	poe_mode	  = "off"
	speed 		  = 1000
	stp_port_mode = false
}
`, name)
}

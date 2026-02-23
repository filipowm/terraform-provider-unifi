package acctest

import (
	"fmt"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPortProfile_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	AcceptanceTest(t, AcceptanceTestCase{
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

func TestAccPortProfile_forwardNative(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	AcceptanceTest(t, AcceptanceTestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccPortProfileConfigForwardNative(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_profile.test", "name", name),
					// forward may read back as "native" or "customize" from the API;
					// DiffSuppressFunc should treat both as equivalent.
				),
			},
			// Re-apply the same config with PlanOnly to verify no spurious drift.
			{
				Config:   testAccPortProfileConfigForwardNative(name),
				PlanOnly: true,
			},
			pt.ImportStep("unifi_port_profile.test"),
		},
	})
}

func testAccPortProfileConfigForwardNative(name string) string {
	return fmt.Sprintf(`
resource "unifi_port_profile" "test" {
	name    = "%s"
	forward = "native"

	poe_mode      = "off"
	speed         = 1000
	stp_port_mode = false
}
`, name)
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

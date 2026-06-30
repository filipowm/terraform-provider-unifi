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
		VersionConstraint: "< 7.4",
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

// TestAccPortProfile_update exercises an in-place update of an already-created port profile.
// This is the lifecycle that triggered issue #98: the second apply (update) failed with
// "Error: not found" because go-unifi v1.9.2 turns a successful-but-empty PUT into
// unifi.ErrNotFound. The update step changes a real attribute (poe_mode) and asserts the apply
// succeeds. NOTE: this acctest is best-effort - the Dockerized controller may echo the object
// (len==1) and therefore not reproduce the bug; the deterministic regression lives in the unit
// test (resource_port_profile_unit_test.go).
func TestAccPortProfile_update(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: "< 7.4",
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccPortProfileConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_profile.test", "poe_mode", "off"),
					resource.TestCheckResourceAttr("unifi_port_profile.test", "name", name),
				),
			},
			{
				Config: testAccPortProfileConfigUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_port_profile.test", "poe_mode", "passthrough"),
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

func testAccPortProfileConfigUpdated(name string) string {
	return fmt.Sprintf(`
resource "unifi_port_profile" "test" {
	name = "%s"

	poe_mode	  = "passthrough"
	speed 		  = 1000
	stp_port_mode = false
}
`, name)
}

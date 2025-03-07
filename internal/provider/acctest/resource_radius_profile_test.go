package acctest

import (
	"fmt"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRadiusProfile_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccRadiusProfileConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_radius_profile.test", "name", name),
				),
			},
			pt.ImportStep("unifi_radius_profile.test"),
		},
	})
}

func TestAccRadiusProfile_servers(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccRadiusProfileConfigServer(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_radius_profile.test", "name", name),
				),
			},
			pt.ImportStep("unifi_radius_profile.test"),
		},
	})
}

func TestAccRadiusProfile_importByName(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Steps: []resource.TestStep{
			// Apply and import network by name.
			{
				Config: testAccRadiusProfileImport(),
			},
			{
				Config:            testAccRadiusProfileImport(),
				ResourceName:      "unifi_radius_profile.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "name=imported",
			},
		},
	})
}

func testAccRadiusProfileConfigServer(name string) string {
	return fmt.Sprintf(`
resource "unifi_radius_profile" "test" {
	name = "%s"
	auth_server {
		ip = "192.168.1.1"
		xsecret = "securepw1"
	}
	auth_server {
		ip = "192.168.10.1"
		port = 8888
		xsecret = "securepw2"
	}
	acct_server {
		ip = "192.168.1.1"
		xsecret = "securepw1"
	}
	acct_server {
		ip = "192.168.10.1"
		port = 9999
		xsecret = "securepw2"
	}
	use_usg_acct_server = false
	use_usg_auth_server = false
}
`, name)
}

func testAccRadiusProfileConfig(name string) string {
	return fmt.Sprintf(`
resource "unifi_radius_profile" "test" {
  	name = "%[1]s"
}
`, name)
}

func testAccRadiusProfileImport() string {
	return `
resource "unifi_radius_profile" "test" {
  	name = "imported"
	auth_server {
		ip = "192.168.1.1"
		port = 1812
		xsecret = "securepw"
	}
	use_usg_auth_server = true
	vlan_enabled = true
	vlan_wlan_mode = "required"
}
`
}

package acctest

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"

	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAccount_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig(name, "secure"),
				Check: resource.ComposeTestCheckFunc(
					// testCheckNetworkExists(t, "name"),
					resource.TestCheckResourceAttr("unifi_account.test", "name", name),
				),
			},
			pt.ImportStep("unifi_account.test"),
		},
	})
}

func TestAccAccount_mac(t *testing.T) {
	mac, unallocateMac := pt.AllocateTestMac(t)
	defer unallocateMac()
	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig(mac, mac),
				Check: resource.ComposeTestCheckFunc(
					// testCheckNetworkExists(t, "name"),
					resource.TestCheckResourceAttr("unifi_account.test", "name", mac),
					resource.TestCheckResourceAttr("unifi_account.test", "password", mac),
				),
			},
			pt.ImportStep("unifi_account.test"),
		},
	})
}

func TestAccAccount_vlan(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	_, vlan := pt.GetTestVLAN(t)
	_, vlan2 := pt.GetTestVLAN(t)
	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfigVLAN(name, "secure", vlan),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_account.test", "name", name),
					resource.TestCheckResourceAttr("unifi_account.test", "vlan", strconv.Itoa(vlan)),
				),
			},
			{
				Config: testAccAccountConfigVLAN(name, "secure", vlan2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_account.test", "vlan", strconv.Itoa(vlan2)),
				),
			},
			pt.ImportStep("unifi_account.test"),
		},
	})
}

func testAccAccountConfig(name, password string) string {
	return fmt.Sprintf(`
resource "unifi_account" "test" {
	name = "%[1]s"
	password = "%[2]s"
}
`, name, password)
}

func testAccAccountConfigVLAN(name, password string, vlan int) string {
	return fmt.Sprintf(`
resource "unifi_account" "test" {
	name = "%[1]s"
	password = "%[2]s"
	vlan = %[3]d
}
`, name, password, vlan)
}

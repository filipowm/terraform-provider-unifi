package acctest

import (
	"fmt"
	"strconv"
	"testing"

	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataAccount_default(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccDataAccountConfig(name, "secure_1234"),
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func TestAccDataAccount_mac(t *testing.T) {
	mac, unallocateMac := pt.AllocateTestMac(t)
	defer unallocateMac()

	AcceptanceTest(t, AcceptanceTestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccDataAccountConfig(mac, mac),
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func TestAccDataAccount_vlan(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	_, vlan := pt.GetTestVLAN(t)
	AcceptanceTest(t, AcceptanceTestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccDataAccountConfigVLAN(name, "secure_1234", vlan),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.unifi_account.test", "vlan", strconv.Itoa(vlan)),
					resource.TestCheckResourceAttrPair("data.unifi_account.test", "vlan", "unifi_account.test", "vlan"),
				),
			},
		},
	})
}

func testAccDataAccountConfig(name, password string) string {
	return fmt.Sprintf(`
resource "unifi_account" "test" {
	name = "%[1]s"
	password = "%[2]s"
}

data "unifi_account" "test" {
	name = "%[1]s"
depends_on = [
    unifi_account.test
  ]
}
`, name, password)
}

func testAccDataAccountConfigVLAN(name, password string, vlan int) string {
	return fmt.Sprintf(`
resource "unifi_account" "test" {
	name = "%[1]s"
	password = "%[2]s"
	vlan = %[3]d
}

data "unifi_account" "test" {
	name = "%[1]s"
depends_on = [
    unifi_account.test
  ]
}
`, name, password, vlan)
}

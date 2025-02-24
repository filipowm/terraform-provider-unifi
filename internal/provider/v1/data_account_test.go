package v1

import (
	"fmt"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccDataAccount_default(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			pt.PreCheck(t)
		},
		ProtoV6ProviderFactories: MuxProviders(t),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			pt.PreCheck(t)
		},
		ProtoV6ProviderFactories: MuxProviders(t),
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccDataAccountConfig(mac, mac),
				Check:  resource.ComposeTestCheckFunc(),
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

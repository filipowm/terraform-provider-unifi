package v1

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataUser_default(t *testing.T) {
	mac, unallocateTestMac := allocateTestMac(t)
	defer unallocateTestMac()
	name := acctest.RandomWithPrefix("tfacc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			//preCheck(t)

			_, err := testClient.CreateUser(context.Background(), "default", &unifi.User{
				MAC:  mac,
				Name: name,
				Note: name,
			})
			if err != nil {
				t.Fatal(err)
			}
		},
		//PreCheck:          func() { preCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataUserConfig_default(mac),
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func testAccDataUserConfig_default(mac string) string {
	return fmt.Sprintf(`
data "unifi_user" "test" {
mac = "%s"
}
`, mac)
}

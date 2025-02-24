package v1

import (
	"fmt"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccDataNetwork_byName(t *testing.T) {
	defaultName := "Default"
	//v, err := version.NewVersion(testClient.Version())
	//if err != nil {
	//	t.Fatalf("error parsing version: %s", err)
	//}
	//if v.LessThan(provider.ControllerV7) {
	//	defaultName = "LAN"
	//}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			pt.PreCheck(t)
		},
		ProtoV6ProviderFactories: MuxProviders(t),
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccDataNetworkConfig_byName(defaultName),
				Check:  resource.ComposeTestCheckFunc(
				// testCheckNetworkExists(t, "name"),
				),
			},
		},
	})
}

func TestAccDataNetwork_byID(t *testing.T) {
	defaultName := "Default"
	//v, err := version.NewVersion(testClient.Version())
	//if err != nil {
	//	t.Fatalf("error parsing version: %s", err)
	//}
	//if v.LessThan(provider.ControllerV7) {
	//	defaultName = "LAN"
	//}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			pt.PreCheck(t)
		},
		ProtoV6ProviderFactories: MuxProviders(t),
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccDataNetworkConfig_byID(defaultName),
				Check:  resource.ComposeTestCheckFunc(
				// testCheckNetworkExists(t, "name"),
				),
			},
		},
	})
}

func testAccDataNetworkConfig_byName(name string) string {
	return fmt.Sprintf(`
data "unifi_network" "lan" {
	name = %q
}
`, name)
}

func testAccDataNetworkConfig_byID(name string) string {
	return fmt.Sprintf(`
data "unifi_network" "lan" {
	name = %q
}

data "unifi_network" "lan_id" {
	id = data.unifi_network.lan.id
}
`, name)
}

package acctest

import (
	"fmt"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccDataNetwork_byName(t *testing.T) {
	defaultName := "Default"
	v, err := version.NewVersion(testClient.Version())
	if err != nil {
		t.Fatalf("error parsing version: %s", err)
	}
	if v.LessThan(base.ControllerV7) {
		defaultName = "LAN"
	}
	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccDataNetworkConfig_byName(defaultName),
				Check: resource.ComposeTestCheckFunc(
					// testCheckNetworkExists(t, "name"),
					// dhcp_guarding is Computed on the data source; assert it is set to a
					// known bool so resource/data-source coverage stays in lockstep (#123).
					resource.TestCheckResourceAttrSet("data.unifi_network.lan", "dhcp_guarding"),
				),
			},
		},
	})
}

func TestAccDataNetwork_byID(t *testing.T) {
	defaultName := "Default"
	v, err := version.NewVersion(testClient.Version())
	if err != nil {
		t.Fatalf("error parsing version: %s", err)
	}
	if v.LessThan(base.ControllerV7) {
		defaultName = "LAN"
	}

	AcceptanceTest(t, AcceptanceTestCase{
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

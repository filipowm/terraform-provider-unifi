package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFirewallGroup_port_group(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: providerFactories,
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallGroupConfig(name, "port-group", nil),
				// Check:  resource.ComposeTestCheckFunc(
				// // testCheckFirewallGroupExists(t, "name"),
				// ),
			},
			importStep("unifi_firewall_group.test"),
			{
				Config: testAccFirewallGroupConfig(name, "port-group", []string{"80", "443"}),
			},
			importStep("unifi_firewall_group.test"),
		},
	})
}

func TestAccFirewallGroup_address_group(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: providerFactories,
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallGroupConfig(name, "address-group", nil),
				// Check:  resource.ComposeTestCheckFunc(
				// // testCheckFirewallGroupExists(t, "name"),
				// ),
			},
			importStep("unifi_firewall_group.test"),
			{
				Config: testAccFirewallGroupConfig(name, "address-group", []string{"10.0.0.1", "10.0.0.2"}),
			},
			importStep("unifi_firewall_group.test"),
			{
				Config: testAccFirewallGroupConfig(name, "address-group", []string{"10.0.0.0/24"}),
			},
			importStep("unifi_firewall_group.test"),
		},
	})
}

func TestAccFirewallGroup_same_name(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: providerFactories,
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config:      testAccFirewallGroupConfig_same_name,
				ExpectError: regexp.MustCompile("firewall groups must have unique names"),
			},
		},
	})
}

func testAccFirewallGroupConfig(name, ty string, members []string) string {
	joined := strings.Join(members, "\",\"")
	if len(joined) > 0 {
		joined = "\"" + joined + "\""
	}

	return fmt.Sprintf(`
resource "unifi_firewall_group" "test" {
	name = "%s"
	type = "%s"
	
	members = [%s]
}
`, name, ty, joined)
}

const testAccFirewallGroupConfig_same_name = `
resource "unifi_firewall_group" "test_a" {
	name = "tf-acc fg"
	type = "address-group"
	
	members = []
}

resource "unifi_firewall_group" "test_b" {
	name = "tf-acc fg"
	type = "address-group"
	
	members = []
}
`

package acctest

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"regexp"
	"strings"
	"testing"

	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFirewallGroup_port_group(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallGroupConfig(name, "port-group", nil),
				// Check:  resource.ComposeTestCheckFunc(
				// // testCheckFirewallGroupExists(t, "name"),
				// ),
			},
			pt.ImportStep("unifi_firewall_group.test"),
			{
				Config: testAccFirewallGroupConfig(name, "port-group", []string{"80", "443"}),
			},
			pt.ImportStep("unifi_firewall_group.test"),
		},
	})
}

func TestAccFirewallGroup_address_group(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc")
	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallGroupConfig(name, "address-group", nil),
				// Check:  resource.ComposeTestCheckFunc(
				// // testCheckFirewallGroupExists(t, "name"),
				// ),
			},
			pt.ImportStep("unifi_firewall_group.test"),
			{
				Config: testAccFirewallGroupConfig(name, "address-group", []string{"10.0.0.1", "10.0.0.2"}),
			},
			pt.ImportStep("unifi_firewall_group.test"),
			{
				Config: testAccFirewallGroupConfig(name, "address-group", []string{"10.0.0.0/24"}),
			},
			pt.ImportStep("unifi_firewall_group.test"),
		},
	})
}

func TestAccFirewallGroup_same_name(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config:      testAccFirewallGroupConfig_same_name,
				ExpectError: regexp.MustCompile("FirewallGroupExisted"),
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

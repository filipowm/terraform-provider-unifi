package acctest

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
)

func TestAccDynamicDNS_dyndns(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccDynamicDNSConfig,
				// Check:  resource.ComposeTestCheckFunc(
				// // testCheckFirewallGroupExists(t, "name"),
				// ),
			},
			pt.ImportStep("unifi_dynamic_dns.test"),
		},
	})
}

const testAccDynamicDNSConfig = `
resource "unifi_dynamic_dns" "test" {
	service = "dyndns"
	
	host_name = "test.example.com"

	server   = "dyndns.example.com"
	login    = "testuser"
	password = "password"
}
`

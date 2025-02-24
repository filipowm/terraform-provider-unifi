package v1

import (
	"testing"

	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDynamicDNS_dyndns(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { pt.PreCheck(t) },
		ProtoV6ProviderFactories: MuxProviders(t),
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

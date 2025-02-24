package v1

import (
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccDataAPGroup_default(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			pt.PreCheck(t)
		},
		ProtoV6ProviderFactories: MuxProviders(t),
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccDataAPGroupConfig_default,
				Check:  resource.ComposeTestCheckFunc(
				// testCheckNetworkExists(t, "name"),
				),
			},
		},
	})
}

const testAccDataAPGroupConfig_default = `
data "unifi_ap_group" "default" {
}
`

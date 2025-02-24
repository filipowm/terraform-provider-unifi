package v1

import (
	"testing"

	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataUserGroup_default(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { pt.PreCheck(t) },
		ProtoV6ProviderFactories: MuxProviders(t),
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccDataUserGroupConfig_default,
				Check:  resource.ComposeTestCheckFunc(
				// testCheckNetworkExists(t, "name"),
				),
			},
		},
	})
}

const testAccDataUserGroupConfig_default = `
data "unifi_user_group" "default" {
}
`

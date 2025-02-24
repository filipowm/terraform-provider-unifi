package v1

import (
	"testing"

	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataPortProfile_default(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			pt.PreCheck(t)
			pt.PreCheckVersionConstraint(t, "< 7.4")
		},
		ProtoV6ProviderFactories: MuxProviders(t),
		// TODO: CheckDestroy: ,
		Steps: []resource.TestStep{
			{
				Config: testAccDataPortProfileConfig_default,
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

const testAccDataPortProfileConfig_default = `
data "unifi_port_profile" "default" {
}
`

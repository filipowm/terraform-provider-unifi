package acctest

import (
	"context"
	"fmt"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccPortalFile_basic(t *testing.T) {

	AcceptanceTest(t, AcceptanceTestCase{
		CheckDestroy: testAccCheckPortalFileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPortalFileConfig("files/testfile.png"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_portal_file.test", "site", "default"),
					resource.TestCheckResourceAttrSet("unifi_portal_file.test", "filename"),
					resource.TestCheckResourceAttrSet("unifi_portal_file.test", "content_type"),
					resource.TestCheckResourceAttrSet("unifi_portal_file.test", "file_size"),
					resource.TestCheckResourceAttrSet("unifi_portal_file.test", "md5"),
					resource.TestCheckResourceAttrSet("unifi_portal_file.test", "url"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_portal_file.test", plancheck.ResourceActionCreate),
			},
			{
				Config: testAccPortalFileConfig("files/testfile2.jpg"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_portal_file.test", "site", "default"),
					resource.TestCheckResourceAttrSet("unifi_portal_file.test", "filename"),
					resource.TestCheckResourceAttrSet("unifi_portal_file.test", "content_type"),
					resource.TestCheckResourceAttrSet("unifi_portal_file.test", "file_size"),
					resource.TestCheckResourceAttrSet("unifi_portal_file.test", "md5"),
					resource.TestCheckResourceAttrSet("unifi_portal_file.test", "url"),
				),
				ConfigPlanChecks: pt.CheckResourceActions("unifi_portal_file.test", plancheck.ResourceActionReplace),
			},
		},
	})
}

func testAccCheckPortalFileDestroy(s *terraform.State) error {
	return pt.CheckDestroy("unifi_portal_file", func(ctx context.Context, site, id string) error {
		_, err := testClient.GetPortalFile(ctx, site, id)
		return err
	})(s)
}

func testAccPortalFileConfig(filePath string) string {
	return fmt.Sprintf(`
resource "unifi_portal_file" "test" {
  file_path = %q
}
`, filepath.ToSlash(filePath))
}

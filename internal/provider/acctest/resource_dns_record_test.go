package acctest

import (
	"fmt"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestDNSRecord_basic(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Steps: Steps{
			{
				Config: testAccDnsRecord("test.com", "192.168.0.128", "A"),
				Check: resource.ComposeTestCheckFunc(
					// testCheckNetworkExists(t, "name"),
					resource.TestCheckResourceAttr("unifi_dns_record.test", "name", "test.com"),
				),
			},
			pt.ImportStep("unifi_dns_record.test"),
		},
	})
}

func testAccDnsRecord(name, record, recordType string) string {
	return fmt.Sprintf(`
resource "unifi_dns_record" "test" {
	name = "%[1]s"
	record = "%[2]s"
	type = "%[3]s"
}
`, name, record, recordType)
}

// muxProviders returns a map of mux servers for the acceptance tests.

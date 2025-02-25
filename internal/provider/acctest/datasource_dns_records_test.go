package acctest

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"strings"
	"testing"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testDnsRecordsDataSourceName = "data.unifi_dns_records.test"

func TestDNSRecordsDataSource_basic(t *testing.T) {
	records := []*dnsRecordTestCase{
		{
			name:       "test1",
			record:     "192.168.1.100",
			recordType: "A",
		},
		{
			name:       "test2",
			record:     "192.168.1.200",
			recordType: "A",
		},
		{
			name:       "mail",
			record:     "mail.example.com",
			recordType: "MX",
			priority:   intPtr(10),
		},
	}

	var configs []string
	var dependencies []string
	for _, record := range records {
		record.recordName = pt.RandHostname()
		resourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
		configs = append(configs, testAccDnsRecordConfigWithResourceName(resourceName, *record))
		dependencies = append(dependencies, fmt.Sprintf("unifi_dns_record.%s", resourceName))
	}
	configs = append(configs, testAccDnsRecordsDataSourceConfig(dependencies))
	AcceptanceTest(t, AcceptanceTestCase{
		MinVersion: base.ControllerVersionDnsRecords,
		Steps: Steps{
			{
				Config: pt.ComposeConfig(configs...),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testDnsRecordsDataSourceName, "result.#", "3"),
					resource.TestCheckResourceAttrSet(testDnsRecordsDataSourceName, "result.0.name"),
					resource.TestCheckResourceAttrSet(testDnsRecordsDataSourceName, "result.0.record"),
					resource.TestCheckResourceAttrSet(testDnsRecordsDataSourceName, "result.0.type"),
					resource.TestCheckResourceAttrSet(testDnsRecordsDataSourceName, "result.1.name"),
					resource.TestCheckResourceAttrSet(testDnsRecordsDataSourceName, "result.1.record"),
					resource.TestCheckResourceAttrSet(testDnsRecordsDataSourceName, "result.1.type"),
					resource.TestCheckResourceAttrSet(testDnsRecordsDataSourceName, "result.2.name"),
					resource.TestCheckResourceAttrSet(testDnsRecordsDataSourceName, "result.2.record"),
					resource.TestCheckResourceAttrSet(testDnsRecordsDataSourceName, "result.2.type"),
				),
			},
		},
	})
}

func TestDNSRecordsDataSource_noRecords(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		MinVersion: base.ControllerVersionDnsRecords,
		Steps: Steps{
			{
				Config: testAccDnsRecordsDataSourceConfig(nil),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testDnsRecordsDataSourceName, "result.#", "0"),
				),
			},
		},
	})
}

func testAccDnsRecordsDataSourceConfig(deps []string) string {
	return `
data "unifi_dns_records" "test" {
	depends_on = [
		` + strings.Join(deps, ",") + `
	]
}`
}

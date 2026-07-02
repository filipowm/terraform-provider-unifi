package acctest

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
)

const testDNSRecordsDataSourceName = "data.unifi_dns_records.test"

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
		configs = append(configs, testAccDNSRecordConfigWithResourceName(resourceName, *record))
		dependencies = append(dependencies, "unifi_dns_record."+resourceName)
	}
	configs = append(configs, testAccDNSRecordsDataSourceConfig(dependencies))
	AcceptanceTest(t, AcceptanceTestCase{
		MinVersion: base.ControllerVersionDNSRecords,
		Lock:       dnsLock,
		Steps: Steps{
			{
				Config: pt.ComposeConfig(configs...),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testDNSRecordsDataSourceName, "result.#", "3"),
					resource.TestCheckResourceAttrSet(testDNSRecordsDataSourceName, "result.0.name"),
					resource.TestCheckResourceAttrSet(testDNSRecordsDataSourceName, "result.0.record"),
					resource.TestCheckResourceAttrSet(testDNSRecordsDataSourceName, "result.0.type"),
					resource.TestCheckResourceAttrSet(testDNSRecordsDataSourceName, "result.1.name"),
					resource.TestCheckResourceAttrSet(testDNSRecordsDataSourceName, "result.1.record"),
					resource.TestCheckResourceAttrSet(testDNSRecordsDataSourceName, "result.1.type"),
					resource.TestCheckResourceAttrSet(testDNSRecordsDataSourceName, "result.2.name"),
					resource.TestCheckResourceAttrSet(testDNSRecordsDataSourceName, "result.2.record"),
					resource.TestCheckResourceAttrSet(testDNSRecordsDataSourceName, "result.2.type"),
				),
			},
		},
	})
}

func TestDNSRecordsDataSource_noRecords(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		MinVersion: base.ControllerVersionDNSRecords,
		Lock:       dnsLock,
		Steps: Steps{
			{
				Config: testAccDNSRecordsDataSourceConfig(nil),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testDNSRecordsDataSourceName, "result.#", "0"),
				),
			},
		},
	})
}

func testAccDNSRecordsDataSourceConfig(deps []string) string {
	return `
data "unifi_dns_records" "test" {
	depends_on = [
		` + strings.Join(deps, ",") + `
	]
}`
}

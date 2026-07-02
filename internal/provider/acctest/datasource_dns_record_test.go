package acctest

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
)

const testDNSRecordDataSourceName = "data.unifi_dns_record.test"

func TestDNSRecordDataSource_basic(t *testing.T) {
	testCases := []struct {
		name         string
		record       string
		recordType   string
		filterByName bool
	}{
		{
			name:         "filter by name",
			record:       "192.168.1.100",
			recordType:   "A",
			filterByName: true,
		},
		{
			name:       "filter by record",
			record:     "192.168.1.200",
			recordType: "A",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recordName := pt.RandHostname()
			r := dnsRecordTestCase{
				recordName: recordName,
				record:     tc.record,
				recordType: tc.recordType,
			}

			AcceptanceTest(t, AcceptanceTestCase{
				MinVersion: base.ControllerVersionDNSRecords,
				Lock:       dnsLock,

				Steps: Steps{
					{
						Config: pt.ComposeConfig(testAccDNSRecordConfig(r), testAccDNSRecordDataSourceConfig(r, tc.filterByName)),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(testDNSRecordDataSourceName, "name", recordName),
							resource.TestCheckResourceAttr(testDNSRecordDataSourceName, "record", tc.record),
							resource.TestCheckResourceAttr(testDNSRecordDataSourceName, "type", tc.recordType),
						),
					},
				},
			})
		})
	}
}

var dnsDataSourceFilterErrorRegex = regexp.MustCompile(`[name,record]`)

func TestDNSRecordDataSource_errorWithoutFilter(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		MinVersion: base.ControllerVersionDNSRecords,

		Steps: Steps{
			{
				Config:      testAccDNSRecordDataSourceWithoutFilter(),
				ExpectError: dnsDataSourceFilterErrorRegex,
			},
		},
	})
}

func testAccDNSRecordDataSourceConfig(tc dnsRecordTestCase, filterByName bool) string {
	filter := ""
	if filterByName {
		filter = "name = \"" + tc.recordName + "\""
	} else {
		filter = "record = \"" + tc.record + "\""
	}

	return fmt.Sprintf(`
data "unifi_dns_record" "test" {
	%s
	depends_on = [unifi_dns_record.test]
}`, filter)
}

func testAccDNSRecordDataSourceWithoutFilter() string {
	return `
data "unifi_dns_record" "test" {
}`
}

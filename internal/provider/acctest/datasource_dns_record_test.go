package acctest

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testDnsRecordDataSourceName = "data.unifi_dns_record.test"

func TestDNSRecordDataSource_basic(t *testing.T) {
	t.Parallel()

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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			recordName := pt.RandHostname()
			r := dnsRecordTestCase{
				recordName: recordName,
				record:     tc.record,
				recordType: tc.recordType,
			}

			AcceptanceTest(t, AcceptanceTestCase{
				MinVersion: base.ControllerVersionDnsRecords,

				Steps: Steps{
					{
						Config: pt.ComposeConfig(testAccDnsRecordConfig(r), testAccDnsRecordDataSourceConfig(r, tc.filterByName)),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(testDnsRecordDataSourceName, "name", recordName),
							resource.TestCheckResourceAttr(testDnsRecordDataSourceName, "record", tc.record),
							resource.TestCheckResourceAttr(testDnsRecordDataSourceName, "type", tc.recordType),
						),
					},
				},
			})
		})
	}
}

var (
	dnsDataSourceFilterErrorRegex = regexp.MustCompile(`[filter.name,filter.record]`)
)

func TestDNSRecordDataSource_errorWithoutFilter(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		MinVersion: base.ControllerVersionDnsRecords,

		Steps: Steps{
			{
				Config:      testAccDnsRecordDataSourceWithoutFilter(),
				ExpectError: dnsDataSourceFilterErrorRegex,
			},
		},
	})
}

func testAccDnsRecordDataSourceConfig(tc dnsRecordTestCase, filterByName bool) string {
	filter := ""
	if filterByName {
		filter = "name = \"" + tc.recordName + "\""
	} else {
		filter = "record = \"" + tc.record + "\""
	}

	return fmt.Sprintf(`
data "unifi_dns_record" "test" {
	filter {
		%s
	}
	depends_on = [unifi_dns_record.test]
}`, filter)
}

func testAccDnsRecordDataSourceWithoutFilter() string {
	return `
data "unifi_dns_record" "test" {
	filter {

 	}
}`
}

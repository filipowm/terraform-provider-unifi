package acctest

import (
	"context"
	"fmt"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"strconv"
	"strings"
	"testing"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const testDnsRecordResourceName = "unifi_dns_record.test"

type dnsRecordTestCase struct {
	name       string
	recordName string
	record     string
	recordType string
	ttl        *int
	enabled    *bool
	priority   *int
	port       *int
	weight     *int
}

func TestDNSRecord_basic(t *testing.T) {
	t.Parallel()
	testCases := []dnsRecordTestCase{
		{
			name:       "A record",
			recordName: "test.com",
			record:     "192.168.0.128",
			recordType: "A",
		},
		{
			name:       "AAAA record",
			recordName: "ipv6.test.com",
			record:     "2001:db8::1",
			recordType: "AAAA",
		},
		{
			name:       "CNAME record",
			recordName: "alias.test.com",
			record:     "target.test.com",
			recordType: "CNAME",
		},
		{
			name:       "NS record",
			recordName: "ns.test.com",
			record:     "127.0.0.1",
			recordType: "NS",
		},
		{
			name:       "MX record with priority",
			recordName: "mail.test.com",
			record:     "mx.test.com",
			recordType: "MX",
			priority:   intPtr(10),
		},
		{
			name:       "disabled A record",
			recordName: "disabled.test.com",
			record:     "192.168.1.100",
			recordType: "A",
			enabled:    boolPtr(false),
		},
		{
			name:       "A record with TTL",
			recordName: "ttl.test.com",
			record:     "192.168.1.100",
			recordType: "A",
			ttl:        intPtr(3600),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			steps := []resource.TestStep{
				{
					Config: testAccDnsRecordConfig(tc),
					Check:  testAccDnsRecordCheckAttrs(tc),
				},
				pt.ImportStep(testDnsRecordResourceName),
			}

			AcceptanceTest(t, AcceptanceTestCase{
				MinVersion:   base.ControllerVersionDnsRecords,
				Steps:        steps,
				CheckDestroy: testAccCheckDNSRecordDestroy,
			})
		})
	}
}

func TestDNSRecord_SRV(t *testing.T) {
	t.Parallel()
	testCases := []dnsRecordTestCase{
		{
			name:       "SRV record with all fields",
			recordName: "_sip._tcp.test.com",
			record:     "sip.test.com",
			recordType: "SRV",
			port:       intPtr(5060),
			priority:   intPtr(10),
			weight:     intPtr(20),
		},
		{
			name:       "SRV record with minimal fields",
			recordName: "_ldap._tcp.test.com",
			record:     "ldap.test.com",
			recordType: "SRV",
			port:       intPtr(389),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			AcceptanceTest(t, AcceptanceTestCase{
				MinVersion:   base.ControllerVersionDnsRecords,
				CheckDestroy: testAccCheckDNSRecordDestroy,
				Steps: Steps{
					{
						Config: testAccDnsRecordConfig(tc),
						Check:  testAccDnsRecordCheckAttrs(tc),
					},
				},
			})
		})
	}
}

func TestDNSRecord_Update(t *testing.T) {
	initial := dnsRecordTestCase{
		name:       "initial",
		recordName: "update.test.com",
		record:     "192.168.1.100",
		recordType: "A",
		ttl:        intPtr(3600),
	}

	updated := dnsRecordTestCase{
		name:       "updated",
		recordName: "update.test.com",
		record:     "192.168.1.200",
		recordType: "A",
		ttl:        intPtr(7200),
	}

	AcceptanceTest(t, AcceptanceTestCase{
		MinVersion:   base.ControllerVersionDnsRecords,
		CheckDestroy: testAccCheckDNSRecordDestroy,
		Steps: Steps{
			{
				Config: testAccDnsRecordConfig(initial),
				Check:  testAccDnsRecordCheckAttrs(initial),
			},
			{
				Config:           testAccDnsRecordConfig(updated),
				Check:            testAccDnsRecordCheckAttrs(updated),
				ConfigPlanChecks: pt.CheckResourceAction(testDnsRecordResourceName, plancheck.ResourceActionUpdate),
			},
		},
	})
}

func TestDNSRecord_MissingAttributes(t *testing.T) {
	t.Parallel()
	testCases := map[string]func() string{
		"name":   testAccDnsRecordConfigMissingName,
		"record": testAccDnsRecordConfigMissingRecord,
		"type":   testAccDnsRecordConfigMissingType,
	}
	for k, v := range testCases {
		t.Run(fmt.Sprintf("missing %s", k), func(t *testing.T) {
			AcceptanceTest(t, AcceptanceTestCase{
				MinVersion: base.ControllerVersionDnsRecords,
				Steps: Steps{
					{
						Config:      v(),
						ExpectError: pt.MissingArgumentErrorRegex(k),
					},
				},
			})
		})
	}
}

func testAccCheckDNSRecordDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "unifi_dns_record" {
			continue
		}

		_, err := testClient.GetDNSRecord(context.Background(), "default", rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("DNS Record %s still exists", rs.Primary.ID)
		}
		// If we get a 404 error, that means the resource was deleted
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			continue
		}
		// For any other error, return it
		return err
	}

	return nil
}

func testAccDnsRecordConfig(tc dnsRecordTestCase) string {
	return testAccDnsRecordConfigWithResourceName("test", tc)
}

func testAccDnsRecordConfigMissingName() string {
	return `
resource "unifi_dns_record" "test" {
	record = "127.0.0.1"
	type = "A"
}
`
}

func testAccDnsRecordConfigMissingRecord() string {
	return `
resource "unifi_dns_record" "test" {
	name = "test.com"
	type = "A"
}
`
}

func testAccDnsRecordConfigMissingType() string {
	return `
resource "unifi_dns_record" "test" {
	name = "test.com"
	record = "127.0.0.1"
}
`
}

func testAccDnsRecordConfigWithResourceName(resourceName string, tc dnsRecordTestCase) string {
	var attrs string

	if tc.ttl != nil {
		attrs += fmt.Sprintf("\tttl = %d\n", *tc.ttl)
	}
	if tc.enabled != nil {
		attrs += fmt.Sprintf("\tenabled = %t\n", *tc.enabled)
	}
	if tc.priority != nil {
		attrs += fmt.Sprintf("\tpriority = %d\n", *tc.priority)
	}
	if tc.port != nil {
		attrs += fmt.Sprintf("\tport = %d\n", *tc.port)
	}
	if tc.weight != nil {
		attrs += fmt.Sprintf("\tweight = %d\n", *tc.weight)
	}

	return fmt.Sprintf(`
resource "unifi_dns_record" "%s" {
	name = "%s"
	record = "%s"
	type = "%s"
%s}
`, resourceName, tc.recordName, tc.record, tc.recordType, attrs)
}

func testAccDnsRecordCheckAttrs(tc dnsRecordTestCase) resource.TestCheckFunc {
	// expected default values
	var (
		ttl      = 0
		enabled  = true
		priority = 0
		port     = 0
		weight   = 0
	)

	if tc.ttl != nil {
		ttl = *tc.ttl
	}
	if tc.enabled != nil {
		enabled = *tc.enabled
	}
	if tc.priority != nil {
		priority = *tc.priority
	}
	if tc.port != nil {
		port = *tc.port
	}
	if tc.weight != nil {
		weight = *tc.weight
	}

	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(testDnsRecordResourceName, "name", tc.recordName),
		resource.TestCheckResourceAttr(testDnsRecordResourceName, "record", tc.record),
		resource.TestCheckResourceAttr(testDnsRecordResourceName, "type", tc.recordType),
		resource.TestCheckResourceAttr(testDnsRecordResourceName, "ttl", strconv.Itoa(ttl)),
		resource.TestCheckResourceAttr(testDnsRecordResourceName, "enabled", strconv.FormatBool(enabled)),
		resource.TestCheckResourceAttr(testDnsRecordResourceName, "priority", strconv.Itoa(priority)),
		resource.TestCheckResourceAttr(testDnsRecordResourceName, "port", strconv.Itoa(port)),
		resource.TestCheckResourceAttr(testDnsRecordResourceName, "weight", strconv.Itoa(weight)),
	}
	return resource.ComposeTestCheckFunc(checks...)
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

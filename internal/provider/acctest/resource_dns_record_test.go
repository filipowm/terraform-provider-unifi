package acctest

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/plancheck"

	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
)

const testDNSRecordResourceName = "unifi_dns_record.test"

// dnsLock serializes all DNS record + DNS data-source acceptance tests. The
// unifi_dns_records data source returns ALL records on the site, so tests that
// create or count records must not run concurrently or global counts become
// non-deterministic.
var dnsLock = &sync.Mutex{}

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
					Config: testAccDNSRecordConfig(tc),
					Check:  testAccDNSRecordCheckAttrs(tc),
				},
				pt.ImportStepWithSite(testDNSRecordResourceName),
			}

			AcceptanceTest(t, AcceptanceTestCase{
				MinVersion:   base.ControllerVersionDNSRecords,
				Lock:         dnsLock,
				Steps:        steps,
				CheckDestroy: testAccCheckDNSRecordDestroy,
			})
		})
	}
}

func TestDNSRecord_SRV(t *testing.T) {
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
				MinVersion:   base.ControllerVersionDNSRecords,
				Lock:         dnsLock,
				CheckDestroy: testAccCheckDNSRecordDestroy,
				Steps: Steps{
					{
						Config: testAccDNSRecordConfig(tc),
						Check:  testAccDNSRecordCheckAttrs(tc),
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
		MinVersion:   base.ControllerVersionDNSRecords,
		Lock:         dnsLock,
		CheckDestroy: testAccCheckDNSRecordDestroy,
		Steps: Steps{
			{
				Config: testAccDNSRecordConfig(initial),
				Check:  testAccDNSRecordCheckAttrs(initial),
			},
			{
				Config:           testAccDNSRecordConfig(updated),
				Check:            testAccDNSRecordCheckAttrs(updated),
				ConfigPlanChecks: pt.CheckResourceActions(testDNSRecordResourceName, plancheck.ResourceActionUpdate),
			},
		},
	})
}

func TestDNSRecord_MissingAttributes(t *testing.T) {
	testCases := map[string]func() string{
		"name":   testAccDNSRecordConfigMissingName,
		"record": testAccDNSRecordConfigMissingRecord,
		"type":   testAccDNSRecordConfigMissingType,
	}
	for k, v := range testCases {
		t.Run("missing "+k, func(t *testing.T) {
			AcceptanceTest(t, AcceptanceTestCase{
				MinVersion: base.ControllerVersionDNSRecords,
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

func testAccDNSRecordConfig(tc dnsRecordTestCase) string {
	return testAccDNSRecordConfigWithResourceName("test", tc)
}

func testAccDNSRecordConfigMissingName() string {
	return `
resource "unifi_dns_record" "test" {
	record = "127.0.0.1"
	type = "A"
}
`
}

func testAccDNSRecordConfigMissingRecord() string {
	return `
resource "unifi_dns_record" "test" {
	name = "test.com"
	type = "A"
}
`
}

func testAccDNSRecordConfigMissingType() string {
	return `
resource "unifi_dns_record" "test" {
	name = "test.com"
	record = "127.0.0.1"
}
`
}

func testAccDNSRecordConfigWithResourceName(resourceName string, tc dnsRecordTestCase) string {
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

func testAccDNSRecordCheckAttrs(tc dnsRecordTestCase) resource.TestCheckFunc {
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
		resource.TestCheckResourceAttr(testDNSRecordResourceName, "name", tc.recordName),
		resource.TestCheckResourceAttr(testDNSRecordResourceName, "record", tc.record),
		resource.TestCheckResourceAttr(testDNSRecordResourceName, "type", tc.recordType),
		resource.TestCheckResourceAttr(testDNSRecordResourceName, "ttl", strconv.Itoa(ttl)),
		resource.TestCheckResourceAttr(testDNSRecordResourceName, "enabled", strconv.FormatBool(enabled)),
		resource.TestCheckResourceAttr(testDNSRecordResourceName, "priority", strconv.Itoa(priority)),
		resource.TestCheckResourceAttr(testDNSRecordResourceName, "port", strconv.Itoa(port)),
		resource.TestCheckResourceAttr(testDNSRecordResourceName, "weight", strconv.Itoa(weight)),
	}
	return resource.ComposeTestCheckFunc(checks...)
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

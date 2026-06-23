package acctest

import (
	"context"
	"fmt"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"regexp"
	"strings"
	"testing"
)

const testAPGroupResourceName = "unifi_ap_group.test"

// FIXME: This test is currently skipped because the test environment does not support AP group adoption.
func TestAccAPGroup_basic(t *testing.T) {
	t.Skip("Skipping test due to lack of support for AP group adoption in test environment")
	rName := acctest.RandomWithPrefix("tfacc-apgroup")
	mac1 := "00:15:6d:00:00:01"

	AcceptanceTest(t, AcceptanceTestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccAPGroupConfig(rName, []string{mac1}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testAPGroupResourceName, "id"),
					resource.TestCheckResourceAttr(testAPGroupResourceName, "site", "default"),
					resource.TestCheckResourceAttr(testAPGroupResourceName, "name", rName),
					pt.TestCheckListResourceAttr(testAPGroupResourceName, "device_macs", mac1),
				),
				ConfigPlanChecks: pt.CheckResourceActions(testAPGroupResourceName, plancheck.ResourceActionCreate),
			},
			pt.ImportStepWithSite(testAPGroupResourceName),
		},
		CheckDestroy: testAccCheckAPGroupDestroy,
	})
}

func TestAccAPGroup_update(t *testing.T) {
	t.Skip("Skipping test due to lack of support for AP group adoption in test environment")
	rName := acctest.RandomWithPrefix("tfacc-apgroup")
	updatedName := acctest.RandomWithPrefix("tfacc-apgroup-updated")
	mac1 := "00:11:22:33:44:55"
	mac2 := "AA:BB:CC:DD:EE:FF"

	AcceptanceTest(t, AcceptanceTestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccAPGroupConfig(rName, []string{mac1}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testAPGroupResourceName, "id"),
					resource.TestCheckResourceAttr(testAPGroupResourceName, "name", rName),
					pt.TestCheckListResourceAttr(testAPGroupResourceName, "device_macs", mac1),
				),
			},
			{
				Config: testAccAPGroupConfig(updatedName, []string{mac1, mac2}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testAPGroupResourceName, "id"),
					resource.TestCheckResourceAttr(testAPGroupResourceName, "name", updatedName),
					pt.TestCheckListResourceAttr(testAPGroupResourceName, "device_macs", mac1, mac2),
				),
				ConfigPlanChecks: pt.CheckResourceActions(testAPGroupResourceName, plancheck.ResourceActionUpdate),
			},
		},
		CheckDestroy: testAccCheckAPGroupDestroy,
	})
}

func TestAccAPGroup_missingName(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Steps: []resource.TestStep{
			{
				Config:      testAccAPGroupConfigMissingName(),
				ExpectError: pt.MissingArgumentErrorRegex("name"),
			},
		},
	})
}

func TestAccAPGroup_missingDeviceMacs(t *testing.T) {
	rName := acctest.RandomWithPrefix("tfacc-apgroup")

	AcceptanceTest(t, AcceptanceTestCase{
		Steps: []resource.TestStep{
			{
				Config:      testAccAPGroupConfigMissingDeviceMacs(rName),
				ExpectError: pt.MissingArgumentErrorRegex("device_macs"),
			},
		},
	})
}

func TestAccAPGroup_emptyDeviceMacs(t *testing.T) {
	rName := acctest.RandomWithPrefix("tfacc-apgroup")

	AcceptanceTest(t, AcceptanceTestCase{
		Steps: []resource.TestStep{
			{
				Config:      testAccAPGroupConfig(rName, []string{}),
				ExpectError: regexp.MustCompile("must contain at least 1 elements"),
			},
		},
	})
}

func TestAccAPGroup_invalidMacAddress(t *testing.T) {
	rName := acctest.RandomWithPrefix("tfacc-apgroup")
	invalidMac := "invalid-mac-address"

	AcceptanceTest(t, AcceptanceTestCase{
		Steps: []resource.TestStep{
			{
				Config:      testAccAPGroupConfig(rName, []string{invalidMac}),
				ExpectError: regexp.MustCompile("invalid MAC address"),
			},
		},
	})
}

func testAccAPGroupConfig(name string, macs []string) string {
	macsStr := "[]"
	if len(macs) > 0 {
		macsStr = fmt.Sprintf("[\"%s\"]", strings.Join(macs, "\", \""))
	}

	return fmt.Sprintf(`
resource "unifi_ap_group" "test" {
	name        = %q
	device_macs = %s
}
`, name, macsStr)
}

func testAccAPGroupConfigMissingName() string {
	return `
resource "unifi_ap_group" "test" {
	device_macs = ["00:11:22:33:44:55"]
}
`
}

func testAccAPGroupConfigMissingDeviceMacs(name string) string {
	return fmt.Sprintf(`
resource "unifi_ap_group" "test" {
	name = %q
}
`, name)
}

func testAccCheckAPGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "unifi_ap_group" {
			continue
		}

		_, err := testClient.GetAPGroup(context.Background(), "default", rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("AP Group %s still exists", rs.Primary.ID)
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

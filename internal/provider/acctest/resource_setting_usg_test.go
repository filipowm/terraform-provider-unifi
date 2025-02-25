package acctest

import (
	"fmt"
	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"
	"regexp"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// using an additional lock to the one around the resource to avoid deadlocking accidentally
var settingUsgLock = sync.Mutex{}

func TestAccSettingUsg_mdns_v6(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			pt.PreCheck(t)
			PreCheckVersionConstraint(t, "< 7")
			settingUsgLock.Lock()
			t.Cleanup(func() {
				settingUsgLock.Unlock()
			})
		},
		ProtoV6ProviderFactories: providers,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgConfig_mdns(true),
				Check:  resource.ComposeTestCheckFunc(),
			},
			pt.ImportStep("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgConfig_mdns(false),
				Check:  resource.ComposeTestCheckFunc(),
			},
			pt.ImportStep("unifi_setting_usg.test"),
			{
				Config: testAccSettingUsgConfig_mdns(true),
				Check:  resource.ComposeTestCheckFunc(),
			},
			pt.ImportStep("unifi_setting_usg.test"),
		},
	})
}

func TestAccSettingUsg_mdns_v7(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			pt.PreCheck(t)
			PreCheckVersionConstraint(t, ">= 7")
			settingUsgLock.Lock()
			t.Cleanup(func() {
				settingUsgLock.Unlock()
			})
		},
		ProtoV6ProviderFactories: providers,
		Steps: []resource.TestStep{
			{
				Config:      testAccSettingUsgConfig_mdns(true),
				ExpectError: regexp.MustCompile("multicast_dns_enabled is not supported"),
			},
		},
	})
}

func TestAccSettingUsg_dhcpRelay(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			pt.PreCheck(t)
			settingUsgLock.Lock()
			t.Cleanup(func() {
				settingUsgLock.Unlock()
			})
		},
		ProtoV6ProviderFactories: providers,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgConfig_dhcpRelay(),
				Check:  resource.ComposeTestCheckFunc(),
			},
			pt.ImportStep("unifi_setting_usg.test"),
		},
	})
}

func TestAccSettingUsg_site(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			pt.PreCheck(t)
			settingUsgLock.Lock()
			t.Cleanup(func() {
				settingUsgLock.Unlock()
			})
		},
		ProtoV6ProviderFactories: providers,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgConfig_site(),
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				ResourceName:      "unifi_setting_usg.test",
				ImportState:       true,
				ImportStateIdFunc: pt.SiteAndIDImportStateIDFunc("unifi_setting_usg.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSettingUsgConfig_mdns(mdns bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_usg" "test" {
	multicast_dns_enabled = %t
}
`, mdns)
}

func testAccSettingUsgConfig_dhcpRelay() string {
	return `
resource "unifi_setting_usg" "test" {
	dhcp_relay_servers = [
		"10.1.2.3",
		"10.1.2.4",
	]
}
`
}

func testAccSettingUsgConfig_site() string {
	return `
resource "unifi_site" "test" {
	description = "test"
}

resource "unifi_setting_usg" "test" {
	site = unifi_site.test.name
}
`
}

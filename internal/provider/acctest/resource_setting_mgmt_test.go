package acctest

import (
	"sync"
	"testing"

	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

var settingMgmtLock = sync.Mutex{}

const testSettingMgmtResourceName = "unifi_setting_mgmt.test"

func TestAccSettingMgmt_basic(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingMgmtLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingMgmtConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testSettingMgmtResourceName, "id"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "site", "default"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "auto_upgrade", "true"),
				),
				ConfigPlanChecks: pt.CheckResourceActions(testSettingMgmtResourceName, plancheck.ResourceActionCreate),
			},
			pt.ImportStepWithSite(testSettingMgmtResourceName),
		},
	})
}

func TestAccSettingMgmt_site(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingMgmtLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingMgmtConfig_site(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testSettingMgmtResourceName, "id"),
					resource.TestCheckResourceAttrPair(testSettingMgmtResourceName, "site", "unifi_site.test", "name"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "auto_upgrade", "true"),
				),
				ConfigPlanChecks: pt.CheckResourceActions(testSettingMgmtResourceName, plancheck.ResourceActionCreate),
			},
			pt.ImportStepWithSite(testSettingMgmtResourceName),
		},
	})
}

func TestAccSettingMgmt_sshKeys(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingMgmtLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingMgmtConfig_sshKeys(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testSettingMgmtResourceName, "id"),
					resource.TestCheckResourceAttrPair(testSettingMgmtResourceName, "site", "unifi_site.test", "name"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.#", "1"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.0.name", "Test key"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.0.type", "ssh-rsa"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.0.comment", "test@example.com"),
				),
				ConfigPlanChecks: pt.CheckResourceActions(testSettingMgmtResourceName, plancheck.ResourceActionCreate),
			},
			pt.ImportStepWithSite(testSettingMgmtResourceName),
		},
	})
}

func TestAccSettingMgmt_fullConfig(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 7.3",
		Lock:              &settingMgmtLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingMgmtConfig_fullConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testSettingMgmtResourceName, "id"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "site", "default"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "auto_upgrade", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "auto_upgrade_hour", "3"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "advanced_feature_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "alert_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "boot_sound", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "debug_tools_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "direct_connect_enabled", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "led_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "outdoor_mode_enabled", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "unifi_idp_enabled", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "wifiman_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_auth_password_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_bind_wildcard", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_username", "admin"),
				),
				ConfigPlanChecks: pt.CheckResourceActions(testSettingMgmtResourceName, plancheck.ResourceActionCreate),
			},
			pt.ImportStepWithSite(testSettingMgmtResourceName),
		},
	})
}

func TestAccSettingMgmt_update(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingMgmtLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingMgmtConfig_initialConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testSettingMgmtResourceName, "id"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "auto_upgrade", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "auto_upgrade_hour", "3"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "led_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_enabled", "true"),
				),
			},
			pt.ImportStepWithSite(testSettingMgmtResourceName),
			{
				Config: testAccSettingMgmtConfig_updatedConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testSettingMgmtResourceName, "id"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "auto_upgrade", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "auto_upgrade_hour", "5"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "led_enabled", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_enabled", "false"),
				),
			},
		},
	})
}

func TestAccSettingMgmt_sshCredentials(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingMgmtLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingMgmtConfig_sshCredentials(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testSettingMgmtResourceName, "id"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_auth_password_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_username", "admin"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_password", "securepassword"),
				),
			},
			pt.ImportStepWithSite(testSettingMgmtResourceName),
		},
	})
}

func TestAccSettingMgmt_cornerCases(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingMgmtLock,
		Steps: []resource.TestStep{
			{
				// Initial configuration with specific values
				Config: testAccSettingMgmtConfig_cornerCasesInitial(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testSettingMgmtResourceName, "id"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "site", "default"),
					// Boolean attributes - initial values
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "auto_upgrade", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "alert_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "boot_sound", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "direct_connect_enabled", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "led_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "outdoor_mode_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "unifi_idp_enabled", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "wifiman_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_auth_password_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_bind_wildcard", "true"),
					// Numeric values - initial
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "auto_upgrade_hour", "3"),
				),
			},
			pt.ImportStepWithSite(testSettingMgmtResourceName),
			{
				// Toggle all boolean values and change numeric values
				Config: testAccSettingMgmtConfig_cornerCasesToggled(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testSettingMgmtResourceName, "id"),
					// Boolean attributes - toggled values
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "auto_upgrade", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "alert_enabled", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "boot_sound", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "direct_connect_enabled", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "led_enabled", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "outdoor_mode_enabled", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "unifi_idp_enabled", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "wifiman_enabled", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_enabled", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_auth_password_enabled", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_bind_wildcard", "false"),
					// Numeric values - changed
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "auto_upgrade_hour", "23"),
				),
			},
			pt.ImportStepWithSite(testSettingMgmtResourceName),
			{
				// Test boundary values for numeric fields and mixed boolean values
				Config: testAccSettingMgmtConfig_cornerCasesBoundary(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testSettingMgmtResourceName, "id"),
					// Mixed boolean values
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "auto_upgrade", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "alert_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "boot_sound", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "direct_connect_enabled", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "led_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "outdoor_mode_enabled", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "unifi_idp_enabled", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "wifiman_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_auth_password_enabled", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_bind_wildcard", "true"),
					// Boundary value for auto_upgrade_hour (1 - minimum value)
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "auto_upgrade_hour", "0"),
				),
			},
		},
	})
}

func TestAccSettingMgmt_sshKeyManagement(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingMgmtLock,
		Steps: []resource.TestStep{
			{
				// Initial configuration with one SSH key
				Config: testAccSettingMgmtConfig_sshKeyManagementInitial(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testSettingMgmtResourceName, "id"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.#", "1"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.0.name", "Initial key"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.0.type", "ssh-rsa"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.0.key", "AAAAB3NzaC1yc2EAAAADAQABAAABAQC0"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.0.comment", "initial@example.com"),
				),
			},
			pt.ImportStepWithSite(testSettingMgmtResourceName),
			{
				// Add a second SSH key and modify the first one
				Config: testAccSettingMgmtConfig_sshKeyManagementModified(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testSettingMgmtResourceName, "id"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.#", "2"),
					// First key is modified
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.0.name", "Modified key"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.0.type", "ssh-rsa"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.0.key", "AAAAB3NzaC1yc2EAAAADAQABAAABAQC1"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.0.comment", "modified@example.com"),
					// Second key is added
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.1.name", "Additional key"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.1.type", "ssh-ed25519"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.1.key", "AAAAC3NzaC1lZDI1NTE5AAAAIG"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.1.comment", "additional@example.com"),
				),
			},
			pt.ImportStepWithSite(testSettingMgmtResourceName),
			{
				// Remove the first key, keep the second key
				Config: testAccSettingMgmtConfig_sshKeyManagementRemoved(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testSettingMgmtResourceName, "id"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.#", "1"),
					// Only the second key remains
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.0.name", "Additional key"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.0.type", "ssh-ed25519"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.0.key", "AAAAC3NzaC1lZDI1NTE5AAAAIG"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.0.comment", "additional@example.com"),
				),
			},
			pt.ImportStepWithSite(testSettingMgmtResourceName),
			{
				// Remove all SSH keys
				Config: testAccSettingMgmtConfig_sshKeyManagementNoKeys(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testSettingMgmtResourceName, "id"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.#", "0"),
				),
			},
			pt.ImportStepWithSite(testSettingMgmtResourceName),
		},
	})
}

func TestAccSettingMgmt_sshAuthModes(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: &settingMgmtLock,
		Steps: []resource.TestStep{
			{
				// Initial configuration with SSH password authentication enabled
				Config: testAccSettingMgmtConfig_sshAuthModesPasswordOnly(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testSettingMgmtResourceName, "id"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_auth_password_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_username", "admin"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_password", "password123"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.#", "0"),
				),
			},
			pt.ImportStepWithSite(testSettingMgmtResourceName),
			{
				// Switch to SSH key authentication only
				Config: testAccSettingMgmtConfig_sshAuthModesKeyOnly(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testSettingMgmtResourceName, "id"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_auth_password_enabled", "false"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.#", "1"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.0.name", "Auth key"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.0.type", "ssh-rsa"),
				),
			},
			pt.ImportStepWithSite(testSettingMgmtResourceName),
			{
				// Enable both authentication methods
				Config: testAccSettingMgmtConfig_sshAuthModesBoth(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testSettingMgmtResourceName, "id"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_auth_password_enabled", "true"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_username", "admin"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_password", "newpassword"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.#", "1"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.0.name", "Auth key"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_key.0.type", "ssh-rsa"),
				),
			},
			pt.ImportStepWithSite(testSettingMgmtResourceName),
			{
				// Disable SSH entirely
				Config: testAccSettingMgmtConfig_sshAuthModesDisabled(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testSettingMgmtResourceName, "id"),
					resource.TestCheckResourceAttr(testSettingMgmtResourceName, "ssh_enabled", "false"),
				),
			},
			pt.ImportStepWithSite(testSettingMgmtResourceName),
		},
	})
}

func testAccSettingMgmtConfig_basic() string {
	return `
resource "unifi_setting_mgmt" "test" {
	auto_upgrade = true
}
`
}

func testAccSettingMgmtConfig_site() string {
	return `
resource "unifi_site" "test" {
	description = "test"
}

resource "unifi_setting_mgmt" "test" {
	site = unifi_site.test.name
	auto_upgrade = true
}
`
}

func testAccSettingMgmtConfig_sshKeys() string {
	return `
resource "unifi_site" "test" {
	description = "test"
}

resource "unifi_setting_mgmt" "test" {
	site = unifi_site.test.name
	ssh_enabled = true
	ssh_key {
		name = "Test key"
		type = "ssh-rsa"
		key = "AAAAB3NzaC1yc2EAAAADAQABAAACAQDNWqT8zvVtmaks7sLlP+hmWmJVmruyNU9uk8JpLTX0oE+r9hjePsXCThTrft7s+vlaj+bLr8Yf5//TT8KS7LB/YIp2O3jPomOz9A4hIsG5R6FLfSggzQP4a7QSlNLCm/6WjKHP9DhRb7trnFz+KkCNmCVKLZgiyeUm2LydVKJ2QncHopA5yomtSpmb6x66zaKr+DbwzHC13WIEms5Ros0N9pEOcAghsSEVL42bfGBfSH37R+Kaw0nhWei4Y25jO66xsbtyZKoiF1+XXXBuEi77Tv7iQGHHOFRqNKKfGI1QhYvwlcjdzh9wu7Gtzeyh/+jpF8mwCLtFKle+W/zSs+lHCuCihvQEQtCIpZL5FapvxfxPZQJWL5RgsL9jieUaoF8EsWAOM83BCSZa/FB1RyfKdy4f7BQtDCKIm3nD5paCJSfS6DSw1TMvaFPeJLG3PuyHRbNvbVLmHRl9lK03na6/R9JX06nBUuPdP+FLjIZsyZz1DOUSDjCWHFk0+Ne2uEinV7SkOoxC6E2NxqlY/SyMnWZS+p95Zx6yOlNqB9sQ+Q4/YLGY5mUmqJrHPlH6LjXfudybKHMZUuVRF1NX3ESue8NSKc0SlJDQUXtJ9wkjjX1wAWvXCDwI72jtC86r/wzw+mcIfpks3jHQrOhpwCRmQL4vAs5DztA3hKxkgElYaw=="
		comment = "test@example.com"
	}
}
`
}

func testAccSettingMgmtConfig_fullConfig() string {
	return `
resource "unifi_setting_mgmt" "test" {
	auto_upgrade = true
	auto_upgrade_hour = 3
	advanced_feature_enabled = true
	alert_enabled = true
	boot_sound = false
	debug_tools_enabled = true
	direct_connect_enabled = false
	led_enabled = true
	outdoor_mode_enabled = false
	unifi_idp_enabled = false
	wifiman_enabled = true
	ssh_enabled = true
	ssh_auth_password_enabled = true
	ssh_bind_wildcard = false
	ssh_username = "admin"
}
`
}

func testAccSettingMgmtConfig_initialConfig() string {
	return `
resource "unifi_setting_mgmt" "test" {
	auto_upgrade = true
	auto_upgrade_hour = 3
	led_enabled = true
	ssh_enabled = true
}
`
}

func testAccSettingMgmtConfig_updatedConfig() string {
	return `
resource "unifi_setting_mgmt" "test" {
	auto_upgrade = false
	auto_upgrade_hour = 5
	led_enabled = false
	ssh_enabled = false
}
`
}

func testAccSettingMgmtConfig_sshCredentials() string {
	return `
resource "unifi_setting_mgmt" "test" {
	ssh_enabled = true
	ssh_auth_password_enabled = true
	ssh_username = "admin"
	ssh_password = "securepassword"
}
`
}

func testAccSettingMgmtConfig_cornerCasesInitial() string {
	return `
resource "unifi_setting_mgmt" "test" {
	auto_upgrade = true
	auto_upgrade_hour = 3
	alert_enabled = true
	boot_sound = true
	direct_connect_enabled = false
	led_enabled = true
	outdoor_mode_enabled = true
	unifi_idp_enabled = false
	wifiman_enabled = true
	ssh_enabled = true
	ssh_auth_password_enabled = true
	ssh_bind_wildcard = true
}
`
}

func testAccSettingMgmtConfig_cornerCasesToggled() string {
	return `
resource "unifi_setting_mgmt" "test" {
	auto_upgrade = false
	auto_upgrade_hour = 23
	alert_enabled = false
	boot_sound = false
	direct_connect_enabled = false
	led_enabled = false
	outdoor_mode_enabled = false
	unifi_idp_enabled = false
	wifiman_enabled = false
	ssh_enabled = false
	ssh_auth_password_enabled = false
	ssh_bind_wildcard = false
}
`
}

func testAccSettingMgmtConfig_cornerCasesBoundary() string {
	return `
resource "unifi_setting_mgmt" "test" {
	auto_upgrade = true
	auto_upgrade_hour = 0
	alert_enabled = true
	boot_sound = false
	direct_connect_enabled = false
	led_enabled = true
	outdoor_mode_enabled = false
	unifi_idp_enabled = false
	wifiman_enabled = true
	ssh_enabled = true
	ssh_auth_password_enabled = false
	ssh_bind_wildcard = true
}
`
}

func testAccSettingMgmtConfig_sshKeyManagementInitial() string {
	return `
resource "unifi_setting_mgmt" "test" {
	ssh_enabled = true
	ssh_key {
		name = "Initial key"
		type = "ssh-rsa"
		key = "AAAAB3NzaC1yc2EAAAADAQABAAABAQC0"
		comment = "initial@example.com"
	}
}
`
}

func testAccSettingMgmtConfig_sshKeyManagementModified() string {
	return `
resource "unifi_setting_mgmt" "test" {
	ssh_enabled = true
	ssh_key {
		name = "Modified key"
		type = "ssh-rsa"
		key = "AAAAB3NzaC1yc2EAAAADAQABAAABAQC1"
		comment = "modified@example.com"
	}
	ssh_key {
		name = "Additional key"
		type = "ssh-ed25519"
		key = "AAAAC3NzaC1lZDI1NTE5AAAAIG"
		comment = "additional@example.com"
	}
}
`
}

func testAccSettingMgmtConfig_sshKeyManagementRemoved() string {
	return `
resource "unifi_setting_mgmt" "test" {
	ssh_enabled = true
	ssh_key {
		name = "Additional key"
		type = "ssh-ed25519"
		key = "AAAAC3NzaC1lZDI1NTE5AAAAIG"
		comment = "additional@example.com"
	}
}
`
}

func testAccSettingMgmtConfig_sshKeyManagementNoKeys() string {
	return `
resource "unifi_setting_mgmt" "test" {
	ssh_enabled = true
}
`
}

func testAccSettingMgmtConfig_sshAuthModesPasswordOnly() string {
	return `
resource "unifi_setting_mgmt" "test" {
	ssh_enabled = true
	ssh_auth_password_enabled = true
	ssh_username = "admin"
	ssh_password = "password123"
}
`
}

func testAccSettingMgmtConfig_sshAuthModesKeyOnly() string {
	return `
resource "unifi_setting_mgmt" "test" {
	ssh_enabled = true
	ssh_auth_password_enabled = false
	ssh_key {
		name = "Auth key"
		type = "ssh-rsa"
		key = "AAAAB3NzaC1yc2EAAAADAQABAAABAQC0"
		comment = "auth@example.com"
	}
}
`
}

func testAccSettingMgmtConfig_sshAuthModesBoth() string {
	return `
resource "unifi_setting_mgmt" "test" {
	ssh_enabled = true
	ssh_auth_password_enabled = true
	ssh_username = "admin"
	ssh_password = "newpassword"
	ssh_key {
		name = "Auth key"
		type = "ssh-rsa"
		key = "AAAAB3NzaC1yc2EAAAADAQABAAABAQC0"
		comment = "auth@example.com"
	}
}
`
}

func testAccSettingMgmtConfig_sshAuthModesDisabled() string {
	return `
resource "unifi_setting_mgmt" "test" {
	ssh_enabled = false
}
`
}

package acctest

import (
	"fmt"
	"sync"
	"testing"

	pt "github.com/filipowm/terraform-provider-unifi/internal/provider/testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var settingGuestAccessLock = &sync.Mutex{}

func TestAccSettingGuestAccess_basic(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "none"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_use_hostname", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_hostname", "guest.example.com"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "template_engine", "angular"),

					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "expire", "60"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "expire_number", "1"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "expire_unit", "60"),

					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "ec_enabled", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_guest_access.test"),
			{
				Config: testAccSettingGuestAccessConfig_basicUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_enabled", "false"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "template_engine", "jsp"),

					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "expire", "1440"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "expire_number", "1"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "expire_unit", "1440"),

					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "ec_enabled", "false"),
				),
			},
		},
	})
}

func TestAccSettingGuestAccess_customAuth(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_customAuth("192.168.1.1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "custom"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "custom_ip", "192.168.1.1"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_customAuth("192.168.1.2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "custom"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "custom_ip", "192.168.1.2"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_auth("none"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "none"),
					resource.TestCheckNoResourceAttr("unifi_setting_guest_access.test", "custom_ip"),
				),
			},
		},
	})
}

func TestAccSettingGuestAccess_password(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_password("pass1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "password", "pass1"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "password_enabled", "true"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_password("pass2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "password", "pass2"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "password_enabled", "true"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_auth("hotspot"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "password_enabled", "false"),
				),
			},
		},
	})
}

func TestAccSettingGuestAccess_voucher(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_voucher(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "voucher_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "voucher_customized", "false"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_voucherCustomized(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "voucher_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "voucher_customized", "true"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_voucher(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "voucher_enabled", "false"),
				),
			},
		},
	})
}

func TestAccSettingGuestAccess_allowedSubnet(t *testing.T) {
	t.Skip("api.err.InvalidPayload; api.err.InvalidKey: ")
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_allowedSubnet("192.168.1.0/24"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "allowed_subnet", "192.168.1.0/24"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_allowedSubnet("10.0.0.0/24"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "allowed_subnet", "10.0.0.0/24"),
				),
			},
		},
	})
}

func TestAccSettingGuestAccess_paymentPaypal(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_paymentPaypal(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "paypal"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "paypal.username", "test@example.com"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "paypal.password", "paypal-password"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "paypal.signature", "paypal-signature"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "paypal.use_sandbox", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_guest_access.test"),
			{
				Config: testAccSettingGuestAccessConfig_paymentPaypal(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "paypal"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "paypal.username", "test@example.com"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "paypal.password", "paypal-password"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "paypal.signature", "paypal-signature"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "paypal.use_sandbox", "false"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_paymentPaypalUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "paypal"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "paypal.username", "updated@example.com"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "paypal.password", "updated-password"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "paypal.signature", "updated-signature"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "paypal.use_sandbox", "true"),
				),
			},
		},
	})
}

func TestAccSettingGuestAccess_paymentStripe(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_paymentStripe("stripe-api-key"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "stripe"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "stripe.api_key", "stripe-api-key"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_guest_access.test"),
			{
				Config: testAccSettingGuestAccessConfig_paymentStripe("updated-stripe-api-key"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "stripe"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "stripe.api_key", "updated-stripe-api-key"),
				),
			},
		},
	})
}

func TestAccSettingGuestAccess_paymentAuthorize(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_paymentAuthorize(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "authorize"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "authorize.login_id", "authorize-login"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "authorize.transaction_key", "authorize-transaction-key"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "authorize.use_sandbox", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_guest_access.test"),
			{
				Config: testAccSettingGuestAccessConfig_paymentAuthorize(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "authorize"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "authorize.login_id", "authorize-login"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "authorize.transaction_key", "authorize-transaction-key"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "authorize.use_sandbox", "false"),
				),
			},
		},
	})
}

func TestAccSettingGuestAccess_paymentQuickpay(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_paymentQuickpay(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "quickpay"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "quickpay.agreement_id", "quickpay-agreement"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "quickpay.api_key", "quickpay-api-key"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "quickpay.merchant_id", "quickpay-merchant"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "quickpay.use_sandbox", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_guest_access.test"),
			{
				Config: testAccSettingGuestAccessConfig_paymentQuickpay(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "quickpay"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "quickpay.agreement_id", "quickpay-agreement"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "quickpay.api_key", "quickpay-api-key"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "quickpay.merchant_id", "quickpay-merchant"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "quickpay.use_sandbox", "false"),
				),
			},
		},
	})
}

func TestAccSettingGuestAccess_paymentMerchantWarrior(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_paymentMerchantWarrior(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "merchantwarrior"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "merchant_warrior.api_key", "mw-api-key"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "merchant_warrior.api_passphrase", "mw-passphrase"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "merchant_warrior.merchant_uuid", "mw-merchant-id"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "merchant_warrior.use_sandbox", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_guest_access.test"),
			{
				Config: testAccSettingGuestAccessConfig_paymentMerchantWarrior(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "merchantwarrior"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "merchant_warrior.api_key", "mw-api-key"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "merchant_warrior.api_passphrase", "mw-passphrase"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "merchant_warrior.merchant_uuid", "mw-merchant-id"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "merchant_warrior.use_sandbox", "false"),
				),
			},
		},
	})
}

func TestAccSettingGuestAccess_paymentIPpay(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_paymentIPpay(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "ippay"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "ippay.terminal_id", "ippay-terminal"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "ippay.use_sandbox", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_guest_access.test"),
			{
				Config: testAccSettingGuestAccessConfig_paymentIPpay(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "ippay"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "ippay.terminal_id", "ippay-terminal"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "ippay.use_sandbox", "false"),
				),
			},
		},
	})
}

func TestAccSettingGuestAccess_paymentSwitchGateways(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_paymentPaypal(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "paypal"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_paymentStripe("stripe-api-key"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "stripe"),
					resource.TestCheckNoResourceAttr("unifi_setting_guest_access.test", "paypal.username"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_paymentAuthorize(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "authorize"),
					resource.TestCheckNoResourceAttr("unifi_setting_guest_access.test", "stripe.api_key"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_paymentQuickpay(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "quickpay"),
					resource.TestCheckNoResourceAttr("unifi_setting_guest_access.test", "authorize.login_id"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_paymentMerchantWarrior(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "merchantwarrior"),
					resource.TestCheckNoResourceAttr("unifi_setting_guest_access.test", "quickpay.api_key"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_paymentIPpay(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "ippay"),
					resource.TestCheckNoResourceAttr("unifi_setting_guest_access.test", "merchant_warrior.api_key"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_auth("hotspot"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_enabled", "false"),
					resource.TestCheckNoResourceAttr("unifi_setting_guest_access.test", "payment_gateway"),
				),
			},
		},
	})
}

func TestAccSettingGuestAccess_redirect(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_redirect("https://example.com", true, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.use_https", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.to_https", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.url", "https://example.com"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_guest_access.test"),
			{
				Config: testAccSettingGuestAccessConfig_redirect("https://updated-example.com", false, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.use_https", "false"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.to_https", "false"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.url", "https://updated-example.com"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_auth("none"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect_enabled", "false"),
					resource.TestCheckNoResourceAttr("unifi_setting_guest_access.test", "redirect"),
				),
			},
		},
	})
}

func TestAccSettingGuestAccess_facebook(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_facebook("facebook-app-id", "facebook-app-secret", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook.app_id", "facebook-app-id"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook.app_secret", "facebook-app-secret"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook.scope_email", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_guest_access.test"),
			{
				Config: testAccSettingGuestAccessConfig_facebook("updated-app-id", "updated-app-secret", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook.app_id", "updated-app-id"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook.app_secret", "updated-app-secret"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook.scope_email", "false"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_auth("none"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "none"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook_enabled", "false"),
					resource.TestCheckNoResourceAttr("unifi_setting_guest_access.test", "facebook"),
				),
			},
		},
	})
}

func TestAccSettingGuestAccess_google(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_google("google-client-id", "google-client-secret", "example.com", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "google_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "google.client_id", "google-client-id"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "google.client_secret", "google-client-secret"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "google.domain", "example.com"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "google.scope_email", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_guest_access.test"),
			{
				Config: testAccSettingGuestAccessConfig_google("updated-client-id", "updated-client-secret", "", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "google_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "google.client_id", "updated-client-id"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "google.client_secret", "updated-client-secret"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "google.domain", ""),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "google.scope_email", "false"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_auth("none"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "none"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "google_enabled", "false"),
					resource.TestCheckNoResourceAttr("unifi_setting_guest_access.test", "google"),
				),
			},
		},
	})
}

func TestAccSettingGuestAccess_radius(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_radius("chap", "radius-profile-id", true, 3799),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "radius_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "radius.auth_type", "chap"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "radius.profile_id", "radius-profile-id"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "radius.disconnect_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "radius.disconnect_port", "3799"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_guest_access.test"),
			{
				Config: testAccSettingGuestAccessConfig_radius("mschapv2", "updated-profile-id", false, 1812),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "radius_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "radius.auth_type", "mschapv2"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "radius.profile_id", "updated-profile-id"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "radius.disconnect_enabled", "false"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "radius.disconnect_port", "1812"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_auth("none"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "none"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "radius_enabled", "false"),
					resource.TestCheckNoResourceAttr("unifi_setting_guest_access.test", "radius"),
				),
			},
		},
	})
}

func TestAccSettingGuestAccess_wechat(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_wechat("wechat-app-id", "wechat-app-secret", "wechat-secret-key", "wechat-shop-id"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "wechat_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "wechat.app_id", "wechat-app-id"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "wechat.app_secret", "wechat-app-secret"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "wechat.secret_key", "wechat-secret-key"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "wechat.shop_id", "wechat-shop-id"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_guest_access.test"),
			{
				Config: testAccSettingGuestAccessConfig_wechat("updated-app-id", "updated-app-secret", "updated-secret-key", "updated-shop-id"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "wechat_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "wechat.app_id", "updated-app-id"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "wechat.app_secret", "updated-app-secret"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "wechat.secret_key", "updated-secret-key"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "wechat.shop_id", "updated-shop-id"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_auth("none"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "none"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "wechat_enabled", "false"),
					resource.TestCheckNoResourceAttr("unifi_setting_guest_access.test", "wechat"),
				),
			},
		},
	})
}

func TestAccSettingGuestAccess_facebookWifi(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_facebookWifi("gateway-id", "gateway-name", "gateway-secret", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "facebook_wifi"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook_wifi.gateway_id", "gateway-id"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook_wifi.gateway_name", "gateway-name"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook_wifi.gateway_secret", "gateway-secret"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook_wifi.block_https", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_guest_access.test"),
			{
				Config: testAccSettingGuestAccessConfig_facebookWifi("updated-gateway-id", "updated-gateway-name", "updated-gateway-secret", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "facebook_wifi"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook_wifi.gateway_id", "updated-gateway-id"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook_wifi.gateway_name", "updated-gateway-name"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook_wifi.gateway_secret", "updated-gateway-secret"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook_wifi.block_https", "false"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_auth("none"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "none"),
					resource.TestCheckNoResourceAttr("unifi_setting_guest_access.test", "facebook_wifi"),
				),
			},
		},
	})
}

func TestAccSettingGuestAccess_restrictedDNS(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_restrictedDNS([]string{"8.8.8.8", "1.1.1.1"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_servers.#", "2"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_servers.0", "8.8.8.8"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_servers.1", "1.1.1.1"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_guest_access.test"),
			{
				Config: testAccSettingGuestAccessConfig_restrictedDNS([]string{"8.8.4.4", "1.0.0.1", "9.9.9.9"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_servers.#", "3"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_servers.0", "8.8.4.4"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_servers.1", "1.0.0.1"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_servers.2", "9.9.9.9"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_restrictedDNS([]string{}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_enabled", "false"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_servers.#", "0"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_enabled", "false"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_servers.#", "0"),
				),
			},
		},
	})
}

func TestAccSettingGuestAccess_portalCustomizationPostVersion74(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 7.4",
		Lock:              settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_portalCustomizationBasicPost74(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.customized", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.bg_type", "color"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.box_radius", "12"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.button_text", "Login"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.authentication_text", "Please authenticate to access the internet"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.success_text", "You are now connected!"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.logo_position", "center"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.logo_size", "150"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_portalCustomizationImagesPost74(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.customized", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.bg_type", "image"),
					resource.TestCheckResourceAttrSet("unifi_setting_guest_access.test", "portal_customization.bg_image_file_id"),
					resource.TestCheckResourceAttrSet("unifi_setting_guest_access.test", "portal_customization.logo_file_id"),
				),
			},
		},
	})
}

func TestAccSettingGuestAccess_portalCustomization(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				// Initial configuration with color theme and basic settings
				Config: testAccSettingGuestAccessConfig_portalCustomizationBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.customized", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.bg_color", "#f5f5f5"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.box_color", "#ffffff"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.box_opacity", "90"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.title", "Guest WiFi Portal"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.tos_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.tos", "By using this WiFi service, you agree to our terms and conditions."),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.box_text_color", "#333333"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.text_color", "#222222"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.link_color", "#0066cc"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.box_link_color", "#0055aa"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.button_color", "#4CAF50"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.button_text_color", "#ffffff"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.languages.#", "3"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.languages.0", "en"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.languages.1", "es"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.languages.2", "fr"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_guest_access.test"),
			{
				// Update with gallery background and text customizations
				Config: testAccSettingGuestAccessConfig_portalCustomizationGallery(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.customized", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.unsplash_author_name", "John Doe"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.unsplash_author_username", "johndoe"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.welcome_text_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.welcome_text", "Welcome to our WiFi network!"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.welcome_text_position", "above_boxes"),
				),
			},
			{
				// Disable customization
				Config: testAccSettingGuestAccessConfig_portalCustomizationDisabled(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.customized", "false"),
				),
			},
			{
				// Back to basic configuration
				Config: testAccSettingGuestAccessConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.customized", "false"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.%", "29"),
				),
			},
		},
	})
}

func testAccSettingGuestAccessConfig_basic() string {
	return `
resource "unifi_setting_guest_access" "test" {
  auth           = "none"
  portal_enabled = true
  portal_use_hostname = true
  portal_hostname    = "guest.example.com"
  template_engine = "angular"
  expire        = 60
  expire_number = 1
  expire_unit   = 60
  ec_enabled = true
}
`
}

func testAccSettingGuestAccessConfig_basicUpdated() string {
	return `
resource "unifi_setting_guest_access" "test" {
  auth           = "hotspot"
  portal_enabled = false
  template_engine = "jsp"
  expire        = 1440
  expire_number = 1
  expire_unit   = 1440
  ec_enabled = false
}
`
}

func testAccSettingGuestAccessConfig_auth(auth string) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth = "%s"
}
`, auth)
}

func testAccSettingGuestAccessConfig_customAuth(ip string) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth     = "custom"
  custom_ip = %q
}
`, ip)
}

func testAccSettingGuestAccessConfig_password(password string) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth     = "hotspot"
  password = %q
}
`, password)
}

func testAccSettingGuestAccessConfig_voucher(enabled bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth            = "hotspot"
  voucher_enabled = %t
}
`, enabled)
}

func testAccSettingGuestAccessConfig_voucherCustomized() string {
	return `
resource "unifi_setting_guest_access" "test" {
  auth               = "hotspot"
  voucher_enabled    = true
  voucher_customized = true
}
`
}

func testAccSettingGuestAccessConfig_allowedSubnet(subnet string) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  allowed_subnet = %q
}
`, subnet)
}

func testAccSettingGuestAccessConfig_paymentPaypal(useSandbox bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth            = "hotspot"
  payment_gateway = "paypal"
  paypal = {
    username    = "test@example.com"
    password    = "paypal-password"
    signature   = "paypal-signature"
    use_sandbox = %t
  }
}
`, useSandbox)
}

func testAccSettingGuestAccessConfig_paymentPaypalUpdated() string {
	return `
resource "unifi_setting_guest_access" "test" {
  auth            = "hotspot"
  payment_gateway = "paypal"
  paypal = {
    username    = "updated@example.com"
    password    = "updated-password"
    signature   = "updated-signature"
    use_sandbox = true
  }
}
`
}

func testAccSettingGuestAccessConfig_paymentStripe(apiKey string) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth            = "hotspot"
  payment_gateway = "stripe"
  stripe = {
    api_key = %q
  }
}
`, apiKey)
}

func testAccSettingGuestAccessConfig_paymentAuthorize(useSandbox bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth            = "hotspot"
  payment_gateway = "authorize"
  authorize = {
    login_id        = "authorize-login"
    transaction_key = "authorize-transaction-key"
    use_sandbox     = %t
  }
}
`, useSandbox)
}

func testAccSettingGuestAccessConfig_paymentQuickpay(useSandbox bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth            = "hotspot"
  payment_gateway = "quickpay"
  quickpay = {
    agreement_id = "quickpay-agreement"
    api_key      = "quickpay-api-key"
    merchant_id  = "quickpay-merchant"
    use_sandbox  = %t
  }
}
`, useSandbox)
}

func testAccSettingGuestAccessConfig_paymentMerchantWarrior(useSandbox bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth            = "hotspot"
  payment_gateway = "merchantwarrior"
  merchant_warrior = {
    api_key = "mw-api-key"
    api_passphrase = "mw-passphrase"
    merchant_uuid = "mw-merchant-id"
    use_sandbox   = %t
  }
}
`, useSandbox)
}

func testAccSettingGuestAccessConfig_paymentIPpay(useSandbox bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth            = "hotspot"
  payment_gateway = "ippay"
  ippay = {
    terminal_id = "ippay-terminal"
    use_sandbox = %t
  }
}
`, useSandbox)
}

func testAccSettingGuestAccessConfig_redirect(url string, useHttps bool, toHttps bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth = "hotspot"
  redirect = {
    url       = %q
    use_https = %t
    to_https  = %t
  }
}
`, url, useHttps, toHttps)
}

func testAccSettingGuestAccessConfig_facebook(appId, appSecret string, scopeEmail bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth = "hotspot"
  facebook = {
    app_id      = %q
    app_secret  = %q
    scope_email = %t
  }
}
`, appId, appSecret, scopeEmail)
}

func testAccSettingGuestAccessConfig_google(clientId, clientSecret, domain string, scopeEmail bool) string {
	domainConfig := ""
	if domain != "" {
		domainConfig = fmt.Sprintf("    domain       = %q", domain)
	}

	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth = "hotspot"
  google = {
    client_id      = %q
    client_secret  = %q
%s
    scope_email    = %t
  }
}
`, clientId, clientSecret, domainConfig, scopeEmail)
}

func testAccSettingGuestAccessConfig_radius(authType, profileId string, disconnectEnabled bool, disconnectPort int) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth = "hotspot"
  radius = {
	auth_type          = %q
	profile_id         = %q
	disconnect_enabled = %t
	disconnect_port    = %d
  }
}
`, authType, profileId, disconnectEnabled, disconnectPort)
}

func testAccSettingGuestAccessConfig_wechat(appId, appSecret, secretKey, shopId string) string {
	shopIdConfig := ""
	if shopId != "" {
		shopIdConfig = fmt.Sprintf("    shop_id      = %q", shopId)
	}

	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth = "hotspot"
  wechat = {
    app_id       = %q
    app_secret   = %q
    secret_key   = %q
%s
  }
}
`, appId, appSecret, secretKey, shopIdConfig)
}

func testAccSettingGuestAccessConfig_facebookWifi(gatewayId, gatewayName, gatewaySecret string, blockHttps bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth = "facebook_wifi"
  facebook_wifi = {
    gateway_id     = %q
    gateway_name   = %q
    gateway_secret = %q
    block_https    = %t
  }
}
`, gatewayId, gatewayName, gatewaySecret, blockHttps)
}

func testAccSettingGuestAccessConfig_restrictedDNS(dnsServers []string) string {
	serversStr := ""
	for i, server := range dnsServers {
		if i > 0 {
			serversStr += ", "
		}
		serversStr += fmt.Sprintf("%q", server)
	}

	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth = "none"
  restricted_dns_servers = [%s]
}
`, serversStr)
}

func testAccSettingGuestAccessConfig_portalCustomizationBasic() string {
	return `
resource "unifi_setting_guest_access" "test" {
  auth = "none"
  portal_customization = {
    customized   = true
    bg_color     = "#f5f5f5"
    box_color    = "#ffffff"
    box_opacity  = 90
    title        = "Guest WiFi Portal"
    tos_enabled        = true
    tos                = "By using this WiFi service, you agree to our terms and conditions."
    box_text_color     = "#333333"
    text_color         = "#222222"
    link_color         = "#0066cc"
    box_link_color     = "#0055aa"
    button_color       = "#4CAF50"
    button_text_color  = "#ffffff"
    languages          = ["en", "es", "fr"]
  }
}
`
}

func testAccSettingGuestAccessConfig_portalCustomizationBasicPost74() string {
	return `
resource "unifi_setting_guest_access" "test" {
  auth = "none"
  portal_customization = {
    customized   = true
    bg_type      = "color"
    box_radius   = 12
    button_text  = "Login",
	authentication_text = "Please authenticate to access the internet",
	success_text = "You are now connected!",
	logo_position = "center",
	logo_size = 150
  }
}
`
}

func testAccSettingGuestAccessConfig_portalCustomizationImagesPost74() string {
	return `
resource "unifi_portal_file" "logo" {
  file_path = "files/testfile.png"
}

resource "unifi_portal_file" "background" {
  file_path = "files/testfile2.jpg"
}

resource "unifi_setting_guest_access" "test" {
  auth = "none"
  portal_customization = {
    customized       = true
    bg_type          = "image"
	bg_image_file_id = unifi_portal_file.background.id
	logo_file_id     = unifi_portal_file.logo.id
  }
}
`
}

func testAccSettingGuestAccessConfig_portalCustomizationGallery() string {
	return `
resource "unifi_setting_guest_access" "test" {
  auth = "none"
  portal_customization = {
    customized               = true
    unsplash_author_name     = "John Doe"
    unsplash_author_username = "johndoe"
    welcome_text_enabled     = true
    welcome_text             = "Welcome to our WiFi network!"
    welcome_text_position    = "above_boxes"
    box_color                = "#ffffff"
    box_opacity              = 90
    title                    = "Guest WiFi Portal"
  }
}
`
}

func testAccSettingGuestAccessConfig_portalCustomizationDisabled() string {
	return `
resource "unifi_setting_guest_access" "test" {
  auth = "none"
  portal_customization = {
    customized = false
  }
}
`
}

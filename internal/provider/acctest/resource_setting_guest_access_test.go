package acctest

import (
	"fmt"
	"strings"
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
				Config: testAccSettingGuestAccessConfigBasic(),
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
				Config: testAccSettingGuestAccessConfigBasicUpdated(),
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
				Config: testAccSettingGuestAccessConfigCustomAuth("192.168.1.1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "custom"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "custom_ip", "192.168.1.1"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfigCustomAuth("192.168.1.2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "custom"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "custom_ip", "192.168.1.2"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfigAuth("none"),
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
				Config: testAccSettingGuestAccessConfigPassword("pass1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "password", "pass1"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "password_enabled", "true"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfigPassword("pass2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "password", "pass2"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "password_enabled", "true"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfigAuth("hotspot"),
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
				Config: testAccSettingGuestAccessConfigVoucher(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "voucher_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "voucher_customized", "false"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfigVoucherCustomized(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "voucher_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "voucher_customized", "true"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfigVoucher(false),
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
				Config: testAccSettingGuestAccessConfigAllowedSubnet("192.168.1.0/24"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "allowed_subnet", "192.168.1.0/24"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfigAllowedSubnet("10.0.0.0/24"),
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
				Config: testAccSettingGuestAccessConfigPaymentPaypal(true),
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
				Config: testAccSettingGuestAccessConfigPaymentPaypal(false),
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
				Config: testAccSettingGuestAccessConfigPaymentPaypalUpdated(),
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
				Config: testAccSettingGuestAccessConfigPaymentStripe("stripe-api-key"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "stripe"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "stripe.api_key", "stripe-api-key"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_guest_access.test"),
			{
				Config: testAccSettingGuestAccessConfigPaymentStripe("updated-stripe-api-key"),
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
				Config: testAccSettingGuestAccessConfigPaymentAuthorize(true),
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
				Config: testAccSettingGuestAccessConfigPaymentAuthorize(false),
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
				Config: testAccSettingGuestAccessConfigPaymentQuickpay(true),
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
				Config: testAccSettingGuestAccessConfigPaymentQuickpay(false),
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
				Config: testAccSettingGuestAccessConfigPaymentMerchantWarrior(true),
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
				Config: testAccSettingGuestAccessConfigPaymentMerchantWarrior(false),
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
				Config: testAccSettingGuestAccessConfigPaymentIPpay(true),
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
				Config: testAccSettingGuestAccessConfigPaymentIPpay(false),
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
				Config: testAccSettingGuestAccessConfigPaymentPaypal(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "paypal"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfigPaymentStripe("stripe-api-key"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "stripe"),
					resource.TestCheckNoResourceAttr("unifi_setting_guest_access.test", "paypal.username"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfigPaymentAuthorize(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "authorize"),
					resource.TestCheckNoResourceAttr("unifi_setting_guest_access.test", "stripe.api_key"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfigPaymentQuickpay(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "quickpay"),
					resource.TestCheckNoResourceAttr("unifi_setting_guest_access.test", "authorize.login_id"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfigPaymentMerchantWarrior(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "merchantwarrior"),
					resource.TestCheckNoResourceAttr("unifi_setting_guest_access.test", "quickpay.api_key"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfigPaymentIPpay(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "payment_gateway", "ippay"),
					resource.TestCheckNoResourceAttr("unifi_setting_guest_access.test", "merchant_warrior.api_key"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfigAuth("hotspot"),
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
				Config: testAccSettingGuestAccessConfigRedirect("https://example.com", true, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.use_https", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.to_https", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.url", "https://example.com"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_guest_access.test"),
			{
				Config: testAccSettingGuestAccessConfigRedirect("https://updated-example.com", false, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.use_https", "false"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.to_https", "false"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.url", "https://updated-example.com"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfigAuth("none"),
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
				Config: testAccSettingGuestAccessConfigFacebook("facebook-app-id", "facebook-app-secret", true),
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
				Config: testAccSettingGuestAccessConfigFacebook("updated-app-id", "updated-app-secret", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook.app_id", "updated-app-id"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook.app_secret", "updated-app-secret"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook.scope_email", "false"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfigAuth("none"),
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
				Config: testAccSettingGuestAccessConfigGoogle("google-client-id", "google-client-secret", "example.com", true),
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
				Config: testAccSettingGuestAccessConfigGoogle("updated-client-id", "updated-client-secret", "", false),
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
				Config: testAccSettingGuestAccessConfigAuth("none"),
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
				Config: testAccSettingGuestAccessConfigRadius("chap", "radius-profile-id", true, 3799),
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
				Config: testAccSettingGuestAccessConfigRadius("mschapv2", "updated-profile-id", false, 1812),
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
				Config: testAccSettingGuestAccessConfigAuth("none"),
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
				Config: testAccSettingGuestAccessConfigWechat("wechat-app-id", "wechat-app-secret", "wechat-secret-key", "wechat-shop-id"),
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
				Config: testAccSettingGuestAccessConfigWechat("updated-app-id", "updated-app-secret", "updated-secret-key", "updated-shop-id"),
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
				Config: testAccSettingGuestAccessConfigAuth("none"),
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
				Config: testAccSettingGuestAccessConfigFacebookWifi("gateway-id", "gateway-name", "gateway-secret", true),
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
				Config: testAccSettingGuestAccessConfigFacebookWifi("updated-gateway-id", "updated-gateway-name", "updated-gateway-secret", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "facebook_wifi"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook_wifi.gateway_id", "updated-gateway-id"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook_wifi.gateway_name", "updated-gateway-name"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook_wifi.gateway_secret", "updated-gateway-secret"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "facebook_wifi.block_https", "false"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfigAuth("none"),
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
				Config: testAccSettingGuestAccessConfigRestrictedDNS([]string{"8.8.8.8", "1.1.1.1"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_servers.#", "2"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_servers.0", "8.8.8.8"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_servers.1", "1.1.1.1"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_guest_access.test"),
			{
				Config: testAccSettingGuestAccessConfigRestrictedDNS([]string{"8.8.4.4", "1.0.0.1", "9.9.9.9"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_servers.#", "3"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_servers.0", "8.8.4.4"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_servers.1", "1.0.0.1"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_servers.2", "9.9.9.9"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfigRestrictedDNS([]string{}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_enabled", "false"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "restricted_dns_servers.#", "0"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfigBasic(),
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
				Config: testAccSettingGuestAccessConfigPortalCustomizationBasicPost74(),
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
				Config: testAccSettingGuestAccessConfigPortalCustomizationImagesPost74(),
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
				Config: testAccSettingGuestAccessConfigPortalCustomizationBasic(),
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
				Config: testAccSettingGuestAccessConfigPortalCustomizationGallery(),
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
				Config: testAccSettingGuestAccessConfigPortalCustomizationDisabled(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.customized", "false"),
				),
			},
			{
				// Back to basic configuration
				Config: testAccSettingGuestAccessConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.customized", "false"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_customization.%", "29"),
				),
			},
		},
	})
}

func testAccSettingGuestAccessConfigBasic() string {
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

func testAccSettingGuestAccessConfigBasicUpdated() string {
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

func testAccSettingGuestAccessConfigAuth(auth string) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth = "%s"
}
`, auth)
}

func testAccSettingGuestAccessConfigCustomAuth(ip string) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth     = "custom"
  custom_ip = %q
}
`, ip)
}

func testAccSettingGuestAccessConfigPassword(password string) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth     = "hotspot"
  password = %q
}
`, password)
}

func testAccSettingGuestAccessConfigVoucher(enabled bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth            = "hotspot"
  voucher_enabled = %t
}
`, enabled)
}

func testAccSettingGuestAccessConfigVoucherCustomized() string {
	return `
resource "unifi_setting_guest_access" "test" {
  auth               = "hotspot"
  voucher_enabled    = true
  voucher_customized = true
}
`
}

func testAccSettingGuestAccessConfigAllowedSubnet(subnet string) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  allowed_subnet = %q
}
`, subnet)
}

func testAccSettingGuestAccessConfigPaymentPaypal(useSandbox bool) string {
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

func testAccSettingGuestAccessConfigPaymentPaypalUpdated() string {
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

func testAccSettingGuestAccessConfigPaymentStripe(apiKey string) string {
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

func testAccSettingGuestAccessConfigPaymentAuthorize(useSandbox bool) string {
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

func testAccSettingGuestAccessConfigPaymentQuickpay(useSandbox bool) string {
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

func testAccSettingGuestAccessConfigPaymentMerchantWarrior(useSandbox bool) string {
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

func testAccSettingGuestAccessConfigPaymentIPpay(useSandbox bool) string {
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

func testAccSettingGuestAccessConfigRedirect(url string, useHTTPS bool, toHTTPS bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth = "hotspot"
  redirect = {
    url       = %q
    use_https = %t
    to_https  = %t
  }
}
`, url, useHTTPS, toHTTPS)
}

func testAccSettingGuestAccessConfigFacebook(appID, appSecret string, scopeEmail bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth = "hotspot"
  facebook = {
    app_id      = %q
    app_secret  = %q
    scope_email = %t
  }
}
`, appID, appSecret, scopeEmail)
}

func testAccSettingGuestAccessConfigGoogle(clientID, clientSecret, domain string, scopeEmail bool) string {
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
`, clientID, clientSecret, domainConfig, scopeEmail)
}

func testAccSettingGuestAccessConfigRadius(authType, profileID string, disconnectEnabled bool, disconnectPort int) string {
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
`, authType, profileID, disconnectEnabled, disconnectPort)
}

func testAccSettingGuestAccessConfigWechat(appID, appSecret, secretKey, shopID string) string {
	shopIDConfig := ""
	if shopID != "" {
		shopIDConfig = fmt.Sprintf("    shop_id      = %q", shopID)
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
`, appID, appSecret, secretKey, shopIDConfig)
}

func testAccSettingGuestAccessConfigFacebookWifi(gatewayID, gatewayName, gatewaySecret string, blockHTTPS bool) string {
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
`, gatewayID, gatewayName, gatewaySecret, blockHTTPS)
}

func testAccSettingGuestAccessConfigRestrictedDNS(dnsServers []string) string {
	serversStr := ""
	var serversStrSb1053 strings.Builder
	for i, server := range dnsServers {
		if i > 0 {
			serversStrSb1053.WriteString(", ")
		}
		fmt.Fprintf(&serversStrSb1053, "%q", server)
	}
	serversStr += serversStrSb1053.String()

	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth = "none"
  restricted_dns_servers = [%s]
}
`, serversStr)
}

func testAccSettingGuestAccessConfigPortalCustomizationBasic() string {
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

func testAccSettingGuestAccessConfigPortalCustomizationBasicPost74() string {
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

func testAccSettingGuestAccessConfigPortalCustomizationImagesPost74() string {
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

func testAccSettingGuestAccessConfigPortalCustomizationGallery() string {
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

func testAccSettingGuestAccessConfigPortalCustomizationDisabled() string {
	return `
resource "unifi_setting_guest_access" "test" {
  auth = "none"
  portal_customization = {
    customized = false
  }
}
`
}

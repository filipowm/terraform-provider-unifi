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
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "template_engine", "angular"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_guest_access.test"),
		},
	})
}

func TestAccSettingGuestAccess_auth(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_auth("none"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "none"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_auth("hotspot"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_customAuth("192.168.1.1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "custom"),
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

func TestAccSettingGuestAccess_ecEnabled(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_ecEnabled(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "ec_enabled", "true"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_ecEnabled(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "ec_enabled", "false"),
				),
			},
		},
	})
}

func TestAccSettingGuestAccess_expiration(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_expiration(60, 1, 60),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "expire", "60"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "expire_number", "1"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "expire_unit", "60"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_expiration(1440, 1, 1440),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "expire", "1440"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "expire_number", "1"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "expire_unit", "1440"),
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

func TestAccSettingGuestAccess_portal(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_portal(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_use_hostname", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_hostname", "guest.example.com"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_portalDisabled(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_enabled", "false"),
				),
			},
		},
	})
}

func TestAccSettingGuestAccess_templateEngine(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_templateEngine("angular"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "template_engine", "angular"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_templateEngine("jsp"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "template_engine", "jsp"),
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

func TestAccSettingGuestAccess_comprehensive(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Lock: settingGuestAccessLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingGuestAccessConfig_comprehensive(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "hotspot"),
					//resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "allowed_subnet", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "ec_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "expire", "60"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "expire_number", "1"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "expire_unit", "60"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "password", "guestpassword"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_use_hostname", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_hostname", "guest.example.com"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "template_engine", "angular"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "voucher_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "voucher_customized", "true"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_guest_access.test"),
			{
				Config: testAccSettingGuestAccessConfig_comprehensiveUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "auth", "custom"),
					//resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "allowed_subnet", "10.0.0.0/24"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "custom_ip", "192.168.1.2"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "ec_enabled", "false"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "expire", "1440"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "expire_number", "1"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "expire_unit", "1440"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "portal_enabled", "false"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "template_engine", "jsp"),
				),
			},
			pt.ImportStepWithSite("unifi_setting_guest_access.test"),
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
				Config: testAccSettingGuestAccessConfig_redirect("https://updated-example.com", true, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.use_https", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.to_https", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.url", "https://updated-example.com"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_redirect("https://example.com", false, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.use_https", "false"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.to_https", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.url", "https://example.com"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_redirect("https://example.com", true, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.use_https", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.to_https", "false"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.url", "https://example.com"),
				),
			},
			{
				Config: testAccSettingGuestAccessConfig_redirect("https://example.com", false, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect_enabled", "true"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.use_https", "false"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.to_https", "false"),
					resource.TestCheckResourceAttr("unifi_setting_guest_access.test", "redirect.url", "https://example.com"),
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

func testAccSettingGuestAccessConfig_basic() string {
	return `
resource "unifi_setting_guest_access" "test" {
  auth           = "none"
  portal_enabled = true
  template_engine = "angular"
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

func testAccSettingGuestAccessConfig_ecEnabled(enabled bool) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  ec_enabled = %t
}
`, enabled)
}

func testAccSettingGuestAccessConfig_expiration(expire, expireNumber, expireUnit int) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  expire        = %d
  expire_number = %d
  expire_unit   = %d
}
`, expire, expireNumber, expireUnit)
}

func testAccSettingGuestAccessConfig_password(password string) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  auth     = "hotspot"
  password = %q
}
`, password)
}

func testAccSettingGuestAccessConfig_portal() string {
	return `
resource "unifi_setting_guest_access" "test" {
  portal_enabled     = true
  portal_use_hostname = true
  portal_hostname    = "guest.example.com"
}
`
}

func testAccSettingGuestAccessConfig_portalDisabled() string {
	return `
resource "unifi_setting_guest_access" "test" {
  portal_enabled = false
}
`
}

func testAccSettingGuestAccessConfig_templateEngine(engine string) string {
	return fmt.Sprintf(`
resource "unifi_setting_guest_access" "test" {
  template_engine = "%s"
}
`, engine)
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

func testAccSettingGuestAccessConfig_comprehensive() string {
	return `
resource "unifi_setting_guest_access" "test" {
  auth               = "hotspot"
  //allowed_subnet     = "192.168.1.0/24"
  ec_enabled         = true
  expire             = 60
  expire_number      = 1
  expire_unit        = 60
  password           = "guestpassword"
  portal_enabled     = true
  portal_use_hostname = true
  portal_hostname    = "guest.example.com"
  template_engine    = "angular"
  voucher_enabled    = true
  voucher_customized = true
}
`
}

func testAccSettingGuestAccessConfig_comprehensiveUpdated() string {
	return `
resource "unifi_setting_guest_access" "test" {
  auth               = "custom"
  //allowed_subnet     = "10.0.0.0/24"
  custom_ip          = "192.168.1.2"
  ec_enabled         = false
  expire             = 1440
  expire_number      = 1
  expire_unit        = 1440
  portal_enabled     = false
  template_engine    = "jsp"
}
`
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

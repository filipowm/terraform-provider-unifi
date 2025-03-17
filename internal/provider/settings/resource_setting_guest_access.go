package settings

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/filipowm/go-unifi/unifi"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
)

// TODO add support for uploading files and configuring logo and background custom images

type guestAccessModel struct {
	base.Model
	AllowedSubnet types.String `tfsdk:"allowed_subnet"`
	Auth          types.String `tfsdk:"auth"`
	AuthUrl       types.String `tfsdk:"auth_url"`

	Authorize types.Object `tfsdk:"authorize"`

	CustomIP  types.String `tfsdk:"custom_ip"`
	EcEnabled types.Bool   `tfsdk:"ec_enabled"`

	Expire       types.Int32 `tfsdk:"expire"`
	ExpireNumber types.Int32 `tfsdk:"expire_number"`
	ExpireUnit   types.Int32 `tfsdk:"expire_unit"`

	FacebookEnabled types.Bool   `tfsdk:"facebook_enabled"`
	Facebook        types.Object `tfsdk:"facebook"`

	FacebookWifi types.Object `tfsdk:"facebook_wifi"`

	GoogleEnabled types.Bool   `tfsdk:"google_enabled"`
	Google        types.Object `tfsdk:"google"`

	IPpay           types.Object `tfsdk:"ippay"`
	MerchantWarrior types.Object `tfsdk:"merchant_warrior"`

	PasswordEnabled types.Bool   `tfsdk:"password_enabled"`
	Password        types.String `tfsdk:"password"`

	PaymentEnabled types.Bool   `tfsdk:"payment_enabled"`
	PaymentGateway types.String `tfsdk:"payment_gateway"`

	Paypal types.Object `tfsdk:"paypal"`

	//PortalCustomization types.Object `tfsdk:"portal_customization"`

	PortalEnabled     types.Bool   `tfsdk:"portal_enabled"`
	PortalHostname    types.String `tfsdk:"portal_hostname"`
	PortalUseHostname types.Bool   `tfsdk:"portal_use_hostname"`

	Quickpay types.Object `tfsdk:"quickpay"`

	RadiusEnabled types.Bool   `tfsdk:"radius_enabled"`
	Radius        types.Object `tfsdk:"radius"`

	RedirectEnabled types.Bool   `tfsdk:"redirect_enabled"`
	Redirect        types.Object `tfsdk:"redirect"`

	//RestrictedDNSEnabled types.Bool   `tfsdk:"restricted_dns_enabled"`
	//RestrictedDNSServers types.List   `tfsdk:"restricted_dns_servers"`
	//RestrictedSubnet     types.String `tfsdk:"restricted_subnet"`

	TemplateEngine types.String `tfsdk:"template_engine"`

	Stripe types.Object `tfsdk:"stripe"`

	VoucherCustomized types.Bool `tfsdk:"voucher_customized"`
	VoucherEnabled    types.Bool `tfsdk:"voucher_enabled"`

	WechatEnabled types.Bool   `tfsdk:"wechat_enabled"`
	Wechat        types.Object `tfsdk:"wechat"`
}

type portalCustomizationModel struct {
	Customized             types.Bool   `tfsdk:"customized"`
	AuthenticationText     types.String `tfsdk:"authentication_text"`
	BgColor                types.String `tfsdk:"bg_color"`
	BgImageTile            types.Bool   `tfsdk:"bg_image_tile"`
	BgType                 types.String `tfsdk:"bg_type"`
	BoxColor               types.String `tfsdk:"box_color"`
	BoxLinkColor           types.String `tfsdk:"box_link_color"`
	BoxOpacity             types.Int32  `tfsdk:"box_opacity"`
	BoxRadius              types.Int32  `tfsdk:"box_radius"`
	BoxTextColor           types.String `tfsdk:"box_text_color"`
	ButtonColor            types.String `tfsdk:"button_color"`
	ButtonText             types.String `tfsdk:"button_text"`
	ButtonTextColor        types.String `tfsdk:"button_text_color"`
	Languages              types.List   `tfsdk:"languages"`
	LinkColor              types.String `tfsdk:"link_color"`
	LogoPosition           types.String `tfsdk:"logo_position"`
	LogoSize               types.Int32  `tfsdk:"logo_size"`
	SuccessText            types.String `tfsdk:"success_text"`
	TextColor              types.String `tfsdk:"text_color"`
	Title                  types.String `tfsdk:"title"`
	Tos                    types.String `tfsdk:"tos"`
	TosEnabled             types.Bool   `tfsdk:"tos_enabled"`
	UnsplashAuthorName     types.String `tfsdk:"unsplash_author_name"`
	UnsplashAuthorUsername types.String `tfsdk:"unsplash_author_username"`
	WelcomeText            types.String `tfsdk:"welcome_text"`
	WelcomeTextEnabled     types.Bool   `tfsdk:"welcome_text_enabled"`
	WelcomeTextPosition    types.String `tfsdk:"welcome_text_position"`
}

func (m *portalCustomizationModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"customized":          types.BoolType,
		"authentication_text": types.StringType,
		"bg_color":            types.StringType,
		"bg_image_tile":       types.BoolType,
		"bg_type":             types.StringType,
		"box_color":           types.StringType,
		"box_link_color":      types.StringType,
		"box_opacity":         types.Int32Type,
		"box_radius":          types.Int32Type,
		"box_text_color":      types.StringType,
		"button_color":        types.StringType,
		"button_text":         types.StringType,
		"button_text_color":   types.StringType,
		"languages": types.ListType{
			ElemType: types.StringType,
		},
		"link_color":               types.StringType,
		"logo_position":            types.StringType,
		"logo_size":                types.Int32Type,
		"success_text":             types.StringType,
		"text_color":               types.StringType,
		"title":                    types.StringType,
		"tos":                      types.StringType,
		"tos_enabled":              types.BoolType,
		"unsplash_author_name":     types.StringType,
		"unsplash_author_username": types.StringType,
		"welcome_text":             types.StringType,
		"welcome_text_enabled":     types.BoolType,
		"welcome_text_position":    types.StringType,
	}
}

type facebookModel struct {
	AppID      types.String `tfsdk:"app_id"`
	AppSecret  types.String `tfsdk:"app_secret"`
	ScopeEmail types.Bool   `tfsdk:"scope_email"`
}

func (m *facebookModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"app_id":      types.StringType,
		"app_secret":  types.StringType,
		"scope_email": types.BoolType,
	}
}

type facebookWifiModel struct {
	BlockHttps types.Bool   `tfsdk:"block_https"`
	GwID       types.String `tfsdk:"gateway_id"`
	GwName     types.String `tfsdk:"gateway_name"`
	GwSecret   types.String `tfsdk:"gateway_secret"`
}

func (m *facebookWifiModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"block_https":    types.BoolType,
		"gateway_id":     types.StringType,
		"gateway_name":   types.StringType,
		"gateway_secret": types.StringType,
	}
}

type googleModel struct {
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	Domain       types.String `tfsdk:"domain"`
	ScopeEmail   types.Bool   `tfsdk:"scope_email"`
}

func (m *googleModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"client_id":     types.StringType,
		"client_secret": types.StringType,
		"domain":        types.StringType,
		"scope_email":   types.BoolType,
	}
}

type paypalModel struct {
	Password   types.String `tfsdk:"password"`
	Username   types.String `tfsdk:"username"`
	UseSandbox types.Bool   `tfsdk:"use_sandbox"`
	Signature  types.String `tfsdk:"signature"`
}

func (m *paypalModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"password":    types.StringType,
		"username":    types.StringType,
		"use_sandbox": types.BoolType,
		"signature":   types.StringType,
	}
}

type ipPayModel struct {
	UseSandbox types.Bool   `tfsdk:"use_sandbox"`
	TerminalID types.String `tfsdk:"terminal_id"`
}

func (m *ipPayModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"use_sandbox": types.BoolType,
		"terminal_id": types.StringType,
	}
}

type quickpayModel struct {
	AgreementID types.String `tfsdk:"agreement_id"`
	ApiKey      types.String `tfsdk:"api_key"`
	MerchantID  types.String `tfsdk:"merchant_id"`
	UseSandbox  types.Bool   `tfsdk:"use_sandbox"`
}

func (m *quickpayModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"agreement_id": types.StringType,
		"api_key":      types.StringType,
		"merchant_id":  types.StringType,
		"use_sandbox":  types.BoolType,
	}
}

type radiusModel struct {
	AuthType          types.String `tfsdk:"auth_type"`
	DisconnectEnabled types.Bool   `tfsdk:"disconnect_enabled"`
	DisconnectPort    types.Int32  `tfsdk:"disconnect_port"`
	ProfileID         types.String `tfsdk:"profile_id"`
}

func (m *radiusModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"auth_type":          types.StringType,
		"disconnect_enabled": types.BoolType,
		"disconnect_port":    types.Int32Type,
		"profile_id":         types.StringType,
	}
}

type redirectModel struct {
	UseHttps types.Bool   `tfsdk:"use_https"`
	ToHttps  types.Bool   `tfsdk:"to_https"`
	Url      types.String `tfsdk:"url"`
}

func (m *redirectModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"use_https": types.BoolType,
		"to_https":  types.BoolType,
		"url":       types.StringType,
	}
}

type wechatModel struct {
	AppID     types.String `tfsdk:"app_id"`
	AppSecret types.String `tfsdk:"app_secret"`
	ShopID    types.String `tfsdk:"shop_id"`
	SecretKey types.String `tfsdk:"secret_key"`
}

func (m *wechatModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"app_id":     types.StringType,
		"app_secret": types.StringType,
		"shop_id":    types.StringType,
		"secret_key": types.StringType,
	}
}

type authorizeModel struct {
	LoginID        types.String `tfsdk:"login_id"`
	TransactionKey types.String `tfsdk:"transaction_key"`
	UseSandbox     types.Bool   `tfsdk:"use_sandbox"`
}

func (m *authorizeModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"login_id":        types.StringType,
		"transaction_key": types.StringType,
		"use_sandbox":     types.BoolType,
	}
}

type merchantWarriorModel struct {
	ApiKey        types.String `tfsdk:"api_key"`
	ApiPassphrase types.String `tfsdk:"api_passphrase"`
	MerchantID    types.String `tfsdk:"merchant_uuid"`
	UseSandbox    types.Bool   `tfsdk:"use_sandbox"`
}

func (m *merchantWarriorModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"api_key":        types.StringType,
		"api_passphrase": types.StringType,
		"merchant_uuid":  types.StringType,
		"use_sandbox":    types.BoolType,
	}
}

type stripeModel struct {
	ApiKey types.String `tfsdk:"api_key"`
}

func (m *stripeModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"api_key": types.StringType,
	}
}

func (d *guestAccessModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &unifi.SettingGuestAccess{
		AllowedSubnet: d.AllowedSubnet.ValueString(),
		Auth:          d.Auth.ValueString(),
		AuthUrl:       d.AuthUrl.ValueString(),
		CustomIP:      d.CustomIP.ValueString(),
		EcEnabled:     d.EcEnabled.ValueBool(),
		Expire:        int(d.Expire.ValueInt32()),
		ExpireNumber:  int(d.ExpireNumber.ValueInt32()),
		ExpireUnit:    int(d.ExpireUnit.ValueInt32()),

		PortalEnabled:     d.PortalEnabled.ValueBool(),
		PortalHostname:    d.PortalHostname.ValueString(),
		PortalUseHostname: d.PortalUseHostname.ValueBool(),
		TemplateEngine:    d.TemplateEngine.ValueString(),
		VoucherCustomized: d.VoucherCustomized.ValueBool(),
		VoucherEnabled:    d.VoucherEnabled.ValueBool(),
	}
	if base.IsEmptyString(d.Password) {
		model.PasswordEnabled = false
	} else {
		model.PasswordEnabled = true
		model.XPassword = d.Password.ValueString()
	}
	diags = d.paymentAsUnifiModel(ctx, model)
	if diags.HasError() {
		return nil, diags
	}
	if base.IsDefined(d.Redirect) {
		var redirect *redirectModel
		diags.Append(d.Redirect.As(ctx, &redirect, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		model.RedirectEnabled = true
		model.RedirectUrl = redirect.Url.ValueString()
		model.RedirectToHttps = redirect.ToHttps.ValueBool()
		model.RedirectHttps = redirect.UseHttps.ValueBool()
	} else {
		model.RedirectEnabled = false
	}

	if base.IsDefined(d.Facebook) {
		var facebook *facebookModel
		diags.Append(d.Facebook.As(ctx, &facebook, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		model.FacebookEnabled = true
		model.FacebookAppID = facebook.AppID.ValueString()
		model.XFacebookAppSecret = facebook.AppSecret.ValueString()
		model.FacebookScopeEmail = facebook.ScopeEmail.ValueBool()
	} else {
		model.FacebookEnabled = false
	}

	if base.IsDefined(d.Google) {
		var google *googleModel
		diags.Append(d.Google.As(ctx, &google, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		model.GoogleEnabled = true
		model.GoogleClientID = google.ClientID.ValueString()
		model.XGoogleClientSecret = google.ClientSecret.ValueString()
		model.GoogleScopeEmail = google.ScopeEmail.ValueBool()
		model.GoogleDomain = google.Domain.ValueString()
	} else {
		model.GoogleEnabled = false
	}

	if base.IsDefined(d.Radius) {
		var radius *radiusModel
		diags.Append(d.Radius.As(ctx, &radius, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		model.RADIUSEnabled = true
		model.RADIUSAuthType = radius.AuthType.ValueString()
		model.RADIUSDisconnectEnabled = radius.DisconnectEnabled.ValueBool()
		model.RADIUSDisconnectPort = int(radius.DisconnectPort.ValueInt32())
		model.RADIUSProfileID = radius.ProfileID.ValueString()
	} else {
		model.RADIUSEnabled = false
	}

	if base.IsDefined(d.Wechat) {
		var wechat *wechatModel
		diags.Append(d.Wechat.As(ctx, &wechat, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		model.WechatEnabled = true
		model.WechatAppID = wechat.AppID.ValueString()
		model.XWechatAppSecret = wechat.AppSecret.ValueString()
		model.WechatShopID = wechat.ShopID.ValueString()
		model.XWechatSecretKey = wechat.SecretKey.ValueString()
	} else {
		model.WechatEnabled = false
	}

	if base.IsDefined(d.FacebookWifi) {
		var facebookWifi *facebookWifiModel
		diags.Append(d.FacebookWifi.As(ctx, &facebookWifi, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		model.FacebookWifiBlockHttps = facebookWifi.BlockHttps.ValueBool()
		model.FacebookWifiGwID = facebookWifi.GwID.ValueString()
		model.FacebookWifiGwName = facebookWifi.GwName.ValueString()
		model.XFacebookWifiGwSecret = facebookWifi.GwSecret.ValueString()
	}

	return model, diags
}

func (d *guestAccessModel) paymentAsUnifiModel(ctx context.Context, model *unifi.SettingGuestAccess) diag.Diagnostics {
	diags := diag.Diagnostics{}
	if base.IsEmptyString(d.PaymentGateway) {
		model.PaymentEnabled = false
	} else {
		gateway := d.PaymentGateway.ValueString()
		model.PaymentEnabled = true
		model.Gateway = gateway
		switch gateway {
		case "authorize":
			var authorize *authorizeModel
			diags.Append(d.Authorize.As(ctx, &authorize, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return diags
			}
			if base.IsDefined(authorize.UseSandbox) {
				model.AuthorizeUseSandbox = authorize.UseSandbox.ValueBool()
			}
			model.XAuthorizeLoginid = authorize.LoginID.ValueString()
			model.XAuthorizeTransactionkey = authorize.TransactionKey.ValueString()
		case "ippay":
			var ippay *ipPayModel
			diags.Append(d.IPpay.As(ctx, &ippay, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return diags
			}
			if base.IsDefined(ippay.UseSandbox) {
				model.IPpayUseSandbox = ippay.UseSandbox.ValueBool()
			}
			model.XIPpayTerminalid = ippay.TerminalID.ValueString()
		case "merchantwarrior":
			var merchantWarrior *merchantWarriorModel
			diags.Append(d.MerchantWarrior.As(ctx, &merchantWarrior, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return diags
			}
			if base.IsDefined(merchantWarrior.UseSandbox) {
				model.MerchantwarriorUseSandbox = merchantWarrior.UseSandbox.ValueBool()
			}
			model.XMerchantwarriorApikey = merchantWarrior.ApiKey.ValueString()
			model.XMerchantwarriorApipassphrase = merchantWarrior.ApiPassphrase.ValueString()
			model.XMerchantwarriorMerchantuuid = merchantWarrior.MerchantID.ValueString()
		case "paypal":
			var paypal *paypalModel
			diags.Append(d.Paypal.As(ctx, &paypal, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return diags
			}
			if base.IsDefined(paypal.UseSandbox) {
				model.PaypalUseSandbox = paypal.UseSandbox.ValueBool()
			}
			model.XPaypalPassword = paypal.Password.ValueString()
			model.XPaypalUsername = paypal.Username.ValueString()
			model.XPaypalSignature = paypal.Signature.ValueString()
		case "quickpay":
			var quickpay *quickpayModel
			diags.Append(d.Quickpay.As(ctx, &quickpay, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return diags
			}
			if base.IsDefined(quickpay.UseSandbox) {
				model.QuickpayTestmode = quickpay.UseSandbox.ValueBool()
			}
			model.XQuickpayAgreementid = quickpay.AgreementID.ValueString()
			model.XQuickpayApikey = quickpay.ApiKey.ValueString()
			model.XQuickpayMerchantid = quickpay.MerchantID.ValueString()
		case "stripe":
			var stripe *stripeModel
			diags.Append(d.Stripe.As(ctx, &stripe, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return diags
			}
			model.XStripeApiKey = stripe.ApiKey.ValueString()
		default:
			diags.AddError("Invalid payment gateway", fmt.Sprintf("Payment gateway %q is not supported", gateway))
		}
	}
	return diags
}

func (d *guestAccessModel) mergePaymentModel(ctx context.Context, model *unifi.SettingGuestAccess) diag.Diagnostics {
	diags := diag.Diagnostics{}
	switch model.Gateway {
	case "authorize":
		authorize := &authorizeModel{
			LoginID:        types.StringValue(model.XAuthorizeLoginid),
			TransactionKey: types.StringValue(model.XAuthorizeTransactionkey),
			UseSandbox:     types.BoolValue(model.AuthorizeUseSandbox),
		}
		d.Authorize, diags = types.ObjectValueFrom(ctx, authorize.AttributeTypes(), authorize)
	case "ippay":
		ippay := &ipPayModel{
			UseSandbox: types.BoolValue(model.IPpayUseSandbox),
			TerminalID: types.StringValue(model.XIPpayTerminalid),
		}
		d.IPpay, diags = types.ObjectValueFrom(ctx, ippay.AttributeTypes(), ippay)
	case "merchantwarrior":
		merchantWarrior := &merchantWarriorModel{
			ApiKey:        types.StringValue(model.XMerchantwarriorApikey),
			ApiPassphrase: types.StringValue(model.XMerchantwarriorApipassphrase),
			MerchantID:    types.StringValue(model.XMerchantwarriorMerchantuuid),
			UseSandbox:    types.BoolValue(model.MerchantwarriorUseSandbox),
		}
		d.MerchantWarrior, diags = types.ObjectValueFrom(ctx, merchantWarrior.AttributeTypes(), merchantWarrior)
	case "paypal":
		paypal := &paypalModel{
			Password:   types.StringValue(model.XPaypalPassword),
			Username:   types.StringValue(model.XPaypalUsername),
			UseSandbox: types.BoolValue(model.PaypalUseSandbox),
			Signature:  types.StringValue(model.XPaypalSignature),
		}
		d.Paypal, diags = types.ObjectValueFrom(ctx, paypal.AttributeTypes(), paypal)
	case "quickpay":
		quickpay := &quickpayModel{
			AgreementID: types.StringValue(model.XQuickpayAgreementid),
			ApiKey:      types.StringValue(model.XQuickpayApikey),
			MerchantID:  types.StringValue(model.XQuickpayMerchantid),
			UseSandbox:  types.BoolValue(model.QuickpayTestmode),
		}
		d.Quickpay, diags = types.ObjectValueFrom(ctx, quickpay.AttributeTypes(), quickpay)
	case "stripe":
		stripe := &stripeModel{
			ApiKey: types.StringValue(model.XStripeApiKey),
		}
		d.Stripe, diags = types.ObjectValueFrom(ctx, stripe.AttributeTypes(), stripe)
	default:
		diags.AddError("Invalid payment gateway", fmt.Sprintf("Payment gateway returned by controller is not supported: %s", model.Gateway))
	}
	return diags
}

func (d *guestAccessModel) Merge(ctx context.Context, unifiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := unifiModel.(*unifi.SettingGuestAccess)
	if !ok {
		diags.AddError("Invalid model type", "Expected *unifi.SettingGuestAccess")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.AllowedSubnet = types.StringValue(model.AllowedSubnet)
	d.Auth = types.StringValue(model.Auth)
	d.AuthUrl = types.StringValue(model.AuthUrl)
	switch model.Auth {
	case "custom":
		d.CustomIP = types.StringValue(model.CustomIP)
	default:
		d.CustomIP = types.StringNull()
	}
	d.EcEnabled = types.BoolValue(model.EcEnabled)
	d.Expire = types.Int32Value(int32(model.Expire))
	d.ExpireNumber = types.Int32Value(int32(model.ExpireNumber))
	d.ExpireUnit = types.Int32Value(int32(model.ExpireUnit))

	d.PaymentEnabled = types.BoolValue(model.PaymentEnabled)
	var od diag.Diagnostics
	d.Authorize, od = base.ObjectNull(&authorizeModel{})
	diags.Append(od...)
	d.Paypal, od = base.ObjectNull(&paypalModel{})
	diags.Append(od...)
	d.IPpay, od = base.ObjectNull(&ipPayModel{})
	diags.Append(od...)
	d.MerchantWarrior, od = base.ObjectNull(&merchantWarriorModel{})
	diags.Append(od...)
	d.Quickpay, od = base.ObjectNull(&quickpayModel{})
	diags.Append(od...)
	d.Stripe, od = base.ObjectNull(&stripeModel{})
	diags.Append(od...)
	if diags.HasError() {
		return diags
	}
	if model.PaymentEnabled {
		d.PaymentGateway = types.StringValue(model.Gateway)
		d.mergePaymentModel(ctx, model)
	} else {
		d.PaymentGateway = types.StringNull()
	}

	d.PasswordEnabled = types.BoolValue(model.PasswordEnabled)
	if model.PasswordEnabled {
		d.Password = types.StringValue(model.XPassword)
	} else {
		d.Password = types.StringNull()
	}

	d.RedirectEnabled = types.BoolValue(model.RedirectEnabled)
	d.Redirect, diags = base.ObjectNull(&redirectModel{})
	if diags.HasError() {
		return diags
	}
	if model.RedirectEnabled {
		redirect := &redirectModel{
			UseHttps: types.BoolValue(model.RedirectHttps),
			ToHttps:  types.BoolValue(model.RedirectToHttps),
			Url:      types.StringValue(model.RedirectUrl),
		}
		d.Redirect, diags = types.ObjectValueFrom(ctx, redirect.AttributeTypes(), redirect)
		if diags.HasError() {
			return diags
		}
	}

	d.FacebookEnabled = types.BoolValue(model.FacebookEnabled)
	d.Facebook, diags = base.ObjectNull(&facebookModel{})
	if diags.HasError() {
		return diags
	}
	if model.FacebookEnabled {
		facebook := &facebookModel{
			AppID:      types.StringValue(model.FacebookAppID),
			AppSecret:  types.StringValue(model.XFacebookAppSecret),
			ScopeEmail: types.BoolValue(model.FacebookScopeEmail),
		}
		d.Facebook, diags = types.ObjectValueFrom(ctx, facebook.AttributeTypes(), facebook)
		if diags.HasError() {
			return diags
		}
	}

	d.GoogleEnabled = types.BoolValue(model.GoogleEnabled)
	d.Google, diags = base.ObjectNull(&googleModel{})
	if diags.HasError() {
		return diags
	}
	if model.GoogleEnabled {
		google := &googleModel{
			ClientID:     types.StringValue(model.GoogleClientID),
			ClientSecret: types.StringValue(model.XGoogleClientSecret),
			Domain:       types.StringValue(model.GoogleDomain),
			ScopeEmail:   types.BoolValue(model.GoogleScopeEmail),
		}
		d.Google, diags = types.ObjectValueFrom(ctx, google.AttributeTypes(), google)
		if diags.HasError() {
			return diags
		}
	}

	d.RadiusEnabled = types.BoolValue(model.RADIUSEnabled)
	d.Radius, diags = base.ObjectNull(&radiusModel{})
	if diags.HasError() {
		return diags
	}
	if model.RADIUSEnabled {
		radius := &radiusModel{
			AuthType:          types.StringValue(model.RADIUSAuthType),
			DisconnectEnabled: types.BoolValue(model.RADIUSDisconnectEnabled),
			DisconnectPort:    types.Int32Value(int32(model.RADIUSDisconnectPort)),
			ProfileID:         types.StringValue(model.RADIUSProfileID),
		}
		d.Radius, diags = types.ObjectValueFrom(ctx, radius.AttributeTypes(), radius)
		if diags.HasError() {
			return diags
		}
	}

	d.WechatEnabled = types.BoolValue(model.WechatEnabled)
	d.Wechat, diags = base.ObjectNull(&wechatModel{})
	if diags.HasError() {
		return diags
	}
	if model.WechatEnabled {
		wechat := &wechatModel{
			AppID:     types.StringValue(model.WechatAppID),
			ShopID:    types.StringValue(model.WechatShopID),
			AppSecret: types.StringValue(model.XWechatAppSecret),
			SecretKey: types.StringValue(model.XWechatSecretKey),
		}
		d.Wechat, diags = types.ObjectValueFrom(ctx, wechat.AttributeTypes(), wechat)
		if diags.HasError() {
			return diags
		}
	}

	d.FacebookWifi, diags = base.ObjectNull(&facebookWifiModel{})
	if diags.HasError() {
		return diags
	}
	if model.Auth == "facebook_wifi" {
		facebookWifi := &facebookWifiModel{
			BlockHttps: types.BoolValue(model.FacebookWifiBlockHttps),
			GwID:       types.StringValue(model.FacebookWifiGwID),
			GwName:     types.StringValue(model.FacebookWifiGwName),
			GwSecret:   types.StringValue(model.XFacebookWifiGwSecret),
		}
		d.FacebookWifi, diags = types.ObjectValueFrom(ctx, facebookWifi.AttributeTypes(), facebookWifi)
		if diags.HasError() {
			return diags
		}
	}

	d.PortalEnabled = types.BoolValue(model.PortalEnabled)
	d.PortalHostname = types.StringValue(model.PortalHostname)
	d.PortalUseHostname = types.BoolValue(model.PortalUseHostname)

	d.TemplateEngine = types.StringValue(model.TemplateEngine)
	d.VoucherCustomized = types.BoolValue(model.VoucherCustomized)
	d.VoucherEnabled = types.BoolValue(model.VoucherEnabled)

	return diags
}

var (
	_ resource.Resource              = &guestAccessResource{}
	_ resource.ResourceWithConfigure = &guestAccessResource{}
	//_ resource.ResourceWithValidateConfig = &guestAccessResource{}
	_ resource.ResourceWithImportState = &guestAccessResource{}
	//_ resource.ResourceWithConfigValidators = &guestAccessResource{}
	_ base.Resource = &guestAccessResource{}
)

type guestAccessResource struct {
	*base.GenericResource[*guestAccessModel]
}

func requiredTogetherIfTrue(condition string, attrs ...string) validators.RequiredTogetherIfValidator {
	var expressions []path.Expression
	for _, attr := range attrs {
		expressions = append(expressions, path.MatchRoot(attr))
	}
	return validators.RequiredTogetherIf(path.MatchRoot(condition), types.BoolValue(true), expressions...)
}

func requiredTogetherIfStringVal(condition, value string, attrs ...string) validators.RequiredTogetherIfValidator {
	var expressions []path.Expression
	for _, attr := range attrs {
		expressions = append(expressions, path.MatchRoot(attr))
	}
	return validators.RequiredTogetherIf(path.MatchRoot(condition), types.StringValue(value), expressions...)
}

func requiredStringValueIfTrue(conditionAttr, targetAttr, targetVal string) validators.RequiredValueIfValidator {
	return validators.RequiredValueIf(path.MatchRoot(conditionAttr), types.BoolValue(true), path.MatchRoot(targetAttr), types.StringValue(targetVal))
}

func (g *guestAccessResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		// Auth validators
		requiredTogetherIfStringVal("auth", "custom", "custom_ip"),
		//requiredTogetherIfStringVal("auth", "facebook_wifi", "facebook_wifi.gateway_id", "facebook_wifi.gateway_name", "facebook_wifi.gateway_secret"),

		// Facebook validators

		// Google validators
		requiredTogetherIfTrue("google.enabled", "google.client_id", "google.client_secret"),
		requiredStringValueIfTrue("google.enabled", "auth", "hotspot"),

		// Password validators
		requiredTogetherIfTrue("password_enabled", "password"),
		requiredStringValueIfTrue("password_enabled", "auth", "hotspot"),

		// Payment validators
		requiredTogetherIfStringVal("payment_gateway", "authorize", "authorize"),
		requiredTogetherIfStringVal("payment_gateway", "ippay", "ippay"),
		requiredTogetherIfStringVal("payment_gateway", "merchantwarrior", "merchant_warrior"),
		requiredTogetherIfStringVal("payment_gateway", "paypal", "paypal"),
		requiredTogetherIfStringVal("payment_gateway", "quickpay", "quickpay"),
		requiredTogetherIfStringVal("payment_gateway", "stripe", "stripe"),

		// Portal validators
		//requiredTogetherIfStringVal("portal_customized_bg_type", "color", "portal_customized_bg_color"),
		//requiredTogetherIfStringVal("portal_customized_bg_type", "gallery", "portal_customized_unsplash_author_name", "portal_customized_unsplash_author_username"),
		//requiredTogetherIfStringVal("portal_customized_bg_type", "image", "portal_customized_bg_image_filename"),
		//requiredTogetherIfTrue("portal_customized_bg_image_enabled", "portal_customized_bg_image_filename"),
		//requiredTogetherIfTrue("portal_customized_logo_enabled", "portal_customized_logo_filename"),
		//requiredTogetherIfTrue("portal_customized_tos_enabled", "portal_customized_tos"),
		//requiredTogetherIfTrue("portal_customized_welcome_text_enabled", "portal_customized_welcome_text"),
		//requiredTogetherIfTrue("portal_use_hostname", "portal_hostname"),

		// RADIUS validators
		//requiredTogetherIfTrue("radius_disconnect_enabled", "radius_disconnect_port"),
		//requiredTogetherIfTrue("radius_enabled", "radius_auth_type", "radius_profile_id"),

		// Restricted DNS validators
		//requiredTogetherIfTrue("restricted_dns_enabled", "restricted_dns_servers"),

		// Voucher validators
		requiredStringValueIfTrue("voucher_enabled", "auth", "hotspot"),
	}
}

func (g *guestAccessResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_guest_access` resource manages the guest access settings in the UniFi controller.\n\nThis resource allows you to configure all aspects of guest network access including authentication methods, portal customization, and payment options.",
		Attributes: map[string]schema.Attribute{
			"id":   base.ID(),
			"site": base.SiteAttribute(),
			"allowed_subnet": schema.StringAttribute{
				MarkdownDescription: "Subnet allowed for guest access.",
				Optional:            true,
				Computed:            true,
			},
			"auth": schema.StringAttribute{
				MarkdownDescription: "Authentication method for guest access. Valid values are:\n" +
					"* `none` - No authentication required\n" +
					"* `hotspot` - Password authentication\n" +
					"* `facebook_wifi` - Facebook auth entication\n" +
					"* `custom` - Custom authentication\n\n" +
					"For password authentication, set `auth` to `hotspot` and `password_enabled` to `true`.\n" +
					"For voucher authentication, set `auth` to `hotspot` and `voucher_enabled` to `true`.\n" +
					"For payment authentication, set `auth` to `hotspot` and `payment_enabled` to `true`.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("none"),
				Validators: []validator.String{
					stringvalidator.OneOf("none", "hotspot", "facebook_wifi", "custom"),
				},
			},
			"auth_url": schema.StringAttribute{
				MarkdownDescription: "URL for authentication. Must be a valid URL including the protocol.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validators.URL(),
				},
			},
			"authorize": schema.SingleNestedAttribute{
				MarkdownDescription: "Authorize.net payment settings.",
				Optional:            true,
				Validators:          []validator.Object{},
				Attributes: map[string]schema.Attribute{
					"use_sandbox": schema.BoolAttribute{
						MarkdownDescription: "Use sandbox mode for Authorize.net payments.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"login_id": schema.StringAttribute{
						MarkdownDescription: "Authorize.net login ID for authentication.",
						Required:            true,
					},
					"transaction_key": schema.StringAttribute{
						MarkdownDescription: "Authorize.net transaction key for authentication.",
						Required:            true,
					},
				},
			},
			"custom_ip": schema.StringAttribute{
				MarkdownDescription: "Custom IP address. Must be a valid IPv4 address (e.g., `192.168.1.1`).",
				Optional:            true,
				Validators: []validator.String{
					validators.IPv4(),
				},
			},
			"ec_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable enterprise controller functionality.",
				Optional:            true,
				Computed:            true,
			},
			"expire": schema.Int32Attribute{
				MarkdownDescription: "Expiration time for guest access.",
				Optional:            true,
				Computed:            true,
			},
			"expire_number": schema.Int32Attribute{
				MarkdownDescription: "Number value for the expiration time.",
				Optional:            true,
				Computed:            true,
			},
			"expire_unit": schema.Int32Attribute{
				MarkdownDescription: "Unit for the expiration time. Valid values are:\n" +
					"* `1` - Minute\n" +
					"* `60` - Hour\n" +
					"* `1440` - Day\n" +
					"* `10080` - Week",
				Optional: true,
				Computed: true,
				Validators: []validator.Int32{
					int32validator.OneOf(1, 60, 1440, 10080),
				},
			},
			"facebook_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether Facebook authentication for guest access is enabled.",
				Computed:            true,
			},
			"facebook": schema.SingleNestedAttribute{
				MarkdownDescription: "Facebook authentication settings.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"app_id": schema.StringAttribute{
						MarkdownDescription: "Facebook application ID for authentication.",
						Required:            true,
					},
					"app_secret": schema.StringAttribute{
						MarkdownDescription: "Facebook application secret for authentication.",
						Required:            true,
						Sensitive:           true,
					},
					"scope_email": schema.BoolAttribute{
						MarkdownDescription: "Request email scope for Facebook authentication.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
					},
				},
			},
			"facebook_wifi": schema.SingleNestedAttribute{
				MarkdownDescription: "Facebook WiFi authentication settings.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"block_https": schema.BoolAttribute{
						MarkdownDescription: "Mode HTTPS for Facebook WiFi.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"gateway_id": schema.StringAttribute{
						MarkdownDescription: "Facebook WiFi gateway ID.",
						Required:            true,
					},
					"gateway_name": schema.StringAttribute{
						MarkdownDescription: "Facebook WiFi gateway name.",
						Required:            true,
					},
					"gateway_secret": schema.StringAttribute{
						MarkdownDescription: "Facebook WiFi gateway secret.",
						Required:            true,
						Sensitive:           true,
					},
				},
			},
			"google_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether Google authentication for guest access is enabled.",
				Computed:            true,
			},
			"google": schema.SingleNestedAttribute{
				MarkdownDescription: "Google authentication settings.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"client_id": schema.StringAttribute{
						MarkdownDescription: "Google client ID for authentication.",
						Required:            true,
						//Sensitive:           true,
					},
					"client_secret": schema.StringAttribute{
						MarkdownDescription: "Google client secret for authentication.",
						Required:            true,
						//Sensitive:           true,
					},
					"domain": schema.StringAttribute{
						MarkdownDescription: "Restrict Google authentication to specific domain.",
						Optional:            true,
						Computed:            true,
						Validators: []validator.String{
							validators.Hostname(),
						},
					},
					"scope_email": schema.BoolAttribute{
						MarkdownDescription: "Request email scope for Google authentication.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
					},
				},
			},
			"ippay": schema.SingleNestedAttribute{
				MarkdownDescription: "IPpay Payments settings.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"terminal_id": schema.StringAttribute{
						MarkdownDescription: "Terminal ID for IP Payments.",
						Required:            true,
						Sensitive:           true,
					},
					"use_sandbox": schema.BoolAttribute{
						MarkdownDescription: "Whether to use sandbox mode for IPPay payments.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
				},
			},
			"merchant_warrior": schema.SingleNestedAttribute{
				MarkdownDescription: "MerchantWarrior payment settings.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"api_key": schema.StringAttribute{
						MarkdownDescription: "MerchantWarrior API key.",
						Required:            true,
						Sensitive:           true,
					},
					"api_passphrase": schema.StringAttribute{
						MarkdownDescription: "MerchantWarrior API passphrase.",
						Required:            true,
						Sensitive:           true,
					},
					"merchant_uuid": schema.StringAttribute{
						MarkdownDescription: "MerchantWarrior merchant UUID.",
						Required:            true,
						Sensitive:           true,
					},
					"use_sandbox": schema.BoolAttribute{
						MarkdownDescription: "Whether to use sandbox mode for MerchantWarrior payments.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
				},
			},
			"paypal": schema.SingleNestedAttribute{
				MarkdownDescription: "PayPal payment settings.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"password": schema.StringAttribute{
						MarkdownDescription: "PayPal password.",
						Required:            true,
						Sensitive:           true,
					},
					"signature": schema.StringAttribute{
						MarkdownDescription: "PayPal signature.",
						Required:            true,
						Sensitive:           true,
					},
					"use_sandbox": schema.BoolAttribute{
						MarkdownDescription: "Whether to use sandbox mode for PayPal payments.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"username": schema.StringAttribute{
						MarkdownDescription: "PayPal username. Must be a valid email address.",
						Required:            true,
						Sensitive:           true,
						Validators: []validator.String{
							validators.Email(),
						},
					},
				},
			},
			"password_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable password authentication for guest access.",
				Computed:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for guest access.",
				Optional:            true,
				Sensitive:           true,
			},
			"payment_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable payment for guest access.",
				Computed:            true,
			},
			"payment_gateway": schema.StringAttribute{
				MarkdownDescription: "Payment gateway. Valid values are:\n" +
					"* `paypal` - PayPal\n" +
					"* `stripe` - Stripe\n" +
					"* `authorize` - Authorize.net\n" +
					"* `quickpay` - QuickPay\n" +
					"* `merchantwarrior` - Merchant Warrior\n" +
					"* `ippay` - IP Payments",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf("paypal", "stripe", "authorize", "quickpay", "merchantwarrior", "ippay"),
				},
			},
			//"portal_customization": schema.SingleNestedAttribute{
			//	MarkdownDescription: "Portal customization settings.",
			//	Optional:            true,
			//	Computed:            true,
			//	Attributes: map[string]schema.Attribute{
			//		"customized": schema.BoolAttribute{
			//			MarkdownDescription: "Whether the portal is customized.",
			//			Optional:            true,
			//			Computed:            true,
			//		},
			//		"authentication_text": schema.StringAttribute{
			//			MarkdownDescription: "Custom authentication text for the portal.",
			//			Optional:            true,
			//			Computed:            true,
			//		},
			//		"bg_color": schema.StringAttribute{
			//			MarkdownDescription: "Background color for the custom portal. Must be a valid hex color code (e.g., #FFF or #FFFFFF).",
			//			Optional:            true,
			//			Computed:            true,
			//			Validators: []validator.String{
			//				validators.HexColor(),
			//			},
			//		},
			//		"bg_image_tile": schema.BoolAttribute{
			//			MarkdownDescription: "Tile the background image.",
			//			Optional:            true,
			//			Computed:            true,
			//		},
			//		"bg_type": schema.StringAttribute{
			//			MarkdownDescription: "Type of portal background. Valid values are:\n" +
			//				"* `color` - Solid color background\n" +
			//				"* `image` - (not yet supported!) Custom image background\n" +
			//				"* `gallery` - Image from Unsplash gallery",
			//			Optional: true,
			//			Computed: true,
			//			Validators: []validator.String{
			//				stringvalidator.OneOf("color", "image", "gallery"),
			//			},
			//		},
			//		"box_color": schema.StringAttribute{
			//			MarkdownDescription: "Color of the login box in the portal. Must be a valid hex color code (e.g., #FFF or #FFFFFF).",
			//			Optional:            true,
			//			Computed:            true,
			//			Validators: []validator.String{
			//				validators.HexColor(),
			//			},
			//		},
			//		"box_link_color": schema.StringAttribute{
			//			MarkdownDescription: "Color of links in the login box. Must be a valid hex color code (e.g., #FFF or #FFFFFF).",
			//			Optional:            true,
			//			Computed:            true,
			//			Validators: []validator.String{
			//				validators.HexColor(),
			//			},
			//		},
			//		"box_opacity": schema.Int32Attribute{
			//			MarkdownDescription: "Opacity of the login box (0-100).",
			//			Optional:            true,
			//			Computed:            true,
			//			Validators: []validator.Int32{
			//				int32validator.Between(0, 100),
			//			},
			//		},
			//		"box_radius": schema.Int32Attribute{
			//			MarkdownDescription: "Border radius of the login box in pixels.",
			//			Optional:            true,
			//			Computed:            true,
			//			Validators: []validator.Int32{
			//				int32validator.AtLeast(0),
			//			},
			//		},
			//		"box_text_color": schema.StringAttribute{
			//			MarkdownDescription: "Text color in the login box. Must be a valid hex color code (e.g., #FFF or #FFFFFF).",
			//			Optional:            true,
			//			Computed:            true,
			//			Validators: []validator.String{
			//				validators.HexColor(),
			//			},
			//		},
			//		"button_color": schema.StringAttribute{
			//			MarkdownDescription: "Button color in the portal. Must be a valid hex color code (e.g., #FFF or #FFFFFF).",
			//			Optional:            true,
			//			Computed:            true,
			//			Validators: []validator.String{
			//				validators.HexColor(),
			//			},
			//		},
			//		"button_text": schema.StringAttribute{
			//			MarkdownDescription: "Custom text for the login button.",
			//			Optional:            true,
			//			Computed:            true,
			//		},
			//		"button_text_color": schema.StringAttribute{
			//			MarkdownDescription: "Button text color. Must be a valid hex color code (e.g., #FFF or #FFFFFF).",
			//			Optional:            true,
			//			Computed:            true,
			//			Validators: []validator.String{
			//				validators.HexColor(),
			//			},
			//		},
			//		"languages": schema.ListAttribute{
			//			MarkdownDescription: "List of enabled languages for the portal.",
			//			Optional:            true,
			//			Computed:            true,
			//			ElementType:         types.StringType,
			//		},
			//		"link_color": schema.StringAttribute{
			//			MarkdownDescription: "Color for links in the portal. Must be a valid hex color code (e.g., #FFF or #FFFFFF).",
			//			Optional:            true,
			//			Computed:            true,
			//			Validators: []validator.String{
			//				validators.HexColor(),
			//			},
			//		},
			//		"logo_position": schema.StringAttribute{
			//			MarkdownDescription: "Position of the logo in the portal. Valid values are: left, center, right.",
			//			Optional:            true,
			//			Computed:            true,
			//			Validators: []validator.String{
			//				stringvalidator.OneOf("left", "center", "right"),
			//			},
			//		},
			//		"logo_size": schema.Int32Attribute{
			//			MarkdownDescription: "Size of the logo in pixels.",
			//			Optional:            true,
			//			Computed:            true,
			//			Validators: []validator.Int32{
			//				int32validator.AtLeast(0),
			//			},
			//		},
			//		"success_text": schema.StringAttribute{
			//			MarkdownDescription: "Text displayed after successful authentication.",
			//			Optional:            true,
			//			Computed:            true,
			//		},
			//		"text_color": schema.StringAttribute{
			//			MarkdownDescription: "Main text color for the portal. Must be a valid hex color code (e.g., #FFF or #FFFFFF).",
			//			Optional:            true,
			//			Computed:            true,
			//			Validators: []validator.String{
			//				validators.HexColor(),
			//			},
			//		},
			//		"title": schema.StringAttribute{
			//			MarkdownDescription: "Title of the portal page.",
			//			Optional:            true,
			//			Computed:            true,
			//		},
			//		"tos": schema.StringAttribute{
			//			MarkdownDescription: "Terms of service text.",
			//			Optional:            true,
			//			Computed:            true,
			//		},
			//		"tos_enabled": schema.BoolAttribute{
			//			MarkdownDescription: "Enable terms of service acceptance requirement.",
			//			Optional:            true,
			//			Computed:            true,
			//		},
			//		"unsplash_author_name": schema.StringAttribute{
			//			MarkdownDescription: "Name of the Unsplash author for gallery background.",
			//			Optional:            true,
			//			Computed:            true,
			//		},
			//		"unsplash_author_username": schema.StringAttribute{
			//			MarkdownDescription: "Username of the Unsplash author for gallery background.",
			//			Optional:            true,
			//			Computed:            true,
			//		},
			//		"welcome_text": schema.StringAttribute{
			//			MarkdownDescription: "Welcome text displayed on the portal.",
			//			Optional:            true,
			//			Computed:            true,
			//		},
			//		"welcome_text_enabled": schema.BoolAttribute{
			//			MarkdownDescription: "Enable welcome text display.",
			//			Optional:            true,
			//			Computed:            true,
			//		},
			//		"welcome_text_position": schema.StringAttribute{
			//			MarkdownDescription: "Position of the welcome text. Valid values are: top, middle, bottom.",
			//			Optional:            true,
			//			Computed:            true,
			//			Validators: []validator.String{
			//				stringvalidator.OneOf("top", "middle", "bottom"),
			//			},
			//		},
			//	},
			//},
			"portal_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable the guest portal.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"portal_hostname": schema.StringAttribute{
				MarkdownDescription: "Hostname to use for the captive portal.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validators.Hostname(),
				},
			},
			"portal_use_hostname": schema.BoolAttribute{
				MarkdownDescription: "Use a custom hostname for the portal.",
				Optional:            true,
				Computed:            true,
			},
			"quickpay": schema.SingleNestedAttribute{
				MarkdownDescription: "QuickPay payment settings.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"agreement_id": schema.StringAttribute{
						MarkdownDescription: "QuickPay agreement ID.",
						Required:            true,
						Sensitive:           true,
					},
					"api_key": schema.StringAttribute{
						MarkdownDescription: "QuickPay API key.",
						Required:            true,
						Sensitive:           true,
					},
					"merchant_id": schema.StringAttribute{
						MarkdownDescription: "QuickPay merchant ID.",
						Required:            true,
						Sensitive:           true,
					},
					"use_sandbox": schema.BoolAttribute{
						MarkdownDescription: "Enable sandbox mode for QuickPay payments.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
				},
			},
			"radius_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether RADIUS authentication for guest access is enabled.",
				Computed:            true,
			},
			"radius": schema.SingleNestedAttribute{
				MarkdownDescription: "RADIUS authentication settings.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"auth_type": schema.StringAttribute{
						MarkdownDescription: "RADIUS authentication type. Valid values are: `chap`, `mschapv2`.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("chap", "mschapv2"),
						},
					},
					"disconnect_enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable RADIUS disconnect messages.",
						Optional:            true,
						Computed:            true,
					},
					"disconnect_port": schema.Int32Attribute{
						MarkdownDescription: "Port for RADIUS disconnect messages.",
						Optional:            true,
						Computed:            true,
						Validators: []validator.Int32{
							int32validator.Between(1, 65535),
						},
					},
					"profile_id": schema.StringAttribute{
						MarkdownDescription: "ID of the RADIUS profile to use.",
						Required:            true,
					},
				},
			},
			"redirect_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether redirect after authentication is enabled.",
				Computed:            true,
			},
			"redirect": schema.SingleNestedAttribute{
				MarkdownDescription: "Redirect after authentication settings.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"use_https": schema.BoolAttribute{
						MarkdownDescription: "Use HTTPS for the redirect URL.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
					},
					"to_https": schema.BoolAttribute{
						MarkdownDescription: "Redirect HTTP requests to HTTPS.",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
					},
					"url": schema.StringAttribute{
						MarkdownDescription: "URL to redirect to after authentication. Must be a valid URL.",
						Required:            true,
						Validators: []validator.String{
							validators.URL(),
						},
					},
				},
			},
			//"restricted_dns_enabled": schema.BoolAttribute{
			//	MarkdownDescription: "Enable restricted DNS servers for guest networks.",
			//	Optional:            true,
			//	Computed:            true,
			//},
			//"restricted_dns_servers": schema.ListAttribute{
			//	MarkdownDescription: "List of restricted DNS servers for guest networks. Each value must be a valid IPv4 address.",
			//	Optional:            true,
			//	Computed:            true,
			//	ElementType:         types.StringType,
			//},
			//"restricted_subnet": schema.StringAttribute{
			//	MarkdownDescription: "Subnet for restricted guest access.",
			//	Optional:            true,
			//	Computed:            true,
			//},
			"stripe": schema.SingleNestedAttribute{
				MarkdownDescription: "Stripe payment settings.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"api_key": schema.StringAttribute{
						MarkdownDescription: "Stripe API key.",
						Required:            true,
						Sensitive:           true,
					},
				},
			},
			"template_engine": schema.StringAttribute{
				MarkdownDescription: "Template engine for the portal. Valid values are: `jsp`, `angular`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("jsp", "angular"),
				},
			},
			"voucher_customized": schema.BoolAttribute{
				MarkdownDescription: "Whether vouchers are customized.",
				Optional:            true,
				Computed:            true,
			},
			"voucher_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable voucher-based authentication for guest access.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"wechat_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether WeChat authentication for guest access is enabled.",
				Computed:            true,
			},
			"wechat": schema.SingleNestedAttribute{
				MarkdownDescription: "WeChat authentication settings.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"app_id": schema.StringAttribute{
						MarkdownDescription: "WeChat App ID for social authentication.",
						Required:            true,
					},
					"app_secret": schema.StringAttribute{
						MarkdownDescription: "WeChat App secret.",
						Required:            true,
						Sensitive:           true,
					},
					"secret_key": schema.StringAttribute{
						MarkdownDescription: "WeChat secret key.",
						Required:            true,
						Sensitive:           true,
					},
					"shop_id": schema.StringAttribute{
						MarkdownDescription: "WeChat Shop ID for payments.",
						Optional:            true,
						Computed:            true,
					},
				},
			},
		},
	}
}

func NewGuestAccessResource() resource.Resource {
	r := &guestAccessResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_guest_access",
		func() *guestAccessModel { return &guestAccessModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingGuestAccess(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			return client.UpdateSettingGuestAccess(ctx, site, body.(*unifi.SettingGuestAccess))
		},
	)
	return r
}

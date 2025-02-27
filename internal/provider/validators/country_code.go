package validators

import (
	"context"
	"github.com/biter777/countries"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func CountryCodeAlpha2() validator.String {
	return countryCodeAlpha2Validator{}
}

type countryCodeAlpha2Validator struct{}

func (c countryCodeAlpha2Validator) Description(_ context.Context) string {
	return "The country code must be a valid ISO 3166-1 alpha-2 code."
}

func (c countryCodeAlpha2Validator) MarkdownDescription(ctx context.Context) string {
	return c.Description(ctx)
}

func (c countryCodeAlpha2Validator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	code := req.ConfigValue
	if base.IsEmptyString(code) {
		return
	}

	codeString := code.ValueString()
	if len(codeString) != 2 || countries.ByName(codeString) == countries.Unknown {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path,
			c.Description(ctx),
			codeString,
		))
	}
}

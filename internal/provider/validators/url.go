package validators

import (
	"context"
	"fmt"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/types"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// URL returns a validator which ensures that the string value is a valid URL.
func URL() validator.String {
	return urlValidator{requireHTTPS: false}
}

// HTTPSUrl returns a validator which ensures that the string value is a valid HTTPS URL.
func HTTPSUrl() validator.String {
	return urlValidator{requireHTTPS: true}
}

type urlValidator struct {
	requireHTTPS bool
}

func (v urlValidator) Description(_ context.Context) string {
	if v.requireHTTPS {
		return "must be a valid HTTPS URL"
	}
	return "must be a valid URL"
}

func (v urlValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v urlValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	value := req.ConfigValue
	if !types.IsDefined(value) {
		return
	}

	val := value.ValueString()
	parsedURL, err := url.Parse(val)

	if err != nil {
		resp.Diagnostics.Append(
			validatordiag.InvalidAttributeValueDiagnostic(
				req.Path,
				v.Description(ctx),
				fmt.Sprintf("%q is not a valid URL: %s", val, err),
			),
		)
		return
	}

	// Check if URL has a scheme
	if parsedURL.Scheme == "" {
		resp.Diagnostics.Append(
			validatordiag.InvalidAttributeValueDiagnostic(
				req.Path,
				v.Description(ctx),
				fmt.Sprintf("%q is missing a scheme (e.g., http:// or https://)", val),
			),
		)
		return
	}

	// Check if HTTPS is required
	if v.requireHTTPS && parsedURL.Scheme != "https" {
		resp.Diagnostics.Append(
			validatordiag.InvalidAttributeValueDiagnostic(
				req.Path,
				v.Description(ctx),
				fmt.Sprintf("%q must use HTTPS scheme", val),
			),
		)
		return
	}

	// Check if URL has a host
	if parsedURL.Host == "" {
		resp.Diagnostics.Append(
			validatordiag.InvalidAttributeValueDiagnostic(
				req.Path,
				v.Description(ctx),
				fmt.Sprintf("%q is missing a host", val),
			),
		)
	}
}

package validators

import (
	"context"
	"fmt"
	"regexp"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// A common regex pattern for validating email addresses
// This is a simplified version and may not catch all edge cases
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Email returns a validator which ensures that the string value is a valid email address.
func Email() validator.String {
	return emailValidator{}
}

type emailValidator struct{}

func (v emailValidator) Description(_ context.Context) string {
	return "must be a valid email address"
}

func (v emailValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v emailValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	value := req.ConfigValue
	if !base.IsDefined(value) {
		return
	}

	val := value.ValueString()
	if !emailRegex.MatchString(val) {
		resp.Diagnostics.Append(
			validatordiag.InvalidAttributeValueDiagnostic(
				req.Path,
				v.Description(ctx),
				fmt.Sprintf("%q is not a valid email address", val),
			),
		)
	}
}

package validators

import (
	"context"
	"fmt"
	"regexp"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var hexColorRegex = regexp.MustCompile("^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$")

// HexColor returns a validator which ensures that the string value is a valid hex color code.
// Valid formats are: #RGB or #RRGGBB
func HexColor() validator.String {
	return hexColorValidator{}
}

type hexColorValidator struct{}

func (v hexColorValidator) Description(_ context.Context) string {
	return "must be a valid hex color code (e.g., #FFF or #FFFFFF)"
}

func (v hexColorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v hexColorValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	value := req.ConfigValue
	if !base.IsDefined(value) {
		return
	}

	val := value.ValueString()
	if !hexColorRegex.MatchString(val) {
		resp.Diagnostics.Append(
			validatordiag.InvalidAttributeValueDiagnostic(
				req.Path,
				v.Description(ctx),
				fmt.Sprintf("%q is not a valid hex color code", val),
			),
		)
	}
}

package validators

import (
	"context"
	"fmt"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/types"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Timezone returns a validator which ensures that the string value is a valid IANA timezone identifier
// according to the time.LoadLocation function.
func Timezone() validator.String {
	return timezoneValidator{}
}

type timezoneValidator struct{}

func (v timezoneValidator) Description(_ context.Context) string {
	return "must be a valid IANA timezone identifier (e.g., 'America/New_York')"
}

func (v timezoneValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v timezoneValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	value := req.ConfigValue
	if !types.IsDefined(value) {
		return
	}

	val := value.ValueString()

	// Check for empty string
	if val == "" {
		resp.Diagnostics.Append(
			validatordiag.InvalidAttributeValueDiagnostic(
				req.Path,
				v.Description(ctx),
				"Timezone cannot be empty. Use a valid IANA timezone identifier like 'America/New_York'",
			),
		)
		return
	}

	// Check for proper case (IANA timezone identifiers are case-sensitive)
	// Regions should start with uppercase
	if val[0] >= 'a' && val[0] <= 'z' {
		resp.Diagnostics.Append(
			validatordiag.InvalidAttributeValueDiagnostic(
				req.Path,
				v.Description(ctx),
				fmt.Sprintf("%q has incorrect case. IANA timezone regions should start with uppercase (e.g., 'America/New_York')", val),
			),
		)
		return
	}

	// Try to load the timezone location
	_, err := time.LoadLocation(val)
	if err != nil {
		// For better error messages, check common mistakes
		if strings.Contains(val, "UTC") && val != "UTC" {
			resp.Diagnostics.Append(
				validatordiag.InvalidAttributeValueDiagnostic(
					req.Path,
					v.Description(ctx),
					fmt.Sprintf("%q is not a valid timezone. For UTC offset use the standard 'UTC' timezone instead.", val),
				),
			)
		} else if strings.Contains(val, " ") {
			resp.Diagnostics.Append(
				validatordiag.InvalidAttributeValueDiagnostic(
					req.Path,
					v.Description(ctx),
					fmt.Sprintf("%q is not a valid timezone. Timezones should not contain spaces.", val),
				),
			)
		} else {
			resp.Diagnostics.Append(
				validatordiag.InvalidAttributeValueDiagnostic(
					req.Path,
					v.Description(ctx),
					fmt.Sprintf("%q is not a valid IANA timezone identifier. Use a value like 'America/New_York'", val),
				),
			)
		}
	}
}

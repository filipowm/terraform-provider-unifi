package validators

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// A regex pattern for validating hostnames without protocol schemes
// This matches hostnames according to RFC 1035 with some limitations
var hostnameRegex = regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$`)

// Hostname returns a validator which ensures that the string value is a valid hostname without protocol schemes.
func Hostname() validator.String {
	return hostnameValidator{}
}

type hostnameValidator struct{}

func (v hostnameValidator) Description(_ context.Context) string {
	return "must be a valid hostname (without protocol scheme)"
}

func (v hostnameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v hostnameValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	value := req.ConfigValue
	if !base.IsDefined(value) {
		return
	}

	val := value.ValueString()

	// Check if the hostname has a scheme (which it shouldn't)
	if strings.Contains(val, "://") {
		resp.Diagnostics.Append(
			validatordiag.InvalidAttributeValueDiagnostic(
				req.Path,
				v.Description(ctx),
				fmt.Sprintf("%q should not include a protocol scheme (e.g., http:// or https://)", val),
			),
		)
		return
	}

	// Convert to lowercase for validation
	hostname := strings.ToLower(val)

	if !hostnameRegex.MatchString(hostname) {
		resp.Diagnostics.Append(
			validatordiag.InvalidAttributeValueDiagnostic(
				req.Path,
				v.Description(ctx),
				fmt.Sprintf("%q is not a valid hostname", val),
			),
		)
	}
}

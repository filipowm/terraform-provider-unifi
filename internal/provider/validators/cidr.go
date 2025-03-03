package validators

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// CIDR returns a validator which ensures that a string value is a valid CIDR notation.
func CIDR() validator.String {
	return cidrValidator{
		allowEmpty: false,
	}
}

// CIDROrEmpty returns a validator which ensures that a string value is either empty or a valid CIDR notation.
func CIDROrEmpty() validator.String {
	return cidrValidator{
		allowEmpty: true,
	}
}

var (
	_ validator.String = cidrValidator{}
)

type cidrValidator struct {
	allowEmpty bool
}

func (v cidrValidator) Description(ctx context.Context) string {
	return "value must be a valid CIDR notation (e.g., '192.168.1.0/24')"
}

func (v cidrValidator) MarkdownDescription(ctx context.Context) string {
	return "value must be a valid CIDR notation (e.g., `192.168.1.0/24`)"
}

func (v cidrValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		if !v.allowEmpty {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid CIDR Notation",
				"CIDR notation cannot be empty",
			)
		}
		return
	}

	_, _, err := net.ParseCIDR(value)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid CIDR Notation",
			fmt.Sprintf("Value %q is not a valid CIDR notation: %v", value, err),
		)
		return
	}
}

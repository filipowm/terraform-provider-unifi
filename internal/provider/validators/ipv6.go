package validators

import (
	"context"
	"fmt"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/utils"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// IPv6 returns a validator which ensures that a string value is a valid IPv6 address.
func IPv6() validator.String {
	return ipv6Validator{}
}

var _ validator.String = ipv6Validator{}

type ipv6Validator struct{}

func (v ipv6Validator) Description(_ context.Context) string {
	return "value must be a valid IPv6 address"
}

func (v ipv6Validator) MarkdownDescription(_ context.Context) string {
	return "value must be a valid IPv6 address"
}

func (v ipv6Validator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		return
	}

	if !utils.IsIPv6(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IPv6 Address",
			fmt.Sprintf("Value %q is not a valid IPv6 address", value),
		)
		return
	}
}

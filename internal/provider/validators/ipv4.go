package validators

import (
	"context"
	"fmt"
	"github.com/filipowm/terraform-provider-unifi/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// IPv4 returns a validator which ensures that a string value is a valid IPv4 address.
func IPv4() validator.String {
	return ipv4Validator{}
}

var _ validator.String = ipv4Validator{}

type ipv4Validator struct{}

func (v ipv4Validator) Description(_ context.Context) string {
	return "value must be a valid IPv4 address"
}

func (v ipv4Validator) MarkdownDescription(_ context.Context) string {
	return "value must be a valid IPv4 address"
}

func (v ipv4Validator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		return
	}

	if !utils.IsIPv4(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IPv4 Address",
			fmt.Sprintf("Value %q is not a valid IPv4 address", value),
		)
		return
	}
}

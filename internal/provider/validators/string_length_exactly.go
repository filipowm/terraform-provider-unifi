package validators

import (
	"context"
	"fmt"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func StringLengthExactly(len int) validator.String {
	return stringLengthExactlyValidator{len: len}
}

type stringLengthExactlyValidator struct {
	len int
}

func (v stringLengthExactlyValidator) invalidUsageMessage() string {
	return fmt.Sprintf("length cannot be less than zero")
}

func (v stringLengthExactlyValidator) Description(_ context.Context) string {
	return fmt.Sprintf("string length must be exactly %d", v.len)
}

func (v stringLengthExactlyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringLengthExactlyValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if v.len < 0 {
		resp.Diagnostics.Append(
			validatordiag.InvalidValidatorUsageDiagnostic(
				req.Path,
				"StringLengthExactly",
				v.invalidUsageMessage(),
			),
		)
		return
	}

	value := req.ConfigValue
	if !base.IsDefined(value) {
		return
	}
	val := value.ValueString()
	if len(val) != v.len {
		resp.Diagnostics.Append(
			validatordiag.InvalidAttributeValueDiagnostic(
				req.Path,
				v.Description(ctx),
				fmt.Sprintf("%s (length: %d)", val, len(val)),
			),
		)
	}
}

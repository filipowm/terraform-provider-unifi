package validators_test

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func newStringValidatorRequestResponse(value string) (validator.StringRequest, *validator.StringResponse) {
	req := validator.StringRequest{
		ConfigValue: types.StringValue(value),
		Path:        path.Empty(),
	}
	resp := validator.StringResponse{
		Diagnostics: []diag.Diagnostic{},
	}
	return req, &resp
}

// Helper function to convert types.String to tftypes.Value
func stringToTfValue(value types.String) tftypes.Value {
	if value.IsNull() {
		return tftypes.NewValue(tftypes.String, nil)
	} else if value.IsUnknown() {
		return tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	}
	return tftypes.NewValue(tftypes.String, value.ValueString())
}

// Helper function to convert types.Bool to tftypes.Value
func boolToTfValue(value types.Bool) tftypes.Value {
	if value.IsNull() {
		return tftypes.NewValue(tftypes.Bool, nil)
	} else if value.IsUnknown() {
		return tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue)
	}
	return tftypes.NewValue(tftypes.Bool, value.ValueBool())
}

package validators

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

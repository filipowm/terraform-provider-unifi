package utils

import (
	"context"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ListElementsAs(list types.List, target interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	if !base.IsDefined(list) {
		return diags
	}
	if diagErr := list.ElementsAs(context.Background(), target, false); diagErr != nil {
		diags = append(diags, diagErr...)
	}
	return diags
}

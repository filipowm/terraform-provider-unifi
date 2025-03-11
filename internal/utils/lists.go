package utils

import (
	"context"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func DefaultEmptyList(elementType attr.Type) defaults.List {
	return listdefault.StaticValue(EmptyList(elementType))
}

func EmptyList(elementType attr.Type) types.List {
	return types.ListValueMust(elementType, []attr.Value{})
}

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

func ListElementsToString(ctx context.Context, list types.List) (string, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	if !base.IsDefined(list) {
		return "", diags
	}
	if list.ElementType(ctx) == types.StringType {
		var target []string
		diags.Append(ListElementsAs(list, &target)...)
		if diags.HasError() {
			return "", diags
		}
		return JoinNonEmpty(target, ","), diags
	}
	diags.AddError("List is not a list of types.StringType", "List is not a list of strings")
	return "", diags
}

func StringToListElements(ctx context.Context, value string) (types.List, diag.Diagnostics) {
	countries := SplitAndTrim(value, ",")
	if len(countries) == 0 {
		return types.ListNull(types.StringType), diag.Diagnostics{}
	}
	list, diags := types.ListValueFrom(ctx, types.StringType, countries)
	if diags.HasError() {
		return types.ListNull(types.StringType), diags
	}
	return list, diags
}

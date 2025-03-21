package types

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"strings"
)

type NestedObject interface {
	AttributeTypes() map[string]attr.Type
}

func ObjectNull(obj interface{}) (basetypes.ObjectValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	if nested, ok := obj.(NestedObject); ok {
		obj := types.ObjectNull(nested.AttributeTypes())
		return obj, diags
	}
	diags.AddError("Invalid object type", fmt.Sprintf("Expected NestedObject, got: %T", obj))
	return types.ObjectNull(map[string]attr.Type{}), diags
}

func ObjectValueMust(ctx context.Context, obj interface{}) basetypes.ObjectValue {
	if nested, ok := obj.(NestedObject); ok {
		val, diags := types.ObjectValueFrom(ctx, nested.AttributeTypes(), obj)
		if diags.HasError() {
			// This could potentially be added to the diag package.
			diagsStrings := make([]string, 0, len(diags))

			for _, diagnostic := range diags {
				diagsStrings = append(diagsStrings, fmt.Sprintf(
					"%s | %s | %s",
					diagnostic.Severity(),
					diagnostic.Summary(),
					diagnostic.Detail()))
			}

			panic("ObjectValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
		}
		return val
	}
	panic(fmt.Sprintf("ObjectValueMust received invalid object type. Expected NestedObject, got: %T", obj))
}

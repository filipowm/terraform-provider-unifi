package types

import (
	"context"
	"fmt"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// MACType is a custom string type for MAC addresses. Two values are treated as
// semantically equal when they refer to the same hardware address regardless of
// case or separator (e.g. "AA-BB-CC-DD-EE-FF" == "aa:bb:cc:dd:ee:ff").
//
// This replaces the previous plan-time NormalizeMAC modifier, which rewrote the
// configured value to the controller's canonical form at plan time. Rewriting a
// known configuration value is rejected by Terraform on create (there is no
// prior state to justify the difference), producing a "Provider produced an
// invalid plan" error. Semantic equality avoids that: the configured value is
// preserved verbatim in state, while a differently-formatted-but-equal value
// from the controller (or a later config edit) does not produce a perpetual
// diff. It is the Plugin Framework analogue of the legacy
// utils.MacDiffSuppressFunc.
type MACType struct {
	basetypes.StringType
}

var _ basetypes.StringTypable = MACType{}

func (t MACType) Equal(o attr.Type) bool {
	other, ok := o.(MACType)
	if !ok {
		return false
	}
	return t.StringType.Equal(other.StringType)
}

func (t MACType) String() string {
	return "types.MACType"
}

func (t MACType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return MACValue{StringValue: in}, nil
}

func (t MACType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

func (t MACType) ValueType(_ context.Context) attr.Value {
	return MACValue{}
}

// MACValue is the value type produced by MACType.
type MACValue struct {
	basetypes.StringValue
}

var _ basetypes.StringValuableWithSemanticEquals = MACValue{}

func (v MACValue) Type(_ context.Context) attr.Type {
	return MACType{}
}

func (v MACValue) Equal(o attr.Value) bool {
	other, ok := o.(MACValue)
	if !ok {
		return false
	}
	return v.StringValue.Equal(other.StringValue)
}

// StringSemanticEquals returns true when both values denote the same MAC address
// after normalization (lowercase, colon-separated). Null/unknown values fall
// back to strict equality since their content cannot be compared.
func (v MACValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(MACValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			fmt.Sprintf("expected value type %T but got %T. Please report this to the provider developers.", v, newValuable),
		)
		return false, diags
	}

	if v.IsNull() || v.IsUnknown() || newValue.IsNull() || newValue.IsUnknown() {
		return v.StringValue.Equal(newValue.StringValue), diags
	}

	return utils.CleanMAC(v.ValueString()) == utils.CleanMAC(newValue.ValueString()), diags
}

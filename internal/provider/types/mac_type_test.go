package types_test

import (
	"context"
	"testing"

	ut "github.com/filipowm/terraform-provider-unifi/internal/provider/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestMACValueStringSemanticEquals(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		a    ut.MACValue
		b    basetypes.StringValuable
		want bool
	}{
		"identical canonical": {
			a:    ut.MACValue{StringValue: types.StringValue("aa:bb:cc:dd:ee:ff")},
			b:    ut.MACValue{StringValue: types.StringValue("aa:bb:cc:dd:ee:ff")},
			want: true,
		},
		"case differs": {
			a:    ut.MACValue{StringValue: types.StringValue("AA:BB:CC:DD:EE:FF")},
			b:    ut.MACValue{StringValue: types.StringValue("aa:bb:cc:dd:ee:ff")},
			want: true,
		},
		"separator differs": {
			a:    ut.MACValue{StringValue: types.StringValue("AA-BB-CC-DD-EE-FF")},
			b:    ut.MACValue{StringValue: types.StringValue("aa:bb:cc:dd:ee:ff")},
			want: true,
		},
		"mixed separators": {
			a:    ut.MACValue{StringValue: types.StringValue("00-11:22:33-44:55")},
			b:    ut.MACValue{StringValue: types.StringValue("00:11:22:33:44:55")},
			want: true,
		},
		"different macs": {
			a:    ut.MACValue{StringValue: types.StringValue("aa:bb:cc:dd:ee:ff")},
			b:    ut.MACValue{StringValue: types.StringValue("11:22:33:44:55:66")},
			want: false,
		},
		"both null": {
			a:    ut.MACValue{StringValue: types.StringNull()},
			b:    ut.MACValue{StringValue: types.StringNull()},
			want: true,
		},
		"null vs value": {
			a:    ut.MACValue{StringValue: types.StringNull()},
			b:    ut.MACValue{StringValue: types.StringValue("aa:bb:cc:dd:ee:ff")},
			want: false,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, diags := test.a.StringSemanticEquals(context.Background(), test.b)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if got != test.want {
				t.Fatalf("StringSemanticEquals(%s, %s) = %v, want %v", test.a, test.b, got, test.want)
			}
		})
	}
}

// TestMACTypeValueFromTerraform proves a tftypes string round-trips into a
// MACValue, which is what lets the framework build set elements of MACType.
func TestMACTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	got, err := ut.MACType{}.ValueFromTerraform(context.Background(), tftypes.NewValue(tftypes.String, "AA-BB-CC-DD-EE-FF"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	macVal, ok := got.(ut.MACValue)
	if !ok {
		t.Fatalf("expected MACValue, got %T", got)
	}
	if macVal.ValueString() != "AA-BB-CC-DD-EE-FF" {
		t.Fatalf("value not preserved verbatim: %q", macVal.ValueString())
	}
}

// TestMACTypeSetElementsAsStrings ensures a Set whose element type is MACType can
// be read into a []string, which the resources rely on when building the API
// model (e.g. apgroup.AsUnifiModel and globalSwitch.overlay).
func TestMACTypeSetElementsAsStrings(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	set, diags := types.SetValue(ut.MACType{}, []attr.Value{
		ut.MACValue{StringValue: types.StringValue("AA-BB-CC-DD-EE-FF")},
		ut.MACValue{StringValue: types.StringValue("00:11:22:33:44:55")},
	})
	if diags.HasError() {
		t.Fatalf("SetValue diagnostics: %v", diags)
	}

	var out []string
	if d := set.ElementsAs(ctx, &out, false); d.HasError() {
		t.Fatalf("ElementsAs diagnostics: %v", d)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 elements, got %d: %v", len(out), out)
	}
}

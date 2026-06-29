package types_test

import (
	"context"
	"testing"

	ut "github.com/filipowm/terraform-provider-unifi/internal/provider/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func macSet(values ...attr.Value) types.Set {
	return types.SetValueMust(types.StringType, values)
}

func TestNormalizeMAC(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		plan types.Set
		want types.Set
	}{
		"uppercase and dash forms are canonicalized": {
			plan: macSet(types.StringValue("AA:BB:CC:DD:EE:FF"), types.StringValue("00-11-22-33-44-55")),
			want: macSet(types.StringValue("aa:bb:cc:dd:ee:ff"), types.StringValue("00:11:22:33:44:55")),
		},
		"mixed separators are canonicalized to colon": {
			plan: macSet(types.StringValue("00-11:22:33-44:55")),
			want: macSet(types.StringValue("00:11:22:33:44:55")),
		},
		"already canonical is unchanged": {
			plan: macSet(types.StringValue("00:15:6d:00:00:01")),
			want: macSet(types.StringValue("00:15:6d:00:00:01")),
		},
		"null is left untouched": {
			plan: types.SetNull(types.StringType),
			want: types.SetNull(types.StringType),
		},
		"unknown is left untouched": {
			plan: types.SetUnknown(types.StringType),
			want: types.SetUnknown(types.StringType),
		},
		"set with an unknown element is left untouched": {
			plan: macSet(types.StringValue("AA:BB:CC:DD:EE:FF"), types.StringUnknown()),
			want: macSet(types.StringValue("AA:BB:CC:DD:EE:FF"), types.StringUnknown()),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			req := planmodifier.SetRequest{PlanValue: test.plan}
			resp := planmodifier.SetResponse{PlanValue: test.plan}
			ut.NormalizeMAC().PlanModifySet(context.Background(), req, &resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected diagnostics: %s", resp.Diagnostics.Errors()[0].Detail())
			}
			if !resp.PlanValue.Equal(test.want) {
				t.Fatalf("PlanValue = %s, want %s", resp.PlanValue, test.want)
			}
		})
	}
}

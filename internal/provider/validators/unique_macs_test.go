package validators_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	ut "github.com/filipowm/terraform-provider-unifi/internal/provider/types"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/validators"
)

func TestUniqueMACs(t *testing.T) {
	t.Parallel()

	macVal := func(s string) attr.Value { return ut.MACValue{StringValue: types.StringValue(s)} }

	tests := map[string]struct {
		set     types.Set
		wantErr bool
	}{
		"distinct macs accepted": {
			set: types.SetValueMust(ut.MACType{}, []attr.Value{macVal("aa:bb:cc:dd:ee:ff"), macVal("00:11:22:33:44:55")}),
		},
		"same mac different separators rejected": {
			set:     types.SetValueMust(ut.MACType{}, []attr.Value{macVal("AA-BB-CC-DD-EE-FF"), macVal("aa:bb:cc:dd:ee:ff")}),
			wantErr: true,
		},
		"same mac different case rejected": {
			set:     types.SetValueMust(ut.MACType{}, []attr.Value{macVal("AA:BB:CC:DD:EE:FF"), macVal("aa:bb:cc:dd:ee:ff")}),
			wantErr: true,
		},
		"single element accepted": {
			set: types.SetValueMust(ut.MACType{}, []attr.Value{macVal("aa:bb:cc:dd:ee:ff")}),
		},
		"null set accepted": {
			set: types.SetNull(ut.MACType{}),
		},
		"unknown element skipped": {
			set: types.SetValueMust(ut.MACType{}, []attr.Value{macVal("aa:bb:cc:dd:ee:ff"), ut.MACValue{StringValue: types.StringUnknown()}}),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			req := validator.SetRequest{Path: path.Root("device_macs"), ConfigValue: test.set}
			resp := &validator.SetResponse{}
			validators.UniqueMACs().ValidateSet(context.Background(), req, resp)
			if got := resp.Diagnostics.HasError(); got != test.wantErr {
				t.Fatalf("HasError() = %v, want %v (%v)", got, test.wantErr, resp.Diagnostics)
			}
		})
	}
}

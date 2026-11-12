package settings

import (
	"context"
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func categoriesSet(t *testing.T, categories ...string) types.Set {
	t.Helper()
	elems := make([]attr.Value, 0, len(categories))
	for _, c := range categories {
		elems = append(elems, types.StringValue(c))
	}
	return types.SetValueMust(types.StringType, elems)
}

func TestIpsSchema_EnabledCategoriesIsUnorderedAndNotCarriedFromState(t *testing.T) {
	t.Parallel()

	resp := &resource.SchemaResponse{}
	(&ipsResource{}).Schema(context.Background(), resource.SchemaRequest{}, resp)

	attribute, ok := resp.Schema.Attributes["enabled_categories"].(schema.SetAttribute)
	require.True(t, ok, "enabled_categories must be a set so controller reordering is not drift")
	assert.True(t, attribute.Optional)
	assert.True(t, attribute.Computed)
	// UseStateForUnknown here would echo state's categories on update, which the
	// controller then normalizes (drops retired ones), causing an inconsistent result.
	assert.Empty(t, attribute.PlanModifiers)
}

func TestIpsModel_AsUnifiModel_EnabledCategories(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		categories types.Set
		want       []string
	}{
		"unknown (unconfigured update) is omitted": {
			categories: types.SetUnknown(types.StringType),
			want:       nil,
		},
		"null is omitted": {
			categories: types.SetNull(types.StringType),
			want:       nil,
		},
		"configured unsorted values are sent as configured": {
			categories: categoriesSet(t, "emerging-worm", "emerging-dos", "emerging-scan"),
			want:       []string{"emerging-worm", "emerging-dos", "emerging-scan"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := ipsModel{
				EnabledCategories: tc.categories,
				EnabledNetworks:   types.ListNull(types.StringType),
			}
			out, diags := model.AsUnifiModel(context.Background())
			require.False(t, diags.HasError(), "%v", diags)

			typed, ok := out.(*unifi.SettingIps)
			require.True(t, ok)
			if tc.want == nil {
				assert.Empty(t, typed.EnabledCategories)
			} else {
				assert.ElementsMatch(t, tc.want, typed.EnabledCategories)
			}
		})
	}
}

func TestIpsModel_Merge_EnabledCategories(t *testing.T) {
	t.Parallel()

	t.Run("controller dropping a category wins over prior state", func(t *testing.T) {
		t.Parallel()

		// State/plan held emerging-icmp; the controller no longer returns it.
		model := ipsModel{EnabledCategories: categoriesSet(t, "emerging-dos", "emerging-icmp", "emerging-imap")}
		diags := model.Merge(context.Background(), &unifi.SettingIps{
			EnabledCategories: []string{"emerging-dos", "emerging-imap"},
		})
		require.False(t, diags.HasError(), "%v", diags)

		assert.True(t, categoriesSet(t, "emerging-dos", "emerging-imap").Equal(model.EnabledCategories))
	})

	t.Run("controller sorting is not a difference", func(t *testing.T) {
		t.Parallel()

		model := ipsModel{}
		diags := model.Merge(context.Background(), &unifi.SettingIps{
			EnabledCategories: []string{"emerging-dos", "emerging-scan", "emerging-worm"},
		})
		require.False(t, diags.HasError(), "%v", diags)

		configured := categoriesSet(t, "emerging-worm", "emerging-dos", "emerging-scan")
		assert.True(t, configured.Equal(model.EnabledCategories))
	})

	t.Run("no categories becomes an empty known set", func(t *testing.T) {
		t.Parallel()

		model := ipsModel{EnabledCategories: types.SetUnknown(types.StringType)}
		diags := model.Merge(context.Background(), &unifi.SettingIps{})
		require.False(t, diags.HasError(), "%v", diags)

		assert.False(t, model.EnabledCategories.IsUnknown())
		assert.False(t, model.EnabledCategories.IsNull())
		assert.Empty(t, model.EnabledCategories.Elements())
	})
}

func TestCheckEnabledCategoriesApplied(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	resp := &resource.SchemaResponse{}
	(&ipsResource{}).Schema(ctx, resource.SchemaRequest{}, resp)
	s := resp.Schema

	tests := map[string]struct {
		planned  types.Set
		applied  types.Set
		wantErr  bool
		wantText string
	}{
		"controller dropping a configured category is a clear error": {
			planned:  categoriesSet(t, "emerging-worm", "emerging-icmp", "emerging-dos"),
			applied:  categoriesSet(t, "emerging-dos", "emerging-worm"),
			wantErr:  true,
			wantText: "emerging-icmp",
		},
		"controller reordering is not an error": {
			planned: categoriesSet(t, "emerging-worm", "emerging-dos"),
			applied: categoriesSet(t, "emerging-dos", "emerging-worm"),
		},
		"unconfigured (unknown) categories are not checked": {
			planned: types.SetUnknown(types.StringType),
			applied: categoriesSet(t, "emerging-dos"),
		},
		"unconfigured (null) categories are not checked": {
			planned: types.SetNull(types.StringType),
			applied: categoriesSet(t, "emerging-dos"),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			plan := tfsdk.Plan{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}
			require.False(t, plan.SetAttribute(ctx, path.Root("enabled_categories"), tc.planned).HasError())
			state := tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}
			require.False(t, state.SetAttribute(ctx, path.Root("enabled_categories"), tc.applied).HasError())

			diags := checkEnabledCategoriesApplied(ctx, plan, state)
			require.Equal(t, tc.wantErr, diags.HasError(), "%v", diags)
			if tc.wantErr {
				require.Len(t, diags.Errors(), 1)
				assert.Contains(t, diags.Errors()[0].Detail(), tc.wantText)
			}
		})
	}
}

func TestKeepUnconfiguredCategoriesWhenUnchanged(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	resp := &resource.SchemaResponse{}
	(&ipsResource{}).Schema(ctx, resource.SchemaRequest{}, resp)
	s := resp.Schema
	categories := path.Root("enabled_categories")
	mode := path.Root("ips_mode")

	build := func(t *testing.T, ipsMode string, cats types.Set) tftypes.Value {
		t.Helper()
		st := tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}
		require.False(t, st.SetAttribute(ctx, mode, types.StringValue(ipsMode)).HasError())
		require.False(t, st.SetAttribute(ctx, categories, cats).HasError())
		return st.Raw
	}
	prior := categoriesSet(t, "emerging-dos", "emerging-worm")

	tests := map[string]struct {
		configured types.Set
		stateNull  bool
		planMode   string
		want       types.Set
	}{
		"unconfigured no-op keeps the prior categories": {
			configured: types.SetNull(types.StringType),
			planMode:   "ids",
			want:       prior,
		},
		"unconfigured real update leaves categories unknown": {
			configured: types.SetNull(types.StringType),
			planMode:   "ips",
			want:       types.SetUnknown(types.StringType),
		},
		"configured categories are left to the config": {
			configured: categoriesSet(t, "emerging-scan"),
			planMode:   "ids",
			want:       types.SetUnknown(types.StringType),
		},
		"create is left unknown": {
			configured: types.SetNull(types.StringType),
			stateNull:  true,
			planMode:   "ids",
			want:       types.SetUnknown(types.StringType),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			config := tfsdk.Config{Schema: s, Raw: build(t, "ids", tc.configured)}
			state := tfsdk.State{Schema: s, Raw: build(t, "ids", prior)}
			if tc.stateNull {
				state.Raw = tftypes.NewValue(s.Type().TerraformType(ctx), nil)
			}
			plan := tfsdk.Plan{Schema: s, Raw: build(t, tc.planMode, types.SetUnknown(types.StringType))}

			diags := keepUnconfiguredCategoriesWhenUnchanged(ctx, config, state, &plan)
			require.False(t, diags.HasError(), "%v", diags)

			var got types.Set
			require.False(t, plan.GetAttribute(ctx, categories, &got).HasError())
			assert.True(t, tc.want.Equal(got), "got %v, want %v", got, tc.want)
		})
	}
}

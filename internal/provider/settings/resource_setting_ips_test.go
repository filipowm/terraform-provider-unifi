package settings

import (
	"context"
	"testing"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

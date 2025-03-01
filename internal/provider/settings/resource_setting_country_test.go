package settings

import (
	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSettingCountry_ProperCountryCodeMappingFromModel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		code                string
		expectedNumericCode int
	}{
		{"Poland", "PL", 616},
		{"United States", "US", 840},
		{"Unknown", "WP", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			model := countryModel{
				Code: types.StringValue(tc.code),
			}
			unifiModel, _ := model.AsUnifiModel()
			typed, ok := unifiModel.(*unifi.SettingCountry)
			assert.True(t, ok)
			assert.Equal(t, tc.expectedNumericCode, typed.Code)
		})
	}
}

func TestSettingCountry_ProperCountryCodeMappingToModel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		numericCode  int
		expectedCode string
	}{
		{"Poland", 616, "PL"},
		{"United States", 840, "US"},
		{"Unknown", 0, "Unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			unifiModel := &unifi.SettingCountry{
				Code: tc.numericCode,
			}
			model := countryModel{}
			model.Merge(unifiModel)
			assert.Equal(t, tc.expectedCode, model.Code.ValueString())
		})
	}
}

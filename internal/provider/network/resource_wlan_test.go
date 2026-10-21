package network

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func TestValidateMinimumDataRates(t *testing.T) {
	tests := []struct {
		name      string
		minRate2g int
		minRate5g int
		wantErr   bool
	}{
		{name: "BothUnset", minRate2g: 0, minRate5g: 0},
		{name: "BothSet", minRate2g: 1000, minRate5g: 24000},
		{name: "Only2g", minRate2g: 1000, minRate5g: 0, wantErr: true},
		{name: "Only5g", minRate2g: 0, minRate5g: 24000, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMinimumDataRates(tt.minRate2g, tt.minRate5g)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResourceWLANPlanRejectsSingleMinimumDataRate(t *testing.T) {
	cfg := terraform.NewResourceConfigRaw(map[string]interface{}{
		"name":                      "test",
		"security":                  "open",
		"minimum_data_rate_5g_kbps": 24000,
	})
	_, err := ResourceWLAN().Diff(t.Context(), nil, cfg, nil)
	assert.ErrorContains(t, err, "you must set minimum data rates on both 2g and 5g")
}

package validators

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCountryCodeValidation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		code             string
		validationFailed bool
	}{
		{"Poland", "PL", false},
		{"United States", "US", false},
		{"Empty", "", false},
		{"Too long", "ABC", true},
		{"Too short", "A", true},
		{"Unknown", "WP", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			v := countryCodeAlpha2Validator{}
			req, resp := newStringValidatorRequestResponse(tc.code)
			v.ValidateString(nil, req, resp)
			assert.Equal(t, tc.validationFailed, resp.Diagnostics.HasError())
		})
	}
}

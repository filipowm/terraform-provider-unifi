package validators

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStringLengthExactlyValidation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		value            string
		length           int
		validationFailed bool
	}{
		{"", 0, false},
		{"", 1, true},
		{"a", 0, true},
		{"a", 1, false},
		{"a", 2, true},
		{"ab", 2, false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s-expected-length-%d", tc.value, tc.length), func(t *testing.T) {
			t.Parallel()
			v := StringLengthExactly(tc.length)
			req, resp := newStringValidatorRequestResponse(tc.value)
			v.ValidateString(context.Background(), req, resp)
			assert.Equal(t, tc.validationFailed, resp.Diagnostics.HasError())
		})
	}
}

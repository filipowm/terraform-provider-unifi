package base_test

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/filipowm/terraform-provider-unifi/internal/provider/base"
	"github.com/stretchr/testify/assert"
)

func TestAsVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		versionString string
		expected      string
	}{
		{
			name:          "simple version",
			versionString: "1.0.0",
			expected:      "1.0.0",
		},
		{
			name:          "complex version",
			versionString: "7.2.95",
			expected:      "7.2.95",
		},
		{
			name:          "version with prerelease",
			versionString: "6.0.0-beta1",
			expected:      "6.0.0-beta1",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := base.AsVersion(tt.versionString)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestCheckMinimumControllerVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		versionString string
		expectError   bool
	}{
		{
			name:          "version equal to minimum",
			versionString: "6.0.0",
			expectError:   false,
		},
		{
			name:          "version greater than minimum",
			versionString: "7.0.0",
			expectError:   false,
		},
		{
			name:          "version less than minimum",
			versionString: "5.9.9",
			expectError:   true,
		},
		{
			name:          "invalid version",
			versionString: "invalid",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := base.CheckMinimumControllerVersion(tt.versionString)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestControllerVersionValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		clientVer    string
		testFunc     func(v base.ControllerVersionValidator) diag.Diagnostics
		expectError  bool
		errorMessage string
	}{
		{
			name:      "min version satisfied",
			clientVer: "7.0.0",
			testFunc: func(v base.ControllerVersionValidator) diag.Diagnostics {
				return v.RequireMinVersion("6.0.0")
			},
			expectError: false,
		},
		{
			name:      "min version not satisfied",
			clientVer: "6.0.0",
			testFunc: func(v base.ControllerVersionValidator) diag.Diagnostics {
				return v.RequireMinVersion("7.0.0")
			},
			expectError:  true,
			errorMessage: "Controller version 6.0.0 is less than minimum required version 7.0.0",
		},
		{
			name:      "max version satisfied",
			clientVer: "6.0.0",
			testFunc: func(v base.ControllerVersionValidator) diag.Diagnostics {
				return v.RequireMaxVersion("7.0.0")
			},
			expectError: false,
		},
		{
			name:      "max version not satisfied",
			clientVer: "8.0.0",
			testFunc: func(v base.ControllerVersionValidator) diag.Diagnostics {
				return v.RequireMaxVersion("7.0.0")
			},
			expectError:  true,
			errorMessage: "Controller version 8.0.0 is greater than maximum required version 7.0.0",
		},
		{
			name:      "between version satisfied",
			clientVer: "7.0.0",
			testFunc: func(v base.ControllerVersionValidator) diag.Diagnostics {
				return v.RequireVersionBetween("6.0.0", "8.0.0")
			},
			expectError: false,
		},
		{
			name:      "between version not satisfied - too low",
			clientVer: "5.0.0",
			testFunc: func(v base.ControllerVersionValidator) diag.Diagnostics {
				return v.RequireVersionBetween("6.0.0", "8.0.0")
			},
			expectError:  true,
			errorMessage: "Controller version 5.0.0 is not between required 6.0.0 and 8.0.0",
		},
		{
			name:      "between version not satisfied - too high",
			clientVer: "9.0.0",
			testFunc: func(v base.ControllerVersionValidator) diag.Diagnostics {
				return v.RequireVersionBetween("6.0.0", "8.0.0")
			},
			expectError:  true,
			errorMessage: "Controller version 9.0.0 is not between required 6.0.0 and 8.0.0",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := &base.Client{
				Version: base.AsVersion(tt.clientVer),
			}
			validator := base.NewControllerVersionValidator(client)

			diags := tt.testFunc(validator)

			if tt.expectError {
				assert.True(t, diags.HasError())
				assert.Contains(t, diags.Errors()[0].Detail(), tt.errorMessage)
			} else {
				assert.False(t, diags.HasError())
			}
		})
	}
}

func TestControllerVersionValidatorNilClient(t *testing.T) {
	t.Parallel()

	validator := base.NewControllerVersionValidator(nil)
	diags := validator.RequireMinVersion("6.0.0")

	assert.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Summary(), "Controller version not available")
}

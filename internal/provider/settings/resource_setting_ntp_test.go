package settings

import (
	"context"
	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNtpModel_AsUnifiModel_Auto(t *testing.T) {
	t.Parallel()

	// Test case for "auto" mode
	model := ntpModel{
		Mode:       types.StringValue("auto"),
		NtpServer1: types.StringValue("time.google.com"),
		NtpServer2: types.StringValue("pool.ntp.org"),
		NtpServer3: types.StringValue("0.pool.ntp.org"),
		NtpServer4: types.StringValue("1.pool.ntp.org"),
	}
	model.ID = types.StringValue("test-id")

	// Convert to UnifiModel
	unifiModel, diags := model.AsUnifiModel(context.Background())

	// Verify no diagnostics errors
	assert.False(t, diags.HasError())

	// Verify correct type conversion
	typed, ok := unifiModel.(*unifi.SettingNtp)
	assert.True(t, ok, "Expected model to be *unifi.SettingNtp")

	// Verify ID and mode are set correctly
	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "auto", typed.SettingPreference)

	// In auto mode, all server fields should be empty regardless of input values
	assert.Equal(t, "", typed.NtpServer1)
	assert.Equal(t, "", typed.NtpServer2)
	assert.Equal(t, "", typed.NtpServer3)
	assert.Equal(t, "", typed.NtpServer4)
}

func TestNtpModel_AsUnifiModel_Manual(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		server1   types.String
		server2   types.String
		server3   types.String
		server4   types.String
		expected1 string
		expected2 string
		expected3 string
		expected4 string
	}{
		{
			name:      "All servers set",
			server1:   types.StringValue("time.google.com"),
			server2:   types.StringValue("pool.ntp.org"),
			server3:   types.StringValue("0.pool.ntp.org"),
			server4:   types.StringValue("1.pool.ntp.org"),
			expected1: "time.google.com",
			expected2: "pool.ntp.org",
			expected3: "0.pool.ntp.org",
			expected4: "1.pool.ntp.org",
		},
		{
			name:      "Only one server set",
			server1:   types.StringValue("time.google.com"),
			server2:   types.StringNull(),
			server3:   types.StringNull(),
			server4:   types.StringNull(),
			expected1: "time.google.com",
			expected2: "",
			expected3: "",
			expected4: "",
		},
		{
			name:      "Mixed null, unknown and empty values",
			server1:   types.StringValue("time.google.com"),
			server2:   types.StringNull(),
			server3:   types.StringUnknown(),
			server4:   types.StringValue(""),
			expected1: "time.google.com",
			expected2: "",
			expected3: "",
			expected4: "",
		},
		{
			name:      "All null values",
			server1:   types.StringNull(),
			server2:   types.StringNull(),
			server3:   types.StringNull(),
			server4:   types.StringNull(),
			expected1: "",
			expected2: "",
			expected3: "",
			expected4: "",
		},
		{
			name:      "All unknown values",
			server1:   types.StringUnknown(),
			server2:   types.StringUnknown(),
			server3:   types.StringUnknown(),
			server4:   types.StringUnknown(),
			expected1: "",
			expected2: "",
			expected3: "",
			expected4: "",
		},
		{
			name:      "All empty string values",
			server1:   types.StringValue(""),
			server2:   types.StringValue(""),
			server3:   types.StringValue(""),
			server4:   types.StringValue(""),
			expected1: "",
			expected2: "",
			expected3: "",
			expected4: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create model with manual mode and test case values
			model := ntpModel{
				Mode:       types.StringValue("manual"),
				NtpServer1: tc.server1,
				NtpServer2: tc.server2,
				NtpServer3: tc.server3,
				NtpServer4: tc.server4,
			}
			model.ID = types.StringValue("test-id")

			// Convert to UnifiModel
			unifiModel, diags := model.AsUnifiModel(context.Background())

			// Verify no diagnostics errors
			assert.False(t, diags.HasError())

			// Verify correct type conversion
			typed, ok := unifiModel.(*unifi.SettingNtp)
			assert.True(t, ok, "Expected model to be *unifi.SettingNtp")

			// Verify ID and mode are set correctly
			assert.Equal(t, "test-id", typed.ID)
			assert.Equal(t, "manual", typed.SettingPreference)

			// Verify server values based on test case expectations
			assert.Equal(t, tc.expected1, typed.NtpServer1)
			assert.Equal(t, tc.expected2, typed.NtpServer2)
			assert.Equal(t, tc.expected3, typed.NtpServer3)
			assert.Equal(t, tc.expected4, typed.NtpServer4)
		})
	}
}

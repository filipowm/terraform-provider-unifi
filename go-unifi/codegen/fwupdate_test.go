package main

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFirmwareUpdateApiFilter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		key      string
		value    string
		expected string
	}{
		{"channel", "channel", "release", "eq~~channel~~release"},
		{"product", "product", "unifi-controller", "eq~~product~~unifi-controller"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			filter := firmwareUpdateApiFilter(tc.key, tc.value)

			a.Equal(tc.expected, filter)
		})
	}
}

func TestMarshalJSONDataLink(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		link         firmwareUpdateApiResponseEmbeddedFirmwareDataLink
		expectedJSON string
	}{
		{
			"nil",
			firmwareUpdateApiResponseEmbeddedFirmwareDataLink{Href: nil},
			"{\"href\":\"\"}",
		},
		{
			"with value",
			func() firmwareUpdateApiResponseEmbeddedFirmwareDataLink {
				u, err := url.Parse("https://example.com/firmware")
				require.NoError(t, err) // error checking in test setup
				return firmwareUpdateApiResponseEmbeddedFirmwareDataLink{Href: u}
			}(),
			"{\"href\":\"https://example.com/firmware\"}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			b, err := json.Marshal(&tc.link)

			require.NoError(t, err)
			a.JSONEq(tc.expectedJSON, string(b))
		})
	}
}

func TestUnmarshalJSONDataLink(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		jsonStr       string
		shouldError   bool
		errorContains string
		expectedHref  *string
	}{
		{
			"valid",
			`{"href": "https://example.com/firmware"}`,
			false,
			"",
			func(s string) *string { return &s }("https://example.com/firmware"),
		},
		{
			"null",
			`{"href": null}`,
			false,
			"",
			nil,
		},
		{
			"missing",
			`{}`,
			false,
			"",
			nil,
		},
		{
			"non-string",
			`{"href": 123}`,
			true,
			"expected string for href",
			nil,
		},
		{
			"invalid json",
			`{"href": }`,
			true,
			"",
			nil,
		},
		{
			"invalid URL",
			`{"href": "://missing"}`,
			true,
			"missing protocol scheme",
			nil,
		},
		{
			"empty",
			`{"href": ""}`,
			false,
			"",
			func(s string) *string { return &s }(""),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			var link firmwareUpdateApiResponseEmbeddedFirmwareDataLink

			err := json.Unmarshal([]byte(tc.jsonStr), &link)

			if tc.shouldError {
				require.Error(t, err)
				if tc.errorContains != "" {
					require.ErrorContains(t, err, tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				if tc.expectedHref == nil {
					a.Nil(link.Href)
				} else {
					require.NotNil(t, link.Href)
					a.Equal(*tc.expectedHref, link.Href.String())
				}
			}
		})
	}
}

func TestFirmwareUpdateApiResponse_Complete(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	ver, err := version.NewVersion("1.2.3")
	require.NoError(t, err)
	u, err := url.Parse("https://example.com/download")
	require.NoError(t, err)

	req := firmwareUpdateApiResponse{
		Embedded: firmwareUpdateApiResponseEmbedded{
			Firmware: []firmwareUpdateApiResponseEmbeddedFirmware{
				{
					Channel:  "release",
					Created:  "2020-01-01T00:00:00Z",
					Id:       "unique-id",
					Platform: "debian",
					Product:  "unifi-controller",
					Version:  ver,
					Links: firmwareUpdateApiResponseEmbeddedFirmwareLinks{
						Data: firmwareUpdateApiResponseEmbeddedFirmwareDataLink{
							Href: u,
						},
					},
				},
			},
		},
	}

	jsonBytes, err := json.Marshal(req)
	require.NoError(t, err)

	var newReq firmwareUpdateApiResponse
	err = json.Unmarshal(jsonBytes, &newReq)
	require.NoError(t, err)

	require.Len(t, newReq.Embedded.Firmware, 1)
	fw := newReq.Embedded.Firmware[0]
	a.Equal("release", fw.Channel)
	a.Equal("debian", fw.Platform)
	a.NotNil(fw.Version)
	a.Equal("https://example.com/download", fw.Links.Data.Href.String())
}

package main

import (
	"archive/zip"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a temporary zip file with given entries. 'entries' maps file names to their content.
func createTempZipFile(t *testing.T, entries map[string]string) string {
	t.Helper()
	tempDir := t.TempDir()
	tempFileName := filepath.Join(tempDir, "test.zip")
	tempFile, err := os.Create(tempFileName)
	require.NoError(t, err, "Failed to create temp zip file")
	// We need to truncate and write zip contents
	w := zip.NewWriter(tempFile)
	for name, content := range entries {
		f, err := w.Create(name)
		require.NoError(t, err, "Failed to add entry %s", name)
		_, err = f.Write([]byte(content))
		require.NoError(t, err, "Failed to write content for %s", name)
	}
	err = w.Close()
	require.NoError(t, err, "Failed to close zip writer")
	err = tempFile.Close()
	require.NoError(t, err, "Failed to close temp file")
	return tempFile.Name()
}

// Test when the output directory already exists. In this case, DownloadAndExtract should not call downloadJarFn or extractJSONFn.
func TestDownloadAndExtract_WithExistingDirectory(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	tempDir := t.TempDir()
	testURL, _ := url.Parse("http://example.com/test.deb")

	err := DownloadAndExtract(*testURL, tempDir)

	r.NoError(err, "Expected no error when directory exists")
}

// // Test when output path is not a directory.
func TestDownloadAndExtract_PathNotDirectory(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	tempDir := t.TempDir()
	tempFilePath := filepath.Join(tempDir, "dummy")
	_, err := os.Create(tempFilePath)
	r.NoError(err, "Failed to create temp file")
	testURL, _ := url.Parse("http://example.com/test.deb")

	err = DownloadAndExtract(*testURL, tempFilePath)

	r.Error(err, "Expected error because tempFilePath is not a directory")
	r.ErrorContains(err, tempFilePath+" isn't a directory")
}

// // Test extractJSON when the jar file cannot be opened.
func TestExtractJSON_OpenJarError(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	err := extractJSON("nonexisting.jar", t.TempDir())

	r.Error(err)
	r.ErrorContains(err, "unable to open jar")
}

// Test extractJSON with a valid zip file that contains a JSON file under api/fields/ and no Setting.json (so splitting is skipped).
func TestExtractJSON_NoSettings(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	tempDir := t.TempDir()
	jarFile := createTempZipFile(t, map[string]string{"api/fields/dummy.json": "{\"key\": \"value\"}"})

	err := extractJSON(jarFile, tempDir)
	r.NoError(err)

	// Check that dummy.json has been extracted
	expectedPath := filepath.Join(tempDir, "dummy.json")
	data, err := os.ReadFile(expectedPath)
	r.NoError(err, "Expected file %s to exist", expectedPath)
	r.JSONEq("{\"key\": \"value\"}", string(data), "Extracted file content mismatch")
}

// Test extractJSON with Setting.json present, so that it splits settings into individual files.
func TestExtractJSON_WithSettings(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	tempDir := t.TempDir()
	entries := map[string]string{"api/fields/Setting.json": "{\"foo\": {\"bar\": 1}}"}
	jarFile := createTempZipFile(t, entries)

	err := extractJSON(jarFile, tempDir)
	r.NoError(err)

	// Check that the split settings file exists
	settingFile := filepath.Join(tempDir, "SettingFoo.json")
	data, err := os.ReadFile(settingFile)
	r.NoError(err)
	r.Contains(string(data), "bar")
}

// Test sanitizeExtractedPath with valid input.
func TestSanitizeExtractedPath_Valid(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	r := require.New(t)

	tempDir := t.TempDir()
	filePath := "api/fields/dummy.json"

	result, err := sanitizeExtractedPath(filePath, tempDir)
	r.NoError(err, "Expected nil error from sanitizeExtractedPath")

	expExpected := filepath.Join(tempDir, "dummy.json")
	absExpected, err := filepath.Abs(expExpected)
	r.NoError(err, "Failed to get abs path")
	a.Equal(absExpected, result, "Sanitized path mismatch")
}

// Test extractJSON with invalid Setting.json content, expecting an unmarshal error.
func TestExtractJSON_InvalidSettings(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	tempDir := t.TempDir()
	jarFile := createTempZipFile(t, map[string]string{"api/fields/Setting.json": "invalid json"})

	err := extractJSON(jarFile, tempDir)

	r.Error(err)
	r.ErrorContains(err, "unable to unmarshal settings")
}

package main

import (
	"os"
	"os/exec"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupLogging(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	setupLogging(false, false)
	a.Equal(logrus.InfoLevel, log.Level)

	setupLogging(true, false)
	a.Equal(logrus.DebugLevel, log.Level)

	setupLogging(false, true)
	a.Equal(logrus.TraceLevel, log.Level)

	setupLogging(true, true)
	a.Equal(logrus.TraceLevel, log.Level)
}

// integration tests for the CLI
// these test require Internet access

func execCli(args ...string) (string, error) {
	in := []string{"run", "."}
	in = append(in, args...)
	cmd := exec.Command("go", in...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func TestHelpFlag(t *testing.T) {
	t.Parallel()

	out, err := execCli("-h")

	require.Error(t, err)
	assert.Contains(t, out, "Usage: codegen [OPTIONS] version")
}

func TestInvalidFlag(t *testing.T) {
	t.Parallel()

	out, err := execCli("-invalid")

	require.Error(t, err)
	assert.Contains(t, out, "flag provided but not defined: -invalid")
}

func TestDefaultVersion(t *testing.T) {
	t.Parallel()

	out, err := execCli("-version-base-dir", t.TempDir(), "-output-dir", t.TempDir())

	require.NoError(t, err)
	assert.Contains(t, out, "UniFi Controller version")
}

func testGenerate(t *testing.T, opts *options) error {
	t.Helper()

	setupLogging(false, false)
	if opts.versionBaseDir == "" {
		opts.versionBaseDir = t.TempDir()
	}
	if opts.outputDir == "" {
		opts.outputDir = t.TempDir()
	}
	if opts.firmwareUpdateApi == "" {
		opts.firmwareUpdateApi = defaultFirmwareUpdateApi
	}
	return generate(*opts)
}

func TestNonExistentVersion(t *testing.T) {
	t.Parallel()

	err := testGenerate(t, &options{version: "1.2.3"})

	require.Error(t, err)
}

func TestInvalidVersion(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	err := testGenerate(t, &options{version: "invalid-version"})

	r.Error(err)
	r.ErrorContains(err, "Malformed")
	r.ErrorContains(err, "invalid-version")
}

func TestGenerateLatest(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	opts := &options{version: LatestVersionMarker}

	err := testGenerate(t, opts)
	r.NoError(err)

	files, err := os.ReadDir(opts.versionBaseDir)
	r.NoError(err)
	assert.NotEmptyf(t, files, "version base dir '%s' should not be empty", opts.versionBaseDir)

	files, err = os.ReadDir(opts.outputDir)
	r.NoError(err)
	assert.NotEmptyf(t, files, "output dir '%s' should not be empty", opts.outputDir)
}

func TestGenerateDownloadOnly(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	opts := &options{version: LatestVersionMarker, downloadOnly: true}

	err := testGenerate(t, opts)
	r.NoError(err)

	files, err := os.ReadDir(opts.versionBaseDir)
	r.NoError(err)
	assert.NotEmptyf(t, files, "version base dir '%s' should not be empty", opts.versionBaseDir)

	files, err = os.ReadDir(opts.outputDir)
	r.NoError(err) // test generated dir
	assert.Emptyf(t, files, "output dir '%s' should be empty", opts.outputDir)
}

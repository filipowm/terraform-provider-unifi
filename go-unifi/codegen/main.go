package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func usage() {
	fmt.Printf("Usage: %s [OPTIONS] version\n", path.Base(os.Args[0]))
	fmt.Printf("version can be a specific version or '%s' (default) for the latest UniFi Controller version\n", LatestVersionMarker)
	flag.PrintDefaults()
}

func setupLogging(debugEnabled, traceEnabled bool) {
	log.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
		ForceColors:            true,
		FullTimestamp:          false,
	})
	if traceEnabled {
		log.SetLevel(logrus.TraceLevel)
	} else if debugEnabled {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}
}

type options struct {
	versionBaseDir     string
	outputDir          string
	downloadOnly       bool
	version            string
	firmwareUpdateApi  string
	customizationsPath string
}

func main() {
	flag.Usage = usage

	versionBaseDirFlag := flag.String("version-base-dir", ".", "The base directory for version JSON files")
	outputDirFlag := flag.String("output-dir", ".", "The output directory of the generated Go code")
	downloadOnly := flag.Bool("download-only", false, "Only download and build the API structures JSON directory, do not generate")
	debugFlag := flag.Bool("debug", false, "Enable debug logging")
	traceFlag := flag.Bool("trace", false, "Enable trace logging")

	flag.CommandLine.Init(os.Args[0], flag.PanicOnError) // set error handling to panic if parse ends with error
	flag.Parse()
	setupLogging(*debugFlag, *traceFlag)
	specifiedVersion := strings.TrimSpace(flag.Arg(0))
	if specifiedVersion == "" {
		specifiedVersion = LatestVersionMarker // default to latest version
	}
	err := generate(options{
		versionBaseDir:     *versionBaseDirFlag,
		outputDir:          *outputDirFlag,
		downloadOnly:       *downloadOnly,
		version:            specifiedVersion,
		firmwareUpdateApi:  defaultFirmwareUpdateApi,
		customizationsPath: "customizations.yml",
	})
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func generate(opts options) error {
	p := NewUnifiVersionProvider(opts.firmwareUpdateApi)
	unifiVersion, err := p.ByVersionMarker(opts.version)
	if err != nil {
		return fmt.Errorf("unable to determine version and download URL for Unifi version %s: %w", opts.version, err)
	}

	log.Infof("UniFi Controller version: %s", unifiVersion.Version)
	log.Infof("UniFi Controller download URL: %s", unifiVersion.DownloadUrl.String())

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable to determine working directory: %w", err)
	}
	var structuresDir string
	if path.IsAbs(opts.versionBaseDir) {
		structuresDir = opts.versionBaseDir
	} else {
		structuresDir = filepath.Join(wd, opts.versionBaseDir)
	}
	structuresDir = filepath.Join(structuresDir, fmt.Sprintf("v%s", unifiVersion.Version))
	log.Infoln("Downloading UniFi Controller API structures definitions...")
	err = DownloadAndExtract(*unifiVersion.DownloadUrl, structuresDir)
	if err != nil {
		return fmt.Errorf("unable to download and extract UniFi Controller API structures definitions: %w", err)
	}
	log.Infof("Downloaded UniFi Controller API structures definitions in %s", structuresDir)

	if opts.downloadOnly {
		log.Infoln("Structure JSONs ready!")
		return nil
	}

	log.Infoln("Generating resources code...")

	var outDir string
	if path.IsAbs(opts.outputDir) {
		outDir = opts.outputDir
	} else {
		outDir = filepath.Join(wd, opts.outputDir)
	}
	customizer, err := NewCodeCustomizer(opts.customizationsPath)
	if err != nil {
		return fmt.Errorf("unable to create code customizer: %w", err)
	}
	if err = generateCode(structuresDir, outDir, *customizer); err != nil {
		return fmt.Errorf("unable to generate resources code: %w", err)
	}

	log.Infof("Writing version file...")
	if err = writeVersionFile(unifiVersion.Version, outDir); err != nil {
		return fmt.Errorf("failed to write version file to %s: %w", outDir, err)
	}

	basepath := filepath.Dir(wd)
	if err = writeVersionRepoMarkerFile(unifiVersion.Version, basepath); err != nil {
		return fmt.Errorf("failed to write version file to %s: %w", basepath, err)
	}

	log.Infof("Generated resources in %s", outDir)
	return nil
}

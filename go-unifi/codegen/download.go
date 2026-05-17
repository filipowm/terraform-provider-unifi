package main

import (
	"archive/tar"
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/ulikunitz/xz"
	"github.com/xor-gate/ar"
)

func DownloadAndExtract(downloadUrl url.URL, outputDir string) error {
	// Check if output directory exists, if not create and perform extraction

	if created, err := ensurePath(outputDir); err != nil {
		return fmt.Errorf("unable to create output directory %s: %w", outputDir, err)
	} else if created {
		log.Debugf("downloading UniFi Controller package from: %s", downloadUrl.String())
		jarFile, err := downloadJar(downloadUrl, outputDir)
		if err != nil {
			return err
		}

		log.Debugf("extracting JSON files with API structures from: %s to: %s", jarFile, outputDir)
		if err = extractJSON(jarFile, outputDir); err != nil {
			return err
		}

		log.Debugf("JSON files extracted to: %s", outputDir)
		_, err = os.Stat(outputDir)
		if err != nil {
			return err
		}
	}
	if targetInfo, err := os.Stat(outputDir); err != nil {
		return err
	} else if !targetInfo.IsDir() {
		return errors.New("fields info isn't a directory")
	}
	return nil
}

func downloadJar(downloadUrl url.URL, outputDir string) (string, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, downloadUrl.String(), nil)
	if err != nil {
		return "", fmt.Errorf("unable to download UniFi Controller deb: %w", err)
	}

	debResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to download UniFi Controller deb: %w", err)
	}
	if debResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unable to download UniFi Controller deb: HTTP%d. Probably it does not exist under %s", debResp.StatusCode, downloadUrl.String())
	}
	defer debResp.Body.Close()

	var uncompressedReader io.Reader
	arReader := ar.NewReader(debResp.Body)
	for {
		header, err := arReader.Next()
		if errors.Is(err, io.EOF) || header == nil {
			break
		}
		if err != nil {
			return "", fmt.Errorf("in ar next: %w", err)
		}
		if header.Name == "data.tar.xz" {
			uncompressedReader, err = xz.NewReader(arReader)
			if err != nil {
				return "", fmt.Errorf("in xz reader: %w", err)
			}
			break
		}
	}
	if uncompressedReader == nil {
		return "", errors.New("unable to find .deb data file")
	}

	tarReader := tar.NewReader(uncompressedReader)
	var aceJar *os.File
	log.Debugln("extracting ace.jar from downloaded controller package")
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("in next: %w", err)
		}
		if header.Typeflag != tar.TypeReg || header.Name != "./usr/lib/unifi/lib/ace.jar" {
			continue
		}
		aceJar, err = os.Create(filepath.Join(outputDir, "ace.jar"))
		if err != nil {
			return "", fmt.Errorf("unable to create temp file: %w", err)
		}
		_, err = io.Copy(aceJar, tarReader)
		if err != nil {
			return "", fmt.Errorf("unable to write ace.jar temp file: %w", err)
		}
	}
	if aceJar == nil {
		return "", errors.New("unable to find ace.jar")
	}
	defer aceJar.Close()
	log.Debugf("ace.jar extracted to: %s", aceJar.Name())
	return aceJar.Name(), nil
}

func extractJSON(jarFile, fieldsDir string) error {
	jarZip, err := zip.OpenReader(jarFile)
	if err != nil {
		return fmt.Errorf("unable to open jar: %w", err)
	}
	defer jarZip.Close()

	log.Tracef("opened jar %s with %d files", jarFile, len(jarZip.File))
	for _, f := range jarZip.File {
		if !strings.HasPrefix(f.Name, "api/fields/") || path.Ext(f.Name) != ".json" {
			continue
		}

		err = func() error {
			log.Tracef("extracting %s", f.Name)
			src, err := f.Open()
			if err != nil {
				return err
			}
			dstPath, err := sanitizeExtractedPath(f.Name, fieldsDir)
			if err != nil {
				return err
			}
			dst, err := os.Create(dstPath)
			if err != nil {
				return err
			}
			defer dst.Close()
			_, err = io.Copy(dst, src)
			log.Debugf("extracted %s", f.Name)
			if err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			return fmt.Errorf("unable to write JSON file: %w", err)
		}
	}

	settingsData, err := os.ReadFile(filepath.Join(fieldsDir, "Setting.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("unable to open settings file: %w", err)
	}

	var settings map[string]interface{}
	err = json.Unmarshal(settingsData, &settings)
	if err != nil {
		return fmt.Errorf("unable to unmarshal settings: %w", err)
	}

	log.Debugf("splitting Settings.json into individual setting files")
	for settingKey, settingValue := range settings {
		settingName := strcase.ToCamel(settingKey)
		fileName := fmt.Sprintf("Setting%s.json", settingName)
		log.Tracef("splitting %s", fileName)

		data, err := json.MarshalIndent(settingValue, "", "  ")
		if err != nil {
			return fmt.Errorf("unable to marshal setting %q: %w", settingKey, err)
		}

		err = os.WriteFile(filepath.Join(fieldsDir, fileName), data, 0o755)
		if err != nil {
			return fmt.Errorf("unable to write new settings file: %w", err)
		}
		log.Tracef("splitted %s into %s", settingKey, fileName)
	}

	return nil
}

func sanitizeExtractedPath(filePath, destinationDir string) (string, error) {
	absDestinationDir, err := filepath.Abs(destinationDir)
	if err != nil {
		return "", err
	}

	absFilePath, err := filepath.Abs(filepath.Join(destinationDir, filepath.Base(filePath)))
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(absFilePath, absDestinationDir) {
		return "", fmt.Errorf("invalid file path: %s", filePath)
	}

	return absFilePath, nil
}

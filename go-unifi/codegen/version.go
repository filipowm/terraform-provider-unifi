package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/format"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
)

const (
	LatestVersionMarker = "latest"
	baseDownloadUrl     = "https://dl.ui.com/unifi/%s/unifi_sysvinit_all.deb"
)

type UnifiVersion struct {
	Version     *version.Version
	DownloadUrl *url.URL
}

func NewUnifiVersion(unifiVersion *version.Version, downloadUrl *url.URL) *UnifiVersion {
	return &UnifiVersion{
		Version:     unifiVersion,
		DownloadUrl: downloadUrl,
	}
}

type UnifiVersionProvider interface {
	Latest() (*UnifiVersion, error)
	ByVersionMarker(versionMarker string) (*UnifiVersion, error)
}

type defaultUnifiVersionProvider struct {
	firmwareUpdateApi string
}

func NewUnifiVersionProvider(firmwareUpdateApi string) UnifiVersionProvider {
	return &defaultUnifiVersionProvider{
		firmwareUpdateApi: firmwareUpdateApi,
	}
}

func (p *defaultUnifiVersionProvider) Latest() (*UnifiVersion, error) {
	url, err := url.Parse(p.firmwareUpdateApi)
	if err != nil {
		return nil, err
	}

	query := url.Query()
	query.Add("filter", firmwareUpdateApiFilter("channel", releaseChannel))
	query.Add("filter", firmwareUpdateApiFilter("product", unifiControllerProduct))
	url.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respData firmwareUpdateApiResponse
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return nil, err
	}

	for _, firmware := range respData.Embedded.Firmware {
		if firmware.Platform != debianPlatform || firmware.Version == nil {
			continue
		}
		return NewUnifiVersion(firmware.Version.Core(), firmware.Links.Data.Href), nil
	}

	return nil, errors.New("no Unifi Controller firmware found")
}

func (p *defaultUnifiVersionProvider) ByVersionMarker(versionMarker string) (*UnifiVersion, error) {
	if versionMarker == LatestVersionMarker {
		return p.Latest()
	} else {
		unifiVersion, err := version.NewVersion(versionMarker)
		if err != nil {
			return nil, err
		}
		unifiVersion = unifiVersion.Core()
		downloadUrl := fmt.Sprintf(baseDownloadUrl, unifiVersion)
		unifiDownloadUrl, err := url.Parse(downloadUrl)
		if err != nil {
			return nil, err
		}
		return NewUnifiVersion(unifiVersion, unifiDownloadUrl), nil
	}
}

func writeVersionFile(version *version.Version, outDir string) error {
	versionGo := []byte(fmt.Sprintf(`
// Generated code. DO NOT EDIT.

package unifi

const UnifiVersion = %q
`, version.Core()))

	versionGo, err := format.Source(versionGo)
	if err != nil {
		return err
	}

	_, err = writeGeneratedFile(outDir, "version", string(versionGo))
	return err
}

func writeVersionRepoMarkerFile(version *version.Version, outDir string) error {
	versionRepoMarker := []byte(version.Core().String())
	return os.WriteFile(filepath.Join(outDir, ".unifi-version"), versionRepoMarker, 0o644)
}

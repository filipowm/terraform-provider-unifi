package main

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/hashicorp/go-version"
)

const defaultFirmwareUpdateApi = "https://fw-update.ubnt.com/api/firmware-latest"

const (
	debianPlatform         = "debian"
	releaseChannel         = "release"
	unifiControllerProduct = "unifi-controller"
)

type firmwareUpdateApiResponse struct {
	Embedded firmwareUpdateApiResponseEmbedded `json:"_embedded"`
}

type firmwareUpdateApiResponseEmbedded struct {
	Firmware []firmwareUpdateApiResponseEmbeddedFirmware `json:"firmware"`
}

type firmwareUpdateApiResponseEmbeddedFirmware struct {
	Channel  string                                         `json:"channel"`
	Created  string                                         `json:"created"`
	Id       string                                         `json:"id"`
	Platform string                                         `json:"platform"`
	Product  string                                         `json:"product"`
	Version  *version.Version                               `json:"version"`
	Links    firmwareUpdateApiResponseEmbeddedFirmwareLinks `json:"_links"`
}

type firmwareUpdateApiResponseEmbeddedFirmwareDataLink struct {
	Href *url.URL `json:"href"`
}

func (l *firmwareUpdateApiResponseEmbeddedFirmwareDataLink) MarshalJSON() ([]byte, error) {
	var href string
	if l.Href != nil {
		href = l.Href.String()
	}

	aux := struct {
		Href string `json:"href"`
	}{
		Href: href,
	}

	return json.Marshal(aux)
}

func (l *firmwareUpdateApiResponseEmbeddedFirmwareDataLink) UnmarshalJSON(j []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(j, &m); err != nil {
		return err
	}
	if href, exists := m["href"]; exists && href != nil {
		strHref, ok := href.(string)
		if !ok {
			return fmt.Errorf("expected string for href, got %T", href)
		}
		u, err := url.Parse(strHref)
		if err != nil {
			return err
		}
		l.Href = u
	}
	return nil
}

type firmwareUpdateApiResponseEmbeddedFirmwareLinks struct {
	Data firmwareUpdateApiResponseEmbeddedFirmwareDataLink `json:"data"`
}

func firmwareUpdateApiFilter(key, value string) string {
	return fmt.Sprintf("%s~~%s~~%s", "eq", key, value)
}

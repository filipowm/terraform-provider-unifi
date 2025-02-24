package provider

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/go-version"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

func IsServerErrorContains(err error, messageContains string) bool {
	if err == nil {
		return false
	}
	var se *unifi.ServerError
	if errors.As(err, &se) {
		if strings.Contains(se.Message, messageContains) {
			return true
		}
		// check details
		if se.Details != nil {
			for _, m := range se.Details {
				if strings.Contains(m.Message, messageContains) {
					return true
				}
			}
		}
	}
	return false
}

type ClientConfig struct {
	Username       string
	Password       string
	ApiKey         string
	Url            string
	Site           string
	Insecure       bool
	HttpConfigurer func() http.RoundTripper
}

func NewClient(cfg *ClientConfig) (*Client, error) {
	unifiClient, err := unifi.NewClient(&unifi.ClientConfig{
		URL:                      cfg.Url,
		User:                     cfg.Username,
		Password:                 cfg.Password,
		APIKey:                   cfg.ApiKey,
		HttpRoundTripperProvider: cfg.HttpConfigurer,
		ValidationMode:           unifi.DisableValidation,
		Logger:                   unifi.NewDefaultLogger(unifi.WarnLevel),
	})

	if err != nil {
		return nil, err
	}
	err = CheckMinimumControllerVersion(unifiClient.Version())
	log.Printf("[TRACE] Unifi controller version: %q", unifiClient.Version())
	if err != nil {
		return nil, err
	}
	c := &Client{
		Client:  unifiClient,
		Site:    cfg.Site,
		Version: version.Must(version.NewVersion(unifiClient.Version())),
	}
	if cfg.ApiKey != "" && !c.SupportsApiKeyAuthentication() {
		return nil, fmt.Errorf("API key authentication is not supported on this controller version: %s, you must be on %s or higher", c.Version, ControllerVersionApiKeyAuth)
	}
	return c, nil
}

type Client struct {
	unifi.Client
	Site    string
	Version *version.Version
}

func CreateHttpTransport(insecure bool) http.RoundTripper {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,

		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}
}

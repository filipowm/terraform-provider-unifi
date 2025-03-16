package base

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"log"
	"net"
	"net/http"
	"time"
)

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

func (c *Client) ResolveSite(res SiteAware) string {
	if res == nil || IsEmptyString(res.GetRawSite()) {
		return c.Site
	}
	return res.GetSite()
}

func (c *Client) ResolveSiteFromConfig(ctx context.Context, config tfsdk.Config) (string, diag.Diagnostics) {
	var site types.String
	diags := config.GetAttribute(ctx, path.Root("site"), &site)
	if diags.HasError() {
		return "", diags
	}
	if IsEmptyString(site) {
		return c.Site, diags
	}
	return site.ValueString(), diags
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

func checkClientConfigured(client *Client) diag.Diagnostics {
	diags := diag.Diagnostics{}
	if client == nil {
		diags.AddError(
			"Client Not Configured",
			"Expected configured client. Please report this issue to the provider developers.",
		)
	}
	return diags
}

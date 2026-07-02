package base

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	ut "github.com/filipowm/terraform-provider-unifi/internal/provider/types"
	"github.com/filipowm/terraform-provider-unifi/internal/provider/utils"
)

type ClientConfig struct {
	Username       string
	Password       string
	APIKey         string
	URL            string
	Site           string
	Insecure       bool
	HTTPConfigurer func() http.RoundTripper
	// MaxRetries controls how many additional attempts the HTTP layer makes for
	// transient controller responses (network errors, HTTP 5xx/429 and
	// HTML-instead-of-JSON bodies). 0 (the default) disables retries entirely,
	// preserving the historical behavior with zero overhead.
	MaxRetries int
}

func NewClient(cfg *ClientConfig) (*Client, error) {
	config := &unifi.ClientConfig{
		URL:                      cfg.URL,
		User:                     cfg.Username,
		Password:                 cfg.Password,
		APIKey:                   cfg.APIKey,
		HttpRoundTripperProvider: cfg.HTTPConfigurer,
		ValidationMode:           unifi.DisableValidation,
		Logger:                   unifi.NewDefaultLogger(unifi.WarnLevel),
	}
	// Opt-in retrying transport for transient controller responses. When
	// MaxRetries == 0 the provider is left untouched so behavior is identical to
	// before this feature existed (the default).
	if cfg.MaxRetries > 0 {
		baseConfigurer := cfg.HTTPConfigurer
		insecure := cfg.Insecure
		maxRetries := cfg.MaxRetries
		config.HttpRoundTripperProvider = func() http.RoundTripper {
			var next http.RoundTripper
			if baseConfigurer != nil {
				next = baseConfigurer()
			} else {
				next = CreateHTTPTransport(insecure)
			}
			return newRetryRoundTripper(next, maxRetries)
		}
	}
	if cfg.Username != "" && cfg.Password != "" {
		config.User = cfg.Username
		config.Password = cfg.Password
		config.RememberMe = true
	} else {
		config.APIKey = cfg.APIKey
	}
	unifiClient, err := unifi.NewClient(config)
	if err != nil {
		return nil, err
	}
	err = CheckMinimumControllerVersion(unifiClient.Version())
	log.Printf("[TRACE] Unifi controller version: %q", unifiClient.Version())
	if err != nil {
		return nil, err
	}
	c := &Client{
		Client:  NewRetryableUnifiClient(unifiClient),
		Site:    cfg.Site,
		Version: version.Must(version.NewVersion(unifiClient.Version())),
	}
	if cfg.APIKey != "" && !c.SupportsAPIKeyAuthentication() {
		return nil, fmt.Errorf("API key authentication is not supported on this controller version: %s, you must be on %s or higher", c.Version, ControllerVersionAPIKeyAuth)
	}
	return c, nil
}

func NewRetryableUnifiClient(client unifi.Client) unifi.Client {
	return &RetryableUnifiClient{
		Client:     client,
		loginMutex: sync.Mutex{},
	}
}

type RetryableUnifiClient struct {
	unifi.Client
	loginMutex sync.Mutex
}

func (c *RetryableUnifiClient) relogin(err error) error {
	c.loginMutex.Lock()
	defer c.loginMutex.Unlock()
	loginErr := c.Login()
	if loginErr != nil {
		return fmt.Errorf("tried relogging in after %w, but failed: %w", err, loginErr)
	}
	return nil
}

func (c *RetryableUnifiClient) Do(ctx context.Context, method string, apiPath string, reqBody interface{}, respBody interface{}) error {
	err := c.Client.Do(ctx, method, apiPath, reqBody, respBody)
	if err != nil && utils.IsServerErrorStatusCode(err, 401) {
		err := c.relogin(err)
		if err != nil {
			return err
		}
		return c.Client.Do(ctx, method, apiPath, reqBody, respBody)
	}
	return err
}

type Client struct {
	unifi.Client
	Site    string
	Version *version.Version
}

func (c *Client) ResolveSite(res SiteAware) string {
	if res == nil || ut.IsEmptyString(res.GetRawSite()) {
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
	if ut.IsEmptyString(site) {
		return c.Site, diags
	}
	return site.ValueString(), diags
}

func CreateHTTPTransport(insecure bool) http.RoundTripper {
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
			InsecureSkipVerify: insecure, //nolint:gosec // insecure TLS is an opt-in provider feature for self-signed UniFi controller certificates
		},
	}
}

// defaultRetryBackoff is the base delay between retry attempts. The effective
// delay grows linearly with the attempt number.
const defaultRetryBackoff = 500 * time.Millisecond

// retryRoundTripper wraps an http.RoundTripper and retries requests that fail
// with transient controller responses: network/connection errors, HTTP 5xx and
// 429 status codes, and HTML-instead-of-JSON bodies (which the controller
// occasionally returns under parallel load). It only retries idempotent
// requests whose body can be replayed, so it never risks duplicate creates.
type retryRoundTripper struct {
	next       http.RoundTripper
	maxRetries int
	backoff    time.Duration
}

// newRetryRoundTripper returns a RoundTripper that retries transient failures.
// When maxRetries <= 0 the original transport is returned unwrapped so there is
// zero overhead and identical behavior to not using retries at all.
func newRetryRoundTripper(next http.RoundTripper, maxRetries int) http.RoundTripper {
	if maxRetries <= 0 {
		return next
	}
	if next == nil {
		next = http.DefaultTransport
	}
	return &retryRoundTripper{
		next:       next,
		maxRetries: maxRetries,
		backoff:    defaultRetryBackoff,
	}
}

// isIdempotentMethod reports whether it is safe to retry the given HTTP method
// without risking unintended side effects (e.g. duplicate resource creation).
func isIdempotentMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodOptions:
		return true
	default:
		return false
	}
}

func (rt *retryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Only retry idempotent requests. For non-idempotent methods (e.g. POST), or
	// requests whose body cannot be replayed, pass through with a single attempt.
	if !isIdempotentMethod(req.Method) || (req.Body != nil && req.GetBody == nil) {
		return rt.next.RoundTrip(req)
	}

	var resp *http.Response
	var err error
	for attempt := 0; ; attempt++ {
		// Ensure a fresh body for each attempt so the request can be replayed.
		if req.GetBody != nil {
			body, gerr := req.GetBody()
			if gerr != nil {
				if resp != nil {
					return resp, err
				}
				return nil, gerr
			}
			req.Body = body
		}

		resp, err = rt.next.RoundTrip(req)

		if (err == nil && !rt.shouldRetryResponse(resp)) || attempt >= rt.maxRetries {
			return resp, err
		}

		// Drain and close the previous response body before retrying so the
		// underlying connection can be reused.
		if resp != nil && resp.Body != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}

		// Backoff while respecting the request context.
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(rt.backoff * time.Duration(attempt+1)):
		}
	}
}

// shouldRetryResponse reports whether a (non-error) response is transient and
// should be retried. It buffers the body when peeking for HTML content so the
// caller still receives a fully-readable response when no retry happens.
func (rt *retryRoundTripper) shouldRetryResponse(resp *http.Response) bool {
	if resp == nil {
		return false
	}
	if resp.StatusCode >= http.StatusInternalServerError || resp.StatusCode == http.StatusTooManyRequests {
		return true
	}
	if strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "text/html") {
		return true
	}
	return responseBodyLooksLikeHTML(resp)
}

// responseBodyLooksLikeHTML buffers the response body, restores it so it remains
// readable, and reports whether it begins with '<' (an HTML/XML document rather
// than the expected JSON payload).
func responseBodyLooksLikeHTML(resp *http.Response) bool {
	if resp.Body == nil {
		return false
	}
	data, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(data))
	resp.ContentLength = int64(len(data))
	if err != nil {
		return false
	}
	trimmed := bytes.TrimSpace(data)
	return len(trimmed) > 0 && trimmed[0] == '<'
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

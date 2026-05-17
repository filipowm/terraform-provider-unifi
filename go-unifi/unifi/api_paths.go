package unifi

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

const (
	apiPath   = "/api"
	apiV2Path = "/v2/api"

	apiPathNew   = "/proxy/network/api"
	apiV2PathNew = "/proxy/network/v2/api"

	loginPath    = "/api/login"
	loginPathNew = "/api/auth/login"

	statusPath    = "/status"
	statusPathNew = "/proxy/network/status"

	uploadPath    = "/upload"
	uploadPathNew = "/proxy/network/upload"

	logoutPath = "/api/logout"

	defaultUserAgent = "go-unifi/0.0.1"

	ApiKeyHeader      = "X-Api-Key"
	CsrfHeader        = "X-Csrf-Token"
	UserAgentHeader   = "User-Agent"
	AcceptHeader      = "Accept"
	ContentTypeHeader = "Content-Type"
)

// APIPaths defines the URL paths used by the client.
type APIPaths struct {
	ApiPath    string
	ApiV2Path  string
	LoginPath  string
	StatusPath string
	LogoutPath string
	UploadPath string
}

var (
	OldStyleAPI = APIPaths{
		ApiPath:    apiPath,
		ApiV2Path:  apiV2Path,
		LoginPath:  loginPath,
		StatusPath: statusPath,
		LogoutPath: logoutPath,
		UploadPath: uploadPath,
	}
	NewStyleAPI = APIPaths{
		ApiPath:    apiPathNew,
		ApiV2Path:  apiV2PathNew,
		LoginPath:  loginPathNew,
		StatusPath: statusPathNew,
		LogoutPath: logoutPath,
		UploadPath: uploadPathNew,
	}
)

// determineApiStyle checks the base URL to decide which API style to use and sets the apiPaths accordingly.
func (c *client) determineApiStyle() error {
	c.Debug("Determining API style")
	ctx, cancel := c.newRequestContext()
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL.String(), nil)
	if err != nil {
		return err
	}

	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: c.http.Transport,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Discard response body to avoid leaks
	_, _ = io.Copy(io.Discard, resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		c.Debug("Using new style API")
		c.apiPaths = &NewStyleAPI
	case http.StatusFound:
		c.Debug("Using old style API")
		c.apiPaths = &OldStyleAPI
	default:
		return fmt.Errorf("expected 200 or 302 status code, but got: %d", resp.StatusCode)
	}

	if c.apiPaths == &OldStyleAPI && c.credentials.IsAPIKey() {
		return errors.New("unable to use API key authentication with old style API. Switch to user/pass authentication or update controller to latest version")
	}

	return nil
}

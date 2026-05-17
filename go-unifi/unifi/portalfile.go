package unifi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// just to fix compile issues with the import.
var (
	_ context.Context
	_ fmt.Formatter
	_ json.Marshaler
)

type PortalFile struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	ContentType  string `json:"content_type,omitempty"`
	LastModified int    `json:"last_modified,omitempty"`
	Filename     string `json:"filename,omitempty"`
	FileSize     int    `json:"filesize,omitempty"`
	MD5          string `json:"md5,omitempty"`
	URL          string `json:"url,omitempty"`
}

func (dst *PortalFile) UnmarshalJSON(b []byte) error {
	type Alias PortalFile
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(dst),
	}

	err := json.Unmarshal(b, &aux)
	if err != nil {
		return fmt.Errorf("unable to unmarshal alias: %w", err)
	}

	return nil
}

func (c *client) UploadPortalFile(ctx context.Context, site string, filepath string) (*PortalFile, error) {
	var respBody struct {
		Meta Meta         `json:"meta"`
		Data []PortalFile `json:"data"`
	}

	err := c.UploadFile(ctx, fmt.Sprintf("%s/s/%s/portalfile", c.apiPaths.UploadPath, site), filepath, "file", &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) == 0 {
		return nil, ErrNotFound
	}
	return &respBody.Data[0], nil
}

func (c *client) UploadPortalFileFromReader(ctx context.Context, site string, reader io.Reader, filename string) (*PortalFile, error) {
	var respBody struct {
		Meta Meta         `json:"meta"`
		Data []PortalFile `json:"data"`
	}

	err := c.UploadFileFromReader(ctx, fmt.Sprintf("%s/s/%s/portalfile", c.apiPaths.UploadPath, site), reader, filename, "file", &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) == 0 {
		return nil, ErrNotFound
	}
	return &respBody.Data[0], nil
}

func (c *client) DeletePortalFile(ctx context.Context, site, id string) error {
	err := c.Delete(ctx, fmt.Sprintf("s/%s/rest/portalfile/%s", site, id), struct{}{}, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *client) ListPortalFiles(ctx context.Context, site string) ([]PortalFile, error) {
	var respBody struct {
		Meta Meta         `json:"meta"`
		Data []PortalFile `json:"data"`
	}

	err := c.Get(ctx, fmt.Sprintf("s/%s/rest/portalfile", site), nil, &respBody)
	if err != nil {
		return nil, err
	}

	return respBody.Data, nil
}

func (c *client) GetPortalFile(ctx context.Context, site, id string) (*PortalFile, error) {
	var respBody struct {
		Meta Meta         `json:"meta"`
		Data []PortalFile `json:"data"`
	}

	err := c.Get(ctx, fmt.Sprintf("s/%s/rest/portalfile/%s", site, id), nil, &respBody)
	if err != nil {
		return nil, err
	}

	if len(respBody.Data) != 1 {
		return nil, ErrNotFound
	}

	return &respBody.Data[0], nil
}

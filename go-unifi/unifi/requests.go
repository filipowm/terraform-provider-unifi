package unifi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

// marshalRequest marshals the request body to an io.Reader. Returns nil if reqBody is nil.
func marshalRequest(reqBody interface{}) (io.Reader, error) {
	if reqBody == nil {
		return nil, nil //nolint: nilnil
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(reqBytes), nil
}

// buildRequestURL constructs the full URL for a given apiPath using the client's baseURL and apiPaths.
func (c *client) buildRequestURL(apiPath string) (*url.URL, error) {
	reqURL, err := url.Parse(apiPath)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(apiPath, "/") && !reqURL.IsAbs() {
		reqURL.Path = path.Join(c.apiPaths.ApiPath, reqURL.Path)
	}
	return c.baseURL.ResolveReference(reqURL), nil
}

// validateRequestBody validates the request body if validation is enabled.
func (c *client) validateRequestBody(reqBody interface{}) error {
	if reqBody != nil && c.validationMode != DisableValidation {
		c.Trace("Validating request body")
		if err := c.validator.Validate(reqBody); err != nil {
			if c.validationMode == HardValidation {
				return fmt.Errorf("failed validating request body: %w", err)
			} else {
				c.Warnf("failed validating request body: %s", err)
			}
		}
	}
	return nil
}

// newRequestContext creates a new context for the request with a timeout if specified.
func (c *client) newRequestContext() (context.Context, context.CancelFunc) {
	ctx := context.Background()
	if c.timeout != 0 {
		return context.WithTimeout(ctx, c.timeout)
	}
	return ctx, func() {}
}

// executeRequest executes an HTTP request with the given context, method, URL, body, and headers.
// It applies interceptors, handles errors, and decodes the response body if provided.
// Returns an error if the request or response handling fails.
func (c *client) executeRequest(ctx context.Context, method, apiPath string, body io.Reader, headers http.Header, respBody interface{}) error {
	url, err := c.buildRequestURL(apiPath)
	if err != nil {
		return fmt.Errorf("unable to create request URL: %w", err)
	}
	c.Debugf("Executing request: %s %s", method, url.String())

	req, err := http.NewRequestWithContext(ctx, method, url.String(), body)
	if err != nil {
		return fmt.Errorf("unable to create request: %s %s %w", method, apiPath, err)
	}

	if c.useLocking {
		c.lock.Lock()
		c.Trace("Acquired lock for request")
		defer c.lock.Unlock()
	}

	c.Trace("Executing request interceptors")
	for _, interceptor := range c.interceptors {
		if err := interceptor.InterceptRequest(req); err != nil {
			return err
		}
	}

	// Set headers if provided overriding any coming from interceptors
	for key, values := range headers {
		// delete headers if already exist to be able to override them
		if req.Header.Get(key) != "" {
			req.Header.Del(key)
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("unable to perform request: %s %s %w", method, apiPath, err)
	}
	defer resp.Body.Close()

	c.Trace("Executing response interceptors")
	for _, interceptor := range c.interceptors {
		if err := interceptor.InterceptResponse(resp); err != nil {
			return err
		}
	}

	c.Trace("Checking for errors in response")
	if err := c.errorHandler.HandleError(resp); err != nil {
		return err
	}

	// If no response body is expected, return
	if respBody == nil || resp.ContentLength == 0 {
		c.Trace("No response body to decode")
		return nil
	}

	c.Trace("Decoding response body")
	err = json.NewDecoder(resp.Body).Decode(respBody)
	if err != nil {
		return fmt.Errorf("unable to decode body: %s %s %w", method, apiPath, err)
	}
	return nil
}

// UploadFile uploads a file to the UniFi controller.
// It takes a context, API path, file path, field name, and additional form fields.
// The file is uploaded as multipart/form-data.
// It returns the response body and an error if the operation fails.
func (c *client) UploadFile(ctx context.Context, apiPath, filePath, fieldName string, respBody interface{}) error {
	c.Tracef("Uploading file: %s to %s", filePath, apiPath)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("unable to open file for upload: %w", err)
	}
	defer file.Close()
	return c.UploadFileFromReader(ctx, apiPath, file, filepath.Base(filePath), fieldName, respBody)
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// NOTE! This is a copy of the function from the mime/multipart package, but allows to set custom mimetype.
func createFormFile(w *multipart.Writer, mimeType, fieldname, filename string) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, escapeQuotes(fieldname), escapeQuotes(filename)))
	h.Set("Content-Type", mimeType)
	return w.CreatePart(h)
}

// UploadFileFromReader uploads a file to the UniFi controller from an io.Reader.
// It takes a context, API path, reader, filename, field name, and additional form fields.
// The file is uploaded as multipart/form-data.
func (c *client) UploadFileFromReader(ctx context.Context, apiPath string, reader io.Reader, filename, fieldName string, respBody interface{}) error {
	c.Tracef("Uploading file: %s to %s", filename, apiPath)

	// Read the entire content into a buffer first to avoid deadlock. I tied using TeeReader and Pipe but ended up in deadlock.
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return fmt.Errorf("unable to read file content into buffer: %w", err)
	}
	contentReader := bytes.NewReader(buf.Bytes())

	if fieldName == "" {
		fieldName = "file"
	}

	// Detect MIME type from the first reader
	mt, err := mimetype.DetectReader(contentReader)
	if err != nil {
		return fmt.Errorf("unable to detect file mimetype: %w", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := createFormFile(writer, mt.String(), fieldName, filename)
	if err != nil {
		return fmt.Errorf("unable to create form file: %w", err)
	}
	// reinit reader
	contentReader = bytes.NewReader(buf.Bytes())
	// Copy the file content to the form field from the second reader
	if _, err = io.Copy(part, contentReader); err != nil {
		return fmt.Errorf("unable to copy file content: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("unable to close multipart writer: %w", err)
	}

	headers := http.Header{}
	headers.Set("Content-Type", writer.FormDataContentType())
	headers.Set("X-Requested-With", "XMLHttpRequest") // TODO if not provided, the response will be 404. UniFi bug?

	return c.executeRequest(ctx, http.MethodPost, apiPath, body, headers, respBody)
}

// Do performs an HTTP request using the given method, apiPath, request body, and decodes the response into respBody.
// It validates the request body, applies interceptors, and decodes the HTTP response into respBody if provided.
// It returns an error if the request or response handling fails.
func (c *client) Do(ctx context.Context, method, apiPath string, reqBody interface{}, respBody interface{}) error {
	c.Tracef("Performing request: %s %s", method, apiPath)

	if err := c.validateRequestBody(reqBody); err != nil {
		return err
	}

	body, err := marshalRequest(reqBody)
	if err != nil {
		return fmt.Errorf("unable to marshal request: %w", err)
	}

	headers := http.Header{}
	if reqBody != nil {
		headers.Set("Content-Type", "application/json")
	}

	return c.executeRequest(ctx, method, apiPath, body, headers, respBody)
}

// Get sends an HTTP GET request to the specified API path with the provided request body,
// and decodes the HTTP response into respBody.
// It is a convenience wrapper around Do.
func (c *client) Get(ctx context.Context, apiPath string, reqBody interface{}, respBody interface{}) error {
	return c.Do(ctx, http.MethodGet, apiPath, reqBody, respBody)
}

// Post sends an HTTP POST request to the specified API path with the provided request body,
// and decodes the HTTP response into respBody.
// It is a convenience wrapper around Do.
func (c *client) Post(ctx context.Context, apiPath string, reqBody interface{}, respBody interface{}) error {
	return c.Do(ctx, http.MethodPost, apiPath, reqBody, respBody)
}

// Put sends an HTTP PUT request to the specified API path with the provided request body,
// and decodes the HTTP response into respBody.
// It is a convenience wrapper around Do.
func (c *client) Put(ctx context.Context, apiPath string, reqBody interface{}, respBody interface{}) error {
	return c.Do(ctx, http.MethodPut, apiPath, reqBody, respBody)
}

// Delete sends an HTTP DELETE request to the specified API path with the provided request body,
// and decodes the HTTP response into respBody.
// It is a convenience wrapper around Do.
func (c *client) Delete(ctx context.Context, apiPath string, reqBody interface{}, respBody interface{}) error {
	return c.Do(ctx, http.MethodDelete, apiPath, reqBody, respBody)
}

# File Uploads in go-unifi

This document describes how to use the file upload functionality in the go-unifi client.

## Overview

The go-unifi client provides two methods for uploading files to the UniFi controller:

1. `UploadFile` - Upload a file from a file path on disk
2. `UploadFileFromReader` - Upload a file from an `io.Reader` (e.g., from memory, network stream, etc.)

Both methods use the `multipart/form-data` format for file uploads, which is required by the UniFi controller.

## Examples

### Uploading a file from disk

```go
package main

import (
	"context"
	"log"

	"github.com/filipowm/go-unifi/unifi"
)

func main() {
	// Create a client
	client, err := unifi.NewClient(&unifi.ClientConfig{
		URL:      "https://your-unifi-controller:8443",
		User:     "your-username",
		Password: "your-password",
	})
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

	// Prepare any additional form fields if needed
	formFields := map[string]string{
		"description": "My uploaded file",
	}

	// Upload the file to the controller
	var response map[string]interface{} // Adjust this type based on the expected response
	err = client.UploadFile(
		context.Background(),
		"/api/s/default/upload", // The API endpoint to upload to
		"/path/to/your/file.txt", // Path to the file on disk
		"file", // Form field name for the file
		formFields, // Additional form fields
		&response, // Response structure to capture the result
	)
	if err != nil {
		log.Fatalf("Error uploading file: %v", err)
	}

	log.Printf("Upload successful: %v", response)
}
```

### Uploading a file from memory

```go
package main

import (
	"bytes"
	"context"
	"log"

	"github.com/paultyng/go-unifi/unifi"
)

func main() {
	// Create a client
	client, err := unifi.NewClient(&unifi.ClientConfig{
		URL:      "https://your-unifi-controller:8443",
		User:     "your-username",
		Password: "your-password",
	})
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

	// Create file content in memory
	fileContent := []byte("This is some test content to upload")
	reader := bytes.NewReader(fileContent)

	// Upload the file from the reader
	var response map[string]interface{} // Adjust this type based on the expected response
	err = client.UploadFileFromReader(
		context.Background(),
		"/api/s/default/upload", // The API endpoint to upload to
		reader, // Reader with the file content
		"myfile.txt", // Filename to use in the upload
		"file", // Form field name for the file
		nil, // No additional form fields
		&response, // Response structure to capture the result
	)
	if err != nil {
		log.Fatalf("Error uploading file: %v", err)
	}

	log.Printf("Upload successful: %v", response)
}
```

## API Reference

### UploadFile

```go
func (c *client) UploadFile(ctx context.Context, apiPath, filePath, fieldName string, formFields map[string]string, respBody interface{}) error
```

Uploads a file to the UniFi controller from a file path.

Parameters:
- `ctx`: The context for the request
- `apiPath`: The API endpoint path to upload the file to
- `filePath`: Path to the file on disk
- `fieldName`: Form field name for the file (defaults to "file" if empty)
- `formFields`: Additional form fields to include in the upload (can be nil)
- `respBody`: Structure to decode the response into (can be nil)

### UploadFileFromReader

```go
func (c *client) UploadFileFromReader(ctx context.Context, apiPath string, reader io.Reader, filename, fieldName string, formFields map[string]string, respBody interface{}) error
```

Uploads a file to the UniFi controller from an io.Reader.

Parameters:
- `ctx`: The context for the request
- `apiPath`: The API endpoint path to upload the file to
- `reader`: Reader with the file content
- `filename`: Name of the file to use in the upload
- `fieldName`: Form field name for the file (defaults to "file" if empty)
- `formFields`: Additional form fields to include in the upload (can be nil)
- `respBody`: Structure to decode the response into (can be nil)

## Notes

- These methods use `POST` requests for file uploads
- The UniFi controller typically expects files to be uploaded with the field name "file", but this can be changed as needed
- The content type for the request is automatically set to "multipart/form-data" with the correct boundary
- All existing client features like interceptors, error handling, and request validation are preserved

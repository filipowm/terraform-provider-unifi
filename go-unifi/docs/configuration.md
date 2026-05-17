# Client Configuration

The UniFi Go SDK client is highly configurable to cater to different needs and environments. This document explains the various configuration options available in the client.

## Authentication Methods

The client supports two authentication protocols:

### API Key Authentication

Use this method for better security and dedicated access. Example:

```go
c, err := unifi.NewClient(&unifi.ClientConfig{
    BaseURL: "https://unifi.localdomain",
    APIKey: "your-api-key",
})
if err != nil {
    log.Fatalf("Error creating client: %v", err)
}
```

### Username/Password Authentication

Alternatively, you can use username and password:

```go
c, err := unifi.NewClient(&unifi.ClientConfig{
    BaseURL: "https://unifi.localdomain",
    Username: "your-username",
    Password: "your-password",
    RememberMe: true, // Optional: prolong the session validity. Might be needed for long-running applications.
})
if err != nil {
    log.Fatalf("Error creating client: %v", err)
}
```

## Validation Modes

The client has three modes of validation for the API models. The modes help to ensure that the data sent to the controller is correct.

- **Soft Validation (`unifi.SoftValidation`)**: Logs warnings for invalid fields, but does not fail the request (default).
- **Hard Validation (`unifi.HardValidation`)**: Returns an error for invalid fields, preventing the request from being sent.
- **Disable Validation (`unifi.DisableValidation`)**: Disables all validations.

Configure the validation mode as follows:

```go
c, err := unifi.NewClient(&unifi.ClientConfig{
    BaseURL: "https://unifi.localdomain",
    APIKey: "your-api-key",
    ValidationMode: unifi.HardValidation,
})
if err != nil {
    log.Fatalf("Error creating client: %v", err)
}
```

## Customizing the HTTP Client

There are two ways to customize the HTTP client used by the UniFi client:
1. Using the `HttpTransportCustomizer`.
2. Using the `HttpRoundTripperProvider`.

Those methods are mutually exclusive, and only one can be used at a time. If both are provided, the `HttpRoundTripperProvider` takes precedence,
unless it returns `nil`, in which case the `HttpTransportCustomizer` is used if defined (or default transport is used).

### Using `HttpTransportCustomizer`

You can provide your own HTTP client transport configuration using the `HttpTransportCustomizer` callback. This is useful if you need to tweak connection settings like timeouts, idle connection settings, 
or TLS configurations:

```go
c, err := unifi.NewClient(&unifi.ClientConfig{
    BaseURL: "https://unifi.localdomain",
    APIKey: "your-api-key",
    HttpTransportCustomizer: func(transport *http.Transport) (*http.Transport, error) {
        transport.MaxIdleConns = 10
        // Customize TLS settings, proxy, etc. as needed
        // You can also create new instance of transport and return it, instead of customizing pre-configured
        return transport, nil
    },
})
if err != nil {
    log.Fatalf("Error creating client: %v", err)
}
```

### Using `HttpRoundTripperProvider`

You can provide your own HTTP client configuration using the `HttpRoundTripperProvider` callback. This is useful if you need to create a custom round tripper, when `http.Transport` is not enough:

```go
c, err := unifi.NewClient(&unifi.ClientConfig{
    BaseURL: "https://unifi.localdomain",
    APIKey: "your-api-key",
    HttpRoundTripperProvider: func() http.RoundTripper {
        // Create a custom HTTP Round Tripper instance
        return &http.Transport{}, nil
    },
})
```

## Using Interceptors

Interceptors let you hook into the request and response flow. They can be used for logging, metrics, or modifying requests/responses.

Implement the [ClientInterceptor](https://pkg.go.dev/github.com/filipowm/go-unifi/unifi#ClientInterceptor) interface:

```go
// LoggingInterceptor logs each request and response
type LoggingInterceptor struct{}

func (l *LoggingInterceptor) InterceptRequest(req *http.Request) error {
    log.Printf("Request: %s %s", req.Method, req.URL)
    return nil
}

func (l *LoggingInterceptor) InterceptResponse(resp *http.Response) error {
    log.Printf("Response status: %d", resp.StatusCode)
    return nil
}

c, err := unifi.NewClient(&unifi.ClientConfig{
    BaseURL: "https://unifi.localdomain",
    APIKey: "your-api-key",
    Interceptors: []unifi.ClientInterceptor{&LoggingInterceptor{}},
})
if err != nil {
    log.Fatalf("Error creating client: %v", err)
}
```

This flexibility allows you to modify client behavior to suit your application's needs.


## Comprehensive Client Configuration Example

The `ClientConfig` struct is the central configuration for initializing the UniFi client. It allows you to
fine-tune every aspect of the client's behavior such as the controller URL, authentication credentials, HTTP timeout,
SSL verification, custom HTTP transport settings, interceptors, error handling, concurrency locking, and request validation modes.

Below is a full example demonstrating how to configure and use all available properties of `ClientConfig` when
initializing the client with `unifi.NewClient`:

```go
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/filipowm/go-unifi/unifi"
)

// customTransportCustomizer customizes the HTTP transport, e.g., setting idle connection limits and TLS options.
func customTransportCustomizer(transport *http.Transport) (*http.Transport, error) {
	transport.MaxIdleConns = 50
	transport.IdleConnTimeout = 120 * time.Second
	// Set a custom TLS configuration
	transport.TLSClientConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	return transport, nil
}

// myErrorHandler implements a custom error handler for HTTP responses.
type myErrorHandler struct{}

func (h *myErrorHandler) HandleError(resp *http.Response) error {
	if resp.StatusCode >= 400 {
		return fmt.Errorf("custom error: received status code %d", resp.StatusCode)
	}
	return nil
}

// customInterceptor is a simple interceptor that adds a custom header to each request.
type customInterceptor struct{}

func (ci *customInterceptor) InterceptRequest(req *http.Request) error {
	req.Header.Set("X-Custom-Header", "CustomValue")
	return nil
}

func (ci *customInterceptor) InterceptResponse(resp *http.Response) error {
	// Additional response processing can be added here if needed.
	return nil
}

func main() {
	// Create a comprehensive client configuration.
	config := &unifi.ClientConfig{
		URL:    "https://unifi.example.com", // Base URL of the UniFi controller (without trailing '/api')
		APIKey: "your-api-key",              // API key for authentication. Alternatively, use User and Pass for user/password auth.
		// User:         "username",                                 // Uncomment and provide if using user/password authentication
		// Password:     "password",                                 // Uncomment and provide if using user/password authentication
		Timeout:        30 * time.Second,                                // Maximum duration to wait for a response
		VerifySSL:      true,                                            // Enable SSL certificate verification
		Interceptors:   []unifi.ClientInterceptor{&customInterceptor{}}, // Custom interceptors for request/response manipulation
		HttpTransportCustomizer: customTransportCustomizer,              // Function to customize the underlying HTTP transport
		UserAgent:      "MyCustomAgent/1.0",                             // Custom User-Agent string
		ErrorHandler:   &myErrorHandler{},                               // Custom error handler for processing HTTP response errors
		UseLocking:     true,                                            // Enable internal locking for safe concurrent request processing
		ValidationMode: unifi.SoftValidation,                            // Validation mode: SoftValidation, HardValidation, or DisableValidation
	}

	// Initialize the UniFi client with the specified configuration.
	client, err := unifi.NewClient(config)
	if err != nil {
		log.Fatalf("Error creating UniFi client: %v", err)
	}

	// Example operation: Retrieve system information from the UniFi controller.
	sysInfo, err := client.GetSystemInformation()
	if err != nil {
		log.Fatalf("Error retrieving system information: %v", err)
	}
	log.Printf("Connected to UniFi Controller version: %s", sysInfo.Version)

	// Further client operations can be performed using the 'client' instance.
	// For example: creating networks, retrieving device information, etc.
}

```

This example demonstrates how to utilize the full range of configuration options provided by `ClientConfig` to create
a highly customizable UniFi client.
# UniFi Go SDK
![GitHub Release](https://img.shields.io/github/v/release/filipowm/go-unifi)
[![Docs](https://img.shields.io/badge/docs-reference-blue)](https://github.com/filipowm/go-unifi/blob/main/docs/readme.md)
[![GoDoc](https://godoc.org/github.com/filipowm/go-unifi?status.svg)](https://godoc.org/github.com/filipowm/go-unifi)
![GitHub branch check runs](https://img.shields.io/github/check-runs/filipowm/go-unifi/main)
![GitHub License](https://img.shields.io/github/license/filipowm/go-unifi)

This SDK provides a Go client for the UniFi Network Controller API. It is used primarily in the [Terraform provider for UniFi](https://github.com/filipowm/terraform-provider-unifi),
but can be used independently for any Go project requiring UniFi Network Controller API integration.

Check out the detailed [documentation](docs/readme.md) for more information.

## Features

- Great UniFi Network Controller API coverage through automated code generation and manually added code for undocumented endpoints
- Easy to use client with support for API Key and username/password authentication
- Generated data models from UniFi Controller API specifications
- Daily automated updates to track the latest UniFi Controller versions
- Support for multiple UniFi Controller versions
- Strong typing for all API models with Go structs

## Supported UniFi Controller Versions

Any version after 5.12.35 is supported as of now. **Latest version: 9.0.114**.
The SDK is updated daily to track the latest UniFi Controller versions. 
If you encounter any issues with the latest UniFi Controller version, please open an issue.

## Code Generation

The data models and basic REST methods are generated from JSON specifications found in the UniFi Controller JAR files. Those JSON specs show all fields and the associated regex/validation information.
This ensures accuracy and completeness of the API coverage. However, code generation is not perfect and some endpoints might be missing, or not covered perfectly by the generated code. We hope to rely on official API specifications as soon as they are available.

To regenerate the code for the latest UniFi Controller version:

```bash
go generate unifi/codegen.go
```

**Note:** While the current code generation approach works, we're exploring better ways to extract API specifications. There is no official API specifications available, and the UniFi Controller JAR is obfuscated, making it
challenging to directly use Java classes. Contributions and suggestions for improvements are welcome!

## Migrating from `paultyng/go-unifi`

If you already use `paultyng/go-unifi`, you can easily migrate to this SDK, because it is a fork and the SDK is fully compatible with the original one. 
Check out the [migration guide](docs/migrating_from_upstream.md) for information on how to migrate from the upstream `paultyng/go-unifi` SDK.

## Usage

Unifi client support both username/password and API Key authentication. It is recommended to use API Key authentication for better security,
as well as dedicated user restricted to local access only.

### Obtaining an API Key
1. Open your Site in UniFi Site Manager
2. Click on `Control Plane -> Admins & Users`.
3. Select your Admin user.
4. Click `Create API Key`.
5. Add a name for your API Key.
6. Copy the key and store it securely, as it will only be displayed once.
7. Click Done to ensure the key is hashed and securely stored.
8. Use the API Key 🎉

### Client Initialization

```go
c, err := unifi.NewClient(&unifi.ClientConfig{
    BaseURL: "https://unifi.localdomain",
    APIKey: "your-api-key",
})
```

Instead of API Key, you can also use username/password for authentication:

```go
c, err := unifi.NewClient(&unifi.ClientConfig{
    BaseURL: "https://unifi.localdomain",
    Username: "your-username",
    Password: "your-password",
})
```

If you are using self-signed certificates on your UniFi Controller, you can disable certificate verification:

```go
c, err := unifi.NewClient(&unifi.ClientConfig{
    ...
    VerifySSL: false,
})
```

List of available client configuration options is available [here](https://pkg.go.dev/github.com/filipowm/go-unifi/unifi#ClientConfig).

### Customizing HTTP Client

You can customize underlying HTTP client by using `HttpTransportCustomizer` interface:

```go
c, err := unifi.NewClient(&unifi.ClientConfig{
    ...
    HttpTransportCustomizer: func(transport *http.Transport) (*http.Transport, error) {
        transport.MaxIdleConns = 10
        return transport, nil
    },
})
```

### Using interceptors

You can use interceptors to modify requests and responses. This gives you more control over the client behavior
and flexibility to add custom logic.

To use interceptor logic, you need to create a struct implementing [ClientInterceptor](https://pkg.go.dev/github.com/filipowm/go-unifi/unifi#ClientInterceptor) interface.
For example, you can use interceptors to log requests and responses:

```go
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
    ...
    Interceptors: []unifi.ClientInterceptor{&LoggingInterceptor{}},
})
```

### Client-side validation

The SDK provides basic validation for the API models. It is recommended to use it to ensure that the data you are sending 
to the UniFi Controller is correct. The validation is based on the regex and validation rules provided in 
the UniFi Controller API specs extracted from the JAR files.

Client supports 3 modes of validation:
- `unifi.SoftValidation` (_default_) - will log a warning if any of the fields are invalid before sending the request, but will not stop the request
- `unifi.HardValidation` - will return an error if any of the fields are invalid before sending the request
- `unifi.DisableValidation` - will disable validation completely

To change the validation mode, you can use the `ValidationMode` field in the client configuration:

```go
c, err := unifi.NewClient(&unifi.ClientConfig{
    ...
    ValidationMode: unifi.HardValidation,
})
```

If you use hard validation, you can get access to `unifi.ValidationError` struct, which contains information about the validation errors:

```go
n := &unifi.Network{
	Name:     "my-network",
	Purpose:  "invalid-purpose",
	IPSubnet: "10.0.0.10/24",
}
_, err = c.CreateNetwork(ctx, "default", n)

if err != nil {
	validationError := &unifi.ValidationError{}
	errors.As(err, &validationError)
	fmt.Printf("Error: %v\n", validationError)
    fmt.Printf("Root: %v\n", validationError.Root)
}
```

`Root` error is `validator.ValidationErrors` struct from [go-playground/validator](https://pkg.go.dev/github.com/go-playground/validator/v10#ValidationErrors), 
which contains detailed information about the validation errors.

### Examples

List all available networks:
```go
network, err := c.ListNetwork(ctx, "site-name")
```

Create user assigned to network:
```go
user, err := c.CreateUser(ctx, "site-name", &unifi.User{
    Name:      "My Network User",
    MAC:       "00:00:00:00:00:00", 
    NetworkID: network[0].ID, 
    IP:        "10.0.21.37",
})
```

## Plans

- [ ] Support Unifi Controller API V2
  - [x] AP Groups
  - [x] DNS Records
  - [x] Zone-based firewalls
  - [ ] Traffic management
  - [ ] other...?
- [x] Increase API coverage, or modify code generation to rely on the official UniFi Controller API specifications
- [x] Improve error handling (currently only basic error handling is implemented and error details are not propagated)
- [x] Improve client code for better usability
- [x] Support API Key authentication
- [x] Generate client code for currently generated API structures, for use within or outside the Terraform provider
- [ ] Increase test coverage
- [x] Implement validation for fields and structures
- [ ] Extend validators for more complex cases
- [x] Add more documentation and examples
- [ ] Bugfixing...

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change. I will be happy to find additional maintainers!

## Acknowledgment

This project is a fork of [paultyng/go-unifi](https://github.com/paultyng/go-unifi). Huge thanks to Paul Tyng together with the rest of maintainers for creating and maintaining the original SDK,
which provided an excellent foundation for this fork, and is great piece of engineering work. The fork was created to introduce several improvements including keeping it up to date with the latest UniFi Controller versions, more dev-friendly client usage, enhanced error handling, additional API endpoints support,
improved documentation, better test coverage, and various bug fixes. It's goal is to provide a stable, up to date and reliable SDK for the UniFi Network Controller API.

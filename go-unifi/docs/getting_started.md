# Getting Started with UniFi Go SDK

This guide will help you get started with the UniFi Go SDK client. It covers prerequisites, installation, and basic client initialization.
I highly recommend to use the latest version of UniFi Go SDK, as well as update your UniFi Controller to the latest version to ensure compatibility.

## Prerequisites

- Go 1.16 or later

## Installation

Install the UniFi Go SDK by running:

```bash
go get github.com/filipowm/go-unifi
```

If you need to regenerate the client code from the API specifications, run:

```bash
go generate unifi/codegen.go
```

## Initialization

The client supports both API Key authentication and username/password authentication. Below are examples for both methods.
Unifi client support both username/password and API Key authentication. It is recommended to use API Key authentication
together with dedicated user restricted to local access only.

**IMPORTANT:** API Key authentication is available only since UniFi Controller version 9.0.108

### Obtaining an API Key

1. Open your Site in UniFi Site Manager
2. Click on Control Plane -> Admins & Users.
3. Select your Admin user.
4. Click Create API Key.
5. Add a name for your API Key.
6. Copy the key and store it securely, as it will only be displayed once.
7. Click Done to ensure the key is hashed and securely stored.
8. Use the API Key 🎉

### API Key Authentication

```go
c, err := unifi.NewClient(&unifi.ClientConfig{
    BaseURL: "https://unifi.localdomain",
    APIKey: "your-api-key",
})
if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}
```

### Username/Password Authentication

```go
c, err := unifi.NewClient(&unifi.ClientConfig{
    BaseURL: "https://unifi.localdomain",
    Username: "your-username",
    Password: "your-password",
})
if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}
```

You can also configure `Remember Me` option, which will prolong the session validity. Might be required for long-running applications, that require authentication only once.

```go
c, err := unifi.NewClient(&unifi.ClientConfig{
    BaseURL: "https://unifi.localdomain",
    Username: "your-username",
    Password: "your-password",
    RememberMe: true,
})
if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}
```

### Bare Client Initialization

You can also use bare client, which creates a `unifi.Client` without initialization like logging in and getting system information. This can be useful in specific scenarios, when doing such initialization might be an uneeded overhead. To create it you can use the `NewBareClient` function provided in the SDK (see `unifi/client.go`).

Example usage:

```go
c, err := unifi.NewBareClient(&unifi.ClientConfig{
    BaseURL: "https://unifi-controller.example.com",
    APIKey: "your-api-key", // or use Username/Password as needed
    // Configuration for a bare client
})
if err != nil {
    log.Fatalf("Error creating bare client: %v", err)
}
err = c.Login()
if err != nil {
    log.Fatalf("Error logging in: %v", err)
}
```

## Generating Client Code

The UniFi Go SDK uses code generation to provide complete API coverage. To regenerate the client based on the latest specifications, run:

```bash
go generate unifi/codegen.go
```

This will update the generated models and REST methods according to the current UniFi Controller API specifications.


## Usage

Once instantiated, the Bare Client provides direct access to the generated API methods. You can perform operations without the extra layers of processing provided by interceptors or custom validations.
If you use a default site and didn't create any new ones, you can use the `default` site ID.

**Example:**

```go
networks, err := c.ListNetwork(ctx, "default")
if err != nil {
    log.Fatalf("Error listing networks: %v", err)
}

for _, network := range networks {
    fmt.Printf("Network: %s\n", network.Name)
}
```

## Checking if features are supported and enabled

The UniFi Go SDK provides a way to check if a feature is supported and enabled/disabled on the UniFi Controller. 
This can be useful when you want to check if a feature is available before using it. Passed feature names are case-insensitive.

**Example:**

```go
if c.IsFeatureEnabled(ctx, "default", "feature-name") {
    // Feature is enabled
} else {
    // Feature is disabled
}
```

Library comes with a set of predefined feature names, which can be found in `github.com/filipowm/go-unifi/unifi/features` module. You can also use custom feature names.

For example, you can check if the `features.ZoneBasedFirewallMigration` is available on the controller (no `unifi.ErrNotFound` raised) and enabled:
```go
f, err := c.GetFeature(ctx, "default", features.ZoneBasedFirewallMigration)
if err != nil {
    if errors.Is(err, unifi.ErrNotFound) {
        log.Printf("Feature %s unavailable (not found)", features.ZoneBasedFirewallMigration)
    } else {
        log.Fatalf("Error getting feature: %v", err)
    }
    return false
}
return f.FeatureExists // `FeatureExists` is a boolean indicating if the feature is enabled
```

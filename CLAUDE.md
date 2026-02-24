# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Terraform provider for managing Ubiquiti UniFi network controllers. Supports UniFi Controller 6.x+ (UDM, UDM-Pro, UCG). Uses the `github.com/filipowm/go-unifi` client library.

## Commands

```bash
# Build (compilation check)
go build ./...

# Install provider binary for local development with Terraform CLI
make build

# Run all acceptance tests (requires running UniFi controller)
make testacc

# Run a single test
TF_ACC=1 go test ./internal/provider/acctest -run TestAccNetwork -v -count 1 -timeout 20m

# Lint (fix issues automatically)
golangci-lint run --fix

# Format
gofmt -w .
```

Acceptance tests require `TF_ACC=1` and a running UniFi controller configured via environment variables: `UNIFI_API`, `UNIFI_USERNAME`, `UNIFI_PASSWORD` (or `UNIFI_API_KEY`), `UNIFI_SITE`, `UNIFI_INSECURE`.

## Architecture

### Dual Provider (SDKv2 + Plugin Framework)

The provider runs **both** SDKv2 and Plugin Framework v2 simultaneously, muxed together in `main.go` via `tf6muxserver`. Legacy resources use SDKv2 (`internal/provider/provider.go`), new resources use Plugin Framework (`internal/provider/provider_v2.go`). **All new resources must use Plugin Framework.**

### Package Structure

- `internal/provider/base/` — Core abstractions: `GenericResource[T]`, `Client` wrapper, `Model` base struct, `ResourceModel`/`DatasourceModel` interfaces, version gating, feature flags
- `internal/provider/{domain}/` — Resources grouped by domain: `firewall/`, `network/`, `dns/`, `device/`, `settings/`, `routing/`, `portal/`, `user/`, `radius/`, `site/`, `apgroup/`
- `internal/provider/validators/` — Reusable validators (CIDR, IPv4, IPv6, hostname, port range, conditional). Shared across both SDKv2 and Plugin Framework resources
- `internal/provider/types/` — Schema attribute builders (`ID()`, `SiteAttribute()`, list helpers)
- `internal/provider/utils/` — Error handling, CIDR utilities, env variable helpers
- `internal/provider/acctest/` — All acceptance tests
- `internal/provider/testing/` — Test infrastructure (`PreCheck`, `ImportStepWithSite`, `CheckResourceActions`, Docker testcontainers setup)

### Plugin Framework Resource Pattern (New Resources)

Reference implementation: `firewall/resource_firewall_zone.go`

1. **Model struct** — Embeds `base.Model` (provides ID + Site fields), implements `ResourceModel` interface
2. **`AsUnifiModel(ctx)`** — Converts Terraform state → go-unifi API model
3. **`Merge(ctx, other)`** — Merges API response → Terraform state
4. **Resource struct** — Wraps `*base.GenericResource[*myModel]` created via `base.NewGenericResource()` with `ResourceFunctions` for CRUD
5. **Schema** — Uses helpers from `types/` package for common attributes
6. **`ModifyPlan()`** — Optional: version/feature gating (see below)

### SDKv2 Resource Pattern (Legacy Resources)

Uses `*schema.Resource` with `CreateContext`/`ReadContext`/`UpdateContext`/`DeleteContext` functions. State access via `d.Get()`/`d.Set()`. Validators use `validation.StringInSlice()`, `validation.IntBetween()`, etc.

### Settings Resources

Settings are singletons (no Create/Delete, only Read/Update). Use `NewSettingResource` helper in `settings/base_setting_resource.go`. Tests must use `Lock: someMutex` to serialize access.

### Site Multi-Tenancy

All resources are site-aware. Resolution order: resource-level `site` attribute → provider default site ("default"). The `base.Model` struct handles this automatically.

### Client (`base/client.go`)

`RetryableUnifiClient` wraps the go-unifi client with automatic re-login on 401 errors, version detection (`client.Version()`), and feature flag checking.

## Key Patterns

### Trust-the-Write (v10+ API Workaround)

UniFi v10+ controllers accept certain array fields on write but **do not echo them back** on read. The trust-the-write pattern returns the user's input model instead of the API response, extracting only the ID from the response. Used for:

- **IPS settings**: `ad_blocking_configurations`, `dns_filters`, `enabled_categories`, `enabled_networks`, `suppression`
- **USG settings**: `dhcp_relay_servers`

Implementation requires `ImportStateVerifyIgnore` for affected fields in tests. See `settings/resource_setting_ips.go` (`updateSettingIps` function) for the reference implementation. The `settingIpsNoOmit` embedded struct pattern forces empty slices to serialize as `[]` not `null` by shadowing fields without `omitempty` tags.

### Read-Modify-Write (Device Updates)

Device updates must read the current controller state first, then overlay only Terraform-managed fields (Name, PortProfileID, OpMode, PoeMode, AggregateNumPorts) via `mergePortOverrides()`. Without this, sending a partial object replaces controller-managed fields (QoS, storm control, dot1x, LAG), causing `api.err.NotSupportQosConfig`.

### Version Gating

**In resources** (ModifyPlan):
```go
resp.Diagnostics.Append(r.RequireMinVersion("7.2")...)
resp.Diagnostics.Append(r.RequireMinVersionForPath("8.5", path.Root("dns_verification"), req.Config)...)
resp.Diagnostics.Append(r.RequireFeaturesEnabled(ctx, site, features.ZoneBasedFirewall)...)
```

**In tests** (VersionConstraint):
```go
AcceptanceTest(t, AcceptanceTestCase{
    VersionConstraint: ">= 7.3",
    // ...
})
```

### State Upgraders

When changing attribute types (e.g., Int32 → String), implement `ResourceWithUpgradeState` with a JSON-based state migration. See `firewall/resource_firewall_zone_policy.go` (`UpgradeState` method) for an example using `tftypes.ValueFromJSON`.

## Conventions

- **Commits**: Conventional commits format (`feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `test:`)
- **Error handling**: Return `diag.Diagnostics`, use `utils.IsServerErrorStatusCode()` and `utils.IsServerErrorContains()` for API error inspection
- **Test structure**: Use `AcceptanceTest(t, AcceptanceTestCase{...})` with `VersionConstraint`, `Lock` (for singleton settings), and multi-step test cases covering create/import/update/destroy. Import step uses `pt.ImportStepWithSite()` for site:id format.

## Test Environment

Tests use `testcontainers-go` with Docker Compose (`docker-compose.yaml`) to spin up a UniFi controller. Demo mode provides simulated devices (1 UGW, 20 USW). CI runs a matrix across controller versions v6.5 through latest. The `TF_ACC_LOCAL=1` flag enables tests that only work against a locally-configured controller (e.g., tests requiring firewall zones with adopted hardware).

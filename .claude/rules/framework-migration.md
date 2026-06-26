---
paths:
  - "internal/provider/**/*.go"
---

# Framework vs SDKv2 (mux) — adding & migrating resources

This provider is muxed (`main.go`): Framework provider `provider.NewV2` (`provider_v2.go`) +
legacy SDKv2 provider `provider.New` (`provider.go`, upgraded via `tf5to6server`).

## Always use the Plugin Framework for new work
- Register new resources in `provider_v2.go` → `Resources()`; new data sources → `DataSources()`.
- NEVER add new resources to `provider.go` (`ResourcesMap`/`DataSourcesMap`). SDKv2 entries there are legacy.
- A resource name must exist in exactly ONE provider. Adding it to both makes the mux server fail.

## How a Framework resource is built (the GenericResource pattern)
- Embed the generic base: `type fooResource struct { *base.GenericResource[*fooModel] }`.
- Constructor `NewFooResource() resource.Resource` calls `base.NewGenericResource(typeName, modelFactory, base.ResourceFunctions{Read,Create,Update,Delete})`,
  each wired to a `client.*` call from go-unifi. CRUD, Configure, Metadata, ImportState are inherited.
- The model embeds `base.Model` (id, site) and implements `base.ResourceModel`: `AsUnifiModel()` (model→API)
  and `Merge()` (API→state). See `internal/provider/dns/` for the canonical example.
- Settings are singletons: use `settings.NewSettingResource(typeName, modelFactory, getter, updater)`
  (no Delete). See `internal/provider/settings/base_setting_resource.go`.
- Only write custom `Create/Read/Update/Delete` when `GenericResource` doesn't fit (e.g. `portal` file upload).

## Gating
- Min controller version: `RequireMinVersion("9.0.0")` in `ModifyPlan` (e.g. `base.ControllerVersionDnsRecords`).
- Feature flags: `RequireFeaturesEnabled(ctx, site, features.X)`.
- Import IDs use `site:id` format — Framework via `base.ImportIDWithSite`, SDKv2 via `base.ImportSiteAndID`.

## Migrating an SDKv2 resource
Reimplement it as a Framework resource, register it in `provider_v2.go`, and REMOVE the old entry
from `provider.go` in the same change so it isn't served by both providers.

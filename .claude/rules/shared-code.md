---
paths:
  - "internal/provider/**/*.go"
---

# Shared code — reuse, don't duplicate

Anything reusable across ≥2 domains lives in a shared package, not copied into `dns/`, `firewall/`,
etc. Before writing a helper, grep the shared packages below. If a near-match exists, extend it; if a
domain package already has logic a second domain now needs, **lift it** into the right shared package
(with a unit test) rather than copy it. Dependency direction is one-way: domain packages import
`base`/`utils`/`types`/`validators`; those never import domain packages.

## Where things belong
- `internal/provider/base/` — Framework resource **infrastructure**: `GenericResource[T]`,
  `NewGenericResource`, `ResourceFunctions`, `base.Model`, `ResourceModel`/`DatasourceModel` interfaces,
  `Client`, `ConfigureResource`/`ConfigureDatasource`, importers (`ImportIDWithSite`,
  `ImportSiteAndID`), and the version/feature validators (`ControllerVersionValidator`,
  `FeatureValidator`, `ControllerVersion*` version vars). CRUD plumbing and gating go here.
- `internal/provider/validators/` — reusable schema/config validators. See `resource-validation.md`.
- `internal/provider/types/` — reusable **schema attributes** and value helpers: `ID()`,
  `SiteAttribute()`; null/empty ctors `StringOrNull`/`Int32OrNull`/`Int64OrNull`; list helpers
  (`ListElementsAs`, `ListElementsToString`, `StringToListElements`, `EmptyList`/`DefaultEmptyList`);
  object helpers (`ObjectNull`, `ObjectValueMust`); predicates (`IsDefined`, `IsEmptyString`,
  `ShouldBeRemoved`); and the semantic-equality `MACType`.
- `internal/provider/utils/` — stack-agnostic helpers: env getters (`GetAnyStringEnv/BoolEnv/IntEnv`),
  MAC (`CleanMAC`), CIDR/IP (`CidrZeroBased`, `CidrOneBased`, `IsIPv4`/`IsIPv6`), strings
  (`JoinNonEmpty`, `SplitAndTrim`, `RemoveElements`, `IsStringValueNotEmpty`), markdown doc builders
  (`MarkdownValueListString/Int`), server-error checks (`IsServerErrorStatusCode`,
  `IsServerErrorContains`), model errors (`ErrorInvalidModelMergeTarget`), and
  `ReReadOnUpdateNotFound` for eventually-consistent updates.

Note: `utils/` and `types/` still carry some SDKv2-only helpers (e.g. `*DiffSuppressFunc`,
`SetToStringSlice`, `CidrValidate`). New Framework code should use the Framework-typed helpers.

## Import aliasing
`internal/provider/types` collides with `terraform-plugin-framework/types`. Convention: alias our
package as `ut` (`ut "…/internal/provider/types"`) and keep `types` for the framework package
(see `dns/resource_dns_record.go`, `firewall/resource_firewall_zone.go`). Follow the file you're in.

## Maintaining shared code
- Add a helper here the moment a second caller appears — don't wait for a third.
- Every shared helper/validator gets a table-driven unit test in the same package (`*_test.go`).
- Keep helpers narrow and Framework-typed; don't add a domain-specific special case to a generic helper —
  compose at the call site instead.

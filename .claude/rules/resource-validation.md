---
paths:
  - "internal/provider/**/*.go"
---

# Resource validation (Framework)

Validate config, not the world. Prefer letting the UniFi API reject truly server-side concerns;
add client-side validation only when it gives the user a **faster, clearer** error at plan time
(bad enum/format, mutually-exclusive attrs, version/feature gating). Don't reimplement API business
rules you can't keep in sync. Every validator MUST have a unit test (see "Verify" below).

## Where validation goes
- **Single attribute** → `Validators: []validator.String{…}` (or `Int32`, `Set`, …) on the schema
  attribute. Compose framework validators (`stringvalidator.OneOf`, `int32validator.Between`) with our
  reusable ones. Example: `dns/resource_dns_record.go` (`OneOf(...)`, `int32validator.Between(1,65535)`).
- **Across attributes** (mutually exclusive, required-together, conditional) → implement
  `ConfigValidators(ctx) []resource.ConfigValidator`, assert `_ resource.ResourceWithConfigValidators`,
  and return validators from `internal/provider/validators/`. Canonical: `dns/resource_dns_record.go`
  uses `validators.RequiredNoneIf(path.MatchRoot("type"), types.StringValue("A"), path.MatchRoot("port"), …)`.
- **Never** hand-roll a validator that already exists — reuse from `validators/` (see `shared-code.md`).

## Reusable validators (`internal/provider/validators/`)
- Strings: `CIDR()`/`CIDROrEmpty()`, `IPv4()`/`IPv6()`, `URL()`/`HTTPSUrl()`, `Hostname()`,
  `Timezone()`, `CountryCodeAlpha2()`, `StringLengthExactly(n)`. Sets: `UniqueMACs()`.
- Regex singletons (vars, not funcs): `validators.Mac`, `TimeFormat`, `DateFormat`, `HexColor`,
  `Email`, `PortRangeV2` (+ their `*Regex` sources).
- Conditional `ConfigValidator`s: `RequiredNoneIf`/`RequiredNoneIfSet`,
  `RequiredTogetherIf`/`RequiredSimpleTogetherIf`/`RequiredTogetherIfSet`, `RequiredValueIf`,
  and `ResourceIf`/`ResourceIfSet` (gate a nested validator on a condition path).
  Signature shape: `Fn(conditionPath path.Expression, conditionValue attr.Value, targets…)`.

## Version gating — `ControllerVersionValidator`
Embedded in `base.GenericResource[T]`, auto-wired in `base.ConfigureResource`
(`NewControllerVersionValidator(client)`) — so `r.RequireMinVersion(...)` is available on any
Framework resource. Data sources embed `base.ControllerVersionValidator` directly and get it wired by
`base.ConfigureDatasource` (see `dns/datasource_dns_record.go`). Call it in `ModifyPlan` and append to
`resp.Diagnostics`; assert `_ resource.ResourceWithModifyPlan`.

- Whole resource: `resp.Diagnostics.Append(r.RequireMinVersion("9.0.0")...)`.
- Also: `RequireMaxVersion(max)`, `RequireVersionBetween(min, max)`.
- **Per attribute**: `RequireMinVersionForPath("7.0", path.Root("geo_ip_filtering"), req.Config)` (also
  `RequireMaxVersionForPath`, `RequireVersionBetweenForPath`). The `…ForPath` variants no-op when the
  attribute is unset, so they only fire when the user actually configures the gated field. Heavy users:
  `settings/resource_setting_usg.go`, `resource_setting_ips.go`.
- Pin versions as `base.ControllerVersion*` vars where a constant already exists; otherwise pass a
  literal string. Do NOT compare `client.Version` by hand.

## Feature gating — `FeatureValidator`
Also embedded/auto-wired. Needs a resolved site:
`site, diags := r.GetClient().ResolveSiteFromConfig(ctx, req.Config)` then
`resp.Diagnostics.Append(r.RequireFeaturesEnabled(ctx, site, features.ZoneBasedFirewall, …)...)`
(`firewall/resource_firewall_zone.go`). Use `features.*` constants from go-unifi.
`…EnabledForPath` gates a single attribute. Results are cached per site.

## Verify
- **Unit test the validator** (required): table-driven `_test.go` in `validators/`, call
  `v.ValidateString(ctx, validator.StringRequest{ConfigValue: …}, &resp)` and assert on
  `resp.Diagnostics` — cover `null`/`unknown` (must NOT error), empty, valid, invalid.
  Template: `validators/cidr_test.go`.
- **Acceptance-test the wiring** (that a rule fires end-to-end): a step with
  `ExpectError: regexp.MustCompile("…")`, or `pt.MissingArgumentErrorRegex("name")` for required attrs
  (see `acceptance-testing.md`). Gate version-specific expectations with `MinVersion`/`VersionConstraint`.

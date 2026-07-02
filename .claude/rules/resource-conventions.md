---
paths:
  - "internal/provider/**/*.go"
---

# Resource & schema conventions

## Layout & naming (within a domain package, e.g. dns/, firewall/)
- `resource_<name>.go`, `datasource_<name>.go`, `<name>_model.go`, `<name>_test.go` lives in `acctest/`.
- Resource/model types are unexported (`dnsRecordResource`, `dnsRecordModel`); constructors are exported (`NewDnsRecordResource`).
- Assert interfaces at package scope: `var _ resource.ResourceWithImportState = &fooResource{}`.

## Schemas
- Use `MarkdownDescription` on every attribute (rendered into docs). Resource/attribute descriptions are
  written inline; only provider-level descriptions are exported consts (see `provider.go`).
- Reuse `types.ID()` and `types.SiteAttribute()` for the standard id/site attributes.
- Validation (attribute vs `ConfigValidators`, version/feature gating, verifying it): `.claude/rules/resource-validation.md`.

## Models
- Embed `base.Model`; fields are Framework types (`types.String`, `types.Int32`, …) with `tfsdk:"..."` tags.
- Implement `AsUnifiModel()` (→ go-unifi struct) and `Merge()` (← go-unifi struct, populate state).
- Null/empty conversion: use `types.StringOrNull`/`Int32OrNull`; list helpers in `types/lists.go`.

## Helpers & diagnostics
- Reuse shared code from `base`/`utils`/`types`/`validators` — do NOT duplicate. What lives where and
  how to lift/extend it: `.claude/rules/shared-code.md`.
- Framework errors: `resp.Diagnostics.AddError(...)` / `AddAttributeError(path, summary, detail)`.
  SDKv2 (legacy only): `diag.FromErr(err)`.

After changing any schema or description, regenerate docs (`.claude/rules/docs-generation.md`).

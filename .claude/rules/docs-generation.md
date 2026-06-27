---
paths:
  - "examples/**"
  - "templates/**"
  - "docs/**"
---

# Documentation generation (tfplugindocs)

`docs/` is GENERATED — never hand-edit it. Regenerate with `go generate ./tools/`
(directive in `tools/tools.go`: tfplugindocs reads the provider schema + examples + templates).

## Pipeline
- Schema `MarkdownDescription`s (in `internal/provider/**`) + `examples/**` + `templates/*.tmpl` → `docs/**`.
- `templates/index.md.tmpl` renders the provider landing page; `templates/guides/*.tmpl` render guides.

## Example/file conventions a resource needs for good docs
- for every resource there must be an examples included with this as part of documentation
- `examples/resources/<unifi_name>/resource.tf` — minimal working example.
- `examples/resources/<unifi_name>/import.sh` — import command, for importable resources.
- `examples/data-sources/<unifi_name>/data-source.tf` — for data sources.
- Provider auth examples: `examples/provider/provider_api_key.tf`, `examples/provider/provider_user_pass.tf`.

## When to regenerate (and commit the result)
After adding/changing any resource/data-source schema, attribute description, example, or template.
CI does NOT check docs freshness — it's the author's responsibility to run `go generate ./tools/` and commit.

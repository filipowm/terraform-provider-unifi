# UniFi Terraform Provider

Terraform provider for Ubiquiti UniFi network controllers (versions 6.x–9.x), built on the
[`filipowm/go-unifi`](https://github.com/filipowm/go-unifi) SDK. Fork of `paultyng/terraform-provider-unifi`.

Stack: Go 1.23.5 · terraform-plugin-framework v1.14.1 · terraform-plugin-sdk/v2 v2.36.1 ·
terraform-plugin-mux v0.18.0 · terraform-plugin-testing v1.12.0 · go-unifi v1.8.0.

## Critical rules

- **IMPORTANT: This is a *muxed* provider, mid-migration to the Terraform Plugin Framework.**
  New resources and data sources MUST use the Plugin Framework and be registered in
  `internal/provider/provider_v2.go`. Do NOT add new ones to the legacy SDKv2 provider
  (`internal/provider/provider.go`). A given resource must live in exactly one of the two — never both
  (the mux server will conflict). Details: `.claude/rules/framework-migration.md`.
- **IMPORTANT: Regenerate docs after any schema/description change** with `go generate ./tools/`.
  Everything in `docs/` is generated — never hand-edit it. Details: `.claude/rules/docs-generation.md`.
- **Commits MUST follow Conventional Commits** (`feat:`, `fix:`, `docs:`, `chore:`, `refactor:`,
  `build(deps):`). Release notes and labels are derived from them.
- use skills from `terraform-provider-development` plugin when making any changes, according to the changes being implemented

## Architecture

`main.go` builds the mux server: the Framework provider (`provider.NewV2`) and the SDKv2 provider
(`provider.New`, upgraded v5→v6 via `tf5to6server`) are combined with `tf6muxserver`.

```
internal/provider/
├── provider_v2.go     # Framework provider — register NEW resources/data sources HERE
├── provider.go        # Legacy SDKv2 provider — existing resources only
├── base/              # Shared Framework infra: GenericResource[T], Client, importer, version/feature gating
├── <domain>/          # dns, firewall, network, radius, routing, device, site, user, apgroup, portal
│   ├── resource_*.go      # resource implementation
│   ├── datasource_*.go    # data source implementation
│   └── *_model.go         # Framework model (tfsdk tags) + AsUnifiModel/Merge
├── settings/          # 16 singleton "setting" resources via shared NewSettingResource()
├── validators/        # reusable Framework validators (CIDR, HTTPSUrl, Timezone, RequiredNoneIf, …)
├── types/             # reusable schema attributes (ID(), SiteAttribute()) + value helpers
├── utils/             # env getters, MAC/CIDR/string helpers, server-error checks and any other shared utilities
├── acctest/           # acceptance tests + the AcceptanceTest() harness
└── testing/           # Dockerized controller env (testcontainers) + test helpers (imported as `pt`)
```

Coding conventions: `.claude/rules/resource-conventions.md`. Tests: `.claude/rules/acceptance-testing.md`.

## Commands

```bash
make build                 # go install
golangci-lint run --fix    # lint (matches CI; no make target exists)
make testacc               # TF_ACC=1 go test ./... — spins up a Dockerized controller, ~20m
# Run a single acceptance test:
make testacc TEST=./internal/provider/acctest TESTARGS='-run TestAccDNSRecord_basic'
go generate ./tools/       # regenerate docs/ from schema + examples/ + templates/
```

Acceptance tests require Docker (a UniFi controller is started automatically). Without `TF_ACC=1`
they are skipped.

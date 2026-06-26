---
paths:
  - "internal/provider/acctest/**"
---

# Acceptance testing

All acceptance tests live in `internal/provider/acctest/` and run against a real (Dockerized) controller.
They are gated by `TF_ACC=1` (set by `make testacc`) and skipped otherwise. There are no SDK sweepers.

## The harness (use it — don't call resource.Test directly)
- Write tests via `AcceptanceTest(t, AcceptanceTestCase{...})` (`acctest/provider_test.go`). It runs
  `resource.ParallelTest` with the muxed ProtoV6 provider factories.
- `AcceptanceTestCase` fields: `Steps`, `CheckDestroy`, `PreCheck`, `MinVersion`, `VersionConstraint`,
  `Lock` (`*sync.Mutex` to serialize singleton/setting tests).
- `TestMain` (in `provider_test.go`) starts the controller and sets the global `testClient` used by
  PreCheck/CheckDestroy for direct API calls.

## Helpers (testing package, imported as `pt`)
- Import steps: `pt.ImportStep(name, ignoreFields...)`, `pt.ImportStepWithSite(name, ...)`.
- Test data: `pt.GetTestVLAN(t)`, `pt.AllocateTestMac(t)`, `pt.RandAlpha/RandHostname/RandIpAddress`.
- Plan checks: `pt.CheckResourceActions(addr, plancheck.ResourceActionCreate)`; destroy: `pt.CheckDestroy(...)`.
- Error matching: `ExpectError: pt.MissingArgumentErrorRegex("name")`.

## Conventions
- Build HCL config inline with `fmt.Sprintf`; concatenate multi-resource configs with `+` (or `pt.ComposeConfig`).
  There are no `testdata/*.tf` fixtures (binary fixtures live in `acctest/files/`).
- Checks via `resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(...), …)`.
- Gate version-specific tests with `MinVersion`/`VersionConstraint`; lock singleton/setting tests.

## Running
```bash
make testacc                                                            # all
make testacc TEST=./internal/provider/acctest TESTARGS='-run TestAccFoo'  # one
```

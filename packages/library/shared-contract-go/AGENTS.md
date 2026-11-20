# Shared Contract Go

Provide the contracts that the validate, audit and eval commands share: findings, schemas, source catalogs and the file-system port.

- Keep the core on the standard library. Only `pkg/jsonschema/tekuri` imports a module outside it, and only `pkg/fs/osfs` reaches the host disk, environment, process or network. `internal/arch` enforces both, and it fails a package that is missing from its dependency table.
- Reach an adapter only through `pkg/wiring`. A core package depends on the port, `pkg/fs` or `pkg/jsonschema`, and never on `memfs`, `osfs` or `tekuri`.
- Pass the instant to a function. No package calls `time.Now`.
- Write every schema as JSON Schema draft 2020-12 under `pkg/schemas` with an `$id` in `schemas.Namespace`, name its id as a constant in `schemas.go`, and add valid and invalid documents for it to `pkg/wiring/contract_test.go`. A schema uses neither `propertyNames` nor `contentSchema`, because the validator reports the pointer of those errors wrongly; a test enforces this.
- Treat a finding as a result and an error as a fault. A broken document yields findings and a nil error. An unreadable file or an unknown schema returns an error, and `findings.ExitCode` maps any error to exit code 2.
- Unit tests use `memfs` and test doubles. Only `go-test-integration` touches a real disk, through the `integration` build tag, and it stays out of `ci-check`.

# Agent Audit Go

Audit the sources that skills and documents draw on, keeping each concern in its own package behind a port.

- Keep the axes independent: the catalog is read by `shared-contract-go`, a transport reads one kind of address, the rules judge an observation, the audit composes them, `compare` reads a base ref, and `render` formats the result. A package imports only what `internal/arch` lists for it, and the test fails a package missing from that table.
- Add a transport as a package under `internal/transport` plus one `Register` call in `internal/transport/plugins`, the only file that names a concrete transport. A transport reports what it observed and never decides whether a source has drifted; `internal/rules` decides.
- Reach the machine only through a port: `internal/process` for programs, `internal/web` for HTTP and `shared-contract-go` `fs` for files. Only `internal/process/osexec`, `internal/web/nethttp` and `cmd/agent-audit` import the host, and only `cmd/agent-audit` reads the clock. Unit tests use `memfs`, `processtest` and `webtest`.
- Treat a finding as a result and an error as a fault. A broken catalog yields findings; an unreadable file, an unknown base ref or a failed write yields exit status 2, so a partial run never reads as a pass.
- Convert a `SOURCES.md` with `agent-audit convert`: it keeps every word as a typed member, remark or note, validates the result against the schema, and refuses to write when it finds a problem. Add a member to the schema in `shared-contract-go` before the converter emits it.
- `ci-check` runs the fast gates. `ci-integration` builds the binary and runs the `integration` tests against a real git repository, an `httptest` server and every `sources.json` of the plugins; it stays out of CI.

---
type: plan
has_subplans: false
parent_plan: plans/agent-sandbox-provisioning-axes.md
parallel_group: 2
status: complete
dependencies:
  plans: [cli-isolator-provisioner-core]
  files:
    - packages/agent/agent-cli-go/internal/sandbox/compose/
    - packages/shared/shared-agent-go/pkg/types/sandbox.go
    - packages/agent/agent-cli-go/internal/sandbox/registry.go
    - packages/agent/agent-cli-go/internal/cmd/run.go
    - packages/shared/shared-agent-go/pkg/sandbox/defaults.go
skills_to_consult: [general-fp-guide, git-guide]
validation:
  type_check: pass        # agent-cli-go + shared-agent-go go-build
  lint: pass              # agent-cli-go go-lint (0 issues)
  build: pass             # agent-cli-go + shared-agent-go go-build
  tests: pass             # go-test green; alias test asserts --sandbox compose is rejected
  integration: n/a        # deletion only
---

# 05 — Remove the compose isolator

## Objective

Delete the `compose` sandbox tier and all its config/flags/defaults — it's unused and out of the new matrix. Per the project rule, remove cleanly (no shim). Isolation axis becomes `{none, bwrap, docker}`.

> Scope note: this is a compose-only removal. The isolation axis stays `{none, bwrap, docker}`; the NEW orthogonal `Network NetworkMethod` axis (`{host, none, proxy}`) + `EgressAllowlist []string` land in subplans 01/03 — nothing compose-related. Do NOT touch those new fields here.

## Tasks

1. **Delete `internal/sandbox/compose/`** entirely (`compose.go`, `findComposeFile`, the `stacks/ai-agent.yaml`/`ai-agent` defaults).
2. **Drop the config fields**: remove `SandboxConfig.{ComposeFile, Service}` from `shared-agent-go/pkg/types/sandbox.go:26-47`, and the legacy `SandboxCompose` `SandboxMethod` value (`:6-13`) if the `--sandbox` alias map (subplan 03) no longer references it. — Leave the new `Network NetworkMethod` (subplan 01 changes its type from the old `Network bool`) + `EgressAllowlist []string` fields untouched; this task removes ONLY the compose-specific `{ComposeFile, Service}`.
3. **Remove registry wiring**: the `compose` branches in `registry.go` `GetExecutor`/`GetAvailableMethods` and any `compose` entry in `tierIsolation`/the curated matrix.
4. **Remove flags/defaults**: the `--sandbox compose` option + any `--compose-file`/`--service` flags in `cmd/run.go`; remove compose defaults from `shared-agent-go/pkg/sandbox/defaults.go` if present.
5. **Update the `--sandbox` alias map** (from subplan 03) so it no longer maps to compose; document the removal in the CLI help / release notes.
6. **Verify clean removal**: `grep -ri compose packages/agent/agent-cli-go packages/shared/shared-agent-go` returns nothing meaningful; build/lint/test green.

## Validation Steps

```bash
grep -rin 'compose\|ComposeFile\|SandboxCompose' packages/agent/agent-cli-go packages/shared/shared-agent-go || echo "clean"
npx moon run agent-cli-go:go-build agent-cli-go:go-test agent-cli-go:go-lint shared-agent-go:go-build shared-agent-go:go-test
```

## Success Criteria

- [x] `internal/sandbox/compose/` deleted; `SandboxConfig.{ComposeFile,Service}` + the `SandboxCompose` const gone.
- [x] No `compose` branches remain in registry/flags/defaults; isolation axis = `{none, bwrap, docker}`. (`registry.go` + the `--compose-file`/`--service` flags were already gone after 03; `defaults.go` had no compose defaults.)
- [x] `grep` confirms no residual compose implementation; build/lint/test green (`agent-cli-go` + `shared-agent-go`).

## Completion Notes

- **Mostly already excised by 03.** The 03 registry/run.go rewrite already deleted `registry.go` (and its `GetExecutor`/`tierIsolation` compose branches) and dropped the `--compose-file`/`--service` flags, so 05 only had to delete the `compose` package, the two config fields, and the `SandboxCompose` const.
- **Clean rejection, no shim.** The `--sandbox compose` alias case was removed entirely; the legacy alias now falls to the generic `default` error (`unknown --sandbox "compose"; use --isolation/--provision/--network`). The one remaining `compose` mention is the alias unit test asserting that rejection — regression protection against compose silently reappearing, not a residual implementation.
- **Isolation axis is now `{none, bwrap, docker}`** end to end. The legacy `SandboxMethod` enum (none/bwrap/docker/nix/nixflake) and `SandboxConfig.Method` remain only as the deprecated-`--sandbox` alias surface.

## Files Modified/Created

- `internal/sandbox/compose/` — deleted.
- `shared-agent-go/pkg/types/sandbox.go` — drop compose fields/value.
- `internal/sandbox/registry.go`, `internal/cmd/run.go`, `shared-agent-go/pkg/sandbox/defaults.go` — remove compose wiring/defaults.

## Dependencies

`03-cli-isolator-provisioner-core` (the new axis + alias map exist, so removing compose doesn't strand the selector).

## Estimated Duration

~0.5 day.

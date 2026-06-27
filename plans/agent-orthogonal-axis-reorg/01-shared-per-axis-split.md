---
type: plan
has_subplans: false
parent_plan: plans/agent-orthogonal-axis-reorg.md
parallel_group: 1
status: pending
dependencies:
  plans: []
  files:
    - packages/shared/shared-agent-go/pkg/types/sandbox.go
    - packages/shared/shared-agent-go/pkg/sandbox/policy.go
    - packages/shared/shared-agent-go/pkg/sandbox/defaults.go
    - packages/shared/shared-agent-go/pkg/nix/source.go
    - packages/shared/shared-agent-go/pkg/nix/nix.go
    - packages/shared/shared-agent-go/pkg/types/terminal.go
    - packages/shared/shared-agent-go/pkg/config/loader.go
    - packages/shared/shared-agent-go/pkg/worktree/naming.go
    - packages/shared/shared-agent-go/go.mod
skills_to_consult:
  - orthogonal-pattern-guide
  - connascence-guide
  - general-fp-guide
  - moon-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Shared-agent-go Per-Axis Split + Delete SandboxConfig

## Objective

Split the overloaded `shared-agent-go/pkg/{types,sandbox,nix}` packages into per-axis
packages (`pkg/isolation`, `pkg/network`, `pkg/provision`, `pkg/policy`, `pkg/workspace`),
delete the flat `SandboxConfig` god-struct, and reduce the shared module's surface to the
one neutral cross-module seam (`Contribution` + per-axis enums + `Capabilities`/`EnforcePolicy`).
This is the foundational gate for the whole reorg; it also evicts CLI-only code and sheds the
go-toml/yaml dependencies that only the moved config loader used.

## Tasks

1. **Split `pkg/types/sandbox.go` enums into per-axis packages.** Create
   `pkg/isolation/types.go` (`IsolationMethod` + `DefaultContainerImage` + `UserConfigPaths`),
   `pkg/network/types.go` (`NetworkMethod` + `DefaultEgressAllowlist` + `EgressIsRestricted`),
   `pkg/provision/types.go` (`ProvisionMethod` + the `Contribution` seam), and
   `pkg/workspace/vcs.go` (`VCSType`, shared by both the CLI and the operator). Keep each enum as
   a typed string with module-level helpers; no shared mutable state.
   ```go
   // pkg/provision/types.go
   package provision

   type ProvisionMethod string

   // Contribution is the single neutral data-coupling handoff across the
   // shared module boundary. It carries no method-specific fields.
   type Contribution struct {
       RoBindPaths  []string
       PathEntries  []string
       Env          map[string]string
       InitCommands []string
   }
   ```

2. **Rename the policy engine `pkg/sandbox` -> `pkg/policy`.** Move `pkg/sandbox/policy.go` to
   `pkg/policy/policy.go` (`Capabilities`, `EnforcePolicy`, `SandboxPolicy`) and relocate the
   `pkg/sandbox/defaults.go` constants into the axis packages they belong to (container image ->
   `isolation`, egress allowlist -> `network`). Removing the overloaded `sandbox` package name from
   shared resolves the CCP/CRP smell. `EnforcePolicy` keeps switching on no concrete variant.
   ```go
   // pkg/policy/policy.go
   package policy

   type Capabilities struct {
       Pinned, HostToolsUnreachable, EgressRestricted, KernelIsolated bool
   }

   func EnforcePolicy(c Capabilities, p SandboxPolicy) error { /* unchanged body */ }
   ```

3. **Delete the flat `SandboxConfig` god-struct from `pkg/types/sandbox.go`.** Remove the file (and
   the now-empty `pkg/types` package directory if nothing else remains). Shared keeps ONLY the
   neutral seam: per-axis enums + `Contribution` + `Capabilities`. The per-type `Options` structs
   that previously read from `SandboxConfig` are created CLI-side in subplan 02 — no shim, no
   stamp-coupled god-struct survives in shared (connascence: collapse stamp coupling to data coupling).

4. **Nest nix under provision.** Move `pkg/nix/source.go` + `pkg/nix/nix.go` to
   `pkg/provision/nix/source.go` + `pkg/provision/nix/nix.go` (package `nix`); pins, `PackageSets`,
   and `ComputeEnvID` are provision-axis concerns. Update the package's own internal references; the
   CLI/operator import-path fixes land in tasks 6 below and in subplans 02/06.
   ```go
   // pkg/provision/nix/nix.go
   package nix
   // import path: .../shared-agent-go/pkg/provision/nix
   ```

5. **Evict CLI-only modules from shared.** Delete `pkg/types/terminal.go`, `pkg/config/loader.go`,
   and `pkg/worktree/naming.go` from shared-agent-go. These have no operator consumer; their CLI
   homes (`internal/terminal`, `internal/config`, `internal/workspace`) are created in subplans 03/04.
   This is what allows the go.mod trim in task 6.

6. **go.mod hygiene + consumer import-path fixes.** In `shared-agent-go/go.mod` drop
   `github.com/pelletier/go-toml/v2` and `gopkg.in/yaml.v3` (only the evicted config loader used them);
   run `go mod tidy` and remove any dangling `replace`. Then update BOTH consumers' shared-IMPORT
   paths to the new packages so the tree still compiles: rewrite `pkg/types`/`pkg/sandbox`/`pkg/nix`
   imports to `pkg/isolation`/`pkg/network`/`pkg/provision`/`pkg/policy`/`pkg/workspace`/`pkg/provision/nix`
   in `agent-cli-go` and `agent-operator-go`. The CLI's switch from `SandboxConfig` to per-type
   `Options` is owned by subplan 02; this subplan only retargets import paths.

## Validation Steps

- `npx moon run shared-agent-go:typecheck` and `:lint` — shared module is self-consistent.
- `cd packages/shared/shared-agent-go && go build ./... && go test ./...` — shared builds/tests green in isolation.
- `cd packages/shared/shared-agent-go && go mod tidy && git diff --exit-code go.mod go.sum` — confirms go-toml/yaml are gone and nothing re-adds them.
- **Combined Phase-1 gate (with subplan 02):** `npx moon run agent-cli-go:build` / `:test` only go green once 02 lands the per-type `Options`; the `SandboxConfig` deletion build-green is asserted at the combined 01+02 landing, not by 01 alone.
- `cd packages/agent/agent-operator-go && go build ./...` — operator compiles against retargeted shared import paths (its axis relocation is Phase 2 / subplan 06).
- No integration help-text tests change here (no flags touched); flag-grammar integration tests live in subplan 04.

## Success Criteria

- [ ] `pkg/isolation`, `pkg/network`, `pkg/provision`, `pkg/policy`, `pkg/workspace` exist with the enums/seam/helpers above.
- [ ] `pkg/provision/nix` holds the moved nix source; old `pkg/nix` is gone.
- [ ] `SandboxConfig` is deleted; shared exposes only `Contribution` + per-axis enums + `Capabilities`/`EnforcePolicy`.
- [ ] `pkg/types/terminal.go`, `pkg/config/loader.go`, `pkg/worktree/naming.go` removed from shared.
- [ ] `go-toml/v2` and `yaml.v3` removed from `go.mod`; `go mod tidy` clean.
- [ ] Both consumers' shared-import paths retargeted; shared module builds/tests green standalone; operator compiles; CLI build-green deferred to the combined 01+02 landing.

## Files Modified/Created

- Created: `pkg/isolation/types.go`, `pkg/network/types.go`, `pkg/provision/types.go`, `pkg/workspace/vcs.go`, `pkg/policy/policy.go`, `pkg/provision/nix/source.go`, `pkg/provision/nix/nix.go`.
- Deleted: `pkg/types/sandbox.go`, `pkg/types/terminal.go`, `pkg/sandbox/policy.go`, `pkg/sandbox/defaults.go`, `pkg/nix/source.go`, `pkg/nix/nix.go`, `pkg/config/loader.go`, `pkg/worktree/naming.go` (and now-empty `pkg/types`, `pkg/sandbox`, `pkg/nix`, `pkg/config`, `pkg/worktree` dirs).
- Modified: `packages/shared/shared-agent-go/go.mod`, `go.sum`; shared-import statements across `agent-cli-go` and `agent-operator-go`.

## Dependencies

- Prerequisite subplans: none (this is the Group 1 foundational gate).
- **Co-landing requirement:** deleting `SandboxConfig` forces the CLI plugins onto per-type `Options`, which is implemented in subplan **02-cli-isolation-provision-network**. This subplan retargets both consumers' shared-import paths so the tree compiles, but the CLI build-green gate for the `SandboxConfig` deletion is the combined Phase-1 landing (01 + 02). Land them together.

## Estimated Duration

~0.5–1 day. Mechanical package moves + import-path retargeting + go.mod trim; the only judgment is placing each evicted constant/enum in its correct axis package and verifying the operator still compiles against the new shared edge.

---
type: plan
status: complete
has_subplans: true
dependencies:
  plans: []
  subplans:
    01-shared-per-axis-split: []
    02-cli-isolation-provision-network: [01-shared-per-axis-split]
    03-cli-workspace-terminal: [01-shared-per-axis-split]
    04-cli-cmd-flags-wiring: [02-cli-isolation-provision-network, 03-cli-workspace-terminal]
    05-cli-fitness-tests: [02-cli-isolation-provision-network, 03-cli-workspace-terminal, 04-cli-cmd-flags-wiring]
    06-operator-axis-relocation: [01-shared-per-axis-split]
    07-operator-policy-symmetry: [06-operator-axis-relocation]
proposed_subplans:
  - 01-shared-per-axis-split
  - 02-cli-isolation-provision-network
  - 03-cli-workspace-terminal
  - 04-cli-cmd-flags-wiring
  - 05-cli-fitness-tests
  - 06-operator-axis-relocation
  - 07-operator-policy-symmetry
parallel_groups:
  - group: 1
    subplans: [01-shared-per-axis-split]
  - group: 2
    subplans: [02-cli-isolation-provision-network, 03-cli-workspace-terminal]
  - group: 3
    subplans: [04-cli-cmd-flags-wiring]
  - group: 4
    subplans: [05-cli-fitness-tests]
  - group: 5
    subplans: [06-operator-axis-relocation]
  - group: 6
    subplans: [07-operator-policy-symmetry]
skills_to_consult:
  - orthogonal-pattern-guide
  - hexagonal-pattern-guide
  - microkernel-pattern-guide
  - connascence-guide
  - general-fp-guide
  - moon-guide
  - kubernetes-guide
research_sources:
  documentation:
    - orthogonal-pattern-guide (package-by-axis, symmetric placement, shared core vs leaves)
    - hexagonal-pattern-guide (ports in axis shared/, single composition root names concretes)
    - microkernel-pattern-guide (registry/open-closed, capabilities/fail-closed, lazy factories, binding time, DI)
    - connascence-guide (grade the Contribution/resolveAxes/flag seams)
  versions:
    go: 1.26.0
    sigs.k8s.io/controller-runtime: v0.24.1
    k8s.io/api: v0.36.2
    github.com/spf13/cobra: v1.10.2
---

# Agent CLI/Operator Orthogonal-Axis Reorg

## Overview

Reorganize `agent-cli-go` and `agent-operator-go` so the package tree mirrors the system's orthogonal axes — `isolation`, `provision`, `network`, `workspace` — each a `shared/` core plus per-type leaves, with the two consumers symmetric where they share function and a single neutral seam (`Contribution`/enums/`Capabilities`) crossing the shared module boundary. This is a relocation + naming + policy-symmetry job, not a rewrite: the registry and the method-agnostic policy engine already exist.

## Goals

- Replace the overloaded `internal/sandbox` with per-axis dirs (`isolation/`, `provision/`, `network/`, `workspace/`), keeping `sandbox/` as the composition/selection layer only.
- Fix the structural smells: `nixprov`→`provision/nix`, dissolve `sandboxutil` into per-axis `shared/`, make `none`/`command` provisioners siblings of `nix`, localize cross-axis glue at bridge files.
- Delete the `SandboxConfig` god-struct; split shared-agent-go into per-axis packages; expose one `pkg/policy.EnforcePolicy` both consumers call.
- Make the operator symmetric at the axis-name level and route admission through the same shared policy engine.
- Adopt the `--<axis>-<type>-<option>` flag grammar; wire the latent kernel-isolation and egress-allowlist capabilities live.
- Enforce the boundaries in code with architecture/import fitness tests.
- Introduce **zero** new dependencies.

## Current State

- **Stack:** Go 1.26.0 across all three modules. `agent-cli-go` on cobra v1.10.2; `agent-operator-go` on controller-runtime v0.24.1 / k8s.io/api v0.36.2; `shared-agent-go` on go-toml v2.3.1 + yaml v3 (both only used by the CLI-only config loader); `shared-core-go` pure stdlib.
- **Bones already correct:** `internal/sandbox/registry.go` is a lazy-factory `Registry` (`map[Method]Factory`, no global state); `internal/sandbox/plugins/plugins.go` is the sole importer of concrete leaves; `shared-agent-go/pkg/sandbox/policy.go` `EnforcePolicy(Capabilities, SandboxPolicy)` names no concrete variant. Three axes are already flag-selected (`--isolation`/`--provision`/`--network`).
- **Smells:** `internal/sandbox/nixprov` (a provisioner misfiled as an isolation sibling); `internal/sandboxutil` (homeless axis utils); `sandbox` names a mechanism not the axis; `none`/`command` provisioners inline while `nix` is a package; a flat `SandboxConfig` god-struct handed whole to every plugin; the operator duplicates admission logic instead of calling the shared engine; several dead fields (`Runtime` hardcoded `""`, `EgressAllowlist` written-never-read).

## Research Findings

Recommended approach is a structural reorg with no new libraries (the full read-only map, prior design, and skill-principle extraction are in the conversation research). Key skill-grounded decisions:

- **hexagonal** — each axis *port* lives in its `shared/`; `plugins/plugins.go` stays the sole concrete importer; the policy gate reads `Capabilities`, switches on no variant.
- **microkernel** — keep the lazy factory registry (no eager instance maps); `none`/`command`/`nix` become symmetric sibling subpackages; the operator webhook is rewired to the *same* shared `policy.EnforcePolicy`.
- **connascence** — `Contribution` stays a neutral data-coupling handoff (reject any variant-specific field); `resolveAxes` returns a named `ResolvedAxes` struct (weaken positional connascence); operator `Network string` → typed `NetworkMode`.
- **orthogonal-pattern / cross-cutting** — `runtime` and `passthrough` are capability knobs folded into isolation, not axes; the policy gate + telemetry are woven once at the composition root.

Alternatives considered and rejected: a registry for the closed `network` enum (over-engineering); forced per-type leaves in the operator (fabricates non-existent variants); keeping a flat shared `types/` (leaves the god-struct and CLI-only dead code in shared); both-modules-in-one-lockstep (un-reviewable diff).

## Agreed Decisions (from plan-clarify)

1. Confine-axis dir = `isolation`; `internal/sandbox/` retained as the composition/selection layer (the composed whole).
2. `runtime` → fold into `isolation/docker` (CLI) / `isolation/shared` `runtimeClassName` (operator), **not an axis**; **wire** via docker-gated `--isolation-docker-runtime`.
3. `passthrough` → fold into `isolation/bwrap.Options` (knob); flag `--host-passthrough` → `--isolation-bwrap-passthrough`.
4. `network` → closed enum, **no registry**; `shared/` realizer + per-type leaves; per-isolator emitters become bridge files (`isolation/{bwrap,docker}/network.go`); document the asymmetry; wire `EgressAllowlist` via `--network-proxy-egress-allow`.
5. `workspace` → `internal/workspace/{shared,git}` now (git only, no fabricated jj); `VCSType` → shared `pkg/workspace`.
6. Flag grammar = `--<axis>-<type>-<option>` for per-type knobs; bare `--<axis>` selectors + bare `--require-*` policy flags; breaking renames, no shims.
7. Sequencing = Phase 1 (shared split + CLI, together) → Phase 2 (operator).
8. Operator symmetry = axis-name level only (one pod isolator; real leaves only where it varies); symmetry checked as axis-name-set intersection.
9. shared-agent-go = full per-axis split + delete `SandboxConfig` + rename engine to `pkg/policy` + evict CLI-only code + nest nix; shed go-toml/yaml.
10. Operator registries = relocate into per-axis dirs, keep package-level (explicit `NewRegistry`+DI deferred).
11. Per-axis shared piece = literal `shared/` subdir.
12. Scope = axes + in-axis wiring + reorg-caused go.mod hygiene; `scriptlib`, registry-pluggability, operator DI conversion deferred (see Out of Scope).

## Proposed Approach

**Phase 1 — shared boundary + CLI (land together, build/tests green):**

1. `shared-agent-go`: per-axis packages (`pkg/{isolation,network,provision,policy,workspace}`), delete `SandboxConfig`, rename engine → `pkg/policy`, evict CLI-only code (terminal / config-loader / worktree-naming) to CLI internal, nest nix under `pkg/provision/nix`, drop go-toml/yaml.
2. CLI isolation/provision/network: `internal/isolation/{shared,none,bwrap,docker}`, `internal/provision/{shared,none,command,nix}`, `internal/network/{shared,host,none,proxy}`; dissolve `sandboxutil` into per-axis `shared/` + isolation network bridge files; `internal/sandbox` keeps only registry/select/plugins.
3. CLI workspace + terminal: `internal/worktree` → `internal/workspace/{shared,git}`; `internal/wrapper` → `internal/terminal/{shared,none,tmux}`.
4. CLI cmd + flags: slim `run.go` into per-axis flag groups, return a named `ResolvedAxes` struct, wire the three new/renamed flags, delete dead `executor.go` + `buildDirect*`, update help-text integration tests.
5. CLI fitness tests: import-direction, no-sibling-reach, shared-purity, policy-gate-purity, composition-root-only-concretes, registry-factory.

**Phase 2 — operator symmetry:**

6. Operator relocation: `builder/*` → per-axis dirs (one `isolation/shared` pod realizer, `network/shared`, `provision/{shared,nix}`, `workspace/{shared,git,jj}`, `harness/{shared,claude,opencode}`, `provider/`); merge `BuildJob`+`BuildWorkspaceJob`; typed CRD enums.
7. Operator policy symmetry: rewire `webhook.enforcePolicy` → shared `pkg/policy.EnforcePolicy`; golden tests of pod/admission specs before/after; regenerate deepcopy + RBAC; cross-module symmetry test.

## Risk Assessment

- **Shared boundary is the gate** — any shared helper still typed on the deleted `SandboxConfig` must move to the `Contribution`/enum/`Capabilities` seam first, or the split blocks (mitigation: Phase 1 lands shared + CLI together).
- **Package-name clash** — CLI `internal/provision/nix` importing shared `pkg/provision/nix` needs an import alias at every call site (the very thing `nixprov` dodged).
- **Import cycles** — bridge files (`isolation/<type>/network.go`) depend on `network/shared` one-way only; `shared/` never reaches into a leaf (mitigation: the no-sibling-reach + import-direction fitness tests).
- **Breaking CLI surface** — flag renames break help-text integration tests with no shim (AGENTS.md); accepted appetite, tests updated.
- **Behavior-affecting operator refactor** — merging the Job builders and rewiring the webhook can change pod specs / admission verdicts (mitigation: golden tests before/after).
- **Codegen/RBAC drift** — `config/rbac/role.yaml` already out of sync; controller-gen hand-maintained — regenerate/validate after package moves.
- **Scope creep** — `scriptlib`, registry-pluggability kept out (see Out of Scope) to keep the diff reviewable/bisectable.

## Proposed Child Plans

Execution groups (▸ = sequential gate, ∥ = parallel):

- **Group 1 ▸** `01-shared-per-axis-split` — foundational; gates everything.
- **Group 2 ∥** `02-cli-isolation-provision-network`, `03-cli-workspace-terminal` (after 01; parallel, disjoint dirs).
- **Group 3 ▸** `04-cli-cmd-flags-wiring` (after 02, 03).
- **Group 4 ▸** `05-cli-fitness-tests` (after 02–04; closes Phase 1).
- **Group 5 ▸** `06-operator-axis-relocation` (Phase 2; after 01, follows Phase 1 per Decision 7).
- **Group 6 ▸** `07-operator-policy-symmetry` (after 06; closes Phase 2).

## Out of Scope (deferred follow-ups, tracked here)

These are real but orthogonal to the axis reorg; deferred to keep this change reviewable, and recorded so they are not lost:

- `shared-core-go/pkg/scriptlib` re-export barrel deletion (AGENTS.md "no re-exports") — its own change touching a third module + both consumers' build graphs.
- Provider / harness / agent registry pluggability (hardcoded registries) — a registry refactor orthogonal to layout.
- Operator registries → explicit `NewRegistry` + DI (Decision 10) — converges the last wiring asymmetry; optional polish.

## Success Criteria

- `internal/{isolation,provision,network,workspace}` exist on the CLI with `shared/` + per-type leaves; `internal/sandbox` holds only registry/select/plugins; `sandboxutil` and `nixprov` are gone.
- `shared-agent-go` exposes per-axis packages and `pkg/policy.EnforcePolicy`; `SandboxConfig` deleted; go-toml/yaml dropped.
- Operator has matching axis-name dirs and its webhook calls the shared `policy.EnforcePolicy`.
- The `--<axis>-<type>-<option>` flags work; `--isolation-docker-runtime` makes kernel isolation reachable; `--network-proxy-egress-allow` is honored.
- All fitness tests pass; `moon run :typecheck/:lint/:build/:test` green across all three modules; operator pod/admission goldens unchanged except the intended merge.

## Estimated Effort

Large — ~7 subplans across 2 phases. Phase 1 (shared + CLI) is the bulk (the shared split + CLI fan-out + flag rewire + fitness tests); Phase 2 (operator) is relocation + the webhook policy rewire + goldens. No new dependencies; effort is in mechanical relocation, import-alias discipline, and test/golden churn rather than new logic.

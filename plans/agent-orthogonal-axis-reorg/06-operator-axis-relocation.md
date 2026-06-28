---
type: plan
has_subplans: false
parent_plan: plans/agent-orthogonal-axis-reorg.md
parallel_group: 5
status: complete
dependencies:
  plans: [01-shared-per-axis-split.md]
  files:
    - packages/agent/agent-operator-go/internal/builder/
    - packages/agent/agent-operator-go/internal/isolation/
    - packages/agent/agent-operator-go/internal/network/
    - packages/agent/agent-operator-go/internal/provision/
    - packages/agent/agent-operator-go/internal/workspace/
    - packages/agent/agent-operator-go/internal/harness/
    - packages/agent/agent-operator-go/internal/provider/
    - packages/agent/agent-operator-go/api/v1alpha1/agentrun_types.go
skills_to_consult:
  - orthogonal-pattern-guide
  - hexagonal-pattern-guide
  - microkernel-pattern-guide
  - kubernetes-guide
  - general-fp-guide
  - moon-guide
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: n/a
---

## Status (complete — golden-safe)

`internal/builder/` fanned out into `internal/{isolation,network,provision,workspace,harness,provider}`
with axis-name symmetry to the CLI. All operator moon tasks green (lint 0 issues); the relocated unit
tests (the pod/admission goldens) pass, including both `BuildJob` paths — confirming the Job merge is
golden-equivalent.

### Acyclic design (resolves the plan's implicit cross-cutting cycle)

The operator has ONE pod realizer, so `isolation/shared` is the pod composition root: it builds all
Jobs/containers and weaves the security/hardening cross-cutting concern, importing the other axes as pure
data providers (`workspace/shared` scripts+strategies+PVCs, `provision/shared` toolchain image,
`harness/shared` command, `provider` env, `network/shared` policy). The data axes never import isolation —
no cycle. This is why container/Job builders live in `isolation/shared` even though the plan's literal file
list put some under `workspace`; placing them there would have created an `isolation ↔ workspace` cycle.

### Key results
- One pod isolator (`isolation/shared`); no fabricated `bwrap`/`docker`/`none` leaves; `runtimeClassName`
  the kernel knob.
- `BuildJob` + `BuildWorkspaceJob` merged into one `BuildJob(..., ws *WorkspaceBinding)` (nil → standalone,
  non-nil → workspace), preserving exact pod-spec output for both.
- `DefaultContainerSecurityContext`/`DefaultPodSecurityContext` are the single hardening source in
  `isolation/shared/security.go`; `applyPodHardening` consumes them.
- Real leaves only where the operator varies: `provision/nix`, `workspace/{git,jj}`,
  `harness/{claude,opencode}`. The three package-level registries (`toolchainFactories`, `harnessCommands`,
  `vcsStrategies`) live in their axis `shared/` dirs.
- `NetworkMode` typed enum replaces the untyped `Network string` field.

### Deviations
- **CRD enum typing**: only `Network` was an untyped string — typed to `NetworkMode`. `WorkspaceType`,
  `AgentType`, `ToolchainType` were already typed enums. No `+kubebuilder:validation:Enum` markers were
  added to those three (and no CRD-manifest enum constraints added for them) — controller-gen is broken on
  Go 1.25+ (operator AGENTS.md) and adding Go-only markers without the matching hand-edited CRD YAML would
  introduce Go↔CRD drift. The `NetworkMode` change is type-only (identical serialization + existing CRD
  enum), so deepcopy/CRD/RBAC need no regeneration — drift-free.
- **`resolver/` partially kept**: only `resolver/provider.go` moved (to `internal/provider`); the CRD-fetch
  resolvers (`ResolveHarness`/`ResolveToolchain`/`ResolveWorkspace`) and `ApplyHarnessDefaults` stay in
  `resolver/` per the plan's Removed list.
---

# Operator Builder -> Per-Axis Dirs (Symmetric)

## Objective

Relocate `agent-operator-go/internal/builder/*` into per-axis package dirs (`isolation/`, `network/`, `provision/`, `workspace/`, `harness/`, `provider/`), mirroring the CLI axis layout so the two consumers are symmetric at the axis-NAME level. Symmetry is name-level only: the operator has ONE pod isolator (`runtimeClassName` is the kernel-boundary knob) — no fabricated `bwrap`/`docker`/`none` leaves — and real leaves only where the operator genuinely varies (workspace git/jj, harness claude/opencode, provision nix). This is mechanical relocation plus consolidation of duplicate Job builders and security-defaults sources, gated golden-equivalent, with typed CRD enums replacing untyped strings.

## Tasks

1. **Create the isolation axis (one realizer, not per-type).** Move `internal/builder/security.go` + `internal/builder/hardening.go` + `internal/builder/container.go` and the isolation parts of `internal/builder/job.go` into `internal/isolation/shared/` (`job.go` single Job builder, `container.go`, `security.go`, `hardening.go`). `runtimeClassName` stays the kernel-boundary knob inside `shared/`; do NOT fabricate `bwrap`/`docker`/`none` leaves. Consolidate the three security-defaults sources of truth into one helper.

   ```go
   // internal/isolation/shared/security.go
   // defaultSecurityContext is the SOLE source of pod/container hardening defaults.
   func defaultSecurityContext(run *v1alpha1.AgentRun) *corev1.SecurityContext { /* runAsNonRoot, seccomp, drop ALL caps */ }
   ```

2. **Merge the duplicate Job builders into one.** Fold `internal/builder/job.go` `BuildJob` and `internal/builder/workspace.go` `BuildWorkspaceJob` into ONE builder in `internal/isolation/shared/job.go`, with the workspace diff passed in as an explicit input value (resolves the duplicate-Job-builder smell). Keep the function pure: state in, `*batchv1.Job` out.

   ```go
   // internal/isolation/shared/job.go
   func BuildJob(run *v1alpha1.AgentRun, ws workspace.Spec, tc provision.Contribution) *batchv1.Job
   ```

3. **Relocate the network and provision axes.** Move `internal/builder/networkpolicy.go` -> `internal/network/shared/networkpolicy.go`. Split `internal/builder/toolchain.go` into `internal/provision/shared/toolchain.go` (the `Toolchain` interface + `ResolveToolchain` registry lookup) and `internal/provision/nix/nix.go` (the `nixToolchain` concrete). Keep the neutral `Contribution` (data coupling) as the only handoff out of provision.

   ```go
   // internal/provision/shared/toolchain.go
   type Toolchain interface{ Contribute(run *v1alpha1.AgentRun) Contribution; Pinned() bool }
   func ResolveToolchain(t v1alpha1.ToolchainType) (Toolchain, error)
   ```

4. **Relocate the workspace axis (two real variants).** Move `internal/builder/workspace_vcs.go` + `workspace_git.go` + `workspace_jj.go` into `internal/workspace/{shared/vcs.go, git/, jj/}`. `shared/vcs.go` holds the `VCSStrategy` port + `GetVCSStrategy`; PVC/clone/worktree helpers stay under `workspace/shared`; `git/` and `jj/` hold only the per-VCS concretes.

   ```go
   // internal/workspace/shared/vcs.go
   type VCSStrategy interface{ CloneCommands(run *v1alpha1.AgentRun) []string; WorktreePath() string }
   func GetVCSStrategy(t v1alpha1.WorkspaceType) (VCSStrategy, error)
   ```

5. **Relocate the harness and provider axes.** Move `internal/builder/harness.go` + `harness_claude.go` + `harness_opencode.go` into `internal/harness/{shared,claude,opencode}/` (`shared/` holds the harness port + command registry lookup; leaves hold concretes). Merge the provider parts of `internal/builder/env.go` with `internal/resolver/provider.go` into `internal/provider/provider.go`.

   ```go
   // internal/harness/shared/harness.go
   type Harness interface{ Command(run *v1alpha1.AgentRun) []string; EnvVars(run *v1alpha1.AgentRun) []corev1.EnvVar }
   func ResolveHarness(a v1alpha1.AgentType) (Harness, error)
   ```

6. **Relocate the package-level registries into their axis dirs (kept package-level).** Move `toolchainFactories` -> `internal/provision/shared`, `harnessCommands` -> `internal/harness/shared`, `vcsStrategies` -> `internal/workspace/shared`. Keep them as package-level lazy-factory maps (no global mutable state mutated at runtime); explicit `NewRegistry`+DI is the deferred follow-up (Decision 10), not done here.

   ```go
   // internal/provision/shared/registry.go
   var toolchainFactories = map[v1alpha1.ToolchainType]func() Toolchain{
       v1alpha1.ToolchainNix: func() Toolchain { return nix.New() },
   }
   ```

7. **Introduce typed CRD enums.** In `api/v1alpha1/agentrun_types.go` add `NetworkMode`, `WorkspaceType`, `ToolchainType`, `AgentType` typed-string enums with `+kubebuilder:validation:Enum` markers, replacing the untyped `string` fields. Regenerate `zz_generated.deepcopy.go` and CRD/RBAC if markers change. No backwards-compat aliases for the old string fields (AGENTS.md).

   ```go
   // api/v1alpha1/agentrun_types.go
   // +kubebuilder:validation:Enum=none;policy
   type NetworkMode string
   // +kubebuilder:validation:Enum=git;jj
   type WorkspaceType string
   ```

## Validation Steps

- `go build ./...` in `packages/agent/agent-operator-go` (all relocated packages compile, import-alias discipline holds where operator `provision/nix` meets shared `pkg/provision/nix`).
- `go test ./...` in `packages/agent/agent-operator-go` — relocate the builder unit tests alongside their new packages; the merged-Job and security-defaults goldens must match the pre-merge spec (golden-equivalent except the intended Job merge).
- `npx moon run agent-operator-go:typecheck`
- `npx moon run agent-operator-go:lint`
- `npx moon run agent-operator-go:build`
- `npx moon run agent-operator-go:test`
- Regenerate codegen: run controller-gen (deepcopy + CRD) and reconcile `config/rbac/role.yaml`; verify no drift after the enum type changes.

## Success Criteria

- [ ] `internal/{isolation,network,provision,workspace,harness,provider}/` exist with axis-name symmetry to the CLI; `internal/builder/` is gone.
- [ ] Exactly ONE pod isolator under `isolation/shared/`; no fabricated `bwrap`/`docker`/`none` leaves; `runtimeClassName` is the kernel knob.
- [ ] `BuildJob` and `BuildWorkspaceJob` are merged into one builder taking the workspace diff as input.
- [ ] The three security-defaults sources are consolidated into one helper.
- [ ] Real leaves exist only where the operator varies: `provision/nix`, `workspace/{git,jj}`, `harness/{claude,opencode}`.
- [ ] Registries (`toolchainFactories`, `harnessCommands`, `vcsStrategies`) live in their axis `shared/` dirs, package-level.
- [ ] `NetworkMode`/`WorkspaceType`/`ToolchainType`/`AgentType` typed enums replace untyped strings; deepcopy/CRD/RBAC regenerated and drift-free.
- [ ] All operator moon validation tasks pass; pod/admission goldens unchanged except the intended Job merge.

## Files Modified/Created

- Created: `internal/isolation/shared/{job.go,container.go,security.go,hardening.go}`
- Created: `internal/network/shared/networkpolicy.go`
- Created: `internal/provision/shared/{toolchain.go,registry.go}`, `internal/provision/nix/nix.go`
- Created: `internal/workspace/shared/{vcs.go,...}`, `internal/workspace/git/`, `internal/workspace/jj/`
- Created: `internal/harness/{shared,claude,opencode}/`, `internal/provider/provider.go`
- Modified: `api/v1alpha1/agentrun_types.go`, `api/v1alpha1/zz_generated.deepcopy.go`, `config/rbac/role.yaml`, CRD manifests
- Removed: `internal/builder/*` (relocated), `internal/resolver/provider.go` (merged)

## Dependencies

- **01-shared-per-axis-split.md** — provides the per-axis `pkg/{isolation,provision,network,workspace,policy}` shared packages and the neutral `Contribution`/`Capabilities` seam the operator axes import; the operator `provision/nix` aliasing against shared `pkg/provision/nix` requires the shared nix nesting to exist first.
- Sequencing: Phase 2 — begins only after Phase 1 (subplans 02-05) lands green, per Decision 7. (07-operator-policy-symmetry follows this subplan.)

## Estimated Duration

Medium-to-large — mechanical relocation across six axis dirs plus two consolidations (Job-builder merge, security-defaults merge) and the CRD enum typing with codegen regeneration. Effort is in golden-equivalence discipline and import-alias hygiene rather than new logic.

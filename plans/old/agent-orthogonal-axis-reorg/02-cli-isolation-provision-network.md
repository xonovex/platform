---
type: plan
has_subplans: false
parent_plan: plans/agent-orthogonal-axis-reorg.md
parallel_group: 2
status: complete
dependencies:
  plans: [01-shared-per-axis-split.md]
  files:
    - packages/agent/agent-cli-go/internal/sandbox/
    - packages/agent/agent-cli-go/internal/sandboxutil/
    - packages/agent/agent-cli-go/internal/isolation/
    - packages/agent/agent-cli-go/internal/provision/
    - packages/agent/agent-cli-go/internal/network/
skills_to_consult:
  - orthogonal-pattern-guide
  - hexagonal-pattern-guide
  - microkernel-pattern-guide
  - connascence-guide
  - general-fp-guide
  - moon-guide
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
---

## Status (complete — landed with 03/04/05 as Phase-1 CLI)

`internal/{isolation,provision,network}` fanned out with `shared/` cores + bare leaves;
`SandboxConfig` replaced CLI-side by `isolation/shared.RunConfig` + per-leaf knobs; `sandboxutil`
dissolved (`SpawnSandbox`/`WrapWithInitCommands` → `isolation/shared`; per-isolator network
emitters → `isolation/{bwrap,docker}/network.go` bridges depending on `network/shared` one-way);
`nixprov` → `provision/nix` (shared engine aliased `sharednix`); registry/select retargeted to
`pkg/policy`; `plugins.go` the sole concrete importer with lazy `func()` factories. `--isolation-docker-runtime`
now emits docker `--runtime` (kernel isolation live). All `agent-cli-go` moon tasks green.

### Deviations from the written sketch (behavior-preserving)

- **Network leaves**: created `network/{shared,proxy}` only — `host`/`none` carry no per-type data,
  so per the orthogonal-guide's over-factoring caution there is no empty `host`/`none` leaf; the
  closed-enum asymmetry is documented in `network/shared`. The `Mode` enum aliases shared `pkg/network`
  (one owner), not a CLI redefinition.
- **Isolator port** kept its `(int, error)` return and per-leaf `New…` constructors (not the sketch's
  `int`-only signature) to avoid a behavior change.
- **`agentcmd` call sites**: threaded `Agent`/`Provider`/`AgentArgs` from `RunConfig` (the 01 data-coupled
  signature), as flagged in Task 5.
- **Egress allowlist** is now plumbed (`network/proxy.Options` → `SANDBOX_EGRESS_ALLOW` env) rather than
  dead, but full enforcement still requires the out-of-scope host proxy integration.
---

# CLI Isolation/Provision/Network Axis Fan-Out

## Objective

Fan the overloaded `internal/sandbox` tree out into three orthogonal axes — `internal/isolation`, `internal/provision`, `internal/network` — each a literal `shared/` port core plus bare per-type leaves, leaving `internal/sandbox` as the composition/selection layer only. Dissolve the homeless `internal/sandboxutil` into the owning axis `shared/` and cross-axis bridge files, and finish the `SandboxConfig` god-struct deletion CLI-side by giving each leaf its own co-located `Options`. This subplan co-lands with `01` and is the build-green gate for Phase 1's structural moves.

## Tasks

1. **Rename the isolation axis (port + leaves).** Move `internal/sandbox/isolator.go` → `internal/isolation/shared/isolator.go` (package `shared`, the `Isolator` port) and move `internal/sandbox/{bwrap,docker,none}/` → `internal/isolation/{bwrap,docker,none}/`. After this, `internal/sandbox` retains only `registry.go`, `select_test.go`, and `plugins/` (the composed whole). The port stays variant-blind:

   ```go
   // internal/isolation/shared/isolator.go
   package shared

   type Isolator interface {
       Available() bool
       Run(cfg RunConfig, contrib provision.Contribution) int
       Command(cfg RunConfig, contrib provision.Contribution) []string
       HidesHost(passthrough bool, image string) bool
       KernelIsolated(runtime string) bool
   }
   ```

2. **Fan out provision symmetrically.** Move the `Provisioner` interface from `internal/sandbox/provisioner.go` → `internal/provision/shared/provisioner.go`; extract the inline none/command impls into `internal/provision/none/none.go` + `internal/provision/command/command.go`; rename `internal/sandbox/nixprov/` → `internal/provision/nix/` (`nix.go`, `resolve.go`). Because CLI `provision/nix` imports shared `pkg/provision/nix`, alias the shared import at every call site:

   ```go
   // internal/provision/nix/nix.go
   import sharednix "github.com/.../shared-agent-go/pkg/provision/nix"
   func Contribute(opts Options) shared.Contribution { /* uses sharednix.Resolve(...) */ }
   ```

3. **Create the network axis as a closed enum (no registry).** Add `internal/network/{shared,host,none,proxy}/`. Move the method-agnostic proxy env builder + `EgressIsRestricted` from `internal/sandboxutil/network.go` into `internal/network/shared/network.go`, and document that the registry-free asymmetry vs isolation/provision is deliberate:

   ```go
   // internal/network/shared/network.go
   package shared

   // Network is a closed enum (host|none|proxy); it carries no Registry by
   // design — unlike isolation/provision, the variant set is fixed.
   func EgressIsRestricted(mode Mode) bool { return mode == ModeProxy }
   func ProxyEnv(allow []string, url string) map[string]string { /* ... */ }
   ```

4. **Dissolve `internal/sandboxutil`.** Move `utils.go` (`SpawnSandbox`, `WrapWithInitCommands`) → `internal/isolation/shared/spawn.go`. Turn the per-isolator emitters (`BwrapNetworkArgs`/`DockerNetworkArgs`) into cross-axis **bridge** files owned by the dependent leaf: `internal/isolation/bwrap/network.go` and `internal/isolation/docker/network.go`, each depending on `network/shared` one-way only (never the reverse, and `shared/` never reaches a leaf):

   ```go
   // internal/isolation/bwrap/network.go  (bridge: isolation -> network/shared)
   package bwrap
   import netshared "github.com/.../internal/network/shared"
   func networkArgs(mode netshared.Mode, opts netshared.Options) []string { /* ... */ }
   ```

5. **Add per-type `Options` co-located with each variant.** Replace fields stripped from the deleted `SandboxConfig` with immutable value structs owned by each leaf — never a god-struct: `internal/isolation/docker/options.go` (`Image, Runtime, Pids, Memory`), `internal/isolation/bwrap/options.go` (`Passthrough`, host binds), `internal/network/proxy` options (egress allowlist + url), `internal/provision/nix` options (`Source`). Each plugin consumes only its own `Options`:

   ```go
   // internal/isolation/docker/options.go
   package docker
   type Options struct { Image, Runtime string; Pids int64; Memory string } // value type, no methods that mutate
   ```

   **`agentcmd` call-site update (from 01).** Subplan 01 weakened
   `agentcmd.BuildAgentCommand`/`BuildProviderEnv` from `*types.SandboxConfig` to explicit
   `(*types.AgentConfig, *types.ModelProvider, agentArgs, …)` (data coupling, since the god-struct
   is gone). The bwrap/docker leaves are the only callers; thread `Agent`/`Provider`/`AgentArgs`
   from the resolved `Request`/exec context — they are NOT part of any per-type `Options` struct.

6. **Slim `internal/sandbox/registry.go`.** Strip the docker-shaped `Image`/`Runtime` out of `Request` into `isolation/docker.Options`; keep `Select` building `Capabilities` and calling shared `pkg/policy.EnforcePolicy` unchanged (it still names no concrete variant):

   ```go
   // internal/sandbox/registry.go
   type Registry = map[Method]Factory // lazy func() factories, no global state
   func Select(reg Registry, req Request, pol policy.SandboxPolicy) (shared.Isolator, error) {
       caps := shared.Capabilities{ /* Pinned, HostToolsUnreachable, EgressRestricted, KernelIsolated */ }
       if err := policy.EnforcePolicy(caps, pol); err != nil { return nil, err }
       // ...
   }
   ```

7. **Rewire `internal/sandbox/plugins/plugins.go` (the sole composition root).** Repoint imports at the new `isolation/*`, `provision/*`, and `network/*` leaves; keep the lazy `func()` factory registry (no eager instances) so binding stays deferred to `Select`. While here, fix the `none` isolator to honor `BindPaths`/`RoBindPaths`:

   ```go
   // internal/sandbox/plugins/plugins.go — only file importing concrete leaves
   func Register(reg sandbox.Registry) {
       reg[sandbox.MethodBwrap]  = func() shared.Isolator { return bwrap.New() }
       reg[sandbox.MethodDocker] = func() shared.Isolator { return docker.New() }
       reg[sandbox.MethodNone]   = func() shared.Isolator { return none.New() }
   }
   ```

## Validation Steps

- `go build ./...` in `packages/agent/agent-cli-go` — green build is the Phase-1 gate for this subplan.
- `go test ./...` in `packages/agent/agent-cli-go` (covers `internal/sandbox/select_test.go` after the move).
- `npx moon run agent-cli-go:typecheck`
- `npx moon run agent-cli-go:lint`
- `npx moon run agent-cli-go:build`
- `npx moon run agent-cli-go:test`
- Integration help-text tests: no flag surface changes land in this subplan (flag rewiring is `04`), so `internal/cmd` help-text fixtures must remain unchanged — run `go test ./internal/cmd/...` and `test/integration` to confirm no drift.

## Success Criteria

- [ ] `internal/isolation/{shared,bwrap,docker,none}` exist; `Isolator` port lives in `isolation/shared`.
- [ ] `internal/provision/{shared,none,command,nix}` exist; `nixprov` is gone; shared `pkg/provision/nix` imported via alias at every call site.
- [ ] `internal/network/{shared,host,none,proxy}` exist with no registry; the no-registry asymmetry is documented in `network/shared`.
- [ ] `internal/sandboxutil` is gone; `SpawnSandbox`/`WrapWithInitCommands` live in `isolation/shared`; network emitters live in `isolation/{bwrap,docker}/network.go` bridge files depending on `network/shared` one-way only.
- [ ] Each leaf consumes its own `Options`; no `SandboxConfig` reference remains CLI-side.
- [ ] `internal/sandbox` holds only `registry.go`, `select_test.go`, `plugins/`; `plugins.go` is the sole importer of concrete leaves; registry stays lazy `func()` factories.
- [ ] `none` isolator honors `BindPaths`/`RoBindPaths`.
- [ ] All validation tasks green; no help-text fixture drift.

## Files Modified/Created

Created:
- `internal/isolation/shared/isolator.go`, `internal/isolation/shared/spawn.go`
- `internal/isolation/{bwrap,docker,none}/` (moved) + `isolation/{bwrap,docker}/network.go` bridges + `isolation/{docker,bwrap}/options.go`
- `internal/provision/shared/provisioner.go`, `internal/provision/none/none.go`, `internal/provision/command/command.go`, `internal/provision/nix/{nix.go,resolve.go,options.go}`
- `internal/network/{shared,host,none,proxy}/` + proxy options

Modified:
- `internal/sandbox/registry.go` (slimmed `Request`)
- `internal/sandbox/plugins/plugins.go` (repointed imports)
- `internal/sandbox/select_test.go` (import paths)

Removed:
- `internal/sandbox/{isolator.go,provisioner.go,nixprov/}`
- `internal/sandboxutil/`

## Dependencies

- **`01-shared-per-axis-split.md`** — must land first (co-landing, Phase 1): provides the per-axis shared packages (`pkg/provision/nix`, `pkg/policy.EnforcePolicy`, `Contribution`, `Capabilities`) and the deleted `SandboxConfig`. This subplan repoints CLI leaves at those shared packages and removes the last CLI references to `SandboxConfig`; the build is not green until both are merged together.

## Estimated Duration

Medium — ~1 day. Mostly mechanical relocation + import-alias discipline; the load-bearing care is the bridge-file import direction and finishing the `SandboxConfig`→per-type-`Options` migration without leaving a god-struct.

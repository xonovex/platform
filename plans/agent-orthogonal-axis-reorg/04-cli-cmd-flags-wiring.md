---
type: plan
has_subplans: false
parent_plan: plans/agent-orthogonal-axis-reorg.md
parallel_group: 3
status: pending
dependencies:
  plans: [02-cli-isolation-provision-network.md, 03-cli-workspace-terminal.md]
  files:
    - packages/agent/agent-cli-go/internal/cmd/run.go
    - packages/agent/agent-cli-go/internal/executor/executor.go
    - packages/agent/agent-cli-go/internal/config
    - packages/agent/agent-cli-go/test/integration/run_test.go
skills_to_consult: [microkernel-pattern-guide, connascence-guide, general-fp-guide, moon-guide]
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# CLI cmd slim + per-type flag grammar

## Objective

Slim `internal/cmd/run.go` into per-axis flag groups that each assemble one variant's
immutable `Options`, replace the positional `resolveAxes` tuple with a named `ResolvedAxes`
struct, and wire the three new/renamed per-type flags (`--isolation-docker-runtime`,
`--isolation-bwrap-passthrough`, `--network-proxy-egress-allow`) so the latent
kernel-isolation and egress-allowlist capabilities go live. Dead code (`executor.go`,
`buildDirect*`) is deleted and the CLI-only config loader evicted from shared in 01 is
re-homed under `internal/config`.

## Tasks

1. **Restructure `internal/cmd/run.go` into per-axis flag groups.** Replace the flat flag
   block with one `register*Flags` + `build*Options` pair per axis, each producing the
   immutable value `Options` for the selected variant (module-level funcs, no shared mutable
   state per general-fp). Keep the bare selectors and bare policy flags unchanged.

   ```go
   // bare selectors (unchanged names)
   fs.StringVar(&f.isolation, "isolation", "none", "...")   // none|bwrap|docker
   fs.StringVar(&f.provision, "provision", "none", "...")
   fs.StringVar(&f.network,   "network",   "host", "...")
   // bare policy flags (unchanged names)
   fs.BoolVar(&f.requireKernelIsolation, "require-kernel-isolation", false, "...")
   fs.BoolVar(&f.requireEgressRestricted, "require-egress-restricted", false, "...")

   func buildIsolationOptions(f flags) (isolation.Variant, any) { ... }
   func buildNetworkOptions(f flags) (network.Variant, any)     { ... }
   ```

2. **Introduce a named `ResolvedAxes` struct, replacing the positional return** (connascence:
   weaken the positional connascence of the 4+ value tuple to connascence of name).

   ```go
   // before
   func resolveAxes(f flags) (iso isolation.Isolator, prov provision.Provisioner,
       net network.Network, passthrough bool, err error)

   // after
   type ResolvedAxes struct {
       Isolation     isolation.Isolator
       Provision     provision.Provisioner
       Network       network.Network
       IsolationOpts isolation.Options
       NetworkOpts   network.Options
   }
   func resolveAxes(f flags) (ResolvedAxes, error)
   ```

3. **Wire the three new/renamed per-type flags under the `--<axis>-<type>-<option>` grammar.**
   - `--isolation-docker-runtime` (docker-gated): sets `isolation/docker.Options.Runtime`,
     making `Isolator.KernelIsolated(runtime)` return true so `RequireKernelIsolation` is
     reachable (previously the hardcoded `Runtime: ""` left it dead).
   - `--isolation-bwrap-passthrough`: renamed from `--host-passthrough`; sets
     `isolation/bwrap.Options.Passthrough` (no shim per AGENTS.md).
   - `--network-proxy-egress-allow` (repeatable): seeds `network/proxy.Options.EgressAllowlist`
     (previously written-never-read).

   ```go
   fs.StringVar(&f.isolationDockerRuntime, "isolation-docker-runtime", "", "runC-class runtime (docker only)")
   fs.BoolVar(&f.isolationBwrapPassthrough, "isolation-bwrap-passthrough", false, "share host paths (bwrap only)")
   fs.StringArrayVar(&f.networkProxyEgressAllow, "network-proxy-egress-allow", nil, "egress allowlist host (proxy only)")
   ```

4. **Delete dead code.** Remove `internal/executor/executor.go` (imported by no non-test
   source; it duplicates `run.go` plus the `none` isolator) and delete the
   `buildDirectCommand`/`buildDirectEnv` helpers in `run.go` that re-implement the `none`
   isolator. Route the no-isolation path through `isolation/none` so the `none` leaf is the
   single source of the direct-exec command/env.

5. **Re-home the CLI-only config loader evicted from shared in 01.** Move
   `shared pkg/config/loader.go` (YAML/TOML) into `agent-cli-go internal/config/loader.go`,
   keeping `go-toml`/`yaml` as CLI-module-only deps (shared dropped them in 01). Update
   `run.go`'s import to the new internal path; run `go mod tidy` in the CLI module so the deps
   land in the CLI `go.mod` only.

6. **Update help-text + capability integration tests** in `test/integration/run_test.go`
   (breaking change, no shim per AGENTS.md). Replace `--host-passthrough` assertions with
   `--isolation-bwrap-passthrough`; add the two new flag strings to the help-text golden; add
   coverage that `--isolation-docker-runtime` drives `KernelIsolated`/`RequireKernelIsolation`
   and `--network-proxy-egress-allow` seeds the proxy allowlist.

   ```go
   require.Contains(t, help, "--isolation-docker-runtime")
   require.Contains(t, help, "--isolation-bwrap-passthrough")
   require.Contains(t, help, "--network-proxy-egress-allow")
   // capability: docker runtime -> kernel isolated
   axes, err := resolveAxes(flags{isolation: "docker", isolationDockerRuntime: "runsc"})
   require.True(t, axes.Isolation.KernelIsolated("runsc"))
   ```

## Validation Steps

- `npx moon run agent-cli-go:typecheck`
- `npx moon run agent-cli-go:lint`
- `npx moon run agent-cli-go:build`
- `npx moon run agent-cli-go:test`
- In `packages/agent/agent-cli-go`: `go build ./...`, `go test ./...`, `go mod tidy`
  (confirm `go-toml`/`yaml` are present in the CLI `go.mod` and absent from shared).
- Integration help-text tests must pass with the new exact flag strings (flags changed):
  `go test ./test/integration/...`.

## Success Criteria

- [ ] `run.go` is organized into per-axis `register*Flags`/`build*Options` pairs; no flat
      catch-all flag block.
- [ ] `resolveAxes` returns `(ResolvedAxes, error)`; no positional 4+ value tuple remains.
- [ ] `--isolation-docker-runtime`, `--isolation-bwrap-passthrough`,
      `--network-proxy-egress-allow` exist and are wired to their variant `Options`.
- [ ] `--host-passthrough` is gone (renamed, no shim).
- [ ] `internal/executor/executor.go` and `buildDirect*` helpers are deleted; the
      no-isolation path goes through `isolation/none`.
- [ ] Config loader lives in `internal/config`; `go-toml`/`yaml` are CLI-module-only.
- [ ] Integration help-text + new capability tests pass.
- [ ] typecheck / lint / build / test green for `agent-cli-go`.

## Files Modified/Created

- Modified: `packages/agent/agent-cli-go/internal/cmd/run.go`
- Created: `packages/agent/agent-cli-go/internal/config/loader.go`
- Deleted: `packages/agent/agent-cli-go/internal/executor/executor.go`
- Modified: `packages/agent/agent-cli-go/test/integration/run_test.go`
- Modified: `packages/agent/agent-cli-go/go.mod` / `go.sum` (deps re-homed via `go mod tidy`)

## Dependencies

- `02-cli-isolation-provision-network.md` — provides `internal/{isolation,provision,network}`
  with the per-variant `Options`, the `none` isolator, the docker `Runtime` knob, the bwrap
  `Passthrough` knob, and the proxy `EgressAllowlist`; this subplan only assembles and wires
  them from the CLI surface.
- `03-cli-workspace-terminal.md` — provides `internal/workspace`/`internal/terminal` that
  `run.go` composes alongside the three confinement axes; the slimmed `run.go` references
  their types.

## Estimated Duration

~0.5 day — cmd slim + struct rename + three flag wirings + dead-code removal + loader move +
integration-test updates; mechanical given 02/03 supply the `Options` surfaces.

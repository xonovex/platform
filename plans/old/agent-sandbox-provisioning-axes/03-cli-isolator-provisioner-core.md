---
type: plan
has_subplans: false
parent_plan: plans/agent-sandbox-provisioning-axes.md
parallel_group: 2
status: complete
dependencies:
  plans: [shared-types-policy-dedups]
  files:
    - packages/agent/agent-cli-go/internal/sandbox/isolator.go
    - packages/agent/agent-cli-go/internal/sandbox/provisioner.go
    - packages/agent/agent-cli-go/internal/sandbox/registry.go
    - packages/agent/agent-cli-go/internal/sandbox/bwrap/bwrap.go
    - packages/agent/agent-cli-go/internal/sandbox/docker/docker.go
    - packages/agent/agent-cli-go/internal/sandbox/none/none.go
    - packages/agent/agent-cli-go/internal/cmd/run.go
    - packages/agent/agent-cli-go/internal/sandboxutil/utils.go
skills_to_consult: [general-fp-guide, debugging-guide]
validation:
  type_check: pass        # agent-cli-go go-build
  lint: pass              # agent-cli-go go-lint (0 issues)
  build: pass             # agent-cli-go go-build (02 break resolved: nix/nixflake collapsed)
  tests: pass             # agent-cli-go go-test (Select/policy/alias, isolator hardening, network, provisioners)
  integration: pass       # gated bwrap --unshare-net egress block verified live
---

# 03 — CLI `Isolator × Provisioner` composition core

## Objective

Replace the fused `SandboxExecutor`-per-method design with the three-axis composition: a small set of **`Isolator`s** (`none`, `bwrap`, `docker`) that each consume a **`Provisioner` contribution** AND apply a **`NetworkMethod`** (`host`/`none`/`proxy`) explicitly, plus the `none` and `command` provisioners and the `hostPassthrough` knob. This subplan also carries the **isolator hardening** (deny-default `bwrap`, docker security defaults) — the `bwrap`/`docker` isolators stop bleeding the host in by default. (The `nix` provisioner is subplan 04; compose removal is 05.)

Threat model (see parent Overview): the agent runs UNTRUSTED model-generated code under possible prompt injection; `bwrap` and default-runc are attack-surface reduction, not kernel trust boundaries. "Host tools unreachable" ≠ "host unreachable" — egress is a SEPARATE guarantee. Fail CLOSED: refuse to run when a requested guarantee can't be established; never silently degrade.

## Tasks

1. **Define the seam** in `internal/sandbox/`:
   ```go
   type Provisioner interface { Contribute(cfg *types.SandboxConfig) (types.Contribution, error) }
   type Isolator    interface {
       Available() (bool, error)
       Run(cfg *types.SandboxConfig, c types.Contribution) (int, error)
       Command(cfg *types.SandboxConfig, c types.Contribution) []string
   }
   ```
   Add `Select(iso types.IsolationMethod, prov types.ProvisioningMethod, net types.NetworkMethod, passthrough bool, runtime string, pol types.SandboxPolicy) (Isolator, Provisioner, error)` that validates the curated matrix + calls `sandbox.EnforcePolicy(iso, prov, net, passthrough, runtime, pol)` (from shared, subplan 01) — threading all four guarantees, not just iso×prov.

2. **Refactor the `bwrap` Isolator** (from `bwrap/bwrap.go:63-183`) to consume a `Contribution`: bind `c.RoBindPaths` read-only, **prepend** `c.PathEntries` to PATH, apply `c.Env`, wrap with `c.InitCommands` (via the shared `WrapWithInitCommands`). Honor `HostPassthrough`: **on** = also ro-bind host `/usr,/lib,/lib64,/bin,/etc` + append host `$PATH` (today's leaky `bwrap` behavior); **off** = deny-default (only the contribution's binds). This collapses the leaky/deny-default split into one isolator.

   **bwrap hardening (deny-default mode)** — replace the blanket `--bind $HOME $HOME` with a **sandbox-local HOME** dir; bind ONLY the workspace (rw) + `RepoDir` (ro) + the curated `sandbox.UserConfigPaths` entries, read-only where the agent doesn't need to write them — **no host-`$HOME` bind**. Replace `--dev-bind /dev /dev` with `--dev /dev` (minimal devtmpfs; the full bind exposed `/dev/sda,/dev/mem,input`). Add `--clearenv` + an explicit `--setenv` allowlist (no host env bleed; API keys injected explicitly). Assert (already bwrap defaults) `no-new-privs` + `cap-drop ALL`. OPTIONAL/residual (document, may defer): `--new-session` (or a seccomp `ioctl(TIOCSTI)` filter) for CVE-2017-5226; `--unshare-user` + `--disable-userns`; cgroup limits via `systemd-run --scope`; Landlock.

3. **Refactor the `docker` Isolator** (from `docker/docker.go:117-122`) to consume a `Contribution`: mount each `RoBindPath` via `-v <p>:<p>:ro`, set PATH/env, run `InitCommands` before the agent. `HostPassthrough` = also expose the base-image tools on PATH (the actual host is never visible).

   **docker hardening (DEFAULTS, not opt-in)** — `--security-opt=no-new-privileges`, `--cap-drop ALL`, keep the **default seccomp** profile (NEVER `seccomp=unconfined`), `--security-opt apparmor=docker-default`, `--read-only` rootfs + `--tmpfs /tmp:rw,noexec,nosuid` (+ a writable workdir bind — the ONLY writable bind), `--pids-limit` + `--memory` + `--cpus`. **Stop mounting whole `$HOME` rw**: mount only the curated config paths **read-only** into a synthetic container HOME; the workdir is the only writable bind. OPTIONAL/residual: `--runtime runsc` (gVisor) to satisfy `RequireKernelIsolation`; rootless docker / userns-remap.

4. **Refactor the `none` Isolator** (from `none/none.go`) to apply `PathEntries`/`Env`/`InitCommands` directly on the host (no bind/namespace). `HostPassthrough` = append the host PATH after the prepended entries.

5. **Apply `NetworkMethod` explicitly in every Isolator** — each isolator reads `cfg.Network` (`host`/`none`/`proxy`) + `cfg.EgressAllowlist` and applies it:
   - **bwrap**: `none`/`proxy` → emit `--unshare-net`; `proxy` ALSO injects `HTTP_PROXY`/`HTTPS_PROXY` (+ `NO_PROXY` for nothing) pointing at the host-side allowlist proxy and blocks metadata (`169.254.169.254`)/RFC1918/link-local; `host` → share net (no `--unshare-net`).
   - **docker**: `none` → `--network none`; `proxy` → custom net + `HTTP(S)_PROXY` to the allowlist proxy; `host` → bridge.
   - **REGRESSION GUARD**: net state MUST be explicit (always emit `--unshare-net` for `none`/`proxy`, never leave it implicit) so collapsing nix/nixflake — which `--unshare-net` today — into the unified `bwrap` isolator can't silently drop today's nix-tier network isolation. `host` does NOT satisfy `RequireEgressRestricted`; `none`/`proxy` do.

6. **Implement the `none` and `command` provisioners**: `none` returns an empty `Contribution` (image/host provides); `command` returns `Contribution{InitCommands: cfg.SandboxInitCommands}` — **activating the dormant hook**.

7. **Wire flags + selection** in `cmd/run.go` (`:167-201`): add `--isolation {none,bwrap,docker}`, `--provision {none,nix,command}`, `--network {host,none,proxy}` (default **`proxy`** when a proxy/allowlist is configured, else `host`), `--egress-allow` (repeatable, extends `DefaultEgressAllowlist`), `--host-passthrough` (default **off**), `--init-command` (repeatable, feeds `SandboxInitCommands`). Keep `--sandbox` as a **deprecated alias** mapping the old methods → `(iso, prov, passthrough)` pairs for one release (`bwrap`→`bwrap×none+passthrough`, `nix`/`nixflake`→`bwrap×nix`, `docker`→`docker×none`); the alias maps ALL legacy methods to **`Network=host`** to preserve today's behavior for one release. Route policy via the split `SandboxPolicy`.

8. **Condition `IsolationHidesHost`** — it must NOT be a naive `docker→true`. Return true only when host tools are off PATH AND not bind-reachable: **closure-only store binds + NO host-`$HOME` bind + (docker) a pinned image**. This is what makes `RequireHostToolsUnreachable` accurate; bwrap with `HostPassthrough` on, or a host-`$HOME` bind, returns false.

9. **Tests**: each isolator applies a `Contribution` correctly (binds/PATH/env/init); the `command` provisioner runs its list; `Select` rejects invalid cells with the policy error; the `--sandbox` alias map resolves to the right `(iso, prov, Network=host)`. **Network parity**: a `connect()` to a known host FAILS under `none` and is blocked-except-allowlist under `proxy`, across BOTH `bwrap` and `docker`. **Hardening**: assert the deny-default flags are present (`--dev /dev`, `--clearenv`, no `--dev-bind /dev`; docker `--read-only`/`--cap-drop ALL`/`no-new-privileges`/default seccomp), and that there is **NO host-`$HOME` bind** in deny-default mode. **`IsolationHidesHost` conditioning**: true only under closure-only binds + no host-`$HOME` + pinned image.

## Validation Steps

```bash
npx moon run agent-cli-go:go-build agent-cli-go:go-test agent-cli-go:go-lint
# Smoke (inside a nix shell so the plugin no-ops):
agent-cli-go run --isolation bwrap --provision command --init-command 'echo hi' -- <agent> --version
# Network: no egress under --network none; allowlist-only under --network proxy:
agent-cli-go run --isolation bwrap --network none -- sh -c 'curl -sS https://example.com; echo $?'   # expect failure
agent-cli-go run --isolation docker --network proxy --egress-allow github.com -- sh -c 'curl -sS https://example.com; echo $?'  # blocked unless allowlisted
```

## Success Criteria

- [x] `Isolator`/`Provisioner` interfaces + `Select` exist (`internal/sandbox/{isolator,provisioner}.go`); the curated matrix + policy are enforced at selection (`Select` calls shared `EnforcePolicy` and fails closed).
- [x] `bwrap`/`docker`/`none` isolators consume a `Contribution`; `hostPassthrough` toggles host/base-image bleed-through; the leaky/deny-default bwrap split is one isolator + the knob.
- [x] Every isolator applies `cfg.Network` EXPLICITLY (`--unshare-net` for none/proxy / `--network none` / proxy env injection); the regression guard holds — unit-asserted for every non-host cell, and a gated live test confirms `--unshare-net` actually blocks egress. `none` fails CLOSED when egress restriction is requested.
- [x] bwrap deny-default hardening: sandbox-local HOME tmpfs (no host-`$HOME` bind), `--dev /dev` (not `--dev-bind`), `--clearenv` + explicit `--setenv` allowlist (asserted). `no-new-privs` + `cap-drop ALL` are bwrap defaults (no flag to emit; documented).
- [x] docker security DEFAULTS: `no-new-privileges`, `--cap-drop ALL`, default seccomp (never `unconfined`), `apparmor=docker-default`, `--read-only` + `--tmpfs /tmp:rw,noexec,nosuid` + workdir-only writable, `--pids-limit`/`--memory`/`--cpus` (cpus clamped to host cores); no whole-`$HOME` rw mount (asserted).
- [x] `HostToolsUnreachable` conditions the shared classifier (closure-only binds + no host-`$HOME` + pinned image), not a naive `docker→true`; tested across the matrix.
- [x] `none` + `command` provisioners work; the dormant `WrapWithInitCommands` is wired (tested).
- [x] `--isolation`/`--provision`/`--network`/`--egress-allow`/`--host-passthrough`/`--init-command` flags work; `--sandbox` still works as a mapped alias (legacy → `Network=host`; tested).
- [x] build/lint/test green (`agent-cli-go` builds again — the 02 channel-renderer break is resolved).

## Completion Notes

- **Resolves the 02 intermediate break.** The old fused `internal/sandbox/{nix,nixflake}` executors are deleted (collapsed into the bwrap isolator + the nix provisioner), so `agent-cli-go` builds again. `internal/sandbox/compose` and `internal/nixenv` are now unreferenced but still compile — left for **05** (compose removal) and **04** (the nix provisioner reuses `nixenv` dirs + the shared `ComputeEnvID`).
- **Nix provisioning is deferred to 04.** `selectProvisioner(nix)` returns `ErrNoProvisioner` (fail closed); `--provision nix` / `--sandbox nix|nixflake` / `--require-pinned-toolchain` therefore error cleanly until 04 registers the `nixprov` provisioner. This matches the wave: 04 implements resolve→closure→mount.
- **Network proxy is a documented residual for bwrap.** The isolators emit the network method explicitly (the tested regression guard) and inject `HTTP(S)_PROXY` env for `proxy`; the egress allowlist is plumbed onto `SandboxConfig.EgressAllowlist`. The actual allowlist-enforcing proxy and the bwrap namespace→proxy wiring (slirp/pasta) are infrastructure, out of scope here — egress fails closed (`--unshare-net`) rather than open.
- **Provider tokens reach bwrap via `--setenv`** (a consequence of `--clearenv`, per the plan's explicit-allowlist requirement).

## Files Modified/Created

- `internal/sandbox/isolator.go`, `internal/sandbox/provisioner.go` — new (interfaces + `Select` + `none`/`command` provisioners).
- `internal/sandbox/{bwrap,docker,none}/*.go` — refactor to consume a `Contribution`; apply `NetworkMethod` explicitly; bwrap deny-default hardening (sandbox-local HOME, `--dev /dev`, `--clearenv`/`--setenv`); docker security defaults; condition `IsolationHidesHost`.
- `internal/sandbox/registry.go` — replace the per-method switch with `Select`.
- `internal/cmd/run.go` — new flags (`--isolation`/`--provision`/`--network`/`--egress-allow`/`--host-passthrough`/`--init-command`) + `--sandbox` alias map (legacy → `Network=host`).
- `internal/sandboxutil/utils.go` — `WrapWithInitCommands` consumed by all isolators; proxy-env helper for `NetworkMethod=proxy`.

## Dependencies

`01-shared-types-policy-dedups` (the three-axis types — `IsolationMethod`/`ProvisioningMethod`/`NetworkMethod` + `EgressAllowlist`/`DefaultEgressAllowlist` — `Contribution`, and the four-guarantee split policy incl. `RequireEgressRestricted`).

## Estimated Duration

~3–3.5 days (now also covers the `NetworkMethod` axis + bwrap/docker hardening).

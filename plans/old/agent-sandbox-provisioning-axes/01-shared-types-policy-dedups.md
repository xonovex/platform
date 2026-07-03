---
type: plan
has_subplans: false
parent_plan: plans/agent-sandbox-provisioning-axes.md
parallel_group: 1
status: complete
dependencies:
  plans: []
  files:
    - packages/shared/shared-agent-go/pkg/types/sandbox.go
    - packages/shared/shared-agent-go/pkg/sandbox/policy.go
    - packages/shared/shared-core-go/pkg/shell/shell.go
    - packages/shared/shared-core-go/pkg/envutil/envutil.go
    - packages/shared/shared-agent-go/pkg/agentcmd/agentcmd.go
    - packages/agent/agent-cli-go/internal/sandboxutil/utils.go
skills_to_consult: [general-fp-guide, debugging-guide]
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: n/a
---

# 01 — Shared three-axis types, decoupled policy, and pure-helper dedups

## Objective

Hoist the **three-axis types** (isolation × provisioning × **network**, plus the `HostPassthrough` knob), the **four-guarantee decoupled policy**, and the pure CLI helpers into `shared-core-go`/`shared-agent-go` — the foundation both the CLI refactor (03–05) and the operator realizer (06) consume. Purely additive to `shared`; the old `SandboxMethod` stays during the transition so nothing breaks yet. Threat model (see parent Overview): the agent runs **untrusted, model-generated code** — fail CLOSED, never silently degrade; "host tools unreachable" ≠ "host unreachable" (egress is a *separate* guarantee).

## Tasks

1. **Add the three-axis + contribution types** to `shared-agent-go/pkg/types/sandbox.go` (alongside the existing `SandboxMethod`, kept for now):
   ```go
   type IsolationMethod string
   const ( IsolationNone IsolationMethod = "none"; IsolationBwrap = "bwrap"; IsolationDocker = "docker" )

   type ProvisioningMethod string
   const ( ProvisionNone ProvisioningMethod = "none"; ProvisionNix = "nix"; ProvisionCommand = "command" )

   // NetworkMethod is the third orthogonal axis (REPLACES the old `Network bool`).
   //   host  = share host net, unrestricted egress — today's de-facto behavior, now an EXPLICIT opt-in; does NOT satisfy RequireEgressRestricted.
   //   none  = no network (bwrap --unshare-net / docker --network none); satisfies RequireEgressRestricted.
   //   proxy = egress ONLY via host-side allowlist HTTP(S) proxy; link-local + metadata (169.254.169.254) + RFC1918 + loopback denied; satisfies RequireEgressRestricted. RECOMMENDED DEFAULT for untrusted code that still needs the model API.
   type NetworkMethod string
   const ( NetworkHost NetworkMethod = "host"; NetworkNone NetworkMethod = "none"; NetworkProxy NetworkMethod = "proxy" )

   // Contribution is what a Provisioner hands an Isolator; the Isolator applies it
   // via its own mechanism (bwrap binds / docker -v). Pure data, no host calls.
   type Contribution struct {
       RoBindPaths  []string          // host paths to mount read-only (e.g. /nix/store, a closure)
       PathEntries  []string          // PATH dirs to PREPEND (the pinned tools)
       Env          map[string]string // extra env (devShell vars)
       InitCommands []string          // run once at init before the agent (the `command` provisioner)
   }

   // DefaultEgressAllowlist seeds NetworkProxy: provider API endpoints + common
   // package registries / git forges. `--egress-allow` (repeatable) EXTENDS it.
   var DefaultEgressAllowlist = []string{ /* api.anthropic.com, registry.npmjs.org, pypi.org, files.pythonhosted.org, crates.io, github.com, codeload.github.com, … */ }
   ```
   Add `Isolation IsolationMethod`, `Provisioning ProvisioningMethod`, `HostPassthrough bool`, `Network NetworkMethod` (REPLACING the old `Network bool` wherever referenced), and `EgressAllowlist []string` to `SandboxConfig` (`:26-47`). Default `Network = proxy` when a proxy/allowlist is available, else `host`; the deprecated `--sandbox` alias maps legacy methods to `Network=host` for one release.

2. **Split the policy.** Replace `SandboxPolicy{RequirePinnedToolchain}` (`sandbox.go:17-23`) with **four** independently-requestable guarantees:
   ```go
   type SandboxPolicy struct {
       RequirePinnedProvisioning   bool // provisioning ∈ {nix, pinned image}; enforced at resolve via --frozen / --no-write-lock-file + committed lock (fail closed)
       RequireHostToolsUnreachable bool // host tools off PATH AND not bind-reachable; CONDITIONED on closure-only store binds + NO host-$HOME bind + (docker) a pinned image
       RequireEgressRestricted     bool // Network ∈ {none, proxy}
       RequireKernelIsolation      bool // kernel boundary: docker --runtime runsc/gVisor, or pod with sandboxed runtimeClass (gVisor/Kata/kata-cc). NOT bwrap / default runc.
   }
   ```

3. **Create `shared-agent-go/pkg/sandbox/policy.go`** as the single source of truth, moving the classification out of `agent-cli-go/internal/sandbox/registry.go:36-95`:
   ```go
   func ProvisioningIsPinned(p ProvisioningMethod) bool { return p == ProvisionNix } // images pinned separately
   func IsolationHidesHost(i IsolationMethod, passthrough bool) bool {
       // NOTE: host-tools-unreachable also DEPENDS on closure-only store binds + NO host-$HOME bind
       // (and a pinned image for docker). That bind discipline is enforced in 03/04 — this classifier
       // only gates the isolation axis; passthrough=true forfeits the guarantee for bwrap.
       switch i { case IsolationBwrap: return !passthrough; case IsolationDocker: return true; default: return false }
   }
   func EgressIsRestricted(n NetworkMethod) bool { return n == NetworkNone || n == NetworkProxy }
   func KernelIsolated(i IsolationMethod, runtime string) bool {
       // docker + a sandboxed runtime (runsc/gVisor) or an operator pod with a sandboxed runtimeClass.
       // bwrap, isolation=none, and default-runc are ATTACK-SURFACE REDUCTION, NOT a kernel boundary → false.
       return i == IsolationDocker && (runtime == "runsc" || runtime == "gvisor")
   }
   func EnforcePolicy(iso IsolationMethod, prov ProvisioningMethod, net NetworkMethod, passthrough bool, runtime string, pol SandboxPolicy) error
   ```
   `EnforcePolicy` returns a clear, **named** error when a requested guarantee is unmet — `RequirePinnedProvisioning` (e.g. `prov=none/command` without a pinned image), `RequireHostToolsUnreachable` (`none × *` or `bwrap` + passthrough), `RequireEgressRestricted` (`Network=host`), `RequireKernelIsolation` (bwrap / default-runc). Fail CLOSED. Encode the curated matrix (`{none,bwrap,docker} × {none,nix,command}`) as a validity check here too.

4. **Fold the shell-quote dup.** Delete `sandboxutil.shellQuote` (`utils.go:128`) and route its callers (`buildShellCommand`, `WrapWithInitCommands`) to `shared-core-go/pkg/shell.Quote` (`shell.go:7`).

5. **Promote pure env helpers** `ParseCustomEnv`/`MergeEnvMaps`/`EnvMapToSlice` (`utils.go:96-126`) into a new `shared-core-go/pkg/envutil`; update the CLI imports.

6. **Promote `BuildAgentCommand`/`BuildProviderEnv`** (`utils.go:16-73`) into a new `shared-agent-go/pkg/agentcmd`; update CLI callers. (The operator adopts it in subplan 06 — note it; don't touch operator files here.)

7. **Unit-test** the policy matrix in `pkg/sandbox/policy_test.go` (every `iso × prov × net × passthrough × runtime` → expected `ProvisioningIsPinned`/`IsolationHidesHost`/`EgressIsRestricted`/`KernelIsolated`/`EnforcePolicy`), and the moved helpers in their new packages. Table-test the **network** guarantee (`host`→unmet, `none`/`proxy`→met) and the **kernel** guarantee (`docker+runsc`→met; `bwrap`/`none`/`docker+default-runc`→unmet) across the matrix; assert each unmet guarantee returns its distinct named error.

## Validation Steps

```bash
npx moon run shared-agent-go:go-build shared-agent-go:go-test shared-agent-go:go-lint
npx moon run shared-core-go:go-build shared-core-go:go-test shared-core-go:go-lint
npx moon run agent-cli-go:go-build   # CLI still compiles against the moved helpers
```

## Success Criteria

- [x] Three-axis types (`IsolationMethod`/`ProvisioningMethod`/`NetworkMethod`) + `Contribution` + `DefaultEgressAllowlist` + the four-guarantee `SandboxPolicy` exist in `shared-agent-go/pkg/types`; `SandboxConfig.Network NetworkMethod` (replacing `Network bool`) + `EgressAllowlist []string` added.
- [x] `pkg/sandbox/policy.go` classifies pinned-provisioning, host-tools-unreachable, egress-restricted (`EgressIsRestricted`), and kernel-isolation (`KernelIsolated`) as independent, table-tested guarantees; `EnforcePolicy(iso, prov, net, passthrough, runtime, pol)` returns a distinct named error per unmet guarantee and the curated matrix validity check lives here.
- [x] `shellQuote` removed (uses `shell.Quote`); env helpers in `shared-core-go/pkg/envutil`; `BuildAgentCommand`/`BuildProviderEnv` in `shared-agent-go/pkg/agentcmd`.
- [x] `shared-*-go` + `agent-cli-go` build/lint/test green; nothing imports the deleted symbols.

## Files Modified/Created

- `shared-agent-go/pkg/types/sandbox.go` — three-axis types (incl. `NetworkMethod`), `Contribution`, `DefaultEgressAllowlist`, `SandboxConfig.Network`/`EgressAllowlist`, four-guarantee policy.
- `shared-agent-go/pkg/sandbox/policy.go` (+ `policy_test.go`) — new; classifiers incl. `EgressIsRestricted`/`KernelIsolated` + extended `EnforcePolicy`.
- `shared-core-go/pkg/envutil/envutil.go` (+ test) — new.
- `shared-agent-go/pkg/agentcmd/agentcmd.go` (+ test) — new.
- `agent-cli-go/internal/sandboxutil/utils.go` — drop dup helpers, import from shared.

## Dependencies

None — this is the gating foundation.

## Estimated Duration

~1.5 days.

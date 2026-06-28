# Agent

- Version bump: the agent release line moves in lockstep — `agent-cli-go`'s `version` and its five `optionalDependencies`, the five `agent-cli-go-<platform>` binary packages, and `agent-cli-go-github`'s platform refs all move to the same version, plus a matching `## <version>` `CHANGELOG.md` entry from the conventional commits since the last release. A release without its `## <version>` CHANGELOG section makes `github-publish` fail or ship stale notes.

## Sandbox model

The sandbox is selected by **three orthogonal axes** plus a `hostPassthrough` knob — not a single fused method:

- **Isolation** `{none, bwrap, docker}` — how the process is confined. `bwrap` and **default-runc** docker are attack-surface reduction, **not** kernel trust boundaries.
- **Provision** `{none, nix, command}` — how tools reach PATH, independent of isolation. `nix` resolves a `flake.lock`/rev-pinned closure on the host and mounts **only the closure's requisites** (`nix path-info -r`) read-only — never the whole `/nix/store`, never the nix daemon socket. `command` runs an init-command list before the agent.
- **Network** `{host, none, proxy}` — egress, applied EXPLICITLY by the isolator. `host` = share host net, unrestricted egress (today's de-facto behavior, now an explicit opt-in; does **not** satisfy `RequireEgressRestricted`). `none` = no network (`--unshare-net` / `--network none`). `proxy` = egress ONLY via a host-side allowlist HTTP(S) proxy; link-local + metadata `169.254.169.254` + RFC1918 + loopback blocked. **Recommended default** for untrusted code that still needs the model API.
- `hostPassthrough` off (default) = deny-default (host tools off PATH and not bind-reachable). On = expose host/base-image tools as a fallback (forfeits host-tools-unreachable).
- Shared `DefaultEgressAllowlist` (provider API + common registries/git forges) seeds `proxy`; extend with `--egress-allow` (repeatable). Default `Network = proxy` when a proxy/allowlist is configured, else `host`.

## Pluggability

Isolators and provisioners are **plugins resolved by an explicit registry**, not a hardcoded switch — the core never names a concrete plugin:

- The CLI `internal/sandbox` package defines only the `Isolator`/`Provisioner` interfaces + a `Registry` (factory maps) + `Select(reg, req, pol)`. Each plugin declares its own guarantees (`Provisioner.Pinned()`, `Isolator.HidesHost()`/`KernelIsolated()`), and the shared `sandbox.EnforcePolicy(Capabilities, policy)` is method-agnostic — it checks computed capability booleans, never an enum.
- The composition root `internal/sandbox/plugins.DefaultRegistry()` is the ONLY place that imports the concrete plugin packages (`bwrap`/`docker`/`none`/`nixprov`). **Adding an isolator or provisioner is a new package plus one `Register` call there** — zero edits to `Select`, the registry, or the policy engine. Tests build their own minimal `Registry` (no global mutable state).
- The operator mirrors this with a `builder.ResolveToolchain` registry (`ToolchainType → Toolchain{Image, Pinned}`), so the controller and pod-hardening never name a concrete toolchain.

## Policy — four independently-requestable guarantees (fail CLOSED)

- `RequirePinnedProvision` — provisioning ∈ {nix, pinned image}, enforced at resolve via `--frozen`/`--no-write-lock-file` against a committed lock.
- `RequireHostToolsUnreachable` — host tools off PATH **and** not bind-reachable. Conditioned on closure-only store binds + NO host-`$HOME` bind + (docker) a pinned image. This is **not** "host unreachable" — filesystem reach and network egress are SEPARATE guarantees.
- `RequireEgressRestricted` — `Network ∈ {none, proxy}`; `host` does not qualify.
- `RequireKernelIsolation` — docker `--runtime runsc`/gVisor, or an operator pod with a sandboxed `runtimeClass` (gVisor/Kata/kata-cc). **Not** satisfied by bwrap or default runc.

**Threat model:** the agent runs UNTRUSTED, model-generated code under possible indirect prompt injection. bwrap and default-runc are attack-surface reduction, not kernel trust boundaries; "host tools unreachable" ≠ "host unreachable". The engine fails CLOSED — it refuses to run when a requested guarantee can't be established, never silently degrades.

## Provisioning core

- One `flake.lock`-pinned declarative core (`nix/agent-env.nix` + vendored `nix/mkAgentImage.nix`) feeds both surfaces. The **CLI** resolves it to a closure and bind-mounts the requisites; the **operator** bakes the SAME content-addressed store-path closure into a `dockerTools.streamLayeredImage` OCI image (identical store-path hashes from the same `flake.lock` — verify with `nix path-info -r`, **not** byte-identical layers). One image builder: `streamLayeredImage` (`maxLayers=100`, `created` defaults to the epoch — never `now`). `nix2container` is only an alternative if pushing on every small change.
- `mkAgentImage` is **vendored/adapted** from `nothingnesses/agent-images` (bus-factor 1): numeric uid-1000 hand-written passwd/group, `/workspace`, pre-created XDG dirs.
- `numtide/llm-agents.nix` is a `flake.lock`-pinned input — **packaging only; isolation out of scope**. Consumers pin via `flake.lock` alone. The binary-cache trusted key NAME is `niks3.numtide.com` (`niks3.numtide.com-1:DTx8wZduET09hRmMtKdQDxNNthLQETkc/yaX7M4qK0g=`); adding `cache.numtide.com` is a TRUST EXPANSION (numtide's CI/signing key builds the agent binaries) — not enabled by default.

## Operator path

- Provisions via a **nix-built OCI image** (no per-pod `nix profile install`); pods start by image pull. Untrusted pods default to a sandboxed `runtimeClassName` via the existing `DefaultRuntimeClassName`/`AllowedRuntimeClassNames` machinery (wires `RequireKernelIsolation`), bind a dedicated **zero-RBAC ServiceAccount** with `automountServiceAccountToken=false`, get default resource requests/limits + a namespace `LimitRange`/`ResourceQuota`, and always get a **default-deny egress `NetworkPolicy`** per `AgentRun` (mapped from `Network`; metadata/RFC1918/loopback blocked, FQDN-aware via Cilium `toFQDNs`/Squid as the upgrade). `readOnlyRootFilesystem=true` is reconciled with writable HOME/XDG `emptyDir`s + `fsGroup=1000`.

---
type: plan
has_subplans: false
parent_plan: plans/agent-sandbox-provisioning-axes.md
parallel_group: 4
status: pending
dependencies:
  plans: [shared-flake-and-resolver, cli-isolator-provisioner-core, cli-nix-provisioner, cli-remove-compose, operator-image-realizer]
  files:
    - packages/agent/AGENTS.md
    - packages/agent/agent-cli-go/README.md
    - packages/agent/agent-operator-go/AGENTS.md
    - packages/diagram/
    - packages/agent/agent-cli-go/CHANGELOG.md
skills_to_consult: [moon-guide, debugging-guide, pull-request-guide, code-review-guide, git-guide]
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 07 — Docs, end-to-end validation, and rollout

## Objective

Document the **three-axis** model (Isolation × Provisioning × Network) + the `hostPassthrough` knob + the **four-guarantee** policy + the per-surface realization, validate the full matrix (incl. the network axis) and the operator image end-to-end, keep the cross-project build green, and ship via PRs (protected `main`; releases via PR per the AGENTS.md rule).

## Tasks

1. **Docs / taxonomy.** Update `packages/agent/AGENTS.md` and the CLI/operator READMEs:
   - The **three orthogonal axes** — `IsolationMethod{none,bwrap,docker} × ProvisioningMethod{none,nix,command} × NetworkMethod{host,none,proxy}` — plus the `hostPassthrough` knob. `NetworkMethod` semantics: `host` = share host net, unrestricted egress (today's de-facto behavior, now an EXPLICIT opt-in; does NOT satisfy `RequireEgressRestricted`); `none` = no network (`--unshare-net` / `--network none`); `proxy` = egress ONLY via a host-side allowlist HTTP(S) proxy, all else blocked incl. link-local + metadata `169.254.169.254` + RFC1918 + loopback (RECOMMENDED DEFAULT for untrusted code that still needs the model API). Document the shared `DefaultEgressAllowlist` (provider API endpoints + common package registries/git forges), extensible via `--egress-allow`; default `Network = proxy` when a proxy/allowlist is available, else `host`; the deprecated `--sandbox` alias maps legacy methods to `Network=host` for one release.
   - The **four-guarantee** policy, independently requestable: `RequirePinnedProvisioning`, `RequireHostToolsUnreachable`, `RequireEgressRestricted` (Network ∈ {none, proxy}), `RequireKernelIsolation` (docker `--runtime runsc`/gVisor, or operator pod with a sandboxed `runtimeClass`; NOT satisfied by bwrap or default runc).
   - The one-paragraph **THREAT MODEL** (mirror the parent Overview): the agent runs UNTRUSTED, model-generated code under possible indirect prompt injection; bwrap and default-runc containers are ATTACK-SURFACE REDUCTION, not kernel trust boundaries; "host tools unreachable" ≠ "host unreachable" (network egress is a SEPARATE guarantee); fail CLOSED — refuse to run when a requested guarantee can't be established, never silently degrade.
   - **Reframe `IsolationHidesHost` docs:** `RequireHostToolsUnreachable` means host tools are off PATH AND not bind-reachable — it does NOT mean the host (or network) is unreachable; that is `RequireEgressRestricted` / `RequireKernelIsolation`.
   - **Operator path = "nix-built OCI image"** (not runtime in-sandbox), and document: the sandboxed-`runtimeClass` default for the untrusted path (via the existing `DefaultRuntimeClassName`/`AllowedRuntimeClassNames`, gVisor/Systrap, wires `RequireKernelIsolation`); `automountServiceAccountToken=false` + a dedicated **zero-RBAC ServiceAccount**; default resource requests/limits + a `LimitRange`/`ResourceQuota`; an always-emitted default-deny egress `NetworkPolicy` per `AgentRun` upgraded to **FQDN-aware egress** (Cilium `toFQDNs` + kube-dns `rules: dns`, or a Squid egress proxy) blocking metadata/RFC1918.
   - **Image wording:** the CLI and operator provision the **SAME content-addressed store-path closure** (identical store-path hashes from the same `flake.lock`) — NOT "byte-identical layers" (the CLI bind-mounts store directories; the image bakes them into TAR layers); verify via `nix path-info -r`. Commit to ONE image builder: `dockerTools.streamLayeredImage` (drop "or nix2container" from the design; mention nix2container only as an alternative). The image build **vendors/adapts** `nothingnesses/agent-images`' `lib/mkAgentImage.nix` (bus-factor 1) into a `streamLayeredImage` build.
   - **`llm-agents.nix` = "packaging only; isolation out of scope"** (do NOT say it "enforces no isolation"); consumers pin via `flake.lock` ALONE. Record the trusted binary-cache key NAME `niks3.numtide.com` (`niks3.numtide.com-1:DTx8wZduET09hRmMtKdQDxNNthLQETkc/yaX7M4qK0g=`), and call out that adding `cache.numtide.com` is a TRUST EXPANSION (numtide's CI/signing key builds the agent binaries).
   - Refresh the sandbox-isolation diagram in `packages/diagram/` — it gains the **network axis** alongside isolation × provisioning (and the four-guarantee policy) — if it depicts the old tiers.

2. **CLI matrix validation.** End-to-end check that every valid cell wraps/runs (use `MOON_DEBUG_WASM` / the run command outside a nix shell where relevant): `bwrap/docker/none × none/nix/command`; `hostPassthrough` on hides nothing / off hides host (bwrap); `none × nix` runs on host with the closure prepended; invalid cells error.
   - **Network axis cells.** Across `bwrap`/`docker` (and the operator): `Network=none` ⇒ egress **BLOCKED** (assert `--unshare-net` / `--network none` emitted EXPLICITLY — the regression guard so collapsing the old `nix`/`nixflake` tiers can't silently drop net isolation); `Network=proxy` ⇒ **allowlist-only** egress (a host in `DefaultEgressAllowlist` reachable via the proxy; metadata `169.254.169.254` + RFC1918 + loopback + an off-allowlist host DENIED); `Network=host` ⇒ unrestricted (and the deprecated `--sandbox` alias resolves to it).
   - **Hardening flags present.** bwrap deny-default: sandbox-local HOME, `--dev /dev` (not `--dev-bind`), `--clearenv` + `--setenv` allowlist, **NO `--bind $HOME $HOME`**, only workspace(rw) + RepoDir(ro) + curated `UserConfigPaths`. docker defaults: `--security-opt=no-new-privileges`, `--cap-drop ALL`, default seccomp (never `seccomp=unconfined`), `apparmor=docker-default`, `--read-only` + `--tmpfs /tmp:rw,noexec,nosuid`, `--pids-limit`/`--memory`/`--cpus`, workdir as the ONLY writable bind.
   - **Closure-only binds.** `nix` provisioning binds ONLY the resolved closure's requisites (`nix path-info -r` of the descriptor), NOT the whole `/nix/store`; assert host tools outside the closure are absent under `RequireHostToolsUnreachable`.
   - **Lock fail-closed.** Resolve with `--frozen` / `--no-write-lock-file`; a dirty/missing lock is REJECTED (fail closed) — assert the run refuses rather than silently re-pinning.

3. **Policy validation (four guarantees).** Table-driven + a couple of live runs:
   - `RequirePinnedProvisioning` accepts `nix` (any cell) and the operator's pinned image; rejects `command`/`none`; enforced at resolve via `--frozen` + a committed lock.
   - `RequireHostToolsUnreachable` accepts `bwrap`-off / `docker` / pod (CONDITIONED on closure-only store binds + NO host-`$HOME` bind + (docker) a pinned image); rejects `none × *` and `bwrap`+passthrough.
   - `RequireEgressRestricted` accepts `Network ∈ {none, proxy}`; rejects `Network=host`.
   - `RequireKernelIsolation` accepts docker `--runtime runsc`/gVisor + the operator's sandboxed `runtimeClass`; rejects `bwrap` and default-runc docker.
   - **Fail-closed:** any requested-but-unestablishable guarantee refuses the run (no silent degrade).

4. **Operator e2e.** The `streamLayeredImage` image builds from the flake at a pinned rev (`created=epoch`, never `now`; a rebuild-digest-equality check passes), an `AgentRun` with a `nix` toolchain runs a pod that starts by image pull (no init container) under the sandboxed `runtimeClass` (`RequireKernelIsolation` wired), with `automountServiceAccountToken=false` + the zero-RBAC SA, default limits + a `LimitRange`/`ResourceQuota`, a default-deny → FQDN-aware egress `NetworkPolicy` (metadata/RFC1918 blocked), and writable `emptyDir` HOME/XDG_* over `readOnlyRootFilesystem=true` (e2e asserts the agent can write config/cache/state). The agent has the pinned tools, and the toolset is the **SAME content-addressed closure** as the CLI for the same lock — verify by comparing `nix path-info -r`, NOT layer bytes.

5. **Cross-project build + rollout.** `moon ci` green across `config → shared → agent` (shared-core-go, shared-agent-go, agent-cli-go, agent-operator-go). Version-bump per the per-dir `AGENTS.md` (agent-cli-go via `npx moon-version-bump`; the operator image; shared libs); update CHANGELOGs; ship via **PRs** (protected `main`; a `version packages` PR for the release).

6. **Dead-code sweep.** Confirm the retired channel-pinned `nixenv` renderer, the old `nix:`/`nixflake:` `--image` prefixes, the replaced `Network bool` field, the blanket `--dev-bind`, the nix **daemon-socket bind**, all `compose` references, and any leftover "or nix2container"/"byte-identical" wording are gone: `grep -rin 'fetchTarball\|nix:\|nixflake:\|compose\|RequirePinnedToolchain\|Network bool\|--dev-bind\|nix-daemon\|daemon.socket\|nix2container\|byte-identical' packages/agent packages/shared` is clean (modulo the deliberate nix2container "alternative" mention in docs).

## Validation Steps

```bash
npx moon ci :ci-check          # whole workspace
grep -rin 'compose\|RequirePinnedToolchain\|fetchTarball\|Network bool\|--dev-bind\|nix-daemon\|nix2container\|byte-identical' packages/agent packages/shared || echo clean
# CLI matrix + network-axis + operator e2e per tasks 2-4
```

## Success Criteria

- [ ] AGENTS.md / READMEs / diagram document the **three axes** (Isolation × Provisioning × Network) + `hostPassthrough`, the **four-guarantee** policy, the THREAT MODEL, the reframed "host tools unreachable ≠ host unreachable", and the operator = nix-built OCI image (sandboxed `runtimeClass` + zero-RBAC SA + limits + FQDN egress).
- [ ] Docs record the image as the **SAME content-addressed closure** (not byte-identical), the single `streamLayeredImage` builder, the vendored/adapted `mkAgentImage`, and `llm-agents.nix` = "packaging only; isolation out of scope" + the `cache.numtide.com` trust-expansion + the `niks3.numtide.com` key.
- [ ] Every valid CLI cell verified; invalid cells rejected; the **four** policy guarantees behave correctly (incl. `none × nix`); network-axis cells pass (egress BLOCKED under `none`, allowlist-only under `proxy`, across bwrap/docker/operator) with `--unshare-net`/`--network none` emitted EXPLICITLY.
- [ ] bwrap/docker hardening flags present (no `--bind $HOME $HOME`, `--dev /dev`, `--clearenv`+allowlist; docker `no-new-privileges`/`cap-drop ALL`/default-seccomp/`--read-only`/limits); `nix` binds **closure requisites only**; lock fail-closed (`--frozen` rejects a dirty/missing lock).
- [ ] Operator image builds (`created=epoch`, rebuild-digest-equal) + a pod runs the pinned toolset via image pull under a sandboxed `runtimeClass`; SAME content-addressed closure as the CLI (verified by `nix path-info -r`); writable HOME/XDG over `readOnlyRootFilesystem`.
- [ ] `moon ci` green across all four projects; version bumps + CHANGELOGs done; shipped via PRs.
- [ ] Dead-code sweep clean (no channel renderer, `nix:`/`nixflake:` prefixes, compose, `RequirePinnedToolchain`, `Network bool`, `--dev-bind`, nix daemon-socket bind, or leftover "or nix2container"/byte-identical wording).

## Files Modified/Created

- `packages/agent/AGENTS.md`, `agent-cli-go/README.md`, `agent-operator-go/AGENTS.md` — model docs (three axes + four-guarantee policy + threat model + image/cache-trust wording).
- `packages/diagram/` — refresh the sandbox-isolation diagram; it gains the **network axis** (Isolation × Provisioning × Network) and the four-guarantee policy.
- CHANGELOGs + `.moon/` version bumps.

## Dependencies

All of `02`–`06` (the model, both realizers, and compose removal must be in).

## Estimated Duration

~1 day.

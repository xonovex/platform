---
type: plan
has_subplans: false
parent_plan: plans/agent-sandbox-provisioning-axes.md
parallel_group: 3
status: pending
dependencies:
  plans: [shared-types-policy-dedups, shared-flake-and-resolver]
  files:
    - packages/agent/agent-operator-go/api/v1alpha1/agentrun_types.go
    - packages/agent/agent-operator-go/internal/builder/toolchain.go
    - packages/agent/agent-operator-go/internal/builder/toolchain_nix.go
    - packages/agent/agent-operator-go/internal/builder/container.go
    - packages/agent/agent-operator-go/internal/builder/security.go
    - packages/agent/agent-operator-go/internal/builder/networkpolicy.go
    - packages/agent/agent-operator-go/internal/webhook/agentrun_webhook.go
    - packages/agent/agent-operator-go/internal/builder/harness_claude.go
    - nix/agent-env.nix
skills_to_consult: [kubernetes-guide, docker-guide, general-fp-guide, debugging-guide]
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 06 — Operator realizer: nix-built OCI image (replace the per-pod install)

## Objective

Converge the operator onto the same pinned flake by **building the agent image with nix ahead of time** (`streamLayeredImage` from `nix/agent-env.nix`, via the **vendored `mkAgentImage`** of subplan 02) and **referencing it as the pod image** — replacing `toolchain_nix.go`'s per-pod `nix profile install` cold-start. The CLI and operator provision the **same content-addressed store-path closure** (identical store-path hashes from the same `flake.lock`); the pinned image satisfies `RequirePinnedProvisioning`. Adopt the agent-images uid-1000/`/workspace`/XDG layout, **and carry the operator-side hardening** that makes the four policy guarantees real on k8s: a sandboxed `runtimeClassName` (gVisor/Systrap) for `RequireKernelIsolation`, a per-`AgentRun` default-deny egress `NetworkPolicy` mapped from `NetworkMethod` for `RequireEgressRestricted`, a zero-RBAC ServiceAccount + `automountServiceAccountToken=false`, resource limits + a `LimitRange`/`ResourceQuota`, and writable HOME/XDG `emptyDir`s reconciled with `readOnlyRootFilesystem=true`. Threat model: the agent runs **untrusted, model-generated code** under possible prompt injection (stated in the parent Overview) — fail **closed** when a requested guarantee can't be established.

## Tasks

1. **Extend `NixSpec`** (`agentrun_types.go:106-114`) with a **pin** (`NixpkgsRev string`) and a **source** (`Packages []string` already exists; add an optional `FlakeRef`/`Shell` for project-flake builds), and an **`Image string`** field for the resolved/pre-built image reference. Webhook (`agentrun_webhook.go:73-93`): validate the rev + mutual exclusion of packages vs flake-ref.

2. **Add the image-build pipeline** (a moon task on a new `agent-cli-image`/`agent-image` project, or extend `agent-cli-go-github`): `nix build .#agentImage --argstr rev <rev> --arg packages '[...]'` → stream + push to the registry (`./result | skopeo copy docker-archive:/dev/stdin docker://…`, or `nix run .#agentImage.copyToRegistry`). Commit to **ONE builder: `dockerTools.streamLayeredImage`** — it never realizes a store tarball (the layers stream straight to the registry). Drop the per-pod/`nix2container` optionality from the committed design; mention `nix2container` only as an alternative IF we end up pushing on every small change (note its `maxLayers` defaults to 1). The image comes from `nix/agent-env.nix` (subplan 02, the **vendored `mkAgentImage`**): `User=1000:1000`, `WorkingDir=/workspace`, hand-written `/etc/passwd`+`/etc/group` for numeric uid 1000, pre-created `$HOME/.config|.cache|.local|.local/state` + `/workspace`. **`maxLayers = 100`** set deliberately (the safe conventional default; 128 sits at overlay2's modern ceiling with zero extension headroom). **Reproducibility invariant**: never set image `created=now` (`dockerTools` defaults to the epoch) — add a **rebuild-digest-equality CI check** (rebuild from the same `flake.lock` → identical digest). Output: a digest-pinned image ref.

3. **Replace the init-container realizer** in `toolchain_nix.go:31-99`: delete the `nix-env` emptyDir (`:51-67`), the `nixos/nix` init container (`:31-49`), and the `cp -a /nix/.` + `nix profile install` `installScript` (`:87-99`). The `NixToolchain` `ToolchainContributor` now contributes **nothing at runtime** — instead the builder sets the agent container's **image** to the nix-built image. Update `Toolchains()` (`toolchain.go:18-23`) / `BuildMainContainers` (`container.go:42-75`) so a `nix` toolchain selects the image rather than an init container + volume.

4. **Adopt the runtime layout + reconcile `readOnlyRootFilesystem=true` with a writable HOME/XDG**: the image already runs as non-root uid 1000 with `/workspace` (baked in subplan 02) — set the pod `securityContext`/`WorkingDir` accordingly. The container's rootfs stays read-only (already set), so mount **`emptyDir`s for HOME + `XDG_CONFIG_HOME`/`XDG_CACHE_HOME`/`XDG_DATA_HOME`/`XDG_STATE_HOME`** (paths **non-overlapping** with the image's baked content so the mounts don't shadow it) and set **`fsGroup=1000`** so uid 1000 owns them. Keep the existing PVC/worktree mounts under the pre-created XDG dirs so no root-owned parents appear. e2e **assert the agent can write config/cache/state** at runtime.

5. **Policy mapping + pod hardening** (the four guarantees made real on k8s — fail **closed** when one can't be met):
   - `RequirePinnedProvisioning` — satisfied by the `flake.lock`-pinned image (not a runtime install). Update the operator's pinned-guarantee check (it no longer needs the init container to prove pinning).
   - `RequireHostToolsUnreachable` — the pod/container mount + PID namespace; the image carries only the resolved closure (no host store, no host `$HOME`).
   - `RequireKernelIsolation` — **default the untrusted path to a sandboxed `runtimeClassName`** (gVisor/Systrap, **no KVM needed**) via the **existing `DefaultRuntimeClassName`/`AllowedRuntimeClassNames`** machinery (`security.go`; the gVisor/Kata/CoCo e2e suites already exist). Wire `RequireKernelIsolation` as **satisfied only by a sandboxed runtimeClass** — not by default `runc`.
   - `RequireEgressRestricted` — satisfied by the per-`AgentRun` egress `NetworkPolicy` (Task 6) when `Network ∈ {none, proxy}`.
   - **ServiceAccount**: bind agent pods to a **dedicated zero-RBAC ServiceAccount** with **`automountServiceAccountToken=false`** — the agent never calls the K8s API.
   - **Resource bounds**: emit default resource **requests/limits** on the agent container + a namespace **`LimitRange`/`ResourceQuota`** (node-DoS bound).

6. **Network: per-`AgentRun` default-deny egress + `NetworkMethod` mapping**: **ALWAYS emit a default-deny egress `NetworkPolicy`** per `AgentRun` (no implicit open egress). Map `Network NetworkMethod` (from the shared `SandboxConfig`):
   - `host` → unrestricted egress (no egress restriction / today's de-facto behavior, now an **explicit** opt-in). Does **not** satisfy `RequireEgressRestricted`.
   - `none` → deny-all egress (allow only what the pod needs to start). Satisfies `RequireEgressRestricted`.
   - `proxy` → allow egress **only** to the allowlist (`EgressAllowlist`/`DefaultEgressAllowlist`) — start L3/L4 (kube-dns + the proxy/CIDRs), with an **upgrade path to FQDN-aware egress** (Cilium `toFQDNs` paired with kube-dns `rules: dns`, or a Squid egress proxy). **Block metadata (169.254.169.254) + link-local + RFC1918 + loopback.** Satisfies `RequireEgressRestricted`.

7. **Adopt shared `BuildAgentCommand`/`BuildProviderEnv`** (from subplan 01's `shared-agent-go/pkg/agentcmd`), replacing the operator's inline re-implementations (`harness_claude.go:13`, `harness_opencode.go:13`, `resolver/provider.go:98`).

8. **Tests + e2e**: the image builds from the flake at a pinned rev; the CLI and operator resolve to the **same store-path closure** (compare `nix path-info -r`, not layer bytes); the rebuild-digest-equality check holds; an `AgentRun` with a `nix` toolchain produces a Job whose agent container uses the built image with **no** init container + **no** `nix-env` volume; pods start by image pull; the pod runs under a **sandboxed `runtimeClassName`**, a **zero-RBAC ServiceAccount** (`automountServiceAccountToken=false`), resource limits + a `LimitRange`/`ResourceQuota`, and a **default-deny egress `NetworkPolicy`**; the agent can **write config/cache/state** into the HOME/XDG `emptyDir`s despite `readOnlyRootFilesystem=true`; the webhook validates the new `NixSpec` fields; the deepcopy/CRD manifests are regenerated (note the controller-gen Go 1.25+ caveat in the operator's AGENTS.md — maintain manually if needed).

## Validation Steps

```bash
npx moon run agent-operator-go:go-build agent-operator-go:go-test agent-operator-go:go-lint
nix build .#agentImage --argstr rev <rev>     # the image realizes from the flake (streamLayeredImage)
nix path-info -r .#agentEnv                    # same closure as the CLI provisions — compare store-path hashes, not layer bytes
nix build .#agentImage && ./result | sha256sum # rebuild-digest equality (created != now; epoch default)
# e2e (kind): apply an AgentRun with a nix toolchain; assert the Job has the built image, no init container,
#   a sandboxed runtimeClassName, a zero-RBAC SA (automountServiceAccountToken=false), resource limits,
#   a default-deny egress NetworkPolicy, and that the agent can write config/cache/state (HOME/XDG emptyDirs)
npx moon run agent-operator-go:test
```

## Success Criteria

- [ ] `NixSpec` carries a pin + source + image ref; webhook validates them.
- [ ] The agent image is `nix`-built from `nix/agent-env.nix` via the **vendored `mkAgentImage`** (`dockerTools.streamLayeredImage`, `maxLayers = 100`) and pushed digest-pinned; the CLI and operator provision the **same content-addressed store-path closure** (identical store-path hashes from the same `flake.lock`) — verified by comparing `nix path-info -r`, **not** layer bytes; `created` is never `now` and the **rebuild-digest-equality** check holds.
- [ ] `toolchain_nix.go`'s emptyDir + `nixos/nix` init container + `nix profile install` are **gone**; the `nix` toolchain selects the pod image instead.
- [ ] Pods run non-root uid 1000 at `/workspace`; start by image pull (no per-pod nix); `RequirePinnedProvisioning` satisfied by the pinned image.
- [ ] The untrusted path defaults to a **sandboxed `runtimeClassName`** (gVisor/Systrap) via `DefaultRuntimeClassName`/`AllowedRuntimeClassNames`; `RequireKernelIsolation` is satisfied by the sandboxed runtimeClass (**not** default `runc`).
- [ ] Agent pods bind a **dedicated zero-RBAC ServiceAccount** with `automountServiceAccountToken=false`.
- [ ] Default resource **requests/limits** + a namespace **`LimitRange`/`ResourceQuota`** bound node-DoS.
- [ ] `readOnlyRootFilesystem=true` is reconciled with writable HOME + `XDG_*` `emptyDir`s (non-overlapping with baked content) + `fsGroup=1000`; e2e asserts the agent writes config/cache/state.
- [ ] A **default-deny egress `NetworkPolicy`** is emitted per `AgentRun`; `Network ∈ {none, proxy}` satisfies `RequireEgressRestricted` (proxy allows only the allowlist; metadata/RFC1918/loopback blocked); FQDN-aware upgrade path documented.
- [ ] Operator uses shared `BuildAgentCommand`/`BuildProviderEnv`; build/lint/test/e2e green.

## Files Modified/Created

- `api/v1alpha1/agentrun_types.go` — `NixSpec` pin/source/image; `Network NetworkMethod` + `EgressAllowlist`; regenerate deepcopy/CRDs.
- `internal/builder/toolchain_nix.go` — replace init-container install with image selection.
- `internal/builder/{toolchain.go,container.go}` — wire image selection; resource requests/limits; HOME/`XDG_*` `emptyDir` mounts + `fsGroup=1000`; zero-RBAC ServiceAccount + `automountServiceAccountToken=false`.
- `internal/builder/security.go` — default the untrusted path to a sandboxed `runtimeClassName` via `DefaultRuntimeClassName`/`AllowedRuntimeClassNames`; wire `RequireKernelIsolation`.
- `internal/builder/networkpolicy.go` — extend the existing `BuildNetworkPolicy` to default-deny egress per `AgentRun`, map `NetworkMethod`, block metadata/RFC1918/loopback, with an FQDN-aware upgrade path.
- RBAC + `LimitRange`/`ResourceQuota` manifests (new) under `config/` (zero-RBAC SA, node-DoS bound).
- `internal/webhook/agentrun_webhook.go` — validate new fields.
- `internal/builder/harness_*.go`, `internal/resolver/provider.go` — adopt shared `agentcmd`.
- image-build moon task (new) + `nix/agent-env.nix` (consumed; vendored `mkAgentImage`, rebuild-digest CI check).

## Dependencies

`01` (shared types/policy/`agentcmd`), `02` (the flake's `streamLayeredImage` target + `NixSource`). Independent of the CLI refactor — runs parallel with group 2.

## Estimated Duration

~3 days.

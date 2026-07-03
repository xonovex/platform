---
type: plan
has_subplans: true
status: pending-approval
feature: agent-sandbox-provisioning-axes
dependencies:
  plans: []
proposed_subplans:
  - shared-types-policy-dedups
  - shared-flake-and-resolver
  - cli-isolator-provisioner-core
  - cli-nix-provisioner
  - cli-remove-compose
  - operator-image-realizer
  - docs-validation-rollout
parallel_groups:
  - group: 1
    plans: [shared-types-policy-dedups, shared-flake-and-resolver]
    note: "Shared foundation in shared-core-go/shared-agent-go + the pinned flake. Gates everything. shared-flake-and-resolver consumes the EnvSpec/policy types from shared-types-policy-dedups, so it lands second."
  - group: 2
    plans: [cli-isolator-provisioner-core, cli-nix-provisioner, cli-remove-compose]
    depends_on: [1]
    note: "agent-cli-go refactor. All three edit internal/sandbox/* and cmd/run.go, so they SERIALIZE on that package (any order), not in parallel worktrees. 03 now ALSO carries the network axis (NetworkMethod application) + the bwrap/docker hardening (deny-default HOME, --dev /dev, --clearenv, no-new-privileges/cap-drop/read-only/limits), so it grows beyond a pure split."
  - group: 3
    plans: [operator-image-realizer]
    depends_on: [1]
    note: "agent-operator-go realizer. Independent of the CLI refactor — runs in PARALLEL with group 2 once the shared foundation lands. 06 now ALSO carries operator hardening (sandboxed runtimeClass default, zero-RBAC SA, resource limits/LimitRange, readOnlyRootFilesystem-vs-writable-XDG reconcile, FQDN-aware default-deny egress)."
  - group: 4
    plans: [docs-validation-rollout]
    depends_on: [2, 3]
    note: "Cross-cutting: taxonomy/AGENTS.md docs, end-to-end validation of the matrix + operator image, moon build green. After the CLI and operator both land."
skills_to_consult:
  - general-fp-guide        # the Go refactor: pure Provisioner contributions, explicit context, composition over the tier switch
  - docker-guide            # the docker isolator + the operator's dockerTools/nix2container OCI image
  - kubernetes-guide        # the operator PodSpec realizer (image pull vs init container)
  - shell-scripting-guide   # the command provisioner + the host-resolve helper + install scripts
  - moon-guide              # cross-project tasks/build (config -> shared -> agent), CI
  - debugging-guide         # determinism: did the right closure mount, did the policy guarantee actually engage
  - git-guide               # conventional commits; cross-repo coordination
  - pull-request-guide      # protected main; sizing/splitting per wave
  - code-review-guide       # xonovex ships via PRs (Conventional Comments)
research_sources:
  documentation:
    - packages/shared/shared-agent-go/pkg/types/sandbox.go        # SandboxMethod/SandboxConfig/SandboxPolicy/SandboxExecutor — the FUSED axes (the thing this plan splits)
    - packages/agent/agent-cli-go/internal/sandbox/registry.go    # GetExecutor/SelectExecutor (16-79), tierIsolation/enforcePinnedToolchain (53-95)
    - packages/agent/agent-cli-go/internal/sandbox/{none,bwrap,docker,compose,nix,nixflake}/  # the six fused tiers; bwrap skeleton ~triplicated in bwrap/nix/nixflake
    - packages/agent/agent-cli-go/internal/nixenv/{render,resolve,build}.go  # host buildEnv, channel-pinned via fetchTarball — to be RETIRED
    - packages/agent/agent-cli-go/internal/sandboxutil/utils.go   # WrapWithInitCommands (161, dormant `command` hook), env helpers, BuildAgentCommand/BuildProviderEnv, shellQuote (dup of shell.Quote)
    - packages/shared/shared-agent-go/pkg/nix/nix.go              # Pins/PackageSets/ExpandPackageSets — already shared, consumed by both binaries
    - packages/agent/agent-operator-go/internal/builder/toolchain.go      # ToolchainContributor interface (the operator's clean seam)
    - packages/agent/agent-operator-go/internal/builder/toolchain_nix.go  # init-container `nix profile install` into a 10Gi emptyDir — to be REPLACED
    - packages/agent/agent-operator-go/api/v1alpha1/agentrun_types.go     # NixSpec/ToolchainSpec (no pin, no flake source today)
    - https://github.com/numtide/llm-agents.nix                   # agent packaging + flake.lock + hashes.json pin + binary cache (cache.numtide.com)
    - https://github.com/nothingnesses/agent-images               # dockerTools layered-image pattern (lib/mkAgentImage.nix), uid-1000/workspace/XDG layout
    - https://github.com/nlewo/nix2container                      # streamed/incremental OCI layers, skip-already-pushed
    - https://ryantm.github.io/nixpkgs/builders/images/dockertools/  # dockerTools.streamLayeredImage (maxLayers)
    - https://ianthehenry.com/posts/how-to-learn-nix/saving-your-shell/  # gc-root via inputDerivation (mid-run GC eviction)
    - https://www.tweag.io/blog/2020-06-25-eval-cache/            # nix eval cache (cold-start)
    - https://www.ranti.dev/blog/securing-ai-agents-with-nix-and-bubblewrap  # nix + bubblewrap agent sandboxing
    - plans/agent-sandbox-provisioning-axes/research-best-practices-audit.md  # best-practices audit: the three-axis canon, the new policy guarantees, bwrap/docker/operator hardening, image content-closure correction
  versions:
    go: "1.26 (agent-cli-go, agent-operator-go, shared-core-go, shared-agent-go)"
    nixpkgs: "pinned by flake.lock to a REV (not a channel branch); resolve to a content-pinned closure"
    llm_agents_nix: "added as a flake input pinned to a rev; agents substituted from cache.numtide.com"
    image_builder: "nixpkgs dockerTools.streamLayeredImage (committed builder; maxLayers set deliberately, 100); nlewo/nix2container only as an alternative if pushing on every small change (its maxLayers defaults to 1)"
    moon: "v2.x (existing workspace)"
---

# Agent Sandbox: Provisioning as an Axis Orthogonal to Isolation

## Overview

The agent sandbox in `agent-cli-go` fuses several concerns into one `SandboxMethod` enum — **how** a process is isolated, **how** its tools are provisioned, and (implicitly) **how** its network egress is constrained (`nix`/`nixflake` are really *bwrap isolation + nix provisioning + `--unshare-net`* welded together). This plan splits them into **three orthogonal axes** — **isolation** ∈ `{none, bwrap, docker}`, **provisioning** ∈ `{none, nix, command}`, **network** ∈ `{host, none, proxy}` — plus a `hostPassthrough` knob; unifies the nix tiers behind **one `flake.lock`-pinned provisioner** that *resolves to a content-pinned closure mounted read-only into the sandbox*, **removes the compose tier**, and converges the `agent-operator-go` K8s operator onto the *same pinned flake* via a **nix-built OCI image** — with the shared declarative core hoisted into `shared-*-go` so both binaries reuse it. The same pass **hardens** the bwrap and docker isolators to deny-default (sandbox-local HOME, minimal `/dev`, `--clearenv`, no-new-privileges/cap-drop/read-only/limits), binds **only the resolved closure** (not the whole `/nix/store`), and enforces the committed lock **at resolve time** (fail closed).

**Threat model.** The agent runs **untrusted, model-generated code** under possible indirect prompt injection. `bwrap` and **default-runc** containers are **attack-surface reduction, not kernel trust boundaries** (kernel isolation is a separate, opt-in guarantee — gVisor/Kata). "Host tools unreachable" is **not** "host unreachable": filesystem reach and network egress are **separate** guarantees. The design **fails closed** — it refuses to run when a requested guarantee can't be established, never silently degrades.

## Goals

- Make **provisioning** and **network** orthogonal axes to isolation: `Isolator{none,bwrap,docker} × Provisioner{none,nix,command} × Network{host,none,proxy}` — the clean three-axis space the current fused enum can't express.
- **Unify** the `nix`/`nixflake` tiers into one provisioner with two sources (rev-pinned package list \| project flake), pinned by `flake.lock`/rev (retire the moving channel pin).
- Provision by **resolving with nix → a content-pinned closure, then mounting it read-only** into the sandbox (no in-sandbox nix daemon) — identical for `bwrap` and `docker`.
- **Bind only the resolved closure's requisites** (`nix path-info -r` of the descriptor), **not** the whole world-readable `/nix/store` (alternative: `nix copy` to a per-sandbox store) — this is what makes `RequireHostToolsUnreachable` accurate.
- **Enforce the committed lock at resolve time** (`--no-write-lock-file` / `--frozen`; reject a dirty/missing lock, **fail closed**) — otherwise the pin is defeated at the moment of resolution.
- Add a **network-egress axis** `NetworkMethod{host,none,proxy}`: `host` = explicit shared-host-network opt-in (today's de-facto behavior); `none` = no network (`--unshare-net` / `--network none`); `proxy` = egress **only** via a host-side allowlist HTTP(S) proxy (metadata `169.254.169.254` + link-local + RFC1918 + loopback denied). Ship a shared `DefaultEgressAllowlist` (provider API endpoints + common package registries/git forges), extensible via repeatable `--egress-allow`; **default `Network = proxy`** when a proxy/allowlist is available, else `host`. The isolator MUST apply the chosen method **explicitly** (regression guard against silently dropping today's `--unshare-net`).
- Add a `command` provisioner (single init-command list) by **activating the already-wired-but-dormant** `WrapWithInitCommands` hook.
- Add a `hostPassthrough` knob that **unifies today's leaky `bwrap` and deny-default `nix`/`nixflake` tiers** into one isolator + a toggle.
- **Harden** the bwrap and docker isolators to **deny-default**: sandbox-local HOME (drop the blanket `--bind $HOME $HOME`; bind only workspace rw + RepoDir ro + curated `sandbox.UserConfigPaths` ro), `--dev /dev` (drop `--dev-bind /dev /dev`), `--clearenv` + an explicit `--setenv` allowlist; docker gets `--security-opt=no-new-privileges` + `--cap-drop ALL` + default seccomp/apparmor + `--read-only` + tmpfs `/tmp` + `--pids-limit`/`--memory`/`--cpus` as **defaults**, not opt-in.
- **Decouple + extend the policy** into four independently-requestable guarantees: `RequirePinnedProvisioning`, `RequireHostToolsUnreachable`, `RequireEgressRestricted` (`Network ∈ {none, proxy}`), and `RequireKernelIsolation` (docker `--runtime runsc`/gVisor or operator sandboxed `runtimeClass` — **not** satisfied by bwrap or default runc).
- **Converge the operator** onto the same pinned flake via a **build-time `streamLayeredImage`** (replacing its per-pod `nix profile install` cold-start), and **harden the untrusted pod path**: default to a sandboxed `runtimeClass` (gVisor/Systrap, no KVM) via the existing `DefaultRuntimeClassName`/`AllowedRuntimeClassNames` machinery (wires `RequireKernelIsolation`); `automountServiceAccountToken=false` + a dedicated **zero-RBAC ServiceAccount**; default resource requests/limits + a `LimitRange`/`ResourceQuota`; reconcile `readOnlyRootFilesystem=true` with writable HOME/XDG via `emptyDir` mounts + `fsGroup=1000`; **always emit a default-deny egress `NetworkPolicy`** per `AgentRun`, upgraded to **FQDN-aware** egress (Cilium `toFQDNs` + kube-dns `rules: dns`, or a Squid proxy) that blocks metadata/RFC1918.
- **Adopt off-the-shelf**: `numtide/llm-agents.nix` (agent packaging + binary cache) and `nothingnesses/agent-images` (operator image pattern).
- **Remove the compose isolator** entirely (unused/deprecated code).

## Current State

**`agent-cli-go`** (`internal/sandbox/*`, Go 1.26):
- One flat `SandboxMethod` enum (`none/bwrap/docker/compose/nix/nixflake`) fuses isolation + provisioning; `SandboxExecutor.Execute` does both.
- The bwrap namespace/bind skeleton is **~triplicated** across `bwrap.go`/`nix.go`/`nixflake.go`, differing only in the env source (host PATH vs `/env` closure vs `nix develop`).
- **Today's plain `bwrap` tier leaks**: a blanket `--bind $HOME $HOME` (whole home exposed), `--dev-bind /dev /dev` (exposes `/dev/sda`, `/dev/mem`, input devices), and host `$PATH` bleed-through — attack surface, not a deny-default sandbox.
- The **docker isolator emits ZERO hardening flags** (no `no-new-privileges`, no `cap-drop`, no read-only rootfs, no resource limits) and **binds the whole `$HOME` rw** — strictly weaker than even the leaky bwrap tier.
- The `nix` tier nix-builds a `buildEnv` from a package list, **channel-pinned via `fetchTarball`** of a moving branch (a real reproducibility gap); the **`nixflake` tier binds the host nix DAEMON SOCKET through the sandbox boundary** to run `nix develop` inside bwrap — the precise anti-pattern this plan removes (resolve-on-host-then-mount instead).
- **Network egress is fused/implicit**: `nix`/`nixflake` `--unshare-net` today, the rest share the host network — there is **no explicit network axis**, so collapsing the tiers risks silently dropping that isolation.
- `SandboxInitCommands` + `WrapWithInitCommands` (`sandboxutil/utils.go:161`) is the **dormant `command` axis** — consumed by every tier, never assigned, no flag.
- `enforcePinnedToolchain` (`registry.go:83`) bundles "pinned source" + "host-tools-unreachable" into one guarantee keyed off the fused method.

**`agent-operator-go`** (K8s operator): already has a clean `ToolchainContributor` seam, but its `NixToolchain` provisions via a `nixos/nix` **init container** running `nix profile install` into a 10Gi emptyDir — **worst-case cold-start per pod** (whole-closure substitution, cold eval cache, no cross-pod dedup); `NixSpec` has **no pin and no flake source**. The pod security context (`security.go`) and `networkpolicy.go` are already **restricted-grade + default-deny** (`readOnlyRootFilesystem=true`, dropped caps, no-privilege-escalation), but the **pod path defaults to runc** — **no kernel isolation** for the untrusted workload, and no sandboxed `runtimeClass` is wired despite the existing `DefaultRuntimeClassName`/`AllowedRuntimeClassNames` machinery.

**Already shared** (`shared-agent-go/pkg/nix`, `pkg/sandbox`, `pkg/types`): the package-set/pin definitions, `DefaultContainerImage`, and the `SandboxExecutor`/policy *types* — consumed by both binaries. Only the host-bound *implementations* are CLI-internal.

## Research Findings

A 2026 adversarial study (workflow, grounded against the two named repos) **refined** the naive "in-sandbox `nix develop` everywhere" idea:

- **CLI — confirmed.** Runtime nix on a host with a durable shared `/nix/store` is correct: cold-start collapses to flake evaluation, bwrap gives deny-default isolation, no image/registry. Two mandatory additions: a **GC-root** (new-style `nix shell`/`develop` results aren't auto-gc-rooted and can be evicted mid-run) and a **committed `flake.lock` + warm binary cache**.
- **Operator — overturned.** A pod has no host `/nix/store`, so in-pod runtime nix is the *worst* path; the robust unit is a **prebuilt nix-built OCI image** (`dockerTools.streamLayeredImage`, `maxLayers` set deliberately — 100 is the safe conventional choice) from the *same flake* → node-cached kubelet image pull, **the same content-addressed store-path closure as the CLI** (identical store-path hashes from the same `flake.lock`; verify with `nix path-info -r`, **not** layer bytes — the CLI bind-mounts store *directories* while the image bakes them into TAR *layers*), no nix daemon in the workload.
- **Realization refinement (final):** the nix provisioner **resolves with nix → a content-pinned closure** (CLI: host nix at run time; operator: nix at build time into image layers), and the isolator **mounts that closure read-only + sets PATH/env**. The sandbox never runs nix itself — which avoids exposing the host nix **daemon socket** through a docker/pod boundary (the trap of "run `nix develop` inside the container").
- **Adopt:** `numtide/llm-agents.nix` as a pinned flake input for agent packaging — it is **packaging only; isolation is out of scope** (it neither provides nor claims a sandbox). Consumers pin via **`flake.lock` alone** (the upstream `hashes.json`/`update.py` is upstream bookkeeping, not the consumer pin surface). The binary-cache trusted **key NAME is `niks3.numtide.com`** (`niks3.numtide.com-1:DTx8wZduET09hRmMtKdQDxNNthLQETkc/yaX7M4qK0g=`), **not** `cache.numtide.com`; adding `cache.numtide.com` is a **trust expansion** (numtide's CI/signing key builds the agent binaries — see Risk Assessment). `nothingnesses/agent-images` (`lib/mkAgentImage.nix`) is the operator's image reference (we **adapt, not reuse** — vendor/fork it; bus-factor 1), including its non-root `agent` uid-1000 / `/workspace` / pre-created-XDG layout.

- **Network egress is the one axis left fused/implicit.** Anthropic's guidance is that **effective sandboxing requires *both* filesystem and network isolation**; today only filesystem reach is (partly) constrained, while egress is `--unshare-net` for the nix tiers and unrestricted for the rest, never an explicit choice. Promoting it to a first-class `NetworkMethod` axis (`host`/`none`/`proxy`) with an explicit-emit regression guard closes the gap and lets untrusted code keep model-API egress through an allowlist proxy.
- **bwrap/docker hardening asymmetry vs the already-restricted operator.** The operator's `security.go`/`networkpolicy.go` are restricted-grade + default-deny, yet the CLI's bwrap leaks (`--bind $HOME`, `--dev-bind /dev`, host PATH) and the docker isolator ships **zero** hardening flags — the two backends that exist to isolate are weaker than the K8s path. This plan brings both CLI isolators up to deny-default parity (and adds the kernel-isolation guarantee the operator still lacks by defaulting to runc).
- See **`plans/agent-sandbox-provisioning-axes/research-best-practices-audit.md`** for the full best-practices audit (the three-axis canon, the four policy guarantees, per-isolator hardening flags, the operator runtimeClass/SA/limits/FQDN-egress corrections, and the image content-closure correction).

**Alternative considered & rejected:** running nix *inside* each sandbox (bind store + daemon socket, `nix develop --command`). Rejected because punching the host nix daemon socket through the docker/pod boundary undermines the very isolation those backends exist to provide; resolve-on-host-then-mount is both simpler and stricter, and unifies cleanly with the operator's build-time image.

## Proposed Approach

1. **Three-axis types + decoupled+extended policy + dedups → `shared-*-go`.** New `IsolationMethod{none,bwrap,docker}` + `ProvisioningMethod{none,nix,command}` + `NetworkMethod{host,none,proxy}` + a `Provisioner` *contribution* type (`{roBindPaths, pathEntries, env, initCommands}`) + a `hostPassthrough` field; `SandboxConfig` gains `Network NetworkMethod` (**replaces** the old `Network bool`) + `EgressAllowlist []string` + a shared `DefaultEgressAllowlist`. Split `RequirePinnedToolchain` → `RequirePinnedProvisioning` + `RequireHostToolsUnreachable`, and **add** `RequireEgressRestricted` (`Network ∈ {none,proxy}`) + `RequireKernelIsolation`; rework `tierIsolation`/`enforce*` as pure `(iso, prov, net, passthrough, runtime, policy) → decision`. Fold `shellQuote` into `shell.Quote`; promote `ParseCustomEnv`/`MergeEnvMaps`/`EnvMapToSlice` (→ `shared-core-go`) and `BuildAgentCommand`/`BuildProviderEnv` (→ `shared-agent-go`, dedups the operator).
2. **The pinned flake + host-resolve helper → `shared`.** One `flake.lock`-pinned flake with `numtide/llm-agents.nix` as an input; a pure helper that, given a *source* (rev-pinned package set \| project flake), produces a **content-pinned closure descriptor** (store paths + PATH/env). **Enforce the committed lock at resolve time** (`--no-write-lock-file` / `--frozen`; reject a dirty/missing lock, **fail closed**) and resolve the closure's **requisites** (`nix path-info -r` of the descriptor) so the isolator binds only those paths — the input both realizers consume. Retire `nixenv/{render,resolve,build}.go`'s channel-pinned `buildEnv` path.
3. **CLI `Isolator × Provisioner × Network` composition + hardening.** Collapse `bwrap`/`nix`/`nixflake` into one bwrap isolator + provisioners; isolators `{none, bwrap, docker}` each *consume a Provisioner contribution* and apply it via their mechanism (bwrap binds / `docker -v`). Wire `--provision` + `--host-passthrough` + `--network{host,none,proxy}` + repeatable `--egress-allow` flags beside `--sandbox` (the deprecated `--sandbox` alias maps legacy methods to `Network=host` for one release); enforce the curated matrix. **Apply `NetworkMethod` EXPLICITLY** (emit `--unshare-net` for `none`/`proxy`; never leave net state implicit — the regression guard against silently dropping today's unshare). **Harden bwrap** (deny-default): sandbox-local HOME with only workspace rw + RepoDir ro + curated `sandbox.UserConfigPaths` ro binds, `--dev /dev` (not `--dev-bind /dev /dev`), `--clearenv` + an explicit `--setenv` allowlist (API keys injected explicitly), assert no-new-privs + cap-drop ALL; residual/optional (document, may defer): `--new-session`/seccomp `ioctl(TIOCSTI)` filter, `--unshare-user` + `--disable-userns`, cgroup limits via `systemd-run --scope`, Landlock. **Harden docker** (defaults, not opt-in): `--security-opt=no-new-privileges`, `--cap-drop ALL`, keep default seccomp (never `seccomp=unconfined`), `--security-opt apparmor=docker-default`, `--read-only` + `--tmpfs /tmp:rw,noexec,nosuid` (+ a writable workdir bind), `--pids-limit`/`--memory`/`--cpus`; curated config paths ro into a synthetic HOME (the workdir is the **only** writable bind); apply `NetworkMethod` (`none → --network none`; `proxy →` custom net + `HTTP(S)_PROXY` to the allowlist proxy; `host → bridge`); residual/optional `--runtime runsc` (gVisor → `RequireKernelIsolation`).
4. **CLI nix provisioner.** Resolve on the host → closure (two sources); contribute read-only binds of **only the closure's requisites** (`nix path-info -r`), **not** the whole `/nix/store` (alternative: `nix copy` to a per-sandbox store) — this is what makes `RequireHostToolsUnreachable` accurate; contribute the closure's PATH/env; **GC-root the FULL dev closure** (`nix develop --profile <gcroot>` or `nix-store --realise --add-root` over the closure paths, `keep-outputs=true`), not just the `mkShell` output drv — regression test: `nix-collect-garbage -d` while a sandbox holds the root; require a committed lock (enforced at resolve in step 2). `hostPassthrough` controls host/base-image PATH bleed-through.
5. **Remove compose.** Delete `internal/sandbox/compose/`, `SandboxConfig.{ComposeFile,Service}`, `findComposeFile`/defaults, the `--sandbox compose` option, and the compose branches in registry; isolation axis collapses to `{none, bwrap, docker}`.
6. **Operator image realizer + hardening.** Replace `toolchain_nix.go`'s init-container install with a `dockerTools.streamLayeredImage` build from the same flake (commit to this **one** builder — it never realizes a store tarball; `nix2container` only if pushing on every small change, noting its `maxLayers` defaults to 1); **vendor/fork** `nothingnesses/agent-images`' `lib/mkAgentImage.nix` (bus-factor 1) and **adapt** its layout (hand-written `/etc/passwd`+`/etc/group` for **numeric** uid 1000; pre-created `/workspace` + XDG dirs) into the `streamLayeredImage` build; `maxLayers` set deliberately (100); never set `created=now` (keep the epoch default) + a rebuild-digest-equality CI check; extend `NixSpec` with a **pin** + **flake/source**; `RequirePinnedProvisioning` satisfied by the pinned layers (the image is the **same content-addressed closure** as the CLI, not byte-identical layers). **Harden the untrusted pod path**: default to a sandboxed `runtimeClass` (gVisor/Systrap) via `DefaultRuntimeClassName`/`AllowedRuntimeClassNames` (wires `RequireKernelIsolation`); `automountServiceAccountToken=false` + a dedicated **zero-RBAC ServiceAccount**; default resource requests/limits + a `LimitRange`/`ResourceQuota`; reconcile `readOnlyRootFilesystem=true` with writable HOME/XDG via `emptyDir` mounts (non-overlapping with baked content) + `fsGroup=1000` (e2e-assert the agent can write config/cache/state); **always emit a default-deny egress `NetworkPolicy`** per `AgentRun`, upgraded to FQDN-aware (Cilium `toFQDNs` + kube-dns `rules: dns`, or a Squid proxy) that blocks metadata/RFC1918.
7. **Docs + validation + rollout.** Update the isolation taxonomy and `AGENTS.md` (three axes + the `hostPassthrough` knob; operator path = "nix-built OCI image"); **document the threat model** (untrusted model-generated code under indirect prompt injection; bwrap/runc = attack-surface reduction, not a kernel boundary; "host tools unreachable" ≠ "host unreachable"; fail closed) and the **four** policy guarantees (`RequirePinnedProvisioning`, `RequireHostToolsUnreachable`, `RequireEgressRestricted`, `RequireKernelIsolation`); validate every CLI matrix cell wraps correctly (incl. explicit `NetworkMethod` emission + the hardening flags present) + the operator image builds and runs + the split/extended policy engages; keep the moon `config → shared → agent` build green; ship via PRs per the protected-main rule.

## Risk Assessment

- **Large blast radius across two binaries + shared libs.** Mitigate with the phased waves: the shared foundation lands first and is purely additive; the CLI refactor and operator realizer build on it independently.
- **Behavioral parity for the CLI nix path.** The resolve-then-mount model replaces both the `nix` (channel buildEnv) and `nixflake` (`nix develop`) tiers — cover with parity tests that the same flake.lock yields the same closure/PATH, and that deny-default (`hostPassthrough` off) still hides host tools.
- **GC eviction mid-run.** New-style nix results aren't auto-gc-rooted; forgetting the GC-root causes intermittent "command not found" under a concurrent `nix-collect-garbage`. Make the GC-root a first-class step, tested.
- **Operator image size / layer ceiling.** `dockerTools` has a max-layer limit; use `streamLayeredImage` (`maxLayers` set deliberately — 100 is the safe conventional choice; 128 sits at overlay2's modern ceiling with zero headroom) and lean on the binary cache; do **not** `follows` nixpkgs to a stable channel against the `llm-agents.nix` input (breaks its cache hits — `flake.lock` still pins the exact rev).
- **`cache.numtide.com` is a trust expansion.** The consumer pin surface is `flake.lock` alone, but substituting prebuilt agent binaries from `cache.numtide.com` means trusting numtide's CI/signing key (`niks3.numtide.com-1:DTx8wZduET09hRmMtKdQDxNNthLQETkc/yaX7M4qK0g=`) to build those binaries. Adding it widens the trusted-key set beyond our own builders — accept deliberately (or build from source) and record the key, don't add it implicitly.
- **Network-default regression.** Collapsing `nix`/`nixflake` (which `--unshare-net` today) into one bwrap isolator can **silently drop network isolation** unless the isolator applies `NetworkMethod` **explicitly**. Make explicit-emit (`--unshare-net` for `none`/`proxy`) the regression guard, with a test asserting the flag is present for every non-`host` cell.
- **Hardening vs functionality.** The agent still needs its curated credential config paths **and** model-API egress, so deny-default must **not** be a blanket lockout: keep the curated read-only config binds and ship a `DefaultEgressAllowlist` that includes the provider endpoint (and let `--egress-allow` extend it). A too-aggressive `--clearenv`/HOME/network clamp that drops the API key or the provider host breaks the run — cover with an e2e that the agent reaches the model API under `Network=proxy` with the curated config present.
- **`readOnlyRootFilesystem` vs writable HOME/XDG (operator).** The image bakes content and sets `readOnlyRootFilesystem=true`, but the agent must write `~/.config`/`~/.cache`/`~/.local/state`; mount **`emptyDir`** for HOME + XDG_* (non-overlapping with baked content) + `fsGroup=1000` and e2e-assert config/cache/state are writable — otherwise the agent fails to start.
- **`llm-agents.nix` / `agent-images` are external, fast-moving inputs.** Pin both to specific revs in our `flake.lock`; mirror the daily-bump pattern as a deliberate freeze/advance lever, not auto-follow.
- **Compose removal is a breaking CLI change.** It's an unused tier; remove cleanly (no shim) per the project rule, and call it out in the release notes.
- **Protected `main` + cross-cutting waves.** Land via PRs in dependency order; the shared foundation must merge before the CLI/operator waves resolve their imports.

## Proposed Child Plans

**Group 1 — shared foundation (gates everything):**
`01-shared-types-policy-dedups` → `02-shared-flake-and-resolver`

**Group 2 — `agent-cli-go` refactor (serialize on `internal/sandbox/*`), after group 1:**
`03-cli-isolator-provisioner-core`, `04-cli-nix-provisioner`, `05-cli-remove-compose`

**Group 3 — `agent-operator-go` realizer (independent Go, parallel with group 2), after group 1:**
`06-operator-image-realizer`

**Group 4 — docs + validation + rollout, after groups 2 & 3:**
`07-docs-validation-rollout`

## Success Criteria

- The sandbox is selected by **three independent axes + a passthrough knob** (`{none,bwrap,docker} × {none,nix,command} × {host,none,proxy}`); the curated matrix is enforced (invalid cells rejected with a clear error) and `NetworkMethod` is applied **explicitly** by the isolator (`--unshare-net` present for every `none`/`proxy` cell — the regression guard).
- The `nix` provisioner **resolves to a `flake.lock`/rev-pinned closure** (lock enforced at resolve via `--frozen`/`--no-write-lock-file`, **fail closed** on a dirty/missing lock) and the isolator **mounts only the closure's requisites read-only** (not the whole `/nix/store`) — no in-sandbox nix daemon; `bwrap × nix` and `docker × nix` yield the **same** toolset; the FULL dev closure is GC-rooted (survives `nix-collect-garbage -d` mid-run).
- `hostPassthrough` off ⇒ host tools off PATH (and unreachable under bwrap/docker); on ⇒ host/base-image tools available as fallback. Today's leaky-bwrap and deny-default-nix tiers are expressible as one isolator + the toggle.
- All **four** policy guarantees — `RequirePinnedProvisioning`, `RequireHostToolsUnreachable`, `RequireEgressRestricted`, `RequireKernelIsolation` — are independently requestable and correctly classify every cell (incl. `none × nix` = pinned-provisioning yes, unreachable no; `Network=host` does **not** satisfy `RequireEgressRestricted`; bwrap/default-runc do **not** satisfy `RequireKernelIsolation`), and the engine **fails closed** when a requested guarantee can't be established.
- `RequireEgressRestricted` holds under `Network ∈ {none, proxy}`: `none` has no network; `proxy` reaches **only** the allowlist (provider API + curated registries) with metadata/RFC1918/loopback denied.
- `RequireKernelIsolation` is satisfied only by docker `--runtime runsc`/gVisor or an operator sandboxed `runtimeClass` (gVisor/Kata/kata-cc) — never by bwrap or default runc.
- The bwrap and docker isolators are **deny-default**: sandbox-local HOME (no blanket `$HOME` bind), `--dev /dev` (not `--dev-bind /dev /dev`), `--clearenv` + an explicit setenv allowlist, and (docker) `no-new-privileges` + `cap-drop ALL` + default seccomp/apparmor + `--read-only` + resource limits — all asserted present.
- The **threat model is documented** (untrusted model-generated code; bwrap/runc = attack-surface reduction, not a kernel boundary; "host tools unreachable" ≠ "host unreachable"; fail closed) alongside the three axes and four guarantees.
- The `command` provisioner runs a single init-command list before the agent under every isolator.
- The **operator** provisions via a **nix-built OCI image** from the same flake (no per-pod `nix profile install`); pods start by image pull; the image is the **same content-addressed store-path closure as the CLI** (identical store-path hashes from the same `flake.lock`; verified with `nix path-info -r`, **not** layer bytes), and defaults to a sandboxed `runtimeClass` + zero-RBAC SA + resource limits + a default-deny/FQDN egress `NetworkPolicy`, with writable HOME/XDG reconciled against `readOnlyRootFilesystem=true`.
- The **compose** tier is gone; `typecheck/lint/build/test` green across `shared-*-go`, `agent-cli-go`, `agent-operator-go`; the moon build is green; both consumers ship via PRs.

## Estimated Effort

- Group 1 (shared foundation): ~2–3 days incl. tests + the flake (the network types + the two added policy guarantees are additive).
- Group 2 (CLI refactor + compose removal): ~4–5 days incl. parity tests — **03 grows** beyond the axis split to carry the network-axis application + the bwrap/docker deny-default hardening.
- Group 3 (operator image realizer): ~3–4 days incl. image build + e2e — **06 grows** to carry the operator hardening (sandboxed runtimeClass, zero-RBAC SA, resource limits/LimitRange, readOnlyRootFilesystem-vs-XDG reconcile, FQDN-aware default-deny egress).
- Group 4 (docs/validation/rollout): ~1–2 days incl. the threat-model doc + the four-guarantee matrix validation.
- **Total: ~10–14 days** (the added network axis + CLI hardening + operator hardening add ~2–3 days over the original split), parallelizable across the CLI (group 2) and operator (group 3) tracks once the shared foundation lands (~7–9 days wall-clock with two workstreams).

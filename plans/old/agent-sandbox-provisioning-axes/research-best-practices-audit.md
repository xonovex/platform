---
type: research
parent_plan: plans/agent-sandbox-provisioning-axes.md
topic: agent isolation (bwrap/docker/k8s) + tool provisioning (host/bwrap/docker/k8s) best-practice audit
status: informational
---

# Best-Practices Audit — Agent Isolation & Tool Provisioning

Read-only audit of `plans/agent-sandbox-provisioning-axes.md` (and subplans 01–07) against
2025/2026 best practice for AI-coding-agent isolation and Nix tool provisioning. Grounded
against the actual current code (`internal/sandbox/{bwrap,docker,nix,nixflake}`, operator
`builder/{security,networkpolicy,toolchain_nix}.go`). No code changed.

## Verdict

**Mostly-modern-with-gaps, and the gaps cluster on one axis the plan never names: network egress.**
The plan's *architecture* is genuinely good — splitting the fused `SandboxMethod` enum into orthogonal
`IsolationMethod` × `ProvisioningMethod` + `hostPassthrough`, resolving the nix toolchain on the host and
mounting `/nix/store` read-only (no in-sandbox daemon, no daemon socket punched through the boundary),
GC-rooting the closure, pinning by `flake.lock`, and replacing the per-pod `nix profile install` with a
prebuilt `streamLayeredImage` are all squarely 2025/2026 best practice. But for an agent that runs
untrusted, model-generated code, the plan reasons about isolation *only* as "are host tools reachable on
PATH/binds." It is silent on network egress (the single highest-value control per Anthropic, NVIDIA,
OpenAI Codex), silent on the bwrap/docker hardening flags the comparable shipping CLIs all set, never
states a threat model, and leaves several pre-existing credential-exposure footguns (full `$HOME` bind,
`--dev-bind /dev /dev`, whole-`/nix/store` bind) untouched. None block the refactor, but several are
"must add to be current best practice," not optional polish. Critically, the refactor risks *regressing*
one thing today's code gets right by accident: the nix tiers currently emit `--unshare-net`, plain bwrap
does not, and collapsing them into one isolator with no stated network contract can silently drop that.

## Isolation best practices

### bwrap — partial → gap

What the plan gets right:
- Unifying the triplicated "leaky bwrap" / "deny-default nix" / nixflake skeletons into ONE bwrap isolator
  + a `hostPassthrough` toggle is a clean, correct architecture; real wrappers parameterize a single bwrap
  core this way.
- Read-only binds for the provisioner contribution, and `--die-with-parent` (already in `bwrap.go`), are correct.
- Read-only `/nix/store` mount with no in-sandbox nix daemon is the strongest part of the design and matches
  the Ranti model (https://www.ranti.dev/blog/securing-ai-agents-with-nix-and-bubblewrap).

Gaps vs 2025/2026 baseline (the bwrap README itself: protection "is entirely determined by the arguments
passed" — https://github.com/containers/bubblewrap/blob/main/README.md):
- **No `--unshare-net` / network contract.** bwrap shares the host net namespace unless explicitly unshared;
  the existing `--share-net` (bwrap.go:76) is a no-op because nothing unshares net first. "Deny-default" hides
  files but leaves full host network egress. Make `--unshare-net` the default for the agent phase.
  (https://developers.openai.com/codex/agent-approvals-security)
- **`--dev-bind /dev /dev` (bwrap.go:100) exposes the whole host device tree** (`/dev/sda`, `/dev/mem`, input
  devices), nullifying every read-only bind. Replace with `--dev /dev`.
  (https://manpages.debian.org/testing/bubblewrap/bwrap.1.en.html)
- **Full `$HOME` read-write bind (bwrap.go:106)** survives even with passthrough OFF, exposing `~/.ssh`/`~/.aws`/
  tokens — the exact anti-pattern agent sandboxes exist to prevent. Set `HOME` to a sandbox-local dir and bind
  only the workspace + RepoDir (ro). (https://blog.palaimon.io/posts/coding-agents-bubblewrap-deep-dive/index.html)
- **No `--clearenv`** — host env (API keys, `SSH_AUTH_SOCK`) inherits into the untrusted sandbox. Add
  `--clearenv` + an explicit `--setenv` allowlist.
- **No `--new-session` / TIOCSTI guard** — CVE-2017-5226 terminal-injection escape unmitigated; bwrap does NOT
  block TIOCSTI by default upstream (the default-block is a distro backport). Add `--new-session` or a seccomp
  filter rejecting `ioctl(TIOCSTI)`. (https://github.com/containers/bubblewrap/issues/142)
- **No `--seccomp`, no `--unshare-user` + `--disable-userns`.** The skeleton uses individual unshares "to avoid
  userns issues," which also blocks `--disable-userns` (bwrap 0.8.0+, requires `--unshare-user`).
  (https://github.com/flatpak/flatpak/pull/5084)
- Optional: cgroup limits via `systemd-run --scope` (fork-bomb bound), Landlock (kernel-enforced fs/net layer),
  fail-closed detection of unprivileged-userns on Ubuntu 24.04+.
  (https://github.com/anthropic-experimental/sandbox-runtime)

`--cap-drop ALL` and no-new-privs are already bwrap defaults — assert, don't add.

### docker — gap

What the plan gets right:
- Running as the current user via `-u uid:gid` (docker.go:62-64) matches OWASP RULE #2.
  (https://cheatsheetseries.owasp.org/cheatsheets/Docker_Security_Cheat_Sheet.html)
- Read-only binds via `-v p:p:ro`, and the explicit refusal to expose the nix daemon socket through the boundary
  (the docker-socket anti-pattern, RULE #1 / MITRE T1611), are correct.

Gaps — `buildDockerArgs` (docker.go:48-130) emits only `run --rm -it`, `-w`, `-u`, `-e`, `-v`, image, cmd, with
**zero hardening flags**, which every cited source rejects for untrusted code:
- Add `--security-opt=no-new-privileges` (OWASP RULE #4 / CIS 5.25), `--cap-drop ALL` (RULE #3), keep default
  seccomp + `--security-opt apparmor=docker-default` (RULE #6), `--read-only` rootfs + `--tmpfs /tmp:rw,noexec,nosuid`
  (RULE #8), and `--pids-limit`/`--memory`/`--cpus` (RULE #7). Defaults, not opt-in.
- **Stop mounting the whole `$HOME` read-write (docker.go:84, no `:ro`)** — runs as the same host uid with write
  access to `~/.ssh`, `~/.aws`, tokens. Mount only needed config paths, read-only, into a synthetic container HOME;
  keep workdir the only writable bind. More impactful than any single flag.
  (https://www.penligent.ai/hackinglabs/claude-code-sandbox-bypass/)
- **No stronger-runtime option.** Default runc is not a trust boundary for model-generated code (Nov-2025 runc CVE
  trio CVE-2025-31133/52565/52881). Allow `--runtime runsc` (gVisor); prefer rootless docker / userns-remap so a
  breakout isn't host-uid. (https://gvisor.dev/docs/architecture_guide/security/,
  https://github.com/opencontainers/runc/security/advisories/GHSA-cgrx-mc8f-2prm)
- **`IsolationHidesHost() → docker:true` (unconditional) is a regression** from current `registry.go` (53-95), which
  treats container tiers as `IsolationContainerPinned` requiring a pinned `--image`. With a whole-`$HOME` mount +
  whole-`/nix/store` bind, "host unreachable" is false. Condition the guarantee on (a) closure-only store mounts,
  (b) no host-home mount, (c) a pinned image.

### k8s (operator) — partial (the strongest surface)

What the plan gets right — these PRE-EXIST, so the plan correctly leaves them intact:
- `security.go` already sets PSS-restricted-grade controls: `allowPrivilegeEscalation=false`, `runAsNonRoot=true`,
  `readOnlyRootFilesystem=true`, seccomp `RuntimeDefault`, drop ALL caps.
  (https://kubernetes.io/docs/concepts/security/pod-security-standards/)
- `networkpolicy.go` already does default-deny ingress + spec-driven egress (empty egress = deny-all, both
  PolicyTypes set) — a genuinely good foundation.
- Prebuilt digest-pinned OCI image is *better* for isolation than the `/nix/store` bind (carries only the closure).

Gaps:
- **`runtimeClassName` "unchanged" = default runc = shared host kernel**, below 2026 K8s Agent Sandbox guidance for
  untrusted model-generated code. The mechanism already exists (`DefaultRuntimeClassName`, `AllowedRuntimeClassNames`,
  gVisor/Kata/CoCo e2e suites) — only the default + a policy binding are missing. Default the untrusted path to gVisor
  (Systrap, no KVM needed). (https://kubernetes.io/blog/2026/03/20/running-agents-on-kubernetes-with-agent-sandbox/)
- **NetworkPolicy is L3/L4 only** — can't express an FQDN allowlist for LLM API/registries behind CDNs, doesn't block
  metadata/RFC1918, and a default-deny that permits kube-dns still leaves DNS-tunnel exfiltration open. Upgrade to
  Cilium `toFQDNs` (paired kube-dns `rules: dns`) or a Squid egress proxy; always emit a default-deny egress policy per
  AgentRun. (https://docs.cilium.io/en/stable/security/dns/,
  https://unit42.paloaltonetworks.com/bypass-of-aws-sandbox-network-isolation-mode/)
- **`automountServiceAccountToken` not disabled / no dedicated zero-RBAC SA** — the agent never calls the K8s API, so
  the auto-mounted `default` SA token is a free cluster foothold. Add `automountServiceAccountToken=false` +
  `serviceAccountName`. (https://kubernetes.io/docs/concepts/security/rbac-good-practices/)
- **No default resource limits / LimitRange / ResourceQuota** — unbounded untrusted container is a trivial node DoS.
- **`readOnlyRootFilesystem=true` vs writable HOME/XDG unreconciled** — the read-only nix image blocks the agent's
  config/cache/state writes; pre-baked XDG dirs in a read-only layer aren't writable, and an emptyDir over a baked path
  shadows it. Add emptyDir mounts for HOME + `XDG_*` with matching env + `fsGroup=1000`.
- Optional backstop: label namespace `pod-security.kubernetes.io/enforce=restricted` + Kyverno/Gatekeeper for what PSA
  omits (readOnlyRootFilesystem, automount, required runtimeClass, limits).

## Provisioning best practices

### host (CLI nix) — modern, with two enforcement gaps

Rulings on the explicit questions:
- **Resolve-on-host-then-mount vs nix-in-sandbox: resolve-on-host is correct and strictly safer.** Read-only
  `/nix/store` means even a fake-root agent can't overwrite binaries; running nix in the jail or punching the daemon
  socket re-grants host store-write/build escalation. (The current `nixflake` tier even binds
  `/nix/var/nix/daemon-socket/socket` — the exact anti-pattern the plan correctly removes.)
- **GC-root: necessary and correctly called out.** Still true in 2025/2026 — only `nix build`'s result symlink
  auto-roots; `nix shell`/`develop`/`--no-link` create no persistent root. *Refinement:* root the FULL dev closure,
  not just the mkShell output drv (rooting the drv doesn't protect its `buildInputs` unless `keep-outputs=true`, default
  false). Use `nix develop --profile <gcroot>` or `nix-store --realise --add-root` over the closure paths. Add a
  `nix-collect-garbage -d`-while-held regression test.
- **`flake.lock` pinning: correct**, and retiring the `fetchTarball` channel renderer drops a fetcher that can't
  substitute from caches.

Gaps:
- **"Require a committed lock" is unenforced without `--no-write-lock-file`/`--frozen`.** By default
  `nix build`/`develop`/`print-dev-env` silently re-lock when an input drifts, defeating the pin at the moment of
  resolution. Resolve with the no-write flag and reject a dirty/missing lock. (subplan 02/04)
- **`RoBindPaths:["/nix/store"]` binds the WHOLE world-readable store** into the untrusted sandbox, leaking every other
  project's source + any in-store secret + path enumeration. Bind only the closure's requisites (`nix path-info -r`) or
  `nix copy` to a per-sandbox store. High vs the untrusted-code threat model; low for single-tenant dev — state the
  trust assumption either way.

### bwrap (provisioning) — modern

`bwrap×nix` and `docker×nix` yielding the same toolset follows by construction: both consume an identical pre-resolved
closure descriptor and differ only in bind mechanism (`--ro-bind` vs `-v :ro`). Directly testable as the plan proposes.
Use `nix print-dev-env --json` (PATH + vars) and `nix build --no-link --print-out-paths`, passing a flakeref not a raw
`.drv`.

### docker (provisioning) — modern (same closure, read-only)

Read-only `/nix/store` bind is the integrity-sound pattern; same closure-scope caveat as host. No provisioning
objection; docker *isolation* hardening gaps above are separate.

### k8s (operator image) — modern, with phrasing/builder corrections

Rulings:
- **Build-time image vs per-pod install: build-time wins decisively.** A pod has no durable `/nix/store`, so
  `nix profile install` re-substitutes the whole closure into an emptyDir on every cold start with a cold eval cache and
  zero cross-pod dedup. Prebuilt image = pull + start, kubelet caches layers per node.
- **streamLayeredImage vs nix2container: pick ONE.** For an image that changes only when `flake.lock` advances,
  `streamLayeredImage` + stock skopeo is simpler (no patched `skopeo-nix2container` input) and never realizes a store
  tarball. Choose nix2container only if you push on every small change (skip-already-pushed) — note its `maxLayers`
  defaults to **1**. (https://github.com/nlewo/nix2container)
- **Read-only `/nix/store`: correct**, same closure-scope note.

Corrections the plan must absorb:
- **"byte-identical layers to the CLI closure" is a category error** (parent lines 60/104/151). The CLI bind-mounts
  store *directories*; the image bakes them into *tar layers* — different representations. Restate as "the CLI and
  operator provision the same content-addressed store-path closure (identical store-path hashes from the same
  `flake.lock`)," and test by comparing `nix path-info -r`, not layer bytes. (subplan 06/07)
- **Builder mismatch:** `nothingnesses/agent-images` `mkAgentImage` is built with `dockerTools.buildLayeredImage`
  (realizes a gzipped store tarball — the cost streamLayeredImage avoids) and uses a *named* `agent` user with no
  `maxLayers`. Don't call it directly; **port its layout** (hand-written `/etc/passwd`+`/etc/group` for uid 1000,
  mkdir+chown of HOME/.config/.cache/.local/workspace) into a `streamLayeredImage` build.
- **`maxLayers=128`: defensible, not a bug**, but Default 100 is the safe conventional choice (overlay2 ceiling, zero
  extension headroom at 128).
- **Reproducibility:** dockerTools defaults `created` to epoch. Make "never set `created=now`" an explicit invariant and
  add a rebuild-digest-equality CI check. (subplan 06)
- Optional supply-chain completion: SBOM (bombon) + cosign sign-by-digest + admission verification.

## Cross-cutting gaps

1. **Network egress is a MISSING AXIS — the biggest structural gap, and real for this plan's purpose.** The plan's
   thesis is "a fused enum hides orthogonal concerns; split them." Network reachability is genuinely orthogonal to
   isolation strength (fs/PATH) and provisioning (toolset), yet it's the one axis left fused/implicit. Anthropic states
   verbatim "effective sandboxing requires both filesystem and network isolation"
   (https://www.anthropic.com/engineering/claude-code-sandboxing); NVIDIA lists egress as one of three *mandatory*
   controls; Codex defaults network OFF. **Add `NetworkMethod ∈ {host, none, allowlist-proxy}`** across
   {none,bwrap,docker} + operator, plus a third independently-requestable guarantee `RequireEgressRestricted`. At
   minimum the unified bwrap isolator must emit `--unshare-net` by default (preserve what nix.go/nixflake.go do today;
   fix plain bwrap), and docker must keep `--network none`. The refactor is the moment to make this first-class — and
   the moment it could silently regress.
2. **microVM/gVisor/Kata escalation — real on k8s, NOT on the CLI.** For the CLI on a dev/CI host, process-level
   isolation (bwrap) is exactly what Claude Code and Codex CLI use; omitting Firecracker/Kata there is correct. For the
   *operator* running pods of untrusted code, defaulting to runc is below 2026 consensus, and the machinery already
   exists — make gVisor/Kata a deliberate default for the untrusted path.
3. **seccomp / Landlock / hardening flags — real gap on both CLI surfaces.** The operator is restricted-grade; the CLI
   bwrap/docker isolators have none of the analogous controls. Asymmetric with no stated reason.
4. **No stated threat model — real, cheap to fix.** The plan never says the agent runs untrusted/model-generated code
   under indirect prompt injection, nor that bwrap is attack-surface reduction rather than a kernel trust boundary
   (CVE-2025-59532 Codex cwd bypass; CVE-2026-5752 Cohere Terrarium). One overview paragraph makes every recommendation
   above legible as load-bearing.
5. **Credential minimization / fail-closed posture — medium, both surfaces.** Long-lived tokens in the agent env
   exfiltrate as easily as they're used; prefer broker-at-proxy injection. The sandbox should fail CLOSED (refuse if the
   requested guarantee can't be established), not silently degrade.

## External-dependency risks

All four references exist, are active/unarchived, and are characterized correctly in substance:
- **numtide/llm-agents.nix** — real, very active (renamed from `numtide/nix-ai-tools`), binary cache
  `https://cache.numtide.com` (trusted key NAME is `niks3.numtide.com`, not `cache.numtide.com`). Plan correctly forbids
  `inputs.nixpkgs.follows`. Fix: (a) parent line 106 "enforces no isolation" — reword to "packaging only; isolation out
  of scope"; (b) "flake.lock + hashes.json pin" conflates consumer pin (`flake.lock`) with upstream `hashes.json`
  bookkeeping — consumers pin via `flake.lock` alone; (c) adding the cache is a *trust expansion* (numtide's signing key
  builds the agent binaries) — add a Risk-Assessment line.
- **nothingnesses/agent-images** — real but young (created 2026-03-15, single maintainer, bus-factor 1), uses
  `buildLayeredImage` + named user + no `maxLayers`. Since the plan borrows one file's layout, **vendor/fork
  `lib/mkAgentImage.nix`** rather than take a live input. State the plan *adapts*, not reuses.
- **nlewo/nix2container** — mature (v1.0.0), archive-less + skip-already-pushed confirmed. Just don't offer it OR
  streamLayeredImage without committing.
- **dockerTools.streamLayeredImage** — current, confirmed not to realize the image into the store. Correct choice.

## Prioritized recommendations

*Must add to be current best practice:*

1. **Network egress as a first-class axis + guarantee.** `NetworkMethod ∈ {host, none, allowlist-proxy}` +
   `RequireEgressRestricted`; default the agent phase to deny-default egress; `allowlist-proxy` routes through a host-side
   HTTP(S)+SOCKS proxy with exact-match domain allowlist, metadata(169.254.169.254)/RFC1918/loopback/link-local blocked,
   DNS pinned, per-connection logging, regression tests for SOCKS5/null-byte/IDN/IPv6 bypasses. Subplans 01+03+06+07.
   Severity: **high.** (https://www.anthropic.com/engineering/claude-code-sandboxing)
2. **Make the unified bwrap isolator's network contract explicit and tested so the refactor can't drop today's
   `--unshare-net`.** Emit `--unshare-net` by default; parity test that with network denied a `connect()` fails across
   bwrap and docker. Subplan 03. Severity: **high.**
3. **Stop the credential-exposure footguns in deny-default mode:** don't bind full `$HOME` (bwrap.go:106 / docker.go:84) —
   set `HOME` to a sandbox-local dir, bind only workspace + RepoDir(ro); replace `--dev-bind /dev /dev` with `--dev /dev`;
   add `--clearenv` + `--setenv` allowlist. Subplan 03. Severity: **high.**
4. **Bind only the resolved closure, not the whole `/nix/store`** (`nix path-info -r`, or `nix copy` to a per-sandbox
   store); makes `RequireHostToolsUnreachable` accurate. Subplans 04+06. Severity: **high** (untrusted) / medium
   (single-tenant).
5. **Enforce the lock at resolve time with `--no-write-lock-file`/`--frozen`** and reject dirty/missing locks, so
   `RequirePinnedProvisioning` is fail-closed. Subplans 02+04. Severity: **high.**
6. **Add docker hardening defaults** to `buildDockerArgs`: `--security-opt=no-new-privileges`, `--cap-drop ALL`, keep
   default seccomp, `--read-only`+`--tmpfs /tmp`, `--pids-limit`/`--memory`/`--cpus`. Subplan 03. Severity: **high.**
7. **GC-root the full dev closure, not the mkShell drv** (`--profile`/`--add-root`; `keep-outputs=true`), with a
   `nix-collect-garbage -d`-while-held regression test. Subplan 04. Severity: **medium.**
8. **Reconcile `readOnlyRootFilesystem=true` with writable HOME/XDG** in the operator image: emptyDir mounts for HOME +
   `XDG_*` + `fsGroup=1000`, non-overlapping with baked content; e2e assert the agent can write config. Subplan 06.
   Severity: **medium.**
9. **Default the operator untrusted path to a sandboxed runtimeClass (gVisor)** + add a `RequireKernelIsolation`
   guarantee; the machinery already exists. Subplans 01+06. Severity: **medium.**
10. **Set `automountServiceAccountToken=false` + dedicated zero-RBAC SA + default resource limits/LimitRange** on agent
    pods; always emit a default-deny egress NetworkPolicy per run; upgrade to FQDN-aware egress + metadata/RFC1918 block.
    Subplan 06. Severity: **medium.**
11. **Add a one-paragraph threat model** (untrusted/model-generated code under indirect prompt injection; bwrap = attack-
    surface reduction, not a kernel boundary) and reframe `IsolationHidesHost()` docs so "host tools unreachable" ≠ "host
    unreachable." Subplan 07 (+ parent overview). Severity: **medium.**
12. **Fix the docs/builder mischaracterizations:** "byte-identical layers" → "same content-addressed closure"; port
    `mkAgentImage` into `streamLayeredImage`; vendor `mkAgentImage.nix`; reword llm-agents "enforces no isolation" /
    "hashes.json pin"; record the cache trusted-key string + trust expansion; commit to ONE image builder; set `maxLayers`
    deliberately. Subplans 02+06+07. Severity: **low–medium** (correctness/clarity).

*Optional hardening (defensible to defer, document as known residual):*

13. bwrap `--new-session`/seccomp TIOCSTI guard, `--unshare-user`+`--disable-userns`, cgroup limits, Landlock, fail-closed
    userns detection. Subplan 03.
14. CLI docker resource limits + rootless/userns-remap + optional `--runtime runsc`. Subplan 03.
15. Credential-broker-at-proxy (keep tokens out of the sandbox env); SBOM + cosign sign-by-digest + admission
    verification; PSA `enforce=restricted` namespace label + Kyverno; nix-snapshotter as a forward-looking watch item.
    Subplans 06+07.

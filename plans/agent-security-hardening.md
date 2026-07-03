---
type: plan
has_subplans: true
status: approved
feature: agent-security-hardening
dependencies:
  plans: []
  subplans:
    01-operator-webhook-deployment: []
    02-cli-egress-fail-closed: []
    03-operator-fail-closed-enforcement: [01-operator-webhook-deployment]
    04-operator-pod-correctness: [03-operator-fail-closed-enforcement]
    05-operator-reconcile-hygiene: [03-operator-fail-closed-enforcement]
    06-cli-host-reach-and-pinning: [02-cli-egress-fail-closed]
    07-cli-cleanup: [06-cli-host-reach-and-pinning]
proposed_subplans:
  - 01-operator-webhook-deployment
  - 02-cli-egress-fail-closed
  - 03-operator-fail-closed-enforcement
  - 04-operator-pod-correctness
  - 05-operator-reconcile-hygiene
  - 06-cli-host-reach-and-pinning
  - 07-cli-cleanup
parallel_groups:
  - group: 1
    plans: [01-operator-webhook-deployment, 02-cli-egress-fail-closed]
    note: "Independent projects, run in parallel. Webhook deployment gates all operator enforcement work; docker egress is the highest-severity CLI hole."
  - group: 2
    plans: [03-operator-fail-closed-enforcement]
    depends_on: [1]
    note: "Edits the webhook + controller packages that group 1 makes deployable; enforcement changes are only provable in e2e once webhooks actually run."
  - group: 3
    plans: [04-operator-pod-correctness, 05-operator-reconcile-hygiene]
    depends_on: [2]
    note: "Both edit internal/controller/* — SERIALIZE (any order), not parallel worktrees. Independent of the CLI groups. 05 owns the shared-reconcile extraction; 04's controller edits land on top of or before it, coordinate on agentrun_controller.go."
  - group: 4
    plans: [06-cli-host-reach-and-pinning]
    depends_on: [1]
    note: "Touches the same isolation plugin files (bwrap.go/docker.go) as 02-cli-egress-fail-closed, so it serializes after it. Independent of operator groups. Also edits shared-agent-go."
  - group: 5
    plans: [07-cli-cleanup]
    depends_on: [4]
    note: "Deletes config fields 06 may have consumed — must run after 06 to see the merged state."
skills_to_consult:
  - kubernetes-guide
  - docker-guide
  - microkernel-pattern-guide
  - testing-guide
  - code-quality-guide
research_sources:
  documentation:
    - https://book.kubebuilder.io/reference/webhook-overview (webhook deployment: Service + cert-manager + *WebhookConfiguration)
    - https://docs.docker.com/engine/network/ (bridge vs internal networks; env-var proxies are advisory only)
    - https://nix.dev/manual/nix/stable/command-ref/new-cli/nix3-flake (--no-update-lock-file errors on dirty lock; --no-write-lock-file only suppresses writing)
  versions:
    go: 1.26.0
    controller-runtime: v0.24.1
    cobra: v1.10.2
---

# Agent Security Hardening

## Overview

A full review of `packages/agent` (session 2026-07-02: two independent code audits
plus spot verification, including live reproduction of the docker duplicate-mount
and bwrap rw-overmount failures) found that the architecture matches AGENTS.md —
orthogonal axes, plugin registries, method-agnostic policy — but the threat-model
guarantees (fail closed, egress restricted, host tools unreachable, pinned
provision) are each undermined at one or more seams. In the operator the
enforcement layer is not deployable at all. This plan closes those seams.

File:line references throughout this plan and its subplans are anchors as of
`main` @ `a6f765e2` (2026-07-02); they will drift as subplans land — always
read the cited file before editing.

## Goals

- Every `Require*` guarantee is enforced end-to-end or the run is refused — no
  silent degradation on any error path (CLI and operator).
- The operator's admission layer actually runs in a real cluster, and e2e proves
  the hardened paths instead of faking or disabling them.
- Secrets never appear in argv, pod-spec literals, or tmux command strings.
- Dead API surface and dead config are removed (repo rule: no deprecated code).

## Current State

- `agent-cli-go` (Go 1.26, cobra): sandbox CLI. `go vet`/`gofmt`/tests green.
  Policy engine (`internal/sandbox`) and registry composition root verify
  cleanly with good unit coverage; the defects are in the plugin leaves and
  cmd wiring.
- `agent-operator-go` (controller-runtime v0.24.1): AgentRun/Workspace/Harness/
  Toolchain/Provider/Policy CRDs. Builder/registry architecture and golden-file
  unit tests are solid; admission and several runtime paths have never executed
  for real (e2e writes statuses by hand or disables hardening via
  `testutil.E2ESecurityOverrides`).

## Research Findings

### agent-cli-go

- **Docker proxy egress is fail-open.** `internal/isolation/docker/network.go:18`
  maps `proxy` to `--network bridge`; restriction is only advisory
  `HTTP_PROXY`/`HTTPS_PROXY` env (`docker.go:200`). With `AGENT_SANDBOX_PROXY`
  unset, `internal/cmd/run.go` `proxyEnv()` returns nil without error — plain
  bridge, yet `RequireEgressRestricted` is satisfied. Bwrap fails closed
  (`--unshare-net`); docker does not.
- **`HidesHost=true` while credentials are bind-mounted.**
  `shared-agent-go/pkg/isolation/types.go:27-41` (`UserConfigPaths`) mounts
  `~/.ssh`, `.npmrc`, `.config`, `.claude.json`, `.cargo`, `.local`, `.npm`
  into the sandbox (`bwrap.go:97-102`, `docker.go:123-127`) while both isolators
  report `HidesHost=true`, granting `RequireHostToolsUnreachable` with host
  tools and tokens literally reachable — and the agent runs with
  `--permission-mode bypassPermissions` (`agents/claude.go:20`).
- **Worktree mode breaks both isolators** (reproduced live). `run.go:239` adds
  the source repo to rw `BindPaths` while `RunConfig.RepoDir` is separately
  ro-bound: docker rejects the duplicate mount at launch; bwrap silently
  overmounts ro with rw, defeating the documented read-only source guarantee.
- **Nix pin is not fail-closed.** `internal/provision/nix/resolve.go:44,76` uses
  `--no-write-lock-file` (warns and proceeds on dirty/missing lock); the
  erroring flag is `--no-update-lock-file`. `Provisioner.Pinned()` returns true
  unconditionally (`nix.go:74`).
- **Provider-credential errors swallowed.** `bwrap.go:180`, `docker.go:189`
  (`err == nil` guard), `none.go:67` (`_ =`) discard the missing-token error —
  `--provider glm` without its token silently runs on mounted `~/.claude` creds.
- **Secrets in argv.** Tokens pass via `--setenv K V` (bwrap), `-e K=V` (docker),
  and are baked into the tmux command string (`tmux/env.go:47-70`), all readable
  via `/proc/*/cmdline` / `docker inspect`; tmux also serializes the entire host
  environ.
- **Composition-root invariant drift.** `internal/cmd/run.go:15` imports the
  concrete `provision/nix` package (for `SourceFromFlags`); the architecture
  test (`architecture_test.go:172`) only bans constructors, so imports drift
  invisibly.
- **Version misreport.** `internal/cmd/root.go:7` hardcodes `0.1.0`; platform
  builds pass only `-s -w` ldflags (`agent-cli-go-*/moon.yml:14`), so 0.1.31
  binaries report 0.1.0; `test/integration/run_test.go:60` asserts the stale
  value.
- **Only 2 of 4 guarantees CLI-reachable.** `run.go:132-137` wires one flag to
  `RequirePinnedProvision`+`RequireHostToolsUnreachable`; no flag sets
  `RequireEgressRestricted` or `RequireKernelIsolation`.
- **Dead surface.** `FileConfig.HomeDir/BindPaths/RoBindPaths/CustomEnv` parsed,
  never consumed; `LoadDefaultConfig` uncalled; default config path still named
  `sandboxed-claude`; `--debug` flag never read; `TerminalConfig.AttachExisting`
  / `IsInside` unused; `buildShellCommand` duplicated (`isolation/shared/
  spawn.go:45`, `terminal/tmux/tmux.go:270`); env-merge duplicated where
  `envutil.MergeEnvMaps` exists unused.

### agent-operator-go

- **Admission layer not deployable.** `cmd/operator/main.go:79` registers only
  `AgentRunWebhook` (four other `*_webhook.go` never wired); no
  `ValidatingWebhookConfiguration`/`MutatingWebhookConfiguration`, webhook
  Service, or cert mount exists anywhere under `config/` — the API server never
  calls any webhook. Registering one makes controller-runtime start a webhook
  server that needs certs at `/tmp/k8s-webhook-server/serving-certs`, which
  `config/manager/manager.yaml` doesn't mount → deployed operator likely
  crash-loops. E2e runs controllers in-process and never registers webhooks.
- **Policy fails open.** `internal/webhook/agentrun_webhook.go:132-135`: an
  errored `AgentPolicyList` lookup admits with a warning; only `Items[0]` is
  enforced (a weaker policy sorting first shadows a stricter one).
- **Inline-harness policy bypass.** `enforcePolicy` checks only `run.Spec.*`,
  but `internal/resolver/defaults.go:29-45` later merges run-author-controlled
  `spec.harness.default{SecurityContext,PodSecurityContext,Image,NetworkPolicy}`
  — bypassing `RequireSecurityContext`, `RequireNetworkPolicy`, and
  `AllowedImages` (toolchain image at `agentrun_controller.go:169-171` is also
  unchecked). `RequireSecurityContext` also ignores
  `ReadOnlyRootFilesystem=false` and `Capabilities.Add`.
- **Resolution errors silently degrade.** `agentrun_controller.go:82-108` (and
  246-270): harness/toolchain `Get` errors only log and proceed — dropping the
  sandboxed runtimeClass (→ default runc) and the pinned image for an untrusted
  pod. Referenced `AgentToolchain` CRDs are never held to pinning rules
  (`agenttoolchain_webhook.go:37-54` validates only shell metachars).
- **Secrets as pod-spec literals.** `internal/provider/provider.go:52-59` copies
  the token into `EnvVar.Value` instead of `ValueFrom.SecretKeyRef` (readable
  via Job get; lands in etcd/audit logs); token only injected when
  `ANTHROPIC_BASE_URL` is present — provider with `authTokenSecretRef` but no
  base URL silently gets no credential.
- **Paths that have never run.** Workspace-init hardcodes root-running
  `alpine/git:latest` under `RunAsNonRoot=true` (kubelet refuses;
  `agentworkspace_e2e_test.go:59-84` writes Job status by hand; image also
  can't run the jj strategy). Default `node:trixie-slim` pod can't start under
  shipped hardening (e2e disables it via `E2ESecurityOverrides`,
  `fixtures.go:291-300`). Claude harness passes nonexistent `--prompt`
  (`harness/claude/claude.go:18`); opencode never receives the prompt (TUI
  hangs until ActiveDeadline). Writable-HOME emptyDir only added for
  image-based toolchains (`hardening.go:85-103`).
- **Reconcile hygiene.** `AgentProvider` self-loops ~1/s on unconditional
  status writes (`agentprovider_controller.go:76-92`, no Secret watch, no
  requeue); NetworkPolicy `Create`+event fire every 10 s requeue
  (`agentrun_controller.go:147-159`, create-only, never updated); transient
  provider/workspace errors set terminal `Failed` (lines 98-102, top-of-loop
  guard 61-65); `resource.MustParse` on unvalidated `spec.workspace.storageSize`
  panics the reconciler (`wsshared/pvc.go:34`); ~140 lines duplicated between
  `reconcileStandalone`/`reconcileWithWorkspace` (75-209 vs 222-367).
- **Dead API surface.** `AgentPolicyEnforced.MaxResources` and all of
  `AgentPolicyDefaults` (`agentpolicy_types.go:30,41-51`) are never read;
  `ResolvedDefaults.NetworkPolicy` computed, never consumed;
  `internal/validator/repository.go` is a pure re-export shim; tracked
  `coverage.out`; `zap.Options{Development: true}` in production `main.go:35`;
  both webhooks hardcode the nix toolchain type instead of using the registry.

### Verified-healthy (no work needed)

- CLI policy engine and registry: method-agnostic `EnforcePolicy`, fail-closed
  `Select`, dense unit coverage; `network=host` correctly fails
  `RequireEgressRestricted`; bwrap/runc correctly fail `RequireKernelIsolation`.
- Release-line lockstep: all agent packages at 0.1.31 with matching CHANGELOG.
- Operator builder/leaf/registry architecture and golden-file tests.

## Key Decisions

1. **Docker proxy enforcement — fail closed now, enforce later.** Immediate:
   `docker` + `Network=proxy` REFUSES to run (capability
   `EgressRestricted=false` for docker proxy until real enforcement exists),
   and a missing proxy URL in proxy mode is an error everywhere. Follow-up in
   the same subplan: real enforcement via a docker `--internal` network with a
   proxy container attached to both the internal and egress networks.
   Alternative rejected: keeping bridge + env vars with a warning — advisory
   env is not a boundary and contradicts the fail-closed posture.
2. **Webhook certs via cert-manager** (standard kubebuilder path): Certificate +
   Issuer in `config/webhook/`, CA injection annotations, manager cert mount.
   Alternative rejected: hand-rolled self-signed bootstrap — more moving parts
   to maintain and rotate.
3. **Enforce policy on the RESOLVED effective spec.** The webhook performs the
   same defaults-merge as the controller (extract the merge into a shared
   resolver call) and validates the result; the controller re-checks
   fail-closed before Job creation (defense in depth against webhook outage —
   `failurePolicy: Fail` on the WebhookConfiguration regardless).
4. **Credential mounts become opt-in.** Split `UserConfigPaths` into
   tool caches (default) vs credentials (`.ssh`, `.npmrc`, `.claude.json`,
   `.config`) mounted only via explicit repeatable flag;
   `HidesHost`/`RequireHostToolsUnreachable` recomputed from the actual mount
   set, not a constant.
5. **Delete `AgentPolicyDefaults`/`MaxResources` now** (repo rule: remove unused
   code immediately); reintroduce only when designed end-to-end. Alternative
   (implement them) deferred — not required by the threat model.
6. **Secrets out of argv**: bwrap/docker pass token names (`-e KEY`, value from
   the launcher process env); tmux writes env to a 0600 tempfile sourced by the
   shell, never into the command string; operator uses `SecretKeyRef`.

## Proposed Approach

Seven subplans in five groups (see frontmatter). Operator track: make admission
deployable → make enforcement fail-closed → fix pod/runtime correctness → fix
reconcile hygiene. CLI track: close the egress hole → close host-reach/pinning
holes → cleanup. Tracks are independent; within each track subplans serialize
on shared files.

## Risk Assessment

- **Docker proxy refusal is a behavior break** for anyone relying on
  docker+proxy today — but that path currently provides no restriction, so the
  break is the fix. Mitigation: clear error naming the follow-up enforcement.
- **Credential-mount opt-in will break flows that silently depended on mounted
  `~/.claude.json`** — the provider-credential error surfacing (same subplan)
  turns that into an actionable message instead of a mystery.
- **Webhook `failurePolicy: Fail`** makes the operator namespace a hard
  dependency for AgentRun admission; acceptable for a security operator, but
  needs a PodDisruptionBudget and ≥1 replica note in `config/`.
- **E2e realism**: enabling webhooks + removing `E2ESecurityOverrides` will
  surface latent failures (that is the point) — budget iteration time; kind or
  envtest-with-webhooks needed in CI.
- Claude harness `--prompt` fix should be verified against the actual CLI
  version pinned in the toolchain image before changing.

## Proposed Child Plans

1. `operator-webhook-deployment` — register all five webhooks in main.go;
   `config/webhook/` (Service, WebhookConfigurations, cert-manager); manager
   cert mount; e2e job that runs with webhooks active; failurePolicy: Fail.
2. `operator-fail-closed-enforcement` — policy lookup error ⇒ deny; enforce all
   policies in namespace; enforce on resolved spec (close inline-harness
   bypass, check toolchain/default images, ReadOnlyRootFilesystem/Capabilities);
   harness/toolchain resolve errors ⇒ requeue not proceed; AgentToolchain CRD
   pinning validation; delete AgentPolicyDefaults/MaxResources.
3. `operator-pod-correctness` — SecretKeyRef injection (and decouple from
   ANTHROPIC_BASE_URL); workspace-init image/hardening/jj support; default
   image vs RunAsNonRoot resolution; claude/opencode command fixes; HOME
   emptyDir for all read-only-rootfs pods.
4. `operator-reconcile-hygiene` — AgentProvider status-write loop + Secret
   watch; netpol create-once/update-on-change; transient vs terminal failure
   classification; MustParse ⇒ validated parse; extract shared reconcile path
   (~140 dup lines); dead-code removal (validator shim, ResolvedDefaults
   .NetworkPolicy, coverage.out, zap Development, webhook toolchain hardcoding).
5. `cli-egress-fail-closed` — docker proxy ⇒ refuse (capability false) then
   internal-network+proxy-container enforcement; missing proxy URL ⇒ error;
   wire `--require-egress-restricted` and `--require-kernel-isolation` flags.
6. `cli-host-reach-and-pinning` — split UserConfigPaths, credential mounts
   opt-in, recompute HidesHost; fix worktree double-bind (single ro bind);
   `--no-update-lock-file` + honest `Pinned()`; surface provider-cred errors;
   secrets out of argv/tmux strings.
7. `cli-cleanup` — version `-X` ldflags in platform builds + fix integration
   test; move `SourceFromFlags` out of the concrete plugin (tighten
   architecture test to imports); remove dead config/flags/terminal surface;
   dedupe buildShellCommand/env merge; bwrap `/bin` symlinks or trim PATH.

## Success Criteria

- All four `Require*` guarantees are requestable from the CLI and refuse to run
  when unmet; no code path grants a capability the runtime doesn't enforce.
- Operator deploys from `config/default` with all webhooks admitted through the
  API server (verified in e2e with `failurePolicy: Fail`), and e2e passes with
  `E2ESecurityOverrides` deleted and no hand-written statuses.
- No secret value appears in argv, tmux command strings, or pod-spec literals
  (grep + `docker inspect`/Job-manifest assertions in tests).
- Worktree mode works under bwrap and docker with a read-only source repo
  (integration test).
- Released binaries report the release version.
- `go vet`, `gofmt`, unit + integration + e2e suites green in both projects;
  architecture tests tightened to imports pass.

## Estimated Effort

- Group 1: 1-2 days (webhook deployment is mostly manifests + e2e wiring;
  docker refusal is small).
- Groups 2-3: 2-3 days (enforcement-point move + pod correctness carry the
  most test surface).
- Groups 4-5: 1-2 days each.

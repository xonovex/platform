---
type: plan
has_subplans: false
parent_plan: plans/agent-security-hardening.md
parallel_group: 3
status: pending
dependencies:
  plans:
    - 03-operator-fail-closed-enforcement
  files:
    - packages/agent/agent-operator-go/internal/provider/provider.go
    - packages/agent/agent-operator-go/internal/controller/agentworkspace_controller.go
    - packages/agent/agent-operator-go/internal/isolation/shared/**
    - packages/agent/agent-operator-go/internal/network/shared/**
    - packages/agent/agent-operator-go/internal/harness/**
    - packages/agent/agent-operator-go/test/**
skills_to_consult:
  - kubernetes-guide
  - docker-guide
  - testing-guide
  - code-quality-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 04 — Operator: Pods That Actually Start, Secrets That Don't Leak

## Objective

Fix the pod-level defects that mean several documented paths have never run:
secrets injected as plaintext pod-spec values, a workspace-init Job that the
kubelet refuses to start, a default image incompatible with the shipped
hardening, wrong agent commands for both harnesses, and e2e fixtures that
disable exactly the hardening they should prove.

## Context (read this first — no other context is assumed)

File:line references are anchors as of `main` @ `a6f765e2` (2026-07-02) and
will drift as earlier subplans land — always read the cited file before
editing; if a line reference doesn't match, locate the named construct
instead.

`packages/agent/agent-operator-go` launches AI-agent workloads as hardened
Jobs. Hardening builders live in `internal/isolation/shared/`
(`hardening.go`, `security.go`, `container.go`, `job.go`) and
`internal/network/shared/networkpolicy.go` — the operator mirrors the CLI's
per-axis package layout, there is no `builder` package.
Harness command construction is `internal/harness/claude/claude.go` and
`internal/harness/opencode/opencode.go`. Provider credentials come from
`internal/provider/provider.go`. Workspace preparation (git clone init Job) is
`internal/controller/agentworkspace_controller.go` + the `wsshared`/`job.go`
builders.

Current defects, verified against source:

1. **Secrets as pod-spec literals.** `internal/provider/provider.go:52-59`
   reads the provider Secret and injects the token as a literal
   `EnvVar.Value` (via `container.go:12-24`). Anyone with `get` on Jobs/Pods
   reads the token; it also lands in etcd and audit logs inside the Job
   manifest. Bonus defect: the token is only injected when
   `ANTHROPIC_BASE_URL` is present in the env map — a provider with
   `authTokenSecretRef` but no base URL silently gets NO credential.
2. **Workspace-init Job cannot start.**
   `agentworkspace_controller.go:99` hardcodes `image := "alpine/git:latest"`
   (mutable tag, runs as root, lacks the `jj` binary the jujutsu strategy
   needs — see `jj.go:36` `jj git init --colocate`).
   `BuildWorkspaceInitJob` (`job.go:121-170`) applies `RunAsNonRoot=true` +
   `ReadOnlyRootFilesystem=true` to that root image, so the kubelet refuses it
   (`CreateContainerConfigError`), and it skips ALL pod hardening: default
   ServiceAccount with automounted token, no NetworkPolicy, no resource
   limits. `test/e2e/agentworkspace_e2e_test.go:59-84` writes the init Job's
   status BY HAND, so this path has never executed for real.
3. **Default pod cannot start under shipped hardening.** The default image
   `node:trixie-slim` runs as root; `RunAsNonRoot=true` without a numeric
   `runAsUser` ⇒ kubelet rejects the container.
   `test/testutil/fixtures.go:291-300` (`E2ESecurityOverrides`) disables
   `RunAsNonRoot` and `ReadOnlyRootFilesystem` so e2e passes — the tests
   bypass the advertised hardening.
4. **Writable HOME only for toolchain pods.** `hardening.go:85-103` adds the
   HOME/XDG `emptyDir`s only for image-based toolchains; a plain run with
   read-only rootfs has no writable HOME for `~/.claude*`.
5. **Harness commands are wrong.** `internal/harness/claude/claude.go:18`
   emits `--print --prompt <p>`; the Claude Code CLI has no `--prompt` flag
   (the prompt is a positional argument after flags). `container_test.go:103`
   merely locks in the wrong args. `internal/harness/opencode/opencode.go:14-25`
   never passes `run.Spec.Prompt` at all — opencode starts its interactive TUI
   inside a Job and hangs until ActiveDeadline.

## Tasks

1. **SecretKeyRef injection.** `provider.go:52-59` + `container.go:12-24`:
   inject the token as
   `EnvVar{Name: ..., ValueFrom: &EnvVarSource{SecretKeyRef: ...}}` — the
   operator should never read the secret VALUE at all for injection (drop the
   Secret read where its only purpose was the value; keep existence checks if
   the Ready condition needs them). Inject whenever `authTokenSecretRef` is
   set, independent of `ANTHROPIC_BASE_URL`. Update golden files/unit tests to
   assert the SecretKeyRef shape and its presence without a base URL.
2. **Workspace-init correctness + hardening.**
   - Replace the hardcoded `alpine/git:latest` with a digest- or
     version-pinned image that runs as uid 1000 and contains `git` (and `jj`
     for the jujutsu strategy — if no single small image fits, make the init
     image a field on the workspace spec/toolchain with a pinned default and
     document the jj requirement; check how `nix/agent-env.nix` provisions
     tools before inventing a new image).
   - Route the init Job through the SAME hardening bundle as agent pods:
     zero-RBAC ServiceAccount + `automountServiceAccountToken=false`, resource
     requests/limits, NetworkPolicy that allows only git egress, and a
     security context consistent with the chosen image (uid 1000, writable
     workspace mount, read-only rootfs with `emptyDir` for `/tmp`).
   - Delete the hand-written status in
     `test/e2e/agentworkspace_e2e_test.go:59-84`; the e2e must clone a real
     (local/file-based or in-cluster) repo end-to-end.
3. **Default image starts under hardening.** Set explicit
   `runAsUser: 1000` / `runAsGroup: 1000` in
   `DefaultContainerSecurityContext` (`security.go:28-50`) — `node` images
   ship the `node` user as uid 1000, matching the existing `fsGroup: 1000`.
   Verify in e2e that a default run reaches Running without overrides.
4. **HOME emptyDir for every read-only-rootfs pod.** Move the HOME/XDG
   emptyDir logic (`hardening.go:85-103`) out of the toolchain-only branch so
   any pod with `readOnlyRootFilesystem: true` gets writable HOME/XDG.
5. **Fix harness commands.**
   - claude: `--print` with the prompt as the positional argument. VERIFY
     against the actual CLI version pinned in the toolchain image before
     changing (run `claude --help` in that image or check its docs); update
     `container_test.go:103`.
   - opencode: pass the prompt through its non-interactive mode (check the
     opencode CLI: `opencode run <prompt>` or equivalent — verify, don't
     guess). A Job must never launch a TUI.
6. **Delete `E2ESecurityOverrides`.** Remove `fixtures.go:291-300` and every
   use; e2e runs with production hardening. Failures this surfaces are defects
   to fix in this subplan, not to re-suppress.

## Validation Steps

Prerequisites: Go toolchain is nix-managed — if `go` is missing, run inside
`nix develop`. Integration/e2e need envtest binaries (see subplan 01's note).
IMPORTANT: envtest is API-server + etcd only — it has NO kubelet, so pods
never actually run there. The "Job runs to completion" / "pod reaches
Running" assertions in this subplan require a real cluster (kind, or the
infrastructure behind the existing `go-test-e2e-gvisor`/`go-test-e2e-kata`
moon tasks — read how those are provisioned and reuse it). Do not fake
statuses to make envtest pass; that is the defect this subplan removes.

```bash
npx moon run agent-operator-go:go-build
npx moon run agent-operator-go:go-lint
npx moon run agent-operator-go:go-test
npx moon run agent-operator-go:go-test-integration
npx moon run agent-operator-go:go-test-e2e
```

## Success Criteria

- [ ] No secret value appears in any built Job manifest (unit test greps the
      serialized Job for the token; only SecretKeyRef present).
- [ ] Token injected when `authTokenSecretRef` is set without a base URL.
- [ ] Workspace-init Job runs to completion in e2e against a real repo, fully
      hardened (SA token off, netpol, limits), including the jj strategy.
- [ ] Default AgentRun pod reaches Running with production hardening; the
      `E2ESecurityOverrides` fixture no longer exists.
- [ ] claude/opencode container args verified against the real CLIs.
- [ ] All validation commands pass.

## Files Modified/Created

- Modified: `internal/provider/provider.go`,
  `internal/isolation/shared/{container,security,hardening,job}.go`,
  `internal/controller/agentworkspace_controller.go`,
  `internal/harness/claude/claude.go`, `internal/harness/opencode/opencode.go`,
  `test/testutil/fixtures.go`, `test/e2e/agentworkspace_e2e_test.go`, golden
  files/unit tests alongside.

## Dependencies

Depends on `03-operator-fail-closed-enforcement`. Serializes with
`05-operator-reconcile-hygiene` (both edit `internal/controller/*`) — any
order.

## Estimated Duration

1–1.5 days.

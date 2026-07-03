---
type: plan
has_subplans: false
parent_plan: plans/agent-security-hardening.md
parallel_group: 2
status: pending
dependencies:
  plans:
    - 01-operator-webhook-deployment
  files:
    - packages/agent/agent-operator-go/internal/webhook/**
    - packages/agent/agent-operator-go/internal/controller/agentrun_controller.go
    - packages/agent/agent-operator-go/internal/resolver/defaults.go
    - packages/agent/agent-operator-go/api/v1alpha1/agentpolicy_types.go
skills_to_consult:
  - kubernetes-guide
  - testing-guide
  - code-quality-guide
  - fp-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 03 — Operator: Fail-Closed Policy and Resolution

## Objective

Make AgentPolicy enforcement fail closed and un-bypassable: deny on lookup
errors, enforce every policy in the namespace, validate the RESOLVED effective
pod spec (not just the raw AgentRun spec), never proceed when a referenced
harness/toolchain can't be fetched, hold referenced AgentToolchains to the same
pinning rules as inline ones, and delete dead policy API surface.

## Context (read this first — no other context is assumed)

File:line references are anchors as of `main` @ `a6f765e2` (2026-07-02) and
will drift as earlier subplans land — always read the cited file before
editing; if a line reference doesn't match, locate the named construct
instead.

`packages/agent/agent-operator-go` runs untrusted AI-agent workloads as Jobs.
Its threat model demands fail-CLOSED behavior: when a requested guarantee can't
be established, refuse — never silently degrade. `AgentPolicy` is a namespaced
CRD an admin creates to force hardening (e.g. `RequireSecurityContext`,
`RequireNetworkPolicy`, `AllowedImages`, `AllowedRuntimeClassNames`).
Admission is `internal/webhook/agentrun_webhook.go`; the controller is
`internal/controller/agentrun_controller.go`; defaults from a referenced or
inline `AgentHarness` are merged in `internal/resolver/defaults.go`.

Current defects, verified against source:

1. **Policy lookup fails open.** `agentrun_webhook.go:132-135`: if listing
   `AgentPolicy` errors (API hiccup, cache not synced, RBAC), the run is
   ADMITTED with only a warning. A transient apiserver error waives namespace
   security policy.
2. **Only `Items[0]` enforced** (`agentrun_webhook.go:137`): with two policies
   in a namespace, whichever sorts first wins; a weaker one shadows a stricter
   one.
3. **Inline-harness bypass.** `enforcePolicy` (`agentrun_webhook.go:170-235`)
   inspects only `run.Spec.{RuntimeClassName,SecurityContext,Image,
   NetworkPolicy}`. But AFTER admission, `internal/resolver/defaults.go:29-45`
   merges `harness.Spec.Default{SecurityContext,PodSecurityContext,Image,
   NetworkPolicy}` into the effective run — including from an INLINE
   `run.Spec.Harness` that the run author fully controls. Concrete bypasses:
   - `RequireSecurityContext`: put `allowPrivilegeEscalation: true` /
     `runAsNonRoot: false` in `spec.harness.defaultSecurityContext`; the
     webhook sees `run.Spec.SecurityContext == nil` and admits.
   - `RequireNetworkPolicy`: webhook checks only `run.Spec.NetworkPolicy.
     Disabled` (line 208); `spec.harness.defaultNetworkPolicy.disabled: true`
     skips NetworkPolicy creation (`agentrun_controller.go:143-148`).
   - `AllowedImages`: only `run.Spec.Image` is checked; `spec.harness.
     defaultImage` and the toolchain image (which OVERRIDES everything at
     `agentrun_controller.go:169-171`) are never checked.
   - Even on the run itself, `RequireSecurityContext`
     (`agentrun_webhook.go:196-204`) ignores `readOnlyRootFilesystem: false`
     and `capabilities.add`.
4. **Resolution errors silently degrade.** `agentrun_controller.go:82-85,
   105-108` (and the workspace variant at 246-249, 267-270): a `Get` error on
   `HarnessRef`/`ToolchainRef` is only logged; reconcile proceeds with
   `harness=nil`/`tc=nil`. Consequences: no `DefaultRuntimeClassName` ⇒ pod
   runs under default runc (no kernel isolation); pinned nix toolchain image
   silently falls back to the unpinned default image (`node:trixie-slim`).
5. **Referenced AgentToolchains unvalidated.** `validateNixSpec`
   (`agentrun_webhook.go:149-168`) enforces pinning for INLINE toolchains only.
   `agenttoolchain_webhook.go:37-54` validates only shell metacharacters — a
   referenced AgentToolchain without a pinned `image` yields
   `Toolchain.Image()==""` and the controller silently uses the unpinned
   default. Both webhooks also hardcode the `nix` toolchain type instead of
   dispatching through the `plugins`/`builder.ResolveToolchain` registry.
6. **Dead policy API.** `api/v1alpha1/agentpolicy_types.go:30`
   (`AgentPolicyEnforced.MaxResources`) and lines 41-51 (the entire
   `AgentPolicyDefaults` block) are read by nothing (repo-wide grep confirms).
   Parent-plan decision: DELETE them now (repo rule: remove unused code
   immediately); reintroduce only when designed end-to-end.

## Tasks

1. **Deny on policy lookup error.** `agentrun_webhook.go:132-135`: return the
   error (deny admission) instead of appending a warning. Add a unit test with
   an erroring fake client.
2. **Enforce ALL policies.** Iterate `policyList.Items` and run
   `enforcePolicy` against each; every policy must pass (strictest-wins by
   conjunction). Deterministic error message naming the failing policy.
3. **Enforce on the resolved effective spec.**
   - Extract the defaults-merge from `internal/resolver/defaults.go:29-45`
     into a pure function the webhook can call (resolver already imports
     nothing controller-specific; keep it side-effect-free per the repo's FP
     style).
   - In the webhook: resolve the harness (referenced via client Get — a Get
     error here also DENIES — or inline) and the toolchain, apply the merge,
     then run `enforcePolicy` on the RESULT. Check: securityContext (including
     `readOnlyRootFilesystem: false` and `capabilities.add`), podSecurityContext,
     networkPolicy.disabled, ALL images (run image, harness default image,
     toolchain image), runtimeClassName against `AllowedRuntimeClassNames`.
   - Unit tests reproducing each bypass in the Context section, asserting
     denial.
4. **Controller re-check (defense in depth).** In `agentrun_controller.go`,
   immediately before Job creation, run the same enforcement against the final
   resolved spec and fail the run if it violates policy. Rationale: webhooks
   can be deleted/bypassed by cluster admins; the controller is the last line.
5. **Never proceed on resolution errors.** All four sites
   (`agentrun_controller.go:82-85, 105-108, 246-249, 267-270`): return the
   error so controller-runtime requeues with backoff. Do not proceed with nil.
   NotFound is also an error here (the ref may appear later; requeue handles
   it) — the run must never launch un-hardened.
6. **Validate referenced AgentToolchains.** Move/reuse the pinning validation
   (`validateNixSpec`) in `agenttoolchain_webhook.go` so a stored AgentToolchain
   must satisfy the same rules as an inline one. Replace the hardcoded nix
   type checks in BOTH webhooks with dispatch through the toolchain registry
   — `ResolveToolchain` in `internal/plugins/plugins.go:30` (the CR-fetch
   variant with the same name in `internal/resolver/toolchain.go:13` is a
   different function) — so adding a toolchain type doesn't require webhook
   edits.
7. **Delete dead API surface.** Remove `MaxResources` and `AgentPolicyDefaults`
   from `agentpolicy_types.go`, regenerate deepcopy + CRD manifests (find the
   generation command in the project's Makefile/moon tasks or mirror how
   `config/crd` was produced), and fix any compile fallout. No deprecation
   markers — remove outright (repo rule).

## Validation Steps

Prerequisites: Go toolchain is nix-managed — if `go` is missing, run inside
`nix develop`. Integration/e2e need envtest binaries (see subplan 01's note).

```bash
npx moon run agent-operator-go:go-build
npx moon run agent-operator-go:go-lint
npx moon run agent-operator-go:go-test
npx moon run agent-operator-go:go-test-integration
npx moon run agent-operator-go:go-test-e2e   # webhooks active (subplan 01)
```

## Success Criteria

- [ ] Policy lookup error ⇒ admission DENIED (unit test with erroring client).
- [ ] Multiple policies all enforced; unit test where policy #2 is stricter.
- [ ] Every bypass in Context item 3 has a test that now results in denial.
- [ ] Controller refuses (requeues) on any harness/toolchain resolve error; no
      code path reaches Job creation with nil harness/toolchain substitutes.
- [ ] Referenced AgentToolchain without pinning is rejected at admission.
- [ ] `MaxResources`/`AgentPolicyDefaults` gone from types, deepcopy, CRDs.
- [ ] e2e (with webhooks from subplan 01) proves one bypass-denial end-to-end.
- [ ] All validation commands pass.

## Files Modified/Created

- Modified: `internal/webhook/agentrun_webhook.go`,
  `internal/webhook/agenttoolchain_webhook.go`,
  `internal/controller/agentrun_controller.go`,
  `internal/resolver/defaults.go`, `api/v1alpha1/agentpolicy_types.go`,
  `api/v1alpha1/zz_generated.deepcopy.go` (regenerated), `config/crd/*`
  (regenerated), tests alongside each.

## Dependencies

Depends on `01-operator-webhook-deployment` (webhooks must actually run for
e2e proof; also edits the same webhook files).

## Estimated Duration

1.5 days.

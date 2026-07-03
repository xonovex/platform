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
    - packages/agent/agent-operator-go/internal/controller/**
    - packages/agent/agent-operator-go/internal/resolver/defaults.go
    - packages/agent/agent-operator-go/internal/validator/repository.go
    - packages/agent/agent-operator-go/cmd/operator/main.go
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

# 05 — Operator: Reconcile Hygiene and Dead Code

## Objective

Stop the reconcile loops from churning (status self-loops, repeated Creates),
classify transient vs terminal failures correctly, remove a reconciler panic
on user input, deduplicate the two ~140-line reconcile paths, and delete dead
code per the repo rule "remove unused/deprecated code immediately".

## Context (read this first — no other context is assumed)

File:line references are anchors as of `main` @ `a6f765e2` (2026-07-02) and
will drift as earlier subplans land — always read the cited file before
editing; if a line reference doesn't match, locate the named construct
instead.

`packages/agent/agent-operator-go` is a controller-runtime v0.24.1 operator.
The AgentRun controller (`internal/controller/agentrun_controller.go`) has two
near-duplicate reconcile paths: `reconcileStandalone` (lines ~75-209) and
`reconcileWithWorkspace` (lines ~222-367). While a run is active it requeues
every 10 s (line ~407). A run whose `Status.Phase` is `Failed` is treated as
terminal by the top-of-loop guard (lines ~61-65) and never reconciled again.

Current defects, verified against source:

1. **AgentProvider controller self-loops.**
   `agentprovider_controller.go:76-92` unconditionally writes
   `Status.Conditions = []{{LastTransitionTime: metav1.Now(), ...}}` every
   reconcile; each status write triggers the next reconcile (no predicate, no
   transition guard) — permanent ~1 write/sec churn per provider. It also has
   no Watch on Secrets and no `RequeueAfter`, so a provider whose secret
   appears later stays `Ready=false` until something else touches it.
2. **NetworkPolicy Create + event every requeue.** The netpol block
   (`agentrun_controller.go:147-159`, duplicated at ~299-311) sits OUTSIDE the
   `Status.JobName == ""` first-pass guard, so an hour-long run attempts ~360
   failed Creates and emits ~360 duplicate `NetworkPolicyCreated` events. The
   policy is also never updated if the desired spec changes (create-only).
3. **Transient errors are terminal.** `agentrun_controller.go:98-102`
   (provider), ~229-233 / ~260-264 (workspace variants): ANY error from
   `ResolveProvider`/`ResolveWorkspace` — including connection timeouts — sets
   phase `Failed`, which the guard at ~61-65 then treats as terminal forever.
   One flaky secret read permanently kills the run.
4. **Reconciler panic on malformed input.** `BuildPVC`
   (`internal/workspace/shared/pvc.go:34`, package `shared`) calls
   `resource.MustParse(storageSize)` on `run.Spec.Workspace.StorageSize`,
   which the AgentRun webhook never validates. `storageSize: "10Gib"`
   (invalid suffix) panics the reconcile loop on every retry.
5. **~140 lines duplicated** between `reconcileStandalone` and
   `reconcileWithWorkspace`: harness/provider/toolchain resolution, netpol
   creation, job creation, status watching. (Subplan 03 already touches the
   resolution blocks; coordinate — this subplan owns the extraction.)
6. **Dead/inconsistent code** (repo rule: remove immediately):
   - `ResolvedDefaults.NetworkPolicy` computed but never read
     (`resolver/defaults.go:19,54-57,67` — the controller re-derives it).
   - `internal/validator/repository.go` is a pure re-export shim (repo rule:
     no re-exports) — inline it at call sites and delete.
   - `var _ = resource.Quantity{}` unused-import hack
     (`api/v1alpha1/agentpolicy_types.go:9-10`) — may already fall to
     subplan 03's type deletions; verify.
   - `coverage.out` is a git-tracked build artifact — delete and gitignore.
   - `zap.Options{Development: true}` hardcoded in `cmd/operator/main.go:35` —
     production default must be non-development; expose via flag.

## Tasks

1. **AgentProvider controller.**
   - Use `meta.SetStatusCondition` (or equivalent guard) so
     `LastTransitionTime` and the write only happen when the condition
     actually changes; skip the update when nothing changed.
   - Add a Watch on Secrets mapping to AgentProviders that reference them
     (`builder.Watches` + a `handler.MapFunc` doing a field/index lookup), so
     a late-created secret flips Ready without manual touches. If a watch is
     disproportionate, a `RequeueAfter` on not-ready is the minimum.
   - Unit test: two consecutive reconciles with unchanged input produce ONE
     status write (use a counting fake client or interceptor).
2. **NetworkPolicy lifecycle.** Replace the raw `Create` with
   create-or-update semantics (`controllerutil.CreateOrUpdate`), inside the
   correct phase guard so steady-state requeues don't attempt writes; emit the
   event only when the object was actually created. Apply to BOTH reconcile
   paths (or once, after task 4's extraction).
3. **Failure classification.** Transient errors (any client/API error during
   resolution) ⇒ return the error for requeue-with-backoff, do NOT set
   `Failed`. Reserve `Failed` for genuinely terminal outcomes (policy
   violation from subplan 03's re-check, Job exceeded backoff/deadline).
   Review every `Phase = Failed` assignment in the controller against this
   rule.
4. **Extract the shared reconcile path.** Factor the common sequence
   (resolve harness/provider/toolchain → build hardening → netpol → job →
   status) into pure helper functions taking explicit inputs (repo FP style:
   no shared mutable state), leaving the standalone/workspace variants as thin
   orchestrations of the shared parts. Findings 2 and 3 must end up fixed in
   ONE place each.
5. **Panic removal.** `internal/workspace/shared/pvc.go:34`:
   `resource.ParseQuantity` with an explicit
   error return; also validate `storageSize` in the AgentRun webhook (subplan
   03 owns webhook structure; add the field validation here if 03 has merged,
   otherwise in the controller path). Unit test with `"10Gib"`.
6. **Dead code deletion.** Each bullet from Context item 6: delete, inline, or
   flag-expose as described; `git rm coverage.out` + add to the appropriate
   `.gitignore`.

## Validation Steps

Prerequisites: Go toolchain is nix-managed — if `go` is missing, run inside
`nix develop`. Integration/e2e need envtest binaries (see subplan 01's note).

```bash
npx moon run agent-operator-go:go-build
npx moon run agent-operator-go:go-lint
npx moon run agent-operator-go:go-test
npx moon run agent-operator-go:go-test-integration
npx moon run agent-operator-go:go-test-e2e
```

## Success Criteria

- [ ] Steady-state reconcile of an unchanged AgentProvider performs zero
      writes (asserted by test).
- [ ] Long-running AgentRun performs exactly one NetworkPolicy create and one
      event; spec drift is corrected (CreateOrUpdate test).
- [ ] A transiently-failing provider/workspace resolution recovers on retry
      (test with a fake client failing once).
- [ ] Invalid `storageSize` yields a run error, not a panic.
- [ ] `reconcileStandalone`/`reconcileWithWorkspace` share the extracted core;
      the duplicated blocks are gone.
- [ ] Dead code from Context item 6 removed; `rg "validator/repository"` and
      `rg "ResolvedDefaults" | grep NetworkPolicy` come back empty.
- [ ] All validation commands pass.

## Files Modified/Created

- Modified: `internal/controller/agentrun_controller.go`,
  `internal/controller/agentprovider_controller.go`,
  `internal/workspace/shared/pvc.go`,
  `internal/resolver/defaults.go`, `cmd/operator/main.go`, tests alongside.
- Deleted: `internal/validator/repository.go`, `coverage.out`.

## Dependencies

Depends on `03-operator-fail-closed-enforcement`. Serializes with
`04-operator-pod-correctness` (both edit `internal/controller/*`) — any order.

## Estimated Duration

1 day.

---
type: plan
has_subplans: false
parent_plan: ../runtime-enforcement-completion.md
parallel_group: 2
status: pending
dependencies:
  plans:
    - governance-decision-point-fail-closed
  files:
    - packages/agent/agent-operator-go/api/v1alpha1/*.go
    - packages/agent/agent-operator-go/internal/controller/*.go
    - packages/agent/agent-operator-go/internal/webhook/*.go
    - packages/skill/skill-agent-governance/agent-governance-guide/references/autonomy.md
skills_to_consult:
  - plan-guide
  - kubernetes-guide
  - microkernel-pattern-guide
  - hexagonal-pattern-guide
  - testing-guide
  - connascence-guide
validation:
  type_check: not_run
  lint: not_run
  build: not_run
  tests: not_run
updated: "2026-07-17"
---

# A3 Unattended-Orchestration Runtime Coupled to the Oversight Invariant

## Objective

Build the A3 components `autonomy.md` names as targets — non-human triggers, admission control,
escalation routing, and per-run provenance — so unattended orchestration is real and cannot run without
the oversight it depends on. A3 is gated on the Phase 1 verdict being enforced at a non-bypassable
point (the admission webhook), and demotes to the highest level whose oversight still holds when a
control stops producing evidence.

## Tasks

1. Add a **trigger surface** to the operator per the settled Decisions below: two CRDs in
   `packages/agent/agent-operator-go/api/v1alpha1/` with reconcilers in
   `packages/agent/agent-operator-go/internal/controller/` that create `AgentRun`s —
   `AgentSchedule` (time-based CronJob analog: cron expression, `AgentRun` template, `suspend`,
   concurrency policy) and `AgentTrigger` (event-based: declared endpoint + bearer token from a
   Secret, served by an HTTP receiver hosted as a runnable in the existing manager,
   NetworkPolicy-gated; an authenticated POST creates the templated run, unauthenticated or
   unmatched posts create nothing) — so "schedules, sensors, or humans trigger runs"
   (`autonomy.md:24`) is real, not only human-initiated reconcile; hand-maintain both CRD
   manifests and `zz_generated.deepcopy.go` per the controller-gen constraint
   (`agent-operator-go/AGENTS.md`).
2. Add **per-run provenance**: an accountable-owner field on `AgentRunSpec`
   (`api/v1alpha1/agentrun_types.go`, which today has no owner field) that admission requires for
   triggered runs, and a run journal recording model/provider/prompt/tools/granted-permissions (the
   AIBOM described in `packages/agent/AGENTS.md`) written by the reconciler in `internal/controller/`.
3. Add an **escalation router**: an unattended run needing a human raises a bounded escalation with a
   declared window and a safe default (pause/abandon) per `autonomy.md:59-71`; an unanswered escalation
   falls back to the safe default on expiry and records the outcome. Encode "A3 is unavailable where the
   route has no accountable recipient" (`autonomy.md:69`) as a fail-closed admission check in
   `internal/webhook/`.
4. **Couple autonomy to oversight**: gate the ability to run at A3 on the Phase 1 verdict being enforced
   at a non-bypassable point (the admission webhook from Phase 1), protected targets, escalation routing
   present, and per-run provenance recorded (`autonomy.md:53` coupling table); demote to the highest
   level whose oversight still holds when a control stops producing evidence (`autonomy.md:55` demotion
   trigger).
5. Add tests under `internal/controller/` and `internal/webhook/`: a triggered run without an
   accountable escalation recipient is refused; an expired escalation takes its declared safe default; a
   run missing provenance is denied admission at A3; an authenticated `AgentTrigger` POST creates the
   templated run while an unauthenticated or unmatched POST creates nothing.
6. Update `autonomy.md` only where the runtime now backs a previously-aspirational claim (e.g. its
   note at line 9 that "`A3` ... triggers, admission control, and escalation routing are targets an
   adopter builds"), without overstating coverage beyond what the operator actually enforces.

## Acceptance criteria

- Both trigger paths — an `AgentSchedule` firing and an authenticated `AgentTrigger` POST — create an
  `AgentRun` with no human in the loop, each carrying an accountable owner and a per-run provenance
  journal, asserted by `go test ./...`; an unauthenticated or unmatched POST creates nothing.
- An unattended run that needs a human raises an escalation with a window and safe default; on expiry the
  safe default is taken and recorded; a test proves silence never advances a gate.
- Admission refuses an A3 run when the governance verdict is not enforced at a non-bypassable point, when
  the escalation route has no accountable recipient, or when provenance is absent — each asserted by a
  webhook test.
- `go test ./...` in `packages/agent/agent-operator-go` (including the new CRD, reconciler, and webhook)
  passes.

## Dependencies

- **`governance-decision-point-fail-closed` (Phase 1, group 1) must land first.** The parent requires
  Phase 4 to "gate the ability to run at A3 on the Phase 1 verdict being enforced at a non-bypassable
  point (the admission webhook)"; the coupling check has nothing to gate on until Phase 1's verdict is
  enforced at admission. Phase 4 otherwise builds on the existing operator (reconcilers + webhooks
  already live), so it does **not** depend on Phase 2 or Phase 3.

This subplan shares parallel group 2 with `claude-code-native-hook-block` (Phase 2): both depend only on
Phase 1 and touch disjoint files (Phase 2 edits skill/harness assets; this edits operator Go), so they
run concurrently.

## Decisions (settled 2026-07-17)

- **Trigger CRD shape: both** (parent Decision 4). `AgentSchedule` (time-based CronJob analog) and
  `AgentTrigger` (event-based) ship together, each kept minimal, accepting the doubled
  hand-maintenance tax (controller-gen broken on Go 1.25+).
- **`AgentTrigger` v1 event source: authenticated webhook receiver** (parent Decision 5). A
  bearer-token-authenticated HTTP receiver hosted as a runnable in the existing operator manager
  (token from a Secret, NetworkPolicy-gated) — "sensor" means anything that can POST. Phase 5's
  drift detector reuses this receiver as its remediation path.

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

1. Add a **trigger surface** to the operator: a schedule/sensor CRD in
   `packages/agent/agent-operator-go/api/v1alpha1/` plus a reconciler in
   `packages/agent/agent-operator-go/internal/controller/` that creates `AgentRun`s, so "schedules,
   sensors, or humans trigger runs" (`autonomy.md:24`) is real, not only human-initiated reconcile;
   hand-maintain the CRD manifest and `zz_generated.deepcopy.go` per the controller-gen constraint
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
   run missing provenance is denied admission at A3.
6. Update `autonomy.md` only where the runtime now backs a previously-aspirational claim (e.g. its
   note at line 9 that "`A3` ... triggers, admission control, and escalation routing are targets an
   adopter builds"), without overstating coverage beyond what the operator actually enforces.

## Acceptance criteria

- A schedule/sensor creates an `AgentRun` with no human in the loop, and that run carries an accountable
  owner and a per-run provenance journal — both asserted by `go test ./...`.
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

## Open decisions

- **Trigger CRD shape.** The parent names the trigger surface as a "new `AgentSchedule`/`AgentTrigger`
  CRD" (Phase 4, Task 1) without deciding whether it is a single time-based `AgentSchedule`, a distinct
  event/sensor `AgentTrigger`, or both. The parent leaves the choice open; decide it before Task 1 given
  the manual CRD/deepcopy maintenance cost (controller-gen broken on Go 1.25+), and keep the chosen
  surface minimal. This subplan does not invent the answer.

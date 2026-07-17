---
type: plan
has_subplans: false
parent_plan: ../runtime-enforcement-completion.md
parallel_group: 4
status: pending
dependencies:
  plans:
    - governance-decision-point-fail-closed
    - claude-code-native-hook-block
    - a3-unattended-orchestration-runtime
  files:
    - packages/agent/agent-operator-go/internal/**/*.go
    - packages/agent/agent-operator-go/test/e2e/*.go
    - packages/command/command-workflow/commands/observe-run.md
    - packages/command/command-workflow/commands/incident-run.md
    - packages/command/command-workflow/commands/workflow-drift.md
skills_to_consult:
  - plan-guide
  - kubernetes-guide
  - testing-guide
  - shell-scripting-guide
  - connascence-guide
validation:
  type_check: not_run
  lint: not_run
  build: not_run
  tests: not_run
updated: "2026-07-17"
---

# Operational Proof: Drift Detection and Incident Containment on a Live Estate

## Objective

Wire `observe-run` / `incident-run` / `workflow-drift` to real telemetry so drift is actually detected
and an incident actually contained on a running estate. The runtime emits OpenTelemetry-compatible
signals, a drift detector recomputes effective autonomy against required oversight and demotes when a
control degrades, and incident containment pauses/kills a policy-breaching run — all demonstrated on a
live kind/e2e estate.

## Tasks

1. Emit **OpenTelemetry-compatible signals** from the runtime (operator reconciler + Phase 1 decision
   point + Phase 2 hook): per-run traces and verdict/enforcement metrics correlated by the run's
   correlation id, with content minimization (no prompts/secrets by default) per
   `composable-workflow-phases.md` plane 7. New telemetry seam under
   `packages/agent/agent-operator-go/internal/` and the decision service — today a grep for
   `opentelemetry|otel|telemetry|trace.Span` across `packages/agent/**/*.go` returns nothing.
2. Implement a **drift detector** fed by those signals that recomputes effective autonomy against
   required oversight (`autonomy.md` coupling table, line 53) and raises drift when a control stops
   producing evidence, fails open, or diverges from the applied reference — wiring the `workflow-drift`
   contract (`packages/command/command-workflow/commands/workflow-drift.md`) to live data instead of a
   prompt.
3. Implement **incident containment**: on a detected drift or anomaly, exercise a tested kill-switch /
   pause on the live run (building on the operator's Job/AgentRun lifecycle and the Phase 4 escalation
   router) and record the containment as evidence — wiring the `incident-run` contract
   (`packages/command/command-workflow/commands/incident-run.md`); wire `observe-run.md` to the emitted
   traces/metrics.
4. Add an **e2e proof** under `packages/agent/agent-operator-go/test/e2e/` (built with the existing
   `-tags=e2e` harness): induce drift by disabling a required oversight control and induce an incident
   with a policy-breaching run, then show detection + containment with evidence and effective-autonomy
   demotion.
5. Assert **no sensitive-content regression**: the telemetry seam does not log prompts/secrets by
   default, verified by an automated check over the emitted signals.

## Acceptance criteria

- The runtime emits OTel-compatible traces/metrics correlated per run without logging sensitive content
  by default (asserted by a test over the emitted signals).
- Disabling a required oversight control is detected as drift on the live estate and demotes the
  effective autonomy level, with a drift evidence record.
- A policy-breaching run is contained (paused/killed) by the incident path on the live estate, with a
  containment evidence record and no manual intervention.
- The e2e proof runs under the operator's existing `-tags=e2e` harness
  (`go test -tags=e2e ./test/e2e/`) and passes; `go test ./...` remains green.

## Dependencies

- **`governance-decision-point-fail-closed` (Phase 1) and `claude-code-native-hook-block` (Phase 2)
  must land first** — the parent requires OTel signals emitted "from the runtime (operator reconciler +
  **Phase 1 decision point** + **Phase 2 hook**)"; there is no verdict/enforcement telemetry to emit or
  correlate until both exist.
- **`a3-unattended-orchestration-runtime` (Phase 4) must land first** — incident containment builds "on
  the operator's Job/AgentRun lifecycle and **the escalation router from Phase 4**," and drift demotion
  recomputes the A3 oversight coupling Phase 4 establishes.
- This is the final group: the parent states "operational proof) depends on the runtime existing, so it
  is last," and "Phases 4 and 5 ... build only on a verdict that is already enforced at a non-bypassable
  point." It runs after group 3 (the runtime-real walking skeleton) so the full runtime is in place.

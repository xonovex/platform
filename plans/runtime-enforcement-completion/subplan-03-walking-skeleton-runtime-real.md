---
type: plan
has_subplans: false
parent_plan: ../runtime-enforcement-completion.md
parallel_group: 3
status: pending
dependencies:
  plans:
    - governance-decision-point-fail-closed
    - claude-code-native-hook-block
  files:
    - packages/skill/skill-agent-governance/agent-governance-guide/assets/walking-skeleton/*
    - packages/command/command-workflow/commands/*.md
    - packages/skill/skill-workflow/workflow-guide/scripts/*.mjs
skills_to_consult:
  - plan-guide
  - shell-scripting-guide
  - testing-guide
  - connascence-guide
validation:
  type_check: not_run
  lint: not_run
  build: not_run
  tests: not_run
updated: "2026-07-17"
---

# The Walking Skeleton Made Runtime-Real

## Objective

Replace the self-contained simulation with a real end-to-end run: an agent takes a task through the
lifecycle (discovery → … → acceptance → integration) with the Phase 1 decision point enforcing at each
governed gate and provider-native evidence recorded at each. The existing deterministic
`run-skeleton.sh` stays as the contract-shape check; a new runtime-real counterpart exercises the real
harness (Phase 2) plus the real decision point (Phase 1), and a governance violation halts the run.

## Tasks

1. Author a real end-to-end scenario driving the lead harness (the Phase 2 registered hook) through a
   minimal but genuine lifecycle, invoking the actual lifecycle commands/skill operations
   (`packages/command/command-workflow/commands/*.md`) rather than a bash mock, and gating the
   high-impact transitions (acceptance, integration) through the Phase 1 verdict.
2. At each gate, record a provider-native evidence record carrying the exact-revision reference and the
   verdict (reusing the Phase 1 verdict evidence shape), and assert the run **cannot advance** past a
   gate whose verdict is `deny`.
3. Add the runtime-real counterpart under
   `packages/skill/skill-agent-governance/agent-governance-guide/assets/walking-skeleton/` alongside the
   existing `run-skeleton.sh` / `guard.sh`, clearly labelled as **live** (real harness + real decision
   point) versus the existing **simulation**, and replace the hardcoded capability literal
   (`run-skeleton.sh:109`, `MATRIX='{"before-tool-use":...}'`) in the live path with the actual probed
   matrix from Phase 2.
4. Prove one negative path: inject a self-approved acceptance or an executor escalation and show it is
   **blocked mid-run** at the offending gate, producing an escalation/deny evidence record with the
   exact failure code.
5. Keep both paths runnable and green — the deterministic `run-skeleton.sh` still passes its
   `PASS/FAIL` tally (`run-skeleton.sh:136-137`), and the live counterpart passes its own checks — and
   document how a maintainer runs each.

## Acceptance criteria

- A single command drives an agent from discovery to integration with a governed gate enforced by the
  Phase 1 verdict at both acceptance and integration; the evidence trail shows one verdict record per
  gate tied to an exact revision.
- Injecting a governance violation (self-approval, out-of-scope emergency access, or executor
  escalation) halts the run at the offending gate and records the exact failure code — verified against
  the real harness + decision point, not simulated.
- The live counterpart is distinguishable in-repo from the simulation (distinct file names/labels), and
  both pass their checks on a clean checkout — the live counterpart as a maintainer-run probe per
  parent Decision 3; CI runs only the deterministic simulation.

## Dependencies

- **`governance-decision-point-fail-closed` (Phase 1, group 1) must land first** — the run gates
  "the high-impact transitions (acceptance, integration) through the Phase 1 verdict" (parent Phase 3),
  so the decision point must exist and fail closed before the skeleton can enforce at a gate.
- **`claude-code-native-hook-block` (Phase 2, group 2) must land first** — the parent orders the phases
  so that "Phase 3 proves them [Phase 1 and Phase 2] together on a real lifecycle," and Phase 3's first
  task is "Author a real end-to-end scenario **driving the lead harness (Phase 2)**." A runtime-real run
  needs Phase 2's registered, block-proven hook; without it the harness leg is still a simulation.

Because Phase 3 depends on Phase 2, it sits in group 3 (after group 2), not concurrent with
Phase 2 — matching the parent's `parallel_groups`.

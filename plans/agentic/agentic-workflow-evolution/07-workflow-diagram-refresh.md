---
type: plan
has_subplans: false
parent_plan: plans/agentic/agentic-workflow-evolution.md
parallel_group: 4
status: pending
dependencies:
  plans:
    - plans/agentic/agentic-workflow-evolution/03-workflow-next-command.md
    - plans/agentic/agentic-workflow-evolution/04-agent-cli-workflow-run.md
  files:
    - packages/diagram/diagram-agent-workflow/workflow-diagram.dot
    - packages/diagram/diagram-agent-workflow/moon.yml
    - packages/diagram/diagram-agent-workflow/*.png
skills_to_consult:
  - moon-guide
  - code-quality-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 07 — Diagram Refresh: Document the Landed Shape

## Objective

Bring `workflow-diagram.dot` in line with the workflow as it now exists
(post 01–04): named gates with policies, the run journal, the trigger
entry point, the three runners, loop caps, and the delivery hand-off;
render every diagram in the package via the moon task.

## Context (read this first — no other context is assumed)

Anchors as of `main` @ `2b276a7f` (2026-07-03). The change list below
was recorded during the 2026-07-03 architecture session; verify each
against the landed 01–04 reality (e.g. actual gate names from
`workflows/feature.yml`) before drawing.

1. `packages/diagram/diagram-agent-workflow/` contains
   `workflow-diagram.dot` (current manual workflow),
   `target-architecture.dot` and `maturity-ladder.dot` (created
   2026-07-03, already fleet-accurate), and rendered PNGs.
2. `moon.yml`'s `graph-build` already renders all three `.dot` files
   (split into graph-build-{workflow,target,maturity} on 2026-07-03).
   The 2026-07-03 lifecycle rename also already added plan-decide,
   plan-critique, and plan-revise nodes to workflow-diagram.dot — the
   gate/journal/runner refresh below still applies on top.
3. The diagram's conventions (legend, shapes, session markers) are
   established — extend, don't replace.

## Tasks

1. **Named gates**: replace each "(HITL or AI)" diamond with the gate
   identity + policy from `workflows/feature.yml` (e.g. "Gate:
   plan-review [manual]", "Gate: merge-review [assisted →
   plan-critique]"); add a gate legend entry (policy meanings).
2. **Journal + results**: add the run-journal artifact node
   (`plans/<plan>/runs/`) fed by validate/update steps; legend entry
   for the step-result contract.
3. **Second entry point**: "Trigger (sensor / schedule) →
   workflow-next" alongside "Run Agent Wrapper", marked as arriving
   paused at an entry gate (06's shape — draw as planned/dashed if 06
   has not landed when this executes).
4. **Runners**: extend the Setup cluster with the three runner variants
   (interactive session / `agent-cli workflow run` headless / AgentRun
   on the ci cluster) and the pinned-plugin-set note (decision 9);
   annotate the parallel-worktrees cluster with its cluster-side
   equivalent (AgentWorkspace).
5. **Loop caps + delivery**: put max-iteration labels on the dashed
   feedback edges (values from feature.yml); route "Feature Complete"
   into "PR → CI → GitOps (see target-architecture)" closing the sensor
   loop.
6. **Render everything**: verify `graph-build` still renders all three
   `.dot` files (wired 2026-07-03); regenerate and commit the PNGs;
   `dot -Tpng` exit 0 for all.

## Validation Steps

- `npx moon run diagram-agent-workflow:graph-build` renders all three
  diagrams (parent success criterion 7).
- Visual check of the rendered workflow-diagram.png: gates named with
  policies, journal node present, both entry points, runner variants,
  loop caps legible.
- Cross-check gate names/caps against `workflows/feature.yml` — the
  diagram must not contradict the definition (the definition is
  normative).

## Success Criteria

- [ ] workflow-diagram.dot shows named policy-carrying gates, journal,
      trigger entry point, three runners, loop caps, delivery hand-off.
- [ ] graph-build renders all three dots; PNGs committed.
- [ ] Diagram agrees with feature.yml (names, caps, gate policies).

## Files Modified/Created

- Modified: `workflow-diagram.dot`, `moon.yml`, regenerated PNGs.
- Created: none.

## Dependencies

Requires 03 + 04 (documents their landed shape). Parallel with 05
(disjoint files).

## Estimated Duration

1–2 days.

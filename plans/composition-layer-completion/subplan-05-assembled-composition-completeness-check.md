---
type: plan
has_subplans: false
parent_plan: ../composition-layer-completion.md
parallel_group: 3
status: pending
dependencies:
  plans: []
  files:
    - packages/skill/skill-workflow/workflow-guide/references/conformance.md
    - packages/skill/skill-workflow/workflow-guide/scripts/*.mjs
    - packages/skill/skill-workflow/workflow-guide/scripts/*.ts
    - packages/skill/skill-workflow/workflow-guide/assets/*.json
    - packages/skill/skill-workflow/package.json
    - packages/skill/skill-workflow/moon.yml
    - packages/skill/skill-agent-governance/agent-governance-guide/references/conformance.md
    - packages/skill/skill-agent-governance/agent-governance-guide/scripts/*.mjs
    - packages/command/command-workflow/commands/workflow-conformance.md
skills_to_consult:
  - plan-guide
  - skill-guide
  - testing-guide
  - hexagonal-pattern-guide
  - connascence-guide
  - moon-guide
validation:
  type_check: not_run
  lint: not_run
  build: not_run
  tests: not_run
updated: "2026-07-17"
---

# Whole-Assembled-Composition Completeness Check

## Objective

Prove "you can't assemble it wrong": a full selection (workflow profile + governance profile +
selected capabilities + providers + modules + methods) is validated as one artifact for
topology completeness and dangling references across both planes, backing the promise
`workflow-conformance` already makes ("validate their composition") with a named contract and
helper instead of prose.

## Tasks

1. Define a composition-completeness contract: extend
   `packages/skill/skill-workflow/workflow-guide/references/conformance.md` and/or
   `packages/skill/skill-agent-governance/agent-governance-guide/references/conformance.md`
   with a "whole composition" section specifying the checks — profile topology complete
   across both planes; every selected capability present and compatible; no dangling
   capability / provider / module / method reference (resolved against the Phase 4 registry);
   adoption-mode absence report present (per `architecture-and-composition.md` lines 14-24);
   mandatory cross-plane guarantees bound to an adequate enforcement point.
2. Implement `validateAssembledComposition` in
   `workflow-guide/scripts/conformance-helpers.mjs` (or a new
   `workflow-guide/scripts/composition-helpers.ts` — new files are TypeScript per the
   parent's Decision 1) that takes a full assembled selection
   and returns the first failure code, composing the existing `validateProfile`,
   `validateComposition` (`agent-governance-guide/scripts/conformance-helpers.mjs`
   lines 219-225), and provider/enforcement validators rather than duplicating them.
3. Add fixtures under `workflow-guide/assets/fixtures/` (test-only, per `conformance.md`
   line 30): a passing integrated composition (built from the Phase 3 integrated reference
   profile and the governance profile it pairs with) plus adversarial cases — a dangling
   method reference, a selected-but-absent capability, an incompatible provider, and a
   missing cross-plane enforcement point.
4. Add `workflow-guide/scripts/validate-assembled-composition.ts` (TypeScript per the
   parent's Decision 1) that exercises the fixtures with mutation guards (dud-guard count
   zero), matching the precedent in the existing `validate-*.mjs` scripts.
5. Wire `packages/command/command-workflow/commands/workflow-conformance.md` so its "validate
   their composition" step delegates to this contract; update the command doc if the
   operation name changes.
6. Wire the new validator into the `skill-workflow` `package.json` `test` script and add the
   new fixtures and helper to `moon.yml` `test` `inputs`.

## Acceptance criteria

- A CI-run script validates a shipped reference profile assembled into a full composition and
  fails on each adversarial fixture (dangling reference, absent capability, incompatible
  provider, missing enforcement) with a distinct failure code.
- `workflow-conformance`'s composition-validation promise is backed by a named contract and a
  helper (`validateAssembledComposition`), not prose alone.
- The completeness check consults the Phase 4 registry so an unshipped, unmarked capability
  in a composition is reported as a dangling reference.
- Mutation guards confirm the check cannot be silently defeated (dud-guard count zero).

## Dependencies

Phase 3 (adopter-facing-reference-profile-library) and Phase 4
(capability-registry-and-placeholder-marking) must both land first. The parent requires this
check to validate "a shipped reference profile assembled into a full composition" (needs the
Phase 3 library) and to "consult the Phase 4 registry so an unshipped, unmarked capability …
is reported as a dangling reference" (needs the Phase 4 registry). This is the last phase and
runs alone in its group.

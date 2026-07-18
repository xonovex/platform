---
type: plan
has_subplans: false
parent_plan: ../composition-layer-completion.md
parallel_group: 3
status: complete
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
  build: pass
  tests: pass
updated: "2026-07-18"
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

## Implementation notes (completed 2026-07-18)

`build` and `tests` pass via `moon` (`skill-workflow:test` includes the assembled-composition
validator); skill packages define no `type_check`/`lint` task. `format-check` passes for all
three affected packages and the cross-package link guard stays at 64/64. The validator proves
a shipped composition assembles and every adversarial case is rejected with its own code (1
shipped integrated composition + 10 fixtures + 4 mutation guards, dud-guard count zero).
End-to-end adversarial verification (3/3): dropping `user-stories` from the capability
registry makes the real shipped integrated composition fail `dangling-capability` (proving the
registry consultation is live, not fixture-only), pointing the integrated profile's governance
facet at a ghost identity yields `dangling-governance-reference`, and both restore to green.

- `workflow-guide/scripts/composition-helpers.ts` — `validateAssembledComposition(selection)`
  is a composition root: it wires the workflow and governance `validateProfile`, governance
  `validateComposition`, and the capability-registry `resolveCapability` rather than
  re-deriving them, and adds only the cross-plane checks no single plane owns — the
  adoption-mode absence report, the integrated profile pairing, required-capability presence,
  provider compatibility, and mandatory cross-plane enforcement. It returns the first failure
  code: `absence-report-missing`, the per-plane profile codes, `dangling-governance-reference`,
  `module-conflict`, `dangling-capability`, `missing-capability`, `incompatible-provider`, or
  `unenforced-mandatory-control`.
- `workflow-guide/assets/assembled-composition-fixtures.json` — a passing integrated
  composition plus nine adversarial cases, one per failure code.
- `workflow-guide/scripts/validate-assembled-composition.ts` — assembles the real Phase 3
  `integrated-profile` + `governance-only-profile` into a full composition (expects valid),
  runs the fixtures, and embeds mutation guards.
- `workflow-guide/references/conformance.md` gains a "Whole composition" section naming the
  contract and its failure codes; `agent-governance-guide/references/conformance.md`
  cross-references it; `workflow-conformance.md` now names `validateAssembledComposition` as the
  backing for its "validate their composition" step — the promise is contract-backed, not prose.
- Wired into the `skill-workflow` `package.json` test and `moon.yml` inputs.

### Deviations (with rationale)

- **Fixtures live in `assets/`, not `assets/fixtures/`.** `assets/fixtures/` is the scenario
  corpus whose `validate-conformance-scenario-fixtures.mjs` enforces an `index.json` bijection
  and per-file scenario schema; a multi-case fixture there fails that validator. Multi-case
  fixtures (`conformance-fixtures.json`, …) live in `assets/` directly, so
  `assembled-composition-fixtures.json` does too. The task's "assets/fixtures/" wording is
  imprecise for this file type.
- **Helper in a new `composition-helpers.ts` imported by the validator via a `.ts` import**
  (parent Decision 1: new files are TypeScript; Node's native type-stripping resolves a
  `./composition-helpers.ts` import), rather than adding a `.mjs` export to
  `conformance-helpers.mjs`.
- **Tests live in the validator** (a CI-run `validate-*.ts` with mutation guards), matching the
  sibling validators, rather than a separate `.test.ts`. Targeted `prettier --write` on the new
  files only; no `npm`.

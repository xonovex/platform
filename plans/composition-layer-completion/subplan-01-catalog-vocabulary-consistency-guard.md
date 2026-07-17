---
type: plan
has_subplans: false
parent_plan: ../composition-layer-completion.md
parallel_group: 1
status: complete
dependencies:
  plans: []
  files:
    - packages/skill/skill-agent-governance/agent-governance-guide/scripts/*.mjs
    - packages/skill/skill-agent-governance/agent-governance-guide/scripts/*.ts
    - packages/skill/skill-agent-governance/agent-governance-guide/references/*.md
    - packages/skill/skill-agent-governance/agent-governance-guide/assets/conformance-fixtures.json
    - packages/skill/skill-agent-governance/package.json
    - packages/skill/skill-agent-governance/moon.yml
    - packages/skill/skill-workflow/workflow-guide/scripts/*.mjs
    - packages/skill/skill-workflow/workflow-guide/scripts/*.ts
    - packages/skill/skill-workflow/workflow-guide/references/development-contracts.md
    - packages/skill/skill-workflow/workflow-guide/references/profiles.md
    - packages/skill/skill-workflow/workflow-guide/assets/development-assurance-fixtures.json
    - packages/skill/skill-workflow/package.json
    - packages/skill/skill-workflow/moon.yml
    - packages/command/command-workflow/docs/architecture-and-composition.md
    - packages/command/command-workflow/docs/adoption-map.md
skills_to_consult:
  - plan-guide
  - skill-guide
  - connascence-guide
  - testing-guide
  - moon-guide
validation:
  type_check: not_run
  lint: not_run
  build: pass
  tests: pass
updated: "2026-07-18"
---

# Catalog-Wide Vocabulary-Consistency Guard

## Objective

Every closed composition vocabulary is declared once and validated against _every_
declaring site (machine-read constants, fixture arrays, prose tables, cross-package doc
tables), so a catalog cannot be assembled with a vocabulary that has drifted between two
artifacts — generalizing the `validate-executor-vocabulary.mjs` pattern to the whole
catalog and giving the workflow plane the vocabulary check it lacks today.

## Tasks

1. Create `packages/skill/skill-agent-governance/agent-governance-guide/scripts/validate-composition-vocabulary.ts`
   (erasable-syntax TypeScript run directly by Node per the parent's Decision 1; the root
   `engines` floor raise to `>=22.18.0` is owned by the runtime plan's Phase 1 — land it first
   if this plan executes before it), modeled on `validate-executor-vocabulary.mjs`: a
   data-driven registry that, for each
   vocabulary, names the canonical owner (`expectedVocabulary` in
   `agent-governance-guide/scripts/conformance-helpers.mjs`) and lists every declaring site
   with an extractor (constant import, JSON key, prose-table regex, prose backtick-list
   regex). Cover `policyOutcomes`, `moduleKinds`, `moduleClassifications`, `authorityZones`,
   `adoptionModes`, and `profileFacets`.
2. Register the verified declaring sites with extractors:
   `agent-governance-guide/references/policy-and-authority.md` line 13 backtick-list
   (`policyOutcomes`); `references/modules.md` line 25 prose table (`moduleClassifications`,
   `moduleKinds`); `references/catalog-and-inventory.md`, `references/architecture.md`,
   `references/external-enforcement-onboarding.md`, and cross-package
   `packages/command/command-workflow/docs/architecture-and-composition.md` adoption-modes
   table (`adoptionModes`); `references/catalog-and-inventory.md` (`authorityZones`); both
   `profiles.md` facet lists (`profileFacets`); and cross-package
   `packages/command/command-workflow/docs/adoption-map.md` — its modes table
   (`adoptionModes`), its axes-section facet list (`profileFacets`), and its
   module-classification prose (`moduleClassifications`). Also register `adoption-map.md`
   as a view site in the two existing guards: the executor-class enumeration joins
   `validate-executor-vocabulary.mjs` and the twelve-family intent list joins
   `validate-event-intent-vocabulary.mjs`.
3. Remove the second independent machine-read declaration: edit
   `agent-governance-guide/scripts/governance-operations-helpers.mjs` lines 17-24 so
   `allowedModuleClassifications` imports `expectedVocabulary.moduleClassifications` from
   `conformance-helpers.mjs` (per the one-owner rule in `packages/skill/AGENTS.md`) rather
   than re-declaring the six values (`knowledge-only`, `advisory`, `evidence-producing`,
   `enforcing`, `configuration-changing`, `privileged`) — collapsing connascence of value
   to connascence of name.
4. Give the workflow plane its own vocabulary check (it has no `validateVocabulary` analog):
   add `packages/skill/skill-workflow/workflow-guide/scripts/validate-composition-vocabulary.ts`
   covering `independenceLevels` (owner `workflow-guide/scripts/independence-helpers.mjs`
   lines 5-10 — export the constant so it has a canonical owner; views
   `agent-governance-guide/references/actors.md` independence table lines 32-36 and 42-48)
   and the work-shape literals `mechanical` / `bounded-transform` / `adaptive` (owner
   `workflow-guide/scripts/development-assurance-helpers.mjs` `selectDevelopmentExecutor`;
   views `references/development-contracts.md` table,
   `assets/development-assurance-fixtures.json`).
5. Add mutation guards to each new validator — an invented value, a renamed value, a dropped
   value, and an unparseable site must each be caught — matching the dud-guard-zero
   precedent in `validate-executor-vocabulary.mjs`; treat "declaration not found" at a site
   as a failure so a reworded doc is caught rather than silently skipped.
6. Make `agent-governance-guide/references/actors.md` line 23 true: edit its
   "declared once and validated against every declaring site" list so it names exactly the
   vocabularies the new guards cover (policy outcomes, module kinds, and module
   classifications become validated); correct the sentence for any vocabulary intentionally
   left single-source.
7. Wire the new validators into the `test` script of
   `packages/skill/skill-agent-governance/package.json` and
   `packages/skill/skill-workflow/package.json`, and add the newly-read cross-package /
   cross-file inputs (`command-workflow/docs/architecture-and-composition.md`,
   `command-workflow/docs/adoption-map.md`, workflow
   `references/development-contracts.md`, `assets/development-assurance-fixtures.json`) to
   the corresponding `moon.yml` `test` `inputs` (governance `moon.yml` already lists
   `command-workflow/docs/*.md`).

## Acceptance criteria

- `npx moon run skill-agent-governance:test` and `npx moon run skill-workflow:test` each
  exit non-zero, naming the diverging site, when any covered vocabulary value is changed in
  exactly one declaring site (verified by temporarily mutating each site once).
- Every mutation guard reports a failure when the comparison is defeated; the dud-guard
  count is zero, matching the two existing validators.
- `governance-operations-helpers.mjs` no longer holds an independent `allowedModuleClassifications`
  array literal (`grep -n 'knowledge-only' packages/skill/skill-agent-governance/agent-governance-guide/scripts/governance-operations-helpers.mjs`
  returns nothing; the constant is derived from an `expectedVocabulary.moduleClassifications`
  import).
- `actors.md` line 23's list of "validated against every declaring site" vocabularies
  matches the set the guards actually cover — no false claim remains.
- Both new validators are present in the respective `package.json` `test` scripts and
  reachable from `ci-check`.

## Dependencies

None. Phase 1 is independent of the other phases — it guards vocabulary drift over the
existing catalog and is cited by no other phase as a prerequisite. It can run concurrently
with the cross-package link guard and the capability registry.

## Implementation notes (completed 2026-07-18)

Two self-contained validators land, modeled on `validate-executor-vocabulary.mjs` (owner
plus declaring sites, set-difference comparison, and derived mutation guards with a
dud-guard-zero assertion), both wired into the skills' `test` scripts and reachable from
CI. `build` and `tests` pass via `moon`; skill packages define no `type_check`/`lint`
task, so those are `not_run`. Adversarial verification: mutating any one covered site
fails the owning `moon` test naming the diverging file (13/13 sites), all validators pass
on `main`, and the dud-guard count is zero.

- `agent-governance-guide/scripts/validate-composition-vocabulary.ts` — 5 vocabularies
  across 9 machine-readable declaring sites, 20 mutation guards. `policyOutcomes`
  (owner `expectedVocabulary` ↔ `policy-and-authority.md`); `moduleClassifications`
  (owner ↔ `catalog-and-inventory.md` bullets ↔ `modules.md` declaration line);
  `adoptionModes` (owner ↔ `catalog-and-inventory.md`, `architecture.md`, and the
  cross-package `architecture-and-composition.md` and `adoption-map.md` mode tables,
  lower-cased to reconcile display capitalization); `authorityZones`
  (owner ↔ `catalog-and-inventory.md`); `profileFacets` (owner ↔ `adoption-map.md` facet
  span).
- `workflow-guide/scripts/validate-composition-vocabulary.ts` — 2 vocabularies across 3
  sites, 8 mutation guards. `independenceLevels` (owner `independence-helpers.mjs`, now
  exported, ↔ `actors.md` levels table, compared against the above-`none` levels) and the
  work shapes (new owner `developmentWorkShapes` in `development-assurance-helpers.mjs`
  ↔ `development-contracts.md` table ↔ `development-assurance-fixtures.json`).
- `validate-event-intent-vocabulary.mjs` gains `adoption-map.md` as a third view.

Supporting edits: `governance-operations-helpers.mjs` `allowedModuleClassifications` now
imports `expectedVocabulary.moduleClassifications` (connascence of value collapses to
connascence of name); `modules.md` adds the missing `knowledge-only` classification so its
enumeration matches the owner; `actors.md` "why no registry" claim is rewritten to name
exactly the machine-validated vocabularies; the root `engines.node` floor is raised to
`>=22.18.0` (parent Decision 1 precursor) so Node runs the `.ts` validators via native
type stripping.

### Scope decisions (sites deliberately left as human-readable views)

Several sites named in the tasks declare a vocabulary in human display form, not machine
tokens. Registering them would couple the guard to display wording or force a lossy
display→token map — re-introducing the connascence the guard removes — so they stay
human-readable views, per the parent plan's prose-brittleness risk mitigation:

- `moduleKinds`: `modules.md` table uses display names (`Bounded model evaluator` ≠
  `model-evaluator`), so it is not machine-validatable; it remains validated by
  `validateVocabulary` (owner ↔ conformance fixture), and `actors.md` now states this
  honestly rather than claiming full cross-site validation.
- The `authorityZones` table in `architecture.md` (`Repository/project`,
  `Session/runtime`), the executor-class prose in `adoption-map.md`
  (`deterministic script/API` ≠ `deterministic`), and the scaffolded
  `moduleClassifications` prose in `adoption-map.md` ("from … through … to …") are display
  views, not registered.
- `external-enforcement-onboarding.md` names only 2 of 5 adoption modes (a subset
  mention, not a closed-set declaration); not registered.

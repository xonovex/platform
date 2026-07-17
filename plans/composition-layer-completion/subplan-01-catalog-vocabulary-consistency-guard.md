---
type: plan
has_subplans: false
parent_plan: ../composition-layer-completion.md
parallel_group: 1
status: pending
dependencies:
  plans: []
  files:
    - packages/skill/skill-agent-governance/agent-governance-guide/scripts/*.mjs
    - packages/skill/skill-agent-governance/agent-governance-guide/references/*.md
    - packages/skill/skill-agent-governance/agent-governance-guide/assets/conformance-fixtures.json
    - packages/skill/skill-agent-governance/package.json
    - packages/skill/skill-agent-governance/moon.yml
    - packages/skill/skill-workflow/workflow-guide/scripts/*.mjs
    - packages/skill/skill-workflow/workflow-guide/references/development-contracts.md
    - packages/skill/skill-workflow/workflow-guide/references/profiles.md
    - packages/skill/skill-workflow/workflow-guide/assets/development-assurance-fixtures.json
    - packages/skill/skill-workflow/package.json
    - packages/skill/skill-workflow/moon.yml
    - packages/command/command-workflow/docs/architecture-and-composition.md
skills_to_consult:
  - plan-guide
  - skill-guide
  - connascence-guide
  - testing-guide
  - moon-guide
validation:
  type_check: not_run
  lint: not_run
  build: not_run
  tests: not_run
updated: "2026-07-17"
---

# Catalog-Wide Vocabulary-Consistency Guard

## Objective

Every closed composition vocabulary is declared once and validated against _every_
declaring site (machine-read constants, fixture arrays, prose tables, cross-package doc
tables), so a catalog cannot be assembled with a vocabulary that has drifted between two
artifacts — generalizing the `validate-executor-vocabulary.mjs` pattern to the whole
catalog and giving the workflow plane the vocabulary check it lacks today.

## Tasks

1. Create `packages/skill/skill-agent-governance/agent-governance-guide/scripts/validate-composition-vocabulary.mjs`,
   modeled on `validate-executor-vocabulary.mjs`: a data-driven registry that, for each
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
   `profiles.md` facet lists (`profileFacets`).
3. Remove the second independent machine-read declaration: edit
   `agent-governance-guide/scripts/governance-operations-helpers.mjs` lines 17-24 so
   `allowedModuleClassifications` imports `expectedVocabulary.moduleClassifications` from
   `conformance-helpers.mjs` (per the one-owner rule in `packages/skill/AGENTS.md`) rather
   than re-declaring the six values (`knowledge-only`, `advisory`, `evidence-producing`,
   `enforcing`, `configuration-changing`, `privileged`) — collapsing connascence of value
   to connascence of name.
4. Give the workflow plane its own vocabulary check (it has no `validateVocabulary` analog):
   add `packages/skill/skill-workflow/workflow-guide/scripts/validate-composition-vocabulary.mjs`
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
   cross-file inputs (`command-workflow/docs/architecture-and-composition.md`, workflow
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
